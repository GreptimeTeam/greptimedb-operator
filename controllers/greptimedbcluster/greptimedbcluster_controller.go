package greptimedbcluster

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/GreptimeTeam/greptimedb-operator/apis/v1alpha1"
	"github.com/GreptimeTeam/greptimedb-operator/cmd/operator/app/options"
)

const (
	greptimeDBApplication      = "greptime.cloud/application"
	greptimedbClusterFinalizer = "greptimedbcluster.greptime.cloud/finalizer"
	lastAppliedResourceSpec    = "greptimedbcluster.greptime.cloud/last-applied-resource-spec"
)

type SyncFunc func(ctx context.Context, cluster *v1alpha1.GreptimeDBCluster) (ready bool, err error)

// Reconciler reconciles a GreptimeDBCluster object
type Reconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

func Setup(mgr ctrl.Manager, option *options.Options) error {
	reconciler := &Reconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Recorder: mgr.GetEventRecorderFor("greptimedbcluster-controller"),
	}

	return reconciler.SetupWithManager(mgr)
}

// SetupWithManager sets up the controller with the Manager.
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.GreptimeDBCluster{}).
		Owns(&corev1.Service{}).
		Owns(&appsv1.StatefulSet{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}

// +kubebuilder:rbac:groups=greptime.cloud,resources=greptimedbclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=greptime.cloud,resources=greptimedbclusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=greptime.cloud,resources=greptimedbclusters/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;create;
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;delete;
// +kubebuilder:rbac:groups=core,resources=persistentvolumeclaims,verbs=get;list;watch;create;update;delete;
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete

// Reconcile is reconciliation loop for GreptimeDBCluster.
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	klog.Infof("Reconciling GreptimeDBCluster: %s", req.NamespacedName)

	cluster := new(v1alpha1.GreptimeDBCluster)
	if err := r.Get(ctx, req.NamespacedName, cluster); err != nil {
		if errors.IsNotFound(err) {
			klog.Infof("Resource not found: %s", req.NamespacedName)
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	cluster.SetDefaults()

	if err := r.handleFinalizers(ctx, cluster); err != nil {
		return ctrl.Result{}, err
	}

	// The controller will execute the following actions in order and the next action will begin to execute when the previous one is finished.
	var actions []SyncFunc
	if cluster.Spec.Meta != nil {
		actions = append(actions, r.syncEtcd, r.syncMeta)
	}
	if cluster.Spec.Datanode != nil {
		actions = append(actions, r.syncDatanode)
	}
	if cluster.Spec.Frontend != nil {
		actions = append(actions, r.syncFrontend)
	}

	clusterIsReady := false
	for index, action := range actions {
		ready, err := action(ctx, cluster)
		if err != nil {
			return ctrl.Result{}, err
		}
		if !ready {
			return ctrl.Result{}, nil
		}
		if index == len(actions)-1 {
			clusterIsReady = true
		}
	}

	if clusterIsReady {
		klog.Infof("The cluster '%s/%s' is ready", cluster.Namespace, cluster.Name)
		setGreptimeDBClusterCondition(&cluster.Status, newCondition(v1alpha1.GreptimeDBClusterReady, corev1.ConditionTrue, "", "GreptimeDB cluster is ready"))
		if err := r.updateStatus(ctx, cluster); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	if err := r.Update(ctx, cluster); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *Reconciler) syncEtcd(ctx context.Context, cluster *v1alpha1.GreptimeDBCluster) (bool, error) {
	klog.Infof("Syncing etcd...")

	newEtcdService, err := r.buildEtcdService(cluster)
	if err != nil {
		return false, err
	}

	_, err = r.createIfNotExist(ctx, new(corev1.Service), newEtcdService)
	if err != nil {
		return false, err
	}

	newEtcdStatefulSet, err := r.buildEtcdStatefulSet(cluster)
	if err != nil {
		return false, err
	}

	etcdStatefulSet, err := r.createIfNotExist(ctx, new(appsv1.StatefulSet), newEtcdStatefulSet)
	if err != nil {
		return false, err
	}

	if statefulSet, ok := etcdStatefulSet.(*appsv1.StatefulSet); ok {
		return r.isStatefulSetReady(statefulSet), nil
	}

	return false, nil
}

func (r *Reconciler) syncMeta(ctx context.Context, cluster *v1alpha1.GreptimeDBCluster) (bool, error) {
	klog.Infof("Syncing meta...")
	newMetaService, err := r.buildMetaService(cluster)
	if err != nil {
		return false, err
	}

	_, err = r.createIfNotExist(ctx, new(corev1.Service), newMetaService)
	if err != nil {
		return false, err
	}

	newMetaDeployment, err := r.buildMetaDeployment(cluster)
	if err != nil {
		return false, err
	}

	metaDeployment, err := r.createIfNotExist(ctx, new(appsv1.Deployment), newMetaDeployment)
	if err != nil {
		return false, err
	}

	if deployment, ok := metaDeployment.(*appsv1.Deployment); ok {
		needToUpdate, err := r.isNeedToUpdate(deployment, new(appsv1.DeploymentSpec), &newMetaDeployment.Spec)
		if err != nil {
			return false, err
		}

		if r.isDeploymentReady(deployment) && !needToUpdate {
			return true, nil
		}

		if needToUpdate {
			if err := r.Update(ctx, newMetaDeployment); err != nil {
				return false, err
			}
		}

		return false, nil
	}

	return false, nil
}

func (r *Reconciler) syncFrontend(ctx context.Context, cluster *v1alpha1.GreptimeDBCluster) (bool, error) {
	klog.Infof("Syncing frontend...")
	newFrontendService, err := r.buildFrontendService(cluster)
	if err != nil {
		return false, err
	}

	_, err = r.createIfNotExist(ctx, new(corev1.Service), newFrontendService)
	if err != nil {
		return false, err
	}

	newFrontendDeployment, err := r.buildFrontendDeployment(cluster)
	if err != nil {
		return false, err
	}

	frontendDeployment, err := r.createIfNotExist(ctx, new(appsv1.Deployment), newFrontendDeployment)
	if err != nil {
		return false, err
	}

	if deployment, ok := frontendDeployment.(*appsv1.Deployment); ok {
		needToUpdate, err := r.isNeedToUpdate(deployment, new(appsv1.DeploymentSpec), &newFrontendDeployment.Spec)
		if err != nil {
			return false, err
		}

		if r.isDeploymentReady(deployment) && !needToUpdate {
			return true, nil
		}

		if needToUpdate {
			if err := r.Update(ctx, newFrontendDeployment); err != nil {
				return false, err
			}
		}

		return false, nil
	}

	return false, nil
}

func (r *Reconciler) syncDatanode(ctx context.Context, cluster *v1alpha1.GreptimeDBCluster) (bool, error) {
	klog.Infof("Syncing datanode...")
	newDatanodeService, err := r.buildDatanodeService(cluster)
	if err != nil {
		return false, err
	}

	_, err = r.createIfNotExist(ctx, new(corev1.Service), newDatanodeService)
	if err != nil {
		return false, err
	}

	newDatanodeStatefulSet, err := r.buildDatanodeStatefulSet(cluster)
	if err != nil {
		return false, err
	}

	datanodeStatefulSet, err := r.createIfNotExist(ctx, new(appsv1.StatefulSet), newDatanodeStatefulSet)
	if err != nil {
		return false, err
	}

	if statefulSet, ok := datanodeStatefulSet.(*appsv1.StatefulSet); ok {
		needToUpdate, err := r.isNeedToUpdate(statefulSet, new(appsv1.StatefulSetSpec), &newDatanodeStatefulSet.Spec)
		if err != nil {
			return false, err
		}

		if r.isStatefulSetReady(statefulSet) && !needToUpdate {
			return true, nil
		}

		if needToUpdate {
			if err := r.Update(ctx, newDatanodeStatefulSet); err != nil {
				return false, err
			}
		}
		return false, nil
	}

	return false, nil
}

// buildEtcdService create the etcd headless service.
func (r *Reconciler) buildEtcdService(cluster *v1alpha1.GreptimeDBCluster) (*corev1.Service, error) {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: cluster.Namespace,
			Name:      cluster.Name + "-etcd",
		},
		Spec: corev1.ServiceSpec{
			Type: "",
			Selector: map[string]string{
				greptimeDBApplication: cluster.Name + "-etcd",
			},
			Ports: []corev1.ServicePort{
				{
					Name:     "client",
					Protocol: corev1.ProtocolTCP,
					Port:     cluster.Spec.Meta.Etcd.ClientPort,
				},
				{
					Name:     "peer",
					Protocol: corev1.ProtocolTCP,
					Port:     cluster.Spec.Meta.Etcd.PeerPort,
				},
			},
		},
	}

	if err := controllerutil.SetControllerReference(cluster, service, r.Scheme); err != nil {
		return nil, err
	}

	return service, nil
}

func (r *Reconciler) buildEtcdStatefulSet(cluster *v1alpha1.GreptimeDBCluster) (*appsv1.StatefulSet, error) {
	statefulset := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.Name + "-etcd",
			Namespace: cluster.Namespace,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas:    &cluster.Spec.Meta.Etcd.ClusterSize,
			ServiceName: cluster.Name + "-etcd",
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					greptimeDBApplication: cluster.Name + "-etcd",
				},
			},
			VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: cluster.Spec.Meta.Etcd.Storage.Name,
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						StorageClassName: cluster.Spec.Meta.Etcd.Storage.StorageClassName,
						AccessModes: []corev1.PersistentVolumeAccessMode{
							corev1.ReadWriteOnce,
						},
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceStorage: resource.MustParse(cluster.Spec.Meta.Etcd.Storage.StorageSize),
							},
						},
					},
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						greptimeDBApplication: cluster.Name + "-etcd",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "etcd",
							Image: cluster.Spec.Meta.Etcd.Image,
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      cluster.Spec.Meta.Etcd.Storage.Name,
									MountPath: cluster.Spec.Meta.Etcd.Storage.MountPath,
								},
							},
							Ports: []corev1.ContainerPort{
								{
									Name:          "client",
									ContainerPort: cluster.Spec.Meta.Etcd.ClientPort,
									Protocol:      corev1.ProtocolTCP,
								},
								{
									Name:          "peer",
									ContainerPort: cluster.Spec.Meta.Etcd.PeerPort,
									Protocol:      corev1.ProtocolTCP,
								},
							},
							Command: []string{
								"etcd",
							},
							Args: []string{
								"--name", "$(HOSTNAME)",
								"--data-dir", cluster.Spec.Meta.Etcd.Storage.MountPath,
								"--initial-advertise-peer-urls", "http://$(HOSTNAME):" + strconv.Itoa(int(cluster.Spec.Meta.Etcd.PeerPort)),
								"--listen-peer-urls", "http://0.0.0.0:2380",
								"--advertise-client-urls", "http://$(HOSTNAME):" + strconv.Itoa(int(cluster.Spec.Meta.Etcd.ClientPort)),
								"--listen-client-urls", "http://0.0.0.0:2379",
								"--initial-cluster",
								generateInitCluster(cluster.Name+"-etcd", cluster.Name+"-etcd", cluster.Namespace,
									int(cluster.Spec.Meta.Etcd.PeerPort), int(cluster.Spec.Meta.Etcd.ClusterSize)),
								"--initial-cluster-state", "new",
								"--initial-cluster-token", cluster.Name,
							},
							Env: []corev1.EnvVar{
								{
									Name: "HOSTNAME",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.name",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	if err := controllerutil.SetControllerReference(cluster, statefulset, r.Scheme); err != nil {
		return nil, err
	}

	return statefulset, nil
}

