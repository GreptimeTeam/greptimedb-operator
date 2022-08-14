/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package greptimedbcluster

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/greptime/greptimedb-operator/apis/v1alpha1"
	"github.com/greptime/greptimedb-operator/cmd/operator/app/options"
)

const (
	greptimeDBApplication = "greptime.cloud/application"

	defaultFrontendHTTPPort = 18080
	defaultFrontendGRPCPort = 19090

	defaultDatanodeGRPCPort = 19091
)

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

// +kubebuilder:rbac:groups=greptime.cloud,resources=greptimedbclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=greptime.cloud,resources=greptimedbclusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=greptime.cloud,resources=greptimedbclusters/finalizers,verbs=update

// Reconcile is reconciliation loop for GreptimeDBCluster.
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	klog.V(4).Infof("Reconciling GreptimeDBCluster: %s", req.NamespacedName)

	var cluster v1alpha1.GreptimeDBCluster
	if err := r.Get(ctx, req.NamespacedName, &cluster); err != nil {
		if errors.IsNotFound(err) {
			klog.Infof("Resource not found: %s", req.NamespacedName)
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	ready, err := r.syncMeta(ctx, &cluster)
	if err != nil {
		return ctrl.Result{}, err
	}

	if !ready {
		return ctrl.Result{}, nil
	}

	ready, err = r.syncFrontend(ctx, &cluster)
	if err != nil {
		return ctrl.Result{}, err
	}

	if !ready {
		return ctrl.Result{}, nil
	}

	// TODO(zyy17): syncFrontend and syncDatanode can be merged into one function.
	ready, err = r.syncDatanode(ctx, &cluster)
	if err != nil {
		return ctrl.Result{}, err
	}

	if !ready {
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}

// TODO(zyy17): implement syncMeta(), we can just deploy etcd cluster for testing.
func (r *Reconciler) syncMeta(ctx context.Context, cluster *v1alpha1.GreptimeDBCluster) (bool, error) {
	return true, nil
}

func (r *Reconciler) syncFrontend(ctx context.Context, cluster *v1alpha1.GreptimeDBCluster) (bool, error) {
	frontendDeployment, err := r.buildFrontendDeployment(cluster)
	if err != nil {
		return false, err
	}

	frontendService, err := r.buildFrontendService(cluster)
	if err != nil {
		return false, err
	}

	var service corev1.Service
	if err := r.Get(ctx, client.ObjectKey{Namespace: cluster.Namespace, Name: cluster.Name + "-frontend"}, &service); err != nil {
		if errors.IsNotFound(err) {
			if err := r.Create(ctx, frontendService); err != nil {
				return false, err
			}
			klog.Infof("Create frontend service...")
			return false, nil
		}

		return false, err
	}

	var deployment appsv1.Deployment
	if err := r.Get(ctx, client.ObjectKey{Namespace: cluster.Namespace, Name: cluster.Name + "-frontend"}, &deployment); err != nil {
		if errors.IsNotFound(err) {
			if err := r.Create(ctx, frontendDeployment); err != nil {
				return false, err
			}
			klog.Infof("Create frontend deployment...")
			return false, nil
		}
		return false, err
	}

	// FIXME(zyy17): the update logic only consider replicas field now.
	if *deployment.Spec.Replicas != *frontendDeployment.Spec.Replicas {
		if err := r.Update(ctx, frontendDeployment); err != nil {
			return false, err
		}
		klog.Infof("Update frontend deployment...")
	}

	if r.isDeploymentReady(&deployment) {
		return true, nil
	}

	return false, nil
}

func (r *Reconciler) syncDatanode(ctx context.Context, cluster *v1alpha1.GreptimeDBCluster) (bool, error) {
	frontendDeployment, err := r.buildDatanodeDeployment(cluster)
	if err != nil {
		return false, err
	}

	frontendService, err := r.buildDatanodeService(cluster)
	if err != nil {
		return false, err
	}

	var service corev1.Service
	if err := r.Get(ctx, client.ObjectKey{Namespace: cluster.Namespace, Name: cluster.Name + "-datanode"}, &service); err != nil {
		if errors.IsNotFound(err) {
			if err := r.Create(ctx, frontendService); err != nil {
				return false, err
			}
			klog.Infof("Create datanode service...")
			return false, nil
		}

		return false, err
	}

	var deployment appsv1.Deployment
	if err := r.Get(ctx, client.ObjectKey{Namespace: cluster.Namespace, Name: cluster.Name + "-datanode"}, &deployment); err != nil {
		if errors.IsNotFound(err) {
			if err := r.Create(ctx, frontendDeployment); err != nil {
				return false, err
			}
			klog.Infof("Create datanode deployment...")
			return false, nil
		}
		return false, err
	}

	// FIXME(zyy17): the update logic only consider replicas field now.
	if *deployment.Spec.Replicas != *frontendDeployment.Spec.Replicas {
		if err := r.Update(ctx, frontendDeployment); err != nil {
			return false, err
		}
		klog.Infof("Update frontend deployment...")
	}

	if r.isDeploymentReady(&deployment) {
		return true, nil
	}

	return false, nil
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

func (r *Reconciler) buildFrontendDeployment(cluster *v1alpha1.GreptimeDBCluster) (*appsv1.Deployment, error) {
	podTemplate := r.overlayPodSpec(cluster.Spec.Base, cluster.Spec.Frontend.Template)
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
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:      "frontend",
							Image:     podTemplate.MainContainer.Image,
							Resources: podTemplate.MainContainer.Resources,
							Command:   podTemplate.MainContainer.Command,
							Args:      podTemplate.MainContainer.Args,
							Ports: []corev1.ContainerPort{
								{
									Name:          "grpc",
									Protocol:      corev1.ProtocolTCP,
									ContainerPort: defaultFrontendGRPCPort,
								},
								{
									Name:          "http",
									Protocol:      corev1.ProtocolTCP,
									ContainerPort: defaultFrontendHTTPPort,
								},
							},
						},
					},
				},
			},
		},
	}

	if err := controllerutil.SetControllerReference(cluster, deployment, r.Scheme); err != nil {
		return nil, err
	}

	return deployment, nil
}

