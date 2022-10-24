package deployers

import (
	"context"

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

type FrontendDeployer struct {
	*CommonDeployer
}

var _ deployer.Deployer = &FrontendDeployer{}

func NewFrontendDeployer(mgr ctrl.Manager) *FrontendDeployer {
	return &FrontendDeployer{
		CommonDeployer: NewFromManager(mgr),
	}
}

func (d *FrontendDeployer) Render(crdObject client.Object) ([]client.Object, error) {
	cluster, err := d.GetCluster(crdObject)
	if err != nil {
		return nil, err
	}

	svc, err := d.generateSvc(cluster)
	if err != nil {
		return nil, err
	}

	deployment, err := d.generateDeployment(cluster)
	if err != nil {
		return nil, err
	}

	return []client.Object{svc, deployment}, nil
}

func (d *FrontendDeployer) CheckAndUpdateStatus(ctx context.Context, crdObject client.Object) (bool, error) {
	cluster, err := d.GetCluster(crdObject)
	if err != nil {
		return false, err
	}

	var (
		deployment = new(appsv1.Deployment)

		objectKey = client.ObjectKey{
			Namespace: cluster.Namespace,
			Name:      d.ResourceName(cluster.Name, v1alpha1.FrontendComponentKind),
		}
	)

	err = d.Get(ctx, objectKey, deployment)
	if errors.IsNotFound(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	cluster.Status.Frontend.Replicas = *deployment.Spec.Replicas
	cluster.Status.Frontend.ReadyReplicas = deployment.Status.ReadyReplicas
	if err := UpdateStatus(ctx, cluster, d.Client); err != nil {
		klog.Errorf("Failed to update status: %s", err)
	}

	return deployer.IsDeploymentReady(deployment), nil
}

func (d *FrontendDeployer) generateSvc(cluster *v1alpha1.GreptimeDBCluster) (*corev1.Service, error) {
	svc := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: cluster.Namespace,
			Name:      d.ResourceName(cluster.Name, v1alpha1.FrontendComponentKind),
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
			Selector: map[string]string{
				GreptimeComponentName: d.ResourceName(cluster.Name, v1alpha1.FrontendComponentKind),
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

func (d *FrontendDeployer) generateDeployment(cluster *v1alpha1.GreptimeDBCluster) (*appsv1.Deployment, error) {
	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      d.ResourceName(cluster.Name, v1alpha1.FrontendComponentKind),
			Namespace: cluster.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &cluster.Spec.Frontend.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					GreptimeComponentName: d.ResourceName(cluster.Name, v1alpha1.FrontendComponentKind),
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						GreptimeComponentName: cluster.Name + "-frontend",
					},
					Annotations: cluster.Spec.Frontend.Template.Annotations,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:      string(v1alpha1.FrontendComponentKind),
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

	if err := deployer.SetControllerAndAnnotation(cluster, deployment, d.Scheme, deployment.Spec); err != nil {
		return nil, err
	}

	return deployment, nil
}