func (r *Reconciler) buildFrontendService(cluster *v1alpha1.GreptimeDBCluster) (*corev1.Service, error) {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: cluster.Namespace,
			Name:      cluster.Name + "-frontend",
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
			Selector: map[string]string{
				greptimeDBApplication: cluster.Name + "-frontend",
			},
			Ports: []corev1.ServicePort{
				{
					Name:     "grpc",
					Protocol: corev1.ProtocolTCP,
					Port:     cluster.Spec.GRPCServicePort,
				},
				{
					Name:     "http",
					Protocol: corev1.ProtocolTCP,
					Port:     cluster.Spec.HTTPServicePort,
				},
				{
					Name:     "mysql",
					Protocol: corev1.ProtocolTCP,
					Port:     cluster.Spec.MySQLServicePort,
				},
			},
		},
	}

	if err := controllerutil.SetControllerReference(cluster, service, r.Scheme); err != nil {
		return nil, err
	}

	return service, nil
}

func (r *Reconciler) buildFrontendDeployment(cluster *v1alpha1.GreptimeDBCluster) (*appsv1.Deployment, error) {
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.Name + "-frontend",
			Namespace: cluster.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &cluster.Spec.Frontend.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					greptimeDBApplication: cluster.Name + "-frontend",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						greptimeDBApplication: cluster.Name + "-frontend",
					},
					Annotations: cluster.Spec.Frontend.Template.Annotations,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:      "frontend",
							Image:     cluster.Spec.Frontend.Template.MainContainer.Image,
							Resources: *cluster.Spec.Frontend.Template.MainContainer.Resources,
							Command:   cluster.Spec.Frontend.Template.MainContainer.Command,
							Args:      cluster.Spec.Frontend.Template.MainContainer.Args,
							Ports: []corev1.ContainerPort{
								{
									Name:          "grpc",
									Protocol:      corev1.ProtocolTCP,
									ContainerPort: cluster.Spec.GRPCServicePort,
								},
								{
									Name:          "http",
									Protocol:      corev1.ProtocolTCP,
									ContainerPort: cluster.Spec.HTTPServicePort,
								},
								{
									Name:          "mysql",
									Protocol:      corev1.ProtocolTCP,
									ContainerPort: cluster.Spec.MySQLServicePort,
								},
							},
						},
					},
				},
			},
		},
	}

	for k, v := range cluster.Spec.Frontend.Template.Labels {
		deployment.Labels[k] = v
	}

	if err := controllerutil.SetControllerReference(cluster, deployment, r.Scheme); err != nil {
		return nil, err
	}

	if err := r.setLastAppliedResourceSpecAnnotation(deployment, deployment.Spec); err != nil {
		return nil, err
	}

	return deployment, nil
}