func (r *Reconciler) buildFrontendService(cluster *v1alpha1.GreptimeDBCluster) (*corev1.Service, error) {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.Name + "-frontend",
			Namespace: cluster.Namespace,
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
					Port:     defaultFrontendGRPCPort,
				},
				{
					Name:     "http",
					Protocol: corev1.ProtocolTCP,
					Port:     defaultFrontendHTTPPort,
				},
			},
		},
	}

	if err := controllerutil.SetControllerReference(cluster, service, r.Scheme); err != nil {
		return nil, err
	}

	return service, nil
}

func (r *Reconciler) buildDatanodeDeployment(cluster *v1alpha1.GreptimeDBCluster) (*appsv1.Deployment, error) {
	podTemplate := r.overlayPodSpec(cluster.Spec.Base, cluster.Spec.Frontend.Template)
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.Name + "-datanode",
			Namespace: cluster.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &cluster.Spec.Frontend.Replicas,
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
				},
				// TODO(zyy17): More arguments.
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:      "datanode",
							Image:     podTemplate.MainContainer.Image,
							Resources: podTemplate.MainContainer.Resources,
							Command:   podTemplate.MainContainer.Command,
							Args:      podTemplate.MainContainer.Args,
							Ports: []corev1.ContainerPort{
								{
									Name:          "grpc",
									Protocol:      corev1.ProtocolTCP,
									ContainerPort: defaultDatanodeGRPCPort,
								},
							},
						},
					},
				},
			},
		},
	}

	if err := controllerutil.SetControllerReference(cluster, deployment, r.Scheme); err != nil {
		return nil, err
	}

	return deployment, nil
}

func (r *Reconciler) buildDatanodeService(cluster *v1alpha1.GreptimeDBCluster) (*corev1.Service, error) {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.Name + "-datanode",
			Namespace: cluster.Namespace,
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
					Port:     defaultDatanodeGRPCPort,
				},
			},
		},
	}

	if err := controllerutil.SetControllerReference(cluster, service, r.Scheme); err != nil {
		return nil, err
	}

	return service, nil
}

// TODO(zyy17): Make it more generic.
func (r *Reconciler) overlayPodSpec(base *v1alpha1.PodTemplateSpec, component *v1alpha1.PodTemplateSpec) *v1alpha1.PodTemplateSpec {
	if component == nil {
		return base
	}

	if base.Annotations != nil {
		if component.Annotations == nil {
			component.Annotations = make(map[string]string)
		}
		for k, v := range base.Annotations {
			if _, ok := component.Annotations[k]; !ok {
				component.Annotations[k] = v
			}
		}
	}

	if base.Labels != nil {
		if component.Labels == nil {
			component.Labels = make(map[string]string)
		}
		for k, v := range base.Labels {
			if _, ok := component.Labels[k]; !ok {
				component.Labels[k] = v
			}
		}
	}

	if base.MainContainer != nil {
		if component.MainContainer == nil {
			component.MainContainer = base.MainContainer
		}
	}

	return component
}

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
