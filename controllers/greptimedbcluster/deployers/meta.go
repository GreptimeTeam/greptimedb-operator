package deployers

import (
	"context"
	"fmt"
	"strings"
	"time"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/GreptimeTeam/greptimedb-operator/apis/v1alpha1"
	"github.com/GreptimeTeam/greptimedb-operator/pkg/deployer"
)

var (
	defaultDialTimeout = 5 * time.Second
)

type EtcdMaintenanceBuilder func(etcdEndpoints []string) (clientv3.Maintenance, error)

type MetaDeployer struct {
	*CommonDeployer

	etcdMaintenanceBuilder func(etcdEndpoints []string) (clientv3.Maintenance, error)
}

type MetaDeployerOption func(*MetaDeployer)

var _ deployer.Deployer = &MetaDeployer{}

func NewMetaDeployer(mgr ctrl.Manager, opts ...MetaDeployerOption) *MetaDeployer {
	md := &MetaDeployer{
		CommonDeployer:         NewFromManager(mgr),
		etcdMaintenanceBuilder: buildEtcdMaintenance,
	}

	for _, opt := range opts {
		opt(md)
	}

	return md
}

func WithEtcdMaintenanceBuilder(builder EtcdMaintenanceBuilder) func(*MetaDeployer) {
	return func(d *MetaDeployer) {
		d.etcdMaintenanceBuilder = builder
	}
}

func (d *MetaDeployer) Render(crdObject client.Object) ([]client.Object, error) {
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

	deployment, err := d.generateDeployment(cluster)
	if err != nil {
		return nil, err
	}
	renderObjects = append(renderObjects, deployment)

	if cluster.Spec.EnablePrometheusMonitor {
		pm, err := d.generatePodMonitor(cluster)
		if err != nil {
			return nil, err
		}
		renderObjects = append(renderObjects, pm)
	}

	return renderObjects, nil
}

func (d *MetaDeployer) PreSyncHooks() []deployer.Hook {
	return []deployer.Hook{
		d.checkEtcdService,
	}
}

func (d *MetaDeployer) CheckAndUpdateStatus(ctx context.Context, highLevelObject client.Object) (bool, error) {
	cluster, err := d.GetCluster(highLevelObject)
	if err != nil {
		return false, err
	}

	var (
		deployment = new(appsv1.Deployment)

		objectKey = client.ObjectKey{
			Namespace: cluster.Namespace,
			Name:      d.ResourceName(cluster.Name, v1alpha1.MetaComponentKind),
		}
	)

	err = d.Get(ctx, objectKey, deployment)
	if errors.IsNotFound(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	cluster.Status.Meta.Replicas = *deployment.Spec.Replicas
	cluster.Status.Meta.ReadyReplicas = deployment.Status.ReadyReplicas
	cluster.Status.Meta.EtcdEndponts = cluster.Spec.Meta.EtcdEndpoints
	if err := UpdateStatus(ctx, cluster, d.Client); err != nil {
		klog.Errorf("Failed to update status: %s", err)
	}

	return deployer.IsDeploymentReady(deployment), nil
}

func (d *MetaDeployer) generateSvc(cluster *v1alpha1.GreptimeDBCluster) (*corev1.Service, error) {
	svc := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: cluster.Namespace,
			Name:      d.ResourceName(cluster.Name, v1alpha1.MetaComponentKind),
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
			Selector: map[string]string{
				GreptimeComponentName: d.ResourceName(cluster.Name, v1alpha1.MetaComponentKind),
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

	if err := deployer.SetControllerAndAnnotation(cluster, svc, d.Scheme, svc.Spec); err != nil {
		return nil, err
	}

	return svc, nil
}

func (d *MetaDeployer) generateDeployment(cluster *v1alpha1.GreptimeDBCluster) (*appsv1.Deployment, error) {
	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      d.ResourceName(cluster.Name, v1alpha1.MetaComponentKind),
			Namespace: cluster.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &cluster.Spec.Meta.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					GreptimeComponentName: d.ResourceName(cluster.Name, v1alpha1.MetaComponentKind),
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						GreptimeComponentName: d.ResourceName(cluster.Name, v1alpha1.MetaComponentKind),
					},
					Annotations: cluster.Spec.Meta.Template.Annotations,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:      string(v1alpha1.MetaComponentKind),
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

	if err := deployer.SetControllerAndAnnotation(cluster, deployment, d.Scheme, deployment.Spec); err != nil {
		return nil, err
	}

	return deployment, nil
}

func (d *MetaDeployer) generatePodMonitor(cluster *v1alpha1.GreptimeDBCluster) (*monitoringv1.PodMonitor, error) {
	pm := &monitoringv1.PodMonitor{
		TypeMeta: metav1.TypeMeta{
			Kind:       monitoringv1.PodMonitorsKind,
			APIVersion: monitoringv1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      d.ResourceName(cluster.Name, v1alpha1.MetaComponentKind),
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
					GreptimeComponentName: d.ResourceName(cluster.Name, v1alpha1.MetaComponentKind),
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

func (d *MetaDeployer) checkEtcdService(ctx context.Context, crdObject client.Object) error {
	cluster, err := d.GetCluster(crdObject)
	if err != nil {
		return nil
	}

	maintainer, err := d.etcdMaintenanceBuilder(cluster.Spec.Meta.EtcdEndpoints)
	if err != nil {
		return err
	}

	rsp, err := maintainer.Status(ctx, strings.Join(cluster.Spec.Meta.EtcdEndpoints, ","))
	if err != nil {
		return err
	}

	if len(rsp.Errors) != 0 {
		return fmt.Errorf("etcd service error: %v", rsp.Errors)
	}

	defer func() {
		etcdClient, ok := maintainer.(*clientv3.Client)
		if ok {
			etcdClient.Close()
		}
	}()

	return nil
}

func buildEtcdMaintenance(etcdEndpoints []string) (clientv3.Maintenance, error) {
	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints:   etcdEndpoints,
		DialTimeout: defaultDialTimeout,
		DialOptions: []grpc.DialOption{grpc.WithBlock()},
	})
	if err != nil {
		return nil, err
	}

	return etcdClient, nil
}
