// Copyright 2024 Greptime Team
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

package common

import (
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/GreptimeTeam/greptimedb-operator/apis/v1alpha1"
	"github.com/GreptimeTeam/greptimedb-operator/controllers/constant"
)

var (
	DatanodeFileStorageLabels = map[string]string{
		"app.greptime.io/fileStorageType": "datanode",
	}

	WALFileStorageLabels = map[string]string{
		"app.greptime.io/fileStorageType": "wal",
	}

	CacheFileStorageLabels = map[string]string{
		"app.greptime.io/fileStorageType": "cache",
	}
)

func ResourceName(name string, componentKind v1alpha1.ComponentKind) string {
	return name + "-" + string(componentKind)
}

func MountConfigDir(name string, kind v1alpha1.ComponentKind, template *corev1.PodTemplateSpec) {
	template.Spec.Volumes = append(template.Spec.Volumes, corev1.Volume{
		Name: constant.ConfigVolumeName,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: ResourceName(name, kind),
				},
			},
		},
	})

	template.Spec.Containers[constant.MainContainerIndex].VolumeMounts =
		append(template.Spec.Containers[constant.MainContainerIndex].VolumeMounts,
			corev1.VolumeMount{
				Name:      constant.ConfigVolumeName,
				MountPath: constant.GreptimeDBConfigDir,
			},
		)
}

func GenerateConfigMap(namespace, name string, kind v1alpha1.ComponentKind, configData []byte) (*corev1.ConfigMap, error) {
	configmap := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      ResourceName(name, kind),
			Namespace: namespace,
		},
		Data: map[string]string{
			constant.GreptimeDBConfigFileName: string(configData),
		},
	}

	return configmap, nil
}

func GeneratePodMonitor(namespace, name string, kind v1alpha1.ComponentKind, promSpec *v1alpha1.PrometheusMonitorSpec) (*monitoringv1.PodMonitor, error) {
	pm := &monitoringv1.PodMonitor{
		TypeMeta: metav1.TypeMeta{
			Kind:       monitoringv1.PodMonitorsKind,
			APIVersion: monitoringv1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      ResourceName(name, kind),
			Namespace: namespace,
			Labels:    promSpec.Labels,
		},
		Spec: monitoringv1.PodMonitorSpec{
			PodMetricsEndpoints: []monitoringv1.PodMetricsEndpoint{
				{
					Path:        "/metrics",
					Port:        "http",
					Interval:    promSpec.Interval,
					HonorLabels: true,
				},
			},
			Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					constant.GreptimeDBComponentName: ResourceName(name, kind),
				},
			},
			NamespaceSelector: monitoringv1.NamespaceSelector{
				MatchNames: []string{
					namespace,
				},
			},
		},
	}

	return pm, nil
}

func GeneratePodTemplateSpec(kind v1alpha1.ComponentKind, template *v1alpha1.PodTemplateSpec) *corev1.PodTemplateSpec {
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
					Name:            string(kind),
					Resources:       template.MainContainer.Resources,
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