func (r *Reconciler) buildMetaService(cluster *v1alpha1.GreptimeDBCluster) (*corev1.Service, error) {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: cluster.Namespace,
			Name:      cluster.Name + "-meta",
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
			Selector: map[string]string{
				greptimeDBApplication: cluster.Name + "-meta",
			},
			Ports: []corev1.ServicePort{
				{
					Name:     "grpc",
					Protocol: corev1.ProtocolTCP,
					Port:     cluster.Spec.GRPCServicePort,
				},
			},
		},
	}

	if err := controllerutil.SetControllerReference(cluster, service, r.Scheme); err != nil {
		return nil, err
	}

	return service, nil
}

func (r *Reconciler) buildMetaDeployment(cluster *v1alpha1.GreptimeDBCluster) (*appsv1.Deployment, error) {
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.Name + "-meta",
			Namespace: cluster.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &cluster.Spec.Meta.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					greptimeDBApplication: cluster.Name + "-meta",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						greptimeDBApplication: cluster.Name + "-meta",
					},
					Annotations: cluster.Spec.Meta.Template.Annotations,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:      "meta",
							Image:     cluster.Spec.Meta.Template.MainContainer.Image,
							Resources: *cluster.Spec.Meta.Template.MainContainer.Resources,
							Command:   cluster.Spec.Meta.Template.MainContainer.Command,
							Args:      cluster.Spec.Meta.Template.MainContainer.Args,
							Ports: []corev1.ContainerPort{
								{
									Name:          "grpc",
									Protocol:      corev1.ProtocolTCP,
									ContainerPort: cluster.Spec.GRPCServicePort,
								},
							},
						},
					},
				},
			},
		},
	}

	for k, v := range cluster.Spec.Meta.Template.Labels {
		deployment.Labels[k] = v
	}

	if err := controllerutil.SetControllerReference(cluster, deployment, r.Scheme); err != nil {
		return nil, err
	}

	if err := r.setLastAppliedResourceSpecAnnotation(deployment, deployment.Spec); err != nil {
		return nil, err
	}

	return deployment, nil
}

