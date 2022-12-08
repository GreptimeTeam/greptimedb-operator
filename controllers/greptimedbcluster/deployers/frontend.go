// Copyright 2022 Greptime Team
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package deployers

import (
	"context"
	"fmt"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
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
	var renderObjects []client.Object

	cluster, err := d.GetCluster(crdObject)
	if err != nil {
		return nil, err
	}

	if cluster.Spec.Frontend != nil {
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

		if len(cluster.Spec.Frontend.Config) > 0 {
			cm, err := d.GenerateConfigMap(cluster, v1alpha1.FrontendComponentKind)
			if err != nil {
				return nil, err
			}
			renderObjects = append(renderObjects, cm)

			for _, object := range renderObjects {
				if deployment, ok := object.(*appsv1.Deployment); ok {
					d.mountConfigMapVolume(deployment, cm.Name)
				}
			}
		}

		if cluster.Spec.EnablePrometheusMonitor {
			pm, err := d.generatePodMonitor(cluster)
			if err != nil {
				return nil, err
			}
			renderObjects = append(renderObjects, pm)
		}
	}

	return renderObjects, nil
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
	ports := []corev1.ServicePort{
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
		{
			Name:     "postgres",
			Protocol: corev1.ProtocolTCP,
			Port:     cluster.Spec.PostgresServicePort,
		},
		{
			Name:     "opentsdb",
			Protocol: corev1.ProtocolTCP,
			Port:     cluster.Spec.OpenTSDBServicePort,
		},
	}

	svc := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   cluster.Namespace,
			Name:        d.ResourceName(cluster.Name, v1alpha1.FrontendComponentKind),
			Annotations: cluster.Spec.Frontend.Service.Annotations,
			Labels:      cluster.Spec.Frontend.Service.Labels,
		},
		Spec: corev1.ServiceSpec{
			Type: cluster.Spec.Frontend.Service.Type,
			Selector: map[string]string{
				GreptimeComponentName: d.ResourceName(cluster.Name, v1alpha1.FrontendComponentKind),
			},
			Ports:             ports,
			LoadBalancerClass: cluster.Spec.Frontend.Service.LoadBalancerClass,
		},
	}

	if err := deployer.SetControllerAndAnnotation(cluster, svc, d.Scheme, svc.Spec); err != nil {
		return nil, err
	}

	return svc, nil
}

func (d *FrontendDeployer) generateDeployment(cluster *v1alpha1.GreptimeDBCluster) (*appsv1.Deployment, error) {
	var args []string
	if len(cluster.Spec.Frontend.Template.MainContainer.Args) > 0 {
		args = cluster.Spec.Frontend.Template.MainContainer.Args
	} else {
		args = d.buildFrontendArgs(cluster)
	}

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
						GreptimeComponentName: d.ResourceName(cluster.Name, v1alpha1.FrontendComponentKind),
					},
					Annotations: cluster.Spec.Frontend.Template.Annotations,
				},
				Spec: corev1.PodSpec{
					ImagePullSecrets: cluster.Spec.Frontend.Template.ImagePullSecrets,
					Containers: []corev1.Container{
						{
							Name:      string(v1alpha1.FrontendComponentKind),
							Image:     cluster.Spec.Frontend.Template.MainContainer.Image,
							Resources: *cluster.Spec.Frontend.Template.MainContainer.Resources,
							Command:   cluster.Spec.Frontend.Template.MainContainer.Command,
							Args:      args,
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
								{
									Name:          "postgres",
									Protocol:      corev1.ProtocolTCP,
									ContainerPort: cluster.Spec.PostgresServicePort,
								},
								{
									Name:          "opentsdb",
									Protocol:      corev1.ProtocolTCP,
									ContainerPort: cluster.Spec.OpenTSDBServicePort,
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

func (d *FrontendDeployer) buildFrontendArgs(cluster *v1alpha1.GreptimeDBCluster) []string {
	return []string{
		"frontend", "start",
		"--grpc-addr", fmt.Sprintf("0.0.0.0:%d", cluster.Spec.GRPCServicePort),
		"--metasrv-addr", fmt.Sprintf("%s.%s:%d", d.ResourceName(cluster.Name, v1alpha1.MetaComponentKind), cluster.Namespace, cluster.Spec.Meta.ServicePort),
		"--http-addr", fmt.Sprintf("0.0.0.0:%d", cluster.Spec.HTTPServicePort),
		"--mysql-addr", fmt.Sprintf("0.0.0.0:%d", cluster.Spec.MySQLServicePort),
		"--postgres-addr", fmt.Sprintf("0.0.0.0:%d", cluster.Spec.PostgresServicePort),
		"--opentsdb-addr", fmt.Sprintf("0.0.0.0:%d", cluster.Spec.OpenTSDBServicePort),
	}
}

func (d *FrontendDeployer) generatePodMonitor(cluster *v1alpha1.GreptimeDBCluster) (*monitoringv1.PodMonitor, error) {
	pm := &monitoringv1.PodMonitor{
		TypeMeta: metav1.TypeMeta{
			Kind:       monitoringv1.PodMonitorsKind,
			APIVersion: monitoringv1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      d.ResourceName(cluster.Name, v1alpha1.FrontendComponentKind),
			Namespace: cluster.Namespace,
		},
		Spec: monitoringv1.PodMonitorSpec{
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
					GreptimeComponentName: d.ResourceName(cluster.Name, v1alpha1.FrontendComponentKind),
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

func (d *FrontendDeployer) mountConfigMapVolume(deployment *appsv1.Deployment, name string) {
	deployment.Spec.Template.Spec.Volumes = append(deployment.Spec.Template.Spec.Volumes, corev1.Volume{
		Name: "config",
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: name,
				},
			},
		},
	})

	for i, container := range deployment.Spec.Template.Spec.Containers {
		if container.Name == string(v1alpha1.FrontendComponentKind) {
			deployment.Spec.Template.Spec.Containers[i].VolumeMounts = append(deployment.Spec.Template.Spec.Containers[i].VolumeMounts, corev1.VolumeMount{
				Name:      "config",
				MountPath: DefaultConfigPath,
			})
		}
	}
}
