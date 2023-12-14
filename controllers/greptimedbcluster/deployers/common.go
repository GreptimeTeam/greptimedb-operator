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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/GreptimeTeam/greptimedb-operator/apis/v1alpha1"
	"github.com/GreptimeTeam/greptimedb-operator/pkg/dbconfig"
	"github.com/GreptimeTeam/greptimedb-operator/pkg/deployer"
)

const (
	GreptimeDBComponentName = "app.greptime.io/component"
	GreptimeDBConfigDir     = "/etc/greptimedb"
	GreptimeDBTLSDir        = "/etc/greptimedb/tls"

	// GreptimeDBInitConfigDir used for greptimedb-initializer.
	GreptimeDBInitConfigDir = "/etc/greptimedb-init"

	GreptimeDBConfigFileName = "config.toml"
	MainContainerIndex       = 0
	ConfigVolumeName         = "config"
	InitConfigVolumeName     = "init-config"
	TLSVolumeName            = "tls"
)

// CommonDeployer is the common deployer for all components of GreptimeDBCluster.
type CommonDeployer struct {
	Scheme *runtime.Scheme

	client.Client
	deployer.DefaultDeployer
}

// NewFromManager creates a new CommonDeployer from controller manager.
func NewFromManager(mgr ctrl.Manager) *CommonDeployer {
	return &CommonDeployer{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),

		DefaultDeployer: deployer.DefaultDeployer{
			Client: mgr.GetClient(),
		},
	}
}

func (c *CommonDeployer) GetCluster(crdObject client.Object) (*v1alpha1.GreptimeDBCluster, error) {
	cluster, ok := crdObject.(*v1alpha1.GreptimeDBCluster)
	if !ok {
		return nil, fmt.Errorf("the object is not GreptimeDBCluster")
	}
	return cluster, nil
}

type CommonBuilder struct {
	Cluster       *v1alpha1.GreptimeDBCluster
	ComponentKind v1alpha1.ComponentKind

	*deployer.DefaultBuilder
}

func (c *CommonDeployer) NewCommonBuilder(crdObject client.Object, componentKind v1alpha1.ComponentKind) *CommonBuilder {
	cb := &CommonBuilder{
		DefaultBuilder: &deployer.DefaultBuilder{
			Scheme: c.Scheme,
			Owner:  crdObject,
		},
		ComponentKind: componentKind,
	}

	cluster, err := c.GetCluster(crdObject)
	if err != nil {
		cb.Err = err
	}
	cb.Cluster = cluster

	return cb
}

func (c *CommonBuilder) GenerateConfigMap() (*corev1.ConfigMap, error) {
	configData, err := dbconfig.FromCluster(c.Cluster, c.ComponentKind)
	if err != nil {
		return nil, err
	}

	configmap := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      ResourceName(c.Cluster.Name, c.ComponentKind),
			Namespace: c.Cluster.Namespace,
		},
		Data: map[string]string{
			GreptimeDBConfigFileName: string(configData),
		},
	}

	return configmap, nil
}

func (c *CommonBuilder) GeneratePodTemplateSpec(template *v1alpha1.PodTemplateSpec) *corev1.PodTemplateSpec {
	if template == nil || template.MainContainer == nil {
		return nil
	}

	spec := &corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: template.Annotations,
			Labels:      template.Labels,
		},
		Spec: corev1.PodSpec{
			// containers[0] is the main container.
			Containers: []corev1.Container{
				{
					// The main container name is the same as the component kind.
					Name:            string(c.ComponentKind),
					Resources:       *template.MainContainer.Resources,
					Image:           template.MainContainer.Image,
					Command:         template.MainContainer.Command,
					Args:            template.MainContainer.Args,
					WorkingDir:      template.MainContainer.WorkingDir,
					Env:             template.MainContainer.Env,
					LivenessProbe:   template.MainContainer.LivenessProbe,
					ReadinessProbe:  template.MainContainer.ReadinessProbe,
					Lifecycle:       template.MainContainer.Lifecycle,
					ImagePullPolicy: template.MainContainer.ImagePullPolicy,
					VolumeMounts:    template.MainContainer.VolumeMounts,
				},
			},
			NodeSelector:                  template.NodeSelector,
			InitContainers:                template.InitContainers,
			RestartPolicy:                 template.RestartPolicy,
			TerminationGracePeriodSeconds: template.TerminationGracePeriodSeconds,
			ActiveDeadlineSeconds:         template.ActiveDeadlineSeconds,
			DNSPolicy:                     template.DNSPolicy,
			ServiceAccountName:            template.ServiceAccountName,
			HostNetwork:                   template.HostNetwork,
			ImagePullSecrets:              template.ImagePullSecrets,
			Affinity:                      template.Affinity,
			SchedulerName:                 template.SchedulerName,
			Volumes:                       template.Volumes,
			Tolerations:                   template.Tolerations,
		},
	}

	if len(template.AdditionalContainers) > 0 {
		spec.Spec.Containers = append(spec.Spec.Containers, template.AdditionalContainers...)
	}

	return spec
}

func (c *CommonBuilder) GeneratePodMonitor() (*monitoringv1.PodMonitor, error) {
	pm := &monitoringv1.PodMonitor{
		TypeMeta: metav1.TypeMeta{
			Kind:       monitoringv1.PodMonitorsKind,
			APIVersion: monitoringv1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      ResourceName(c.Cluster.Name, c.ComponentKind),
			Namespace: c.Cluster.Namespace,
			Labels:    c.Cluster.Spec.PrometheusMonitor.Labels,
		},
		Spec: monitoringv1.PodMonitorSpec{
			PodMetricsEndpoints: []monitoringv1.PodMetricsEndpoint{
				{
					Path:        "/metrics",
					Port:        "http",
					Interval:    c.Cluster.Spec.PrometheusMonitor.Interval,
					HonorLabels: true,
				},
			},
			Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					GreptimeDBComponentName: ResourceName(c.Cluster.Name, c.ComponentKind),
				},
			},
			NamespaceSelector: monitoringv1.NamespaceSelector{
				MatchNames: []string{
					c.Cluster.Namespace,
				},
			},
		},
	}

	return pm, nil
}

// MountConfigDir mounts the configmap to the main container as '/etc/greptimedb/config.toml'.
func (c *CommonBuilder) MountConfigDir(template *corev1.PodTemplateSpec) {
	template.Spec.Volumes = append(template.Spec.Volumes, corev1.Volume{
		Name: ConfigVolumeName,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: ResourceName(c.Cluster.Name, c.ComponentKind),
				},
			},
		},
	})

	template.Spec.Containers[MainContainerIndex].VolumeMounts =
		append(template.Spec.Containers[MainContainerIndex].VolumeMounts,
			corev1.VolumeMount{
				Name:      ConfigVolumeName,
				MountPath: GreptimeDBConfigDir,
			},
		)
}

func UpdateStatus(ctx context.Context, input *v1alpha1.GreptimeDBCluster, kc client.Client, opts ...client.UpdateOption) error {
	cluster := input.DeepCopy()
	status := cluster.Status
	return retry.RetryOnConflict(retry.DefaultBackoff, func() (err error) {
		objectKey := client.ObjectKey{Namespace: cluster.Namespace, Name: cluster.Name}
		if err = kc.Get(ctx, objectKey, cluster); err != nil {
			return
		}
		cluster.Status = status
		return kc.Status().Update(ctx, cluster, opts...)
	})
}

func ResourceName(clusterName string, componentKind v1alpha1.ComponentKind) string {
	return clusterName + "-" + string(componentKind)
}