func (r *Reconciler) buildDatanodeService(cluster *v1alpha1.GreptimeDBCluster) (*corev1.Service, error) {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: cluster.Namespace,
			Name:      cluster.Name + "-datanode",
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
			Selector: map[string]string{
				greptimeDBApplication: cluster.Name + "-datanode",
			},
			Ports: []corev1.ServicePort{
				{
					Name:     "grpc",
					Protocol: corev1.ProtocolTCP,
					Port:     cluster.Spec.GRPCServicePort,
				},
				{
					Name:     "http",
					Protocol: corev1.ProtocolTCP,
					Port:     cluster.Spec.HTTPServicePort,
				},
				{
					Name:     "mysql",
					Protocol: corev1.ProtocolTCP,
					Port:     cluster.Spec.MySQLServicePort,
				},
			},
		},
	}

	if err := controllerutil.SetControllerReference(cluster, service, r.Scheme); err != nil {
		return nil, err
	}

	return service, nil
}

func (r *Reconciler) buildDatanodeStatefulSet(cluster *v1alpha1.GreptimeDBCluster) (*appsv1.StatefulSet, error) {
	statefulset := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.Name + "-datanode",
			Namespace: cluster.Namespace,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: &cluster.Spec.Datanode.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					greptimeDBApplication: cluster.Name + "-datanode",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						greptimeDBApplication: cluster.Name + "-datanode",
					},
					Annotations: cluster.Spec.Datanode.Template.Annotations,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:      "datanode",
							Image:     cluster.Spec.Datanode.Template.MainContainer.Image,
							Resources: *cluster.Spec.Datanode.Template.MainContainer.Resources,
							Command:   cluster.Spec.Datanode.Template.MainContainer.Command,
							Args:      cluster.Spec.Datanode.Template.MainContainer.Args,
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      cluster.Spec.Datanode.Storage.Name,
									MountPath: cluster.Spec.Datanode.Storage.MountPath,
								},
							},
							Ports: []corev1.ContainerPort{
								{
									Name:          "grpc",
									Protocol:      corev1.ProtocolTCP,
									ContainerPort: cluster.Spec.GRPCServicePort,
								},
								{
									Name:          "http",
									Protocol:      corev1.ProtocolTCP,
									ContainerPort: cluster.Spec.HTTPServicePort,
								},
								{
									Name:          "mysql",
									Protocol:      corev1.ProtocolTCP,
									ContainerPort: cluster.Spec.MySQLServicePort,
								},
							},
						},
					},
				},
			},
			VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: cluster.Spec.Datanode.Storage.Name,
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						StorageClassName: cluster.Spec.Datanode.Storage.StorageClassName,
						AccessModes: []corev1.PersistentVolumeAccessMode{
							corev1.ReadWriteOnce,
						},
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceStorage: resource.MustParse(cluster.Spec.Datanode.Storage.StorageSize),
							},
						},
					},
				},
			},
		},
	}
	for k, v := range cluster.Spec.Datanode.Template.Labels {
		statefulset.Labels[k] = v
	}

	if err := controllerutil.SetControllerReference(cluster, statefulset, r.Scheme); err != nil {
		return nil, err
	}

	if err := r.setLastAppliedResourceSpecAnnotation(statefulset, statefulset.Spec); err != nil {
		return nil, err
	}

	return statefulset, nil
}

