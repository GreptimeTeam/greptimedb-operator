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
	"context"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/GreptimeTeam/greptimedb-operator/apis/v1alpha1"
	"github.com/GreptimeTeam/greptimedb-operator/controllers/constant"
	"github.com/GreptimeTeam/greptimedb-operator/pkg/util"
)

const (
	// FileStorageTypeLabelKey is the key for PVC labels that indicate the type of file storage.
	FileStorageTypeLabelKey = "app.greptime.io/fileStorageType"
)

type FileStorageType string

const (
	FileStorageTypeDatanode FileStorageType = "datanode"
	FileStorageTypeWAL      FileStorageType = "wal"
	FileStorageTypeCache    FileStorageType = "cache"
)

func ResourceName(name string, componentKind v1alpha1.ComponentKind) string {
	return name + "-" + string(componentKind)
}

func AdditionalResourceName(name, additionalName string, componentKind v1alpha1.ComponentKind) string {
	if len(additionalName) == 0 {
		return name + "-" + string(componentKind)
	}
	return name + "-" + string(componentKind) + "-" + additionalName
}

func MountConfigDir(name string, kind v1alpha1.ComponentKind, template *corev1.PodTemplateSpec, additionalName string) {
	resourceName := ResourceName(name, kind)
	if len(additionalName) != 0 {
		resourceName = AdditionalResourceName(name, additionalName, kind)
	}

	template.Spec.Volumes = append(template.Spec.Volumes, corev1.Volume{
		Name: constant.ConfigVolumeName,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: resourceName,
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

func GenerateConfigMap(namespace, name string, kind v1alpha1.ComponentKind, configData []byte, additionalName string) (*corev1.ConfigMap, error) {
	resourceName := ResourceName(name, kind)
	if len(additionalName) != 0 {
		resourceName = AdditionalResourceName(name, additionalName, kind)
	}

	configmap := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      resourceName,
			Namespace: namespace,
		},
		Data: map[string]string{
			constant.GreptimeDBConfigFileName: string(configData),
		},
	}

	return configmap, nil
}

func GeneratePodMonitor(namespace, name string, kind v1alpha1.ComponentKind, promSpec *v1alpha1.PrometheusMonitorSpec, additionalName string) (*monitoringv1.PodMonitor, error) {
	resourceName := ResourceName(name, kind)
	if len(additionalName) != 0 {
		resourceName = AdditionalResourceName(name, additionalName, kind)
	}

	pm := &monitoringv1.PodMonitor{
		TypeMeta: metav1.TypeMeta{
			Kind:       monitoringv1.PodMonitorsKind,
			APIVersion: monitoringv1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      resourceName,
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
					constant.GreptimeDBComponentName: resourceName,
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
					StartupProbe:    template.MainContainer.StartupProbe,
					LivenessProbe:   template.MainContainer.LivenessProbe,
					ReadinessProbe:  template.MainContainer.ReadinessProbe,
					Lifecycle:       template.MainContainer.Lifecycle,
					ImagePullPolicy: template.MainContainer.ImagePullPolicy,
					VolumeMounts:    template.MainContainer.VolumeMounts,
					SecurityContext: template.MainContainer.SecurityContext,
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
			SecurityContext:               template.SecurityContext,
		},
	}

	if len(template.AdditionalContainers) > 0 {
		spec.Spec.Containers = append(spec.Spec.Containers, template.AdditionalContainers...)
	}

	return spec
}

func FileStorageToPVC(name string, fs v1alpha1.FileStorageAccessor, fsType FileStorageType) *corev1.PersistentVolumeClaim {
	var (
		labels      map[string]string
		annotations map[string]string
	)

	switch fsType {
	case FileStorageTypeWAL:
		labels = map[string]string{
			FileStorageTypeLabelKey:          string(FileStorageTypeWAL),
			constant.GreptimeDBComponentName: ResourceName(name, v1alpha1.DatanodeComponentKind),
		}
	case FileStorageTypeCache:
		labels = map[string]string{
			FileStorageTypeLabelKey:          string(FileStorageTypeCache),
			constant.GreptimeDBComponentName: ResourceName(name, v1alpha1.DatanodeComponentKind),
		}
	default:
		// Add common label: 'app.greptime.io/component: ${CLUSTER_NAME}-${RESOURCE_KIND}'.
		labels = map[string]string{
			constant.GreptimeDBComponentName: ResourceName(name, v1alpha1.DatanodeComponentKind),
		}
	}

	if fs.GetLabels() != nil {
		labels = util.MergeStringMap(labels, fs.GetLabels())
	}

	if fs.GetAnnotations() != nil {
		annotations = util.MergeStringMap(annotations, fs.GetAnnotations())
	}

	return &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:        fs.GetName(),
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			StorageClassName: fs.GetStorageClassName(),
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse(fs.GetSize()),
				},
			},
		},
	}
}

func GetPVCs(ctx context.Context, k8sClient client.Client, namespace, name string, kind v1alpha1.ComponentKind, fsType FileStorageType) ([]corev1.PersistentVolumeClaim, error) {
	var labelSelector *metav1.LabelSelector
	switch fsType {
	case FileStorageTypeDatanode:
		// Every PVC has the label of the datanode component.
		// If we want to delete the datanode PVCs, we need to filter out the WAL and cache PVCs from the datanode PVCs.
		labelSelector = &metav1.LabelSelector{
			MatchExpressions: []metav1.LabelSelectorRequirement{
				{
					Key:      FileStorageTypeLabelKey,
					Operator: metav1.LabelSelectorOpNotIn,
					Values:   []string{string(FileStorageTypeWAL), string(FileStorageTypeCache)},
				},
				{
					Key:      constant.GreptimeDBComponentName,
					Operator: metav1.LabelSelectorOpIn,
					Values:   []string{ResourceName(name, kind)},
				},
			},
		}
	case FileStorageTypeWAL:
		labelSelector = &metav1.LabelSelector{
			MatchLabels: map[string]string{
				FileStorageTypeLabelKey: string(FileStorageTypeWAL),
			},
		}
	case FileStorageTypeCache:
		labelSelector = &metav1.LabelSelector{
			MatchLabels: map[string]string{
				FileStorageTypeLabelKey: string(FileStorageTypeCache),
			},
		}
	}

	selector, err := metav1.LabelSelectorAsSelector(labelSelector)
	if err != nil {
		return nil, err
	}

	claims := new(corev1.PersistentVolumeClaimList)

	err = k8sClient.List(ctx, claims, client.InNamespace(namespace), client.MatchingLabelsSelector{Selector: selector})
	if errors.IsNotFound(err) {
		return nil, err
	}
	if err != nil {
		return nil, err
	}

	return claims.Items, nil
}
func MonitoringServiceName(name string) string {
	return name + "-monitor"
}

func LogsPipelineName(namespace, name string) string {
	return namespace + "-" + name + "-logs"
}
