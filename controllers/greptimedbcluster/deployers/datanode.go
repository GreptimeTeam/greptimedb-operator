package deployers

import (
	"context"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/GreptimeTeam/greptimedb-operator/apis/v1alpha1"
	"github.com/GreptimeTeam/greptimedb-operator/pkg/deployer"
)

// DatanodeDeployer is the deployer for datanode.
type DatanodeDeployer struct {
	*CommonDeployer
}

var _ deployer.Deployer = &DatanodeDeployer{}

func NewDatanodeDeployer(mgr ctrl.Manager) *DatanodeDeployer {
	return &DatanodeDeployer{
		CommonDeployer: NewFromManager(mgr),
	}
}

func (d *DatanodeDeployer) Render(crdObject client.Object) ([]client.Object, error) {
	var renderObjects []client.Object

	cluster, err := d.GetCluster(crdObject)
	if err != nil {
		return nil, err
	}

	svc, err := d.generateSvc(cluster)
	if err != nil {
		return nil, err
	}
	renderObjects = append(renderObjects, svc)

	sts, err := d.generateSts(cluster)
	if err != nil {
		return nil, err
	}
	renderObjects = append(renderObjects, sts)

	if cluster.Spec.EnablePrometheusMonitor {
		pm, err := d.generatePodMonitor(cluster)
		if err != nil {
			return nil, err
		}
		renderObjects = append(renderObjects, pm)
	}

	return renderObjects, nil
}

func (d *DatanodeDeployer) CleanUp(ctx context.Context, crdObject client.Object) error {
	cluster, err := d.GetCluster(crdObject)
	if err != nil {
		return err
	}

	if cluster.Spec.Datanode != nil {
		if cluster.Spec.Datanode.Storage.StorageRetainPolicy == v1alpha1.RetainStorageRetainPolicyTypeDelete {
			if err := d.deleteStorage(ctx, cluster); err != nil {
				return err
			}
		}
	}

	return nil
}

func (d *DatanodeDeployer) CheckAndUpdateStatus(ctx context.Context, crdObject client.Object) (bool, error) {
	cluster, err := d.GetCluster(crdObject)
	if err != nil {
		return false, err
	}

	var (
		sts = new(appsv1.StatefulSet)

		objectKey = client.ObjectKey{
			Namespace: cluster.Namespace,
			Name:      d.ResourceName(cluster.Name, v1alpha1.DatanodeComponentKind),
		}
	)

	err = d.Get(ctx, objectKey, sts)
	if errors.IsNotFound(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	cluster.Status.Datanode.Replicas = *sts.Spec.Replicas
	cluster.Status.Datanode.ReadyReplicas = sts.Status.ReadyReplicas
	if err := UpdateStatus(ctx, cluster, d.Client); err != nil {
		klog.Errorf("Failed to update status: %s", err)
	}

	return deployer.IsStatefulSetReady(sts), nil
}

func (d *DatanodeDeployer) deleteStorage(ctx context.Context, cluster *v1alpha1.GreptimeDBCluster) error {
	klog.Infof("Deleting datanode storage...")

	selector, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{
		MatchLabels: map[string]string{
			GreptimeComponentName: d.ResourceName(cluster.Name, v1alpha1.DatanodeComponentKind),
		},
	})
	if err != nil {
		return err
	}

	pvcList := new(corev1.PersistentVolumeClaimList)

	err = d.List(ctx, pvcList, client.InNamespace(cluster.Namespace), client.MatchingLabelsSelector{Selector: selector})
	if errors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}

	for _, pvc := range pvcList.Items {
		klog.Infof("Deleting datanode PVC: %s", pvc.Name)
		if err := d.Delete(ctx, &pvc); err != nil {
			return err
		}
	}

	return nil
}

func (d *DatanodeDeployer) generateSvc(cluster *v1alpha1.GreptimeDBCluster) (*corev1.Service, error) {
	svc := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: cluster.Namespace,
			Name:      d.ResourceName(cluster.Name, v1alpha1.DatanodeComponentKind),
		},
		Spec: corev1.ServiceSpec{
			ClusterIP: corev1.ClusterIPNone,
			Selector: map[string]string{
				GreptimeComponentName: d.ResourceName(cluster.Name, v1alpha1.DatanodeComponentKind),
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

	if err := deployer.SetControllerAndAnnotation(cluster, svc, d.Scheme, svc.Spec); err != nil {
		return nil, err
	}

	return svc, nil
}

func (d *DatanodeDeployer) generateSts(cluster *v1alpha1.GreptimeDBCluster) (*appsv1.StatefulSet, error) {
	sts := &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "StatefulSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      d.ResourceName(cluster.Name, v1alpha1.DatanodeComponentKind),
			Namespace: cluster.Namespace,
		},
		Spec: appsv1.StatefulSetSpec{
			ServiceName: d.ResourceName(cluster.Name, v1alpha1.DatanodeComponentKind),
			Replicas:    &cluster.Spec.Datanode.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					GreptimeComponentName: d.ResourceName(cluster.Name, v1alpha1.DatanodeComponentKind),
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						GreptimeComponentName: d.ResourceName(cluster.Name, v1alpha1.DatanodeComponentKind),
					},
					Annotations: cluster.Spec.Datanode.Template.Annotations,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:      string(v1alpha1.DatanodeComponentKind),
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
		sts.Labels[k] = v
	}

	if err := deployer.SetControllerAndAnnotation(cluster, sts, d.Scheme, sts.Spec); err != nil {
		return nil, err
	}

	return sts, nil
}

func (d *DatanodeDeployer) generatePodMonitor(cluster *v1alpha1.GreptimeDBCluster) (*monitoringv1.PodMonitor, error) {
	pm := &monitoringv1.PodMonitor{
		TypeMeta: metav1.TypeMeta{
			Kind:       monitoringv1.PodMonitorsKind,
			APIVersion: monitoringv1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      d.ResourceName(cluster.Name, v1alpha1.DatanodeComponentKind),
			Namespace: cluster.Namespace,
		},
		Spec: monitoringv1.PodMonitorSpec{
			PodTargetLabels: nil,
			PodMetricsEndpoints: []monitoringv1.PodMetricsEndpoint{
				{
					Path:        DefaultMetricPath,
					Port:        DefaultMetricPortName,
					Interval:    DefaultScapeInterval,
					HonorLabels: true,
				},
			},
			Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					GreptimeComponentName: d.ResourceName(cluster.Name, v1alpha1.DatanodeComponentKind),
				},
			},
			NamespaceSelector: monitoringv1.NamespaceSelector{
				MatchNames: []string{
					cluster.Namespace,
				},
			},
		},
	}

	if err := deployer.SetControllerAndAnnotation(cluster, pm, d.Scheme, pm.Spec); err != nil {
		return nil, err
	}

	return pm, nil
}