func (r *Reconciler) createIfNotExist(ctx context.Context, source, newObject client.Object) (client.Object, error) {
	var err error
	if err = r.Get(ctx, client.ObjectKey{Namespace: newObject.GetNamespace(), Name: newObject.GetName()}, source); err != nil && errors.IsNotFound(err) {
		if err = r.Create(ctx, newObject); err != nil {
			return nil, err
		}
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return source, nil
}

// TODO(zyy17): Maybe it's not a accurate way to detect the statefulset is ready.
func (r *Reconciler) isStatefulSetReady(statefulSet *appsv1.StatefulSet) bool {
	if statefulSet == nil {
		return false
	}

	return statefulSet.Status.ReadyReplicas == *statefulSet.Spec.Replicas
}

// TODO(zyy17): Maybe it's not a accurate way to detect the deployment is ready.
func (r *Reconciler) isDeploymentReady(deployment *appsv1.Deployment) bool {
	if deployment == nil {
		return false
	}

	for _, cond := range deployment.Status.Conditions {
		if cond.Type == appsv1.DeploymentProgressing {
			if cond.Reason == "NewReplicaSetAvailable" &&
				deployment.Status.ReadyReplicas == *deployment.Spec.Replicas {
				return true
			}
		}
	}

	return false
}

func (r *Reconciler) isNeedToUpdate(source client.Object, oldSpec interface{}, newSpec interface{}) (bool, error) {
	annotations := source.GetAnnotations()
	if annotations == nil {
		return false, fmt.Errorf("resource annotations is nil")
	}

	lastAppliedDeploymentSpec, ok := annotations[lastAppliedResourceSpec]
	if !ok {
		return false, fmt.Errorf("last applied source spec is not found")
	}

	if err := json.Unmarshal([]byte(lastAppliedDeploymentSpec), oldSpec); err != nil {
		return false, err
	}

	return !apiequality.Semantic.DeepEqual(oldSpec, newSpec), nil
}

func (r *Reconciler) handleFinalizers(ctx context.Context, cluster *v1alpha1.GreptimeDBCluster) error {
	// Determine if the object is under deletion.
	if cluster.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object. This is equivalent
		// registering our finalizer.
		if cluster.Spec.Meta != nil {
			if !controllerutil.ContainsFinalizer(cluster, greptimedbClusterFinalizer) {
				controllerutil.AddFinalizer(cluster, greptimedbClusterFinalizer)
			}
		}
		return nil
	}

	// The object is being deleted.
	if controllerutil.ContainsFinalizer(cluster, greptimedbClusterFinalizer) {
		if err := r.deleteEtcdStorage(ctx, cluster); err != nil {
			return err
		}

		// remove our finalizer from the list.
		controllerutil.RemoveFinalizer(cluster, greptimedbClusterFinalizer)
	}

	return nil
}

func (r *Reconciler) setLastAppliedResourceSpecAnnotation(object client.Object, spec interface{}) error {
	data, err := json.Marshal(spec)
	if err != nil {
		return err
	}

	object.SetAnnotations(map[string]string{lastAppliedResourceSpec: string(data)})

	return nil
}

// TODO(zyy17): Should use the more elegant way to manage etcd storage.
func (r *Reconciler) deleteEtcdStorage(ctx context.Context, cluster *v1alpha1.GreptimeDBCluster) error {
	klog.Infof("Deleting etcd storage...")

	var (
		etcdStoragePVC = new(corev1.PersistentVolumeClaimList)
		listOptions    []client.ListOption
		selector       labels.Selector
	)

	selector, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{
		MatchLabels: map[string]string{
			greptimeDBApplication: cluster.Name + "-etcd",
		},
	})
	if err != nil {
		return err
	}

	listOptions = append(listOptions, client.MatchingLabelsSelector{Selector: selector})
	if err := r.List(ctx, etcdStoragePVC, listOptions...); err != nil {
		return err
	}

	for _, pvc := range etcdStoragePVC.Items {
		klog.Infof("Deleting etcd PVC: %s", pvc.Name)
		if err := r.Delete(ctx, &pvc); err != nil {
			return err
		}
	}

	return nil
}

func (r *Reconciler) updateStatus(ctx context.Context, cluster *v1alpha1.GreptimeDBCluster, opts ...client.UpdateOption) error {
	status := cluster.DeepCopy().Status
	return retry.RetryOnConflict(retry.DefaultBackoff, func() (err error) {
		if err = r.Client.Get(ctx, client.ObjectKey{
			Namespace: cluster.Namespace,
			Name:      cluster.Name,
		}, cluster); err != nil {
			return
		}
		cluster.Status = status
		return r.Status().Update(ctx, cluster, opts...)
	})
}

func newCondition(conditionType v1alpha1.GreptimeDBConditionType, conditionStatus corev1.ConditionStatus, reason, message string) v1alpha1.GreptimeDBClusterCondition {
	return v1alpha1.GreptimeDBClusterCondition{
		Type:               conditionType,
		Status:             conditionStatus,
		LastTransitionTime: metav1.Now(),
		Reason:             reason,
		Message:            message,
	}
}

func getGreptimeDBClusterConditition(status v1alpha1.GreptimeDBClusterStatus, conditionType v1alpha1.GreptimeDBConditionType) *v1alpha1.GreptimeDBClusterCondition {
	for i := range status.Conditions {
		c := status.Conditions[i]
		if c.Type == conditionType {
			return &c
		}
	}
	return nil
}

func setGreptimeDBClusterCondition(status *v1alpha1.GreptimeDBClusterStatus, condition v1alpha1.GreptimeDBClusterCondition) {
	currentCondition := getGreptimeDBClusterConditition(*status, condition.Type)
	if currentCondition != nil &&
		currentCondition.Status == condition.Status &&
		currentCondition.Reason == condition.Reason {
		return
	}

	if currentCondition != nil && currentCondition.Status == condition.Status {
		condition.LastTransitionTime = currentCondition.LastTransitionTime
	}

	newConditions := filterOutCondition(status.Conditions, condition.Type)
	status.Conditions = append(newConditions, condition)
}

func filterOutCondition(conditions []v1alpha1.GreptimeDBClusterCondition, conditionType v1alpha1.GreptimeDBConditionType) []v1alpha1.GreptimeDBClusterCondition {
	var newCondititions []v1alpha1.GreptimeDBClusterCondition
	for _, c := range conditions {
		if c.Type == conditionType {
			continue
		}
		newCondititions = append(newCondititions, c)
	}
	return newCondititions
}

// generateInitCluster will generare the init cluster string like 'etcd-0=http://etcd-0.etcd.default:2380,etcd-1=http://etcd-1.etcd.default:2380,etcd-2=http://etcd-2.etcd.default:2380'.
func generateInitCluster(prefix, svc, namespace string, port, clusterSize int) string {
	var s []string

	for i := 0; i < clusterSize; i++ {
		s = append(s, fmt.Sprintf("%s-%d=http://%s-%d.%s.%s:%d", prefix, i, prefix, i, svc, namespace, port))
	}

	return strings.Join(s, ",")
}
