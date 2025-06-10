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
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/avast/retry-go"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
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

// ResourceName returns the resource name for the given name and role kind.
// The format will be `${name}-${roleKind}-${extraNames[0]}-${extraNames[1]}...`.
// If extraNames are provided, they will be appended to the resource name. For example,
// - If the name is `my-cluster` and the role kind is `datanode`, the resource name will be `my-cluster-datanode`.
// - If the name is `my-cluster` and the role kind is `datanode` and the extra names `read`, the resource name will be `my-cluster-datanode-read`.
func ResourceName(name string, roleKind v1alpha1.RoleKind, extraNames ...string) string {
	const separator = "-"

	baseName := strings.Join([]string{name, string(roleKind)}, separator)
	for _, extraName := range extraNames {
		if extraName != "" {
			baseName = strings.Join([]string{baseName, extraName}, separator)
		}
	}

	return baseName
}

func MountConfigDir(template *corev1.PodTemplateSpec, configMapName string) {
	template.Spec.Volumes = append(template.Spec.Volumes, corev1.Volume{
		Name: constant.ConfigVolumeName,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: configMapName,
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

func GenerateConfigMap(namespace, resourceName string, configData []byte) (*corev1.ConfigMap, error) {
	configmap := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      resourceName,
		},
		Data: map[string]string{
			constant.GreptimeDBConfigFileName: string(configData),
		},
	}

	return configmap, nil
}

func GeneratePodMonitor(namespace, resourceName string, promSpec *v1alpha1.PrometheusMonitorSpec) (*monitoringv1.PodMonitor, error) {
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

func GeneratePodTemplateSpec(kind v1alpha1.RoleKind, template *v1alpha1.PodTemplateSpec) *corev1.PodTemplateSpec {
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

func FileStorageToPVC(name string, fs v1alpha1.FileStorageAccessor, fsType FileStorageType, kind v1alpha1.RoleKind) *corev1.PersistentVolumeClaim {
	var (
		labels      map[string]string
		annotations map[string]string
	)

	switch fsType {
	case FileStorageTypeWAL:
		labels = map[string]string{
			FileStorageTypeLabelKey:          string(FileStorageTypeWAL),
			constant.GreptimeDBComponentName: ResourceName(name, kind),
		}
	case FileStorageTypeCache:
		labels = map[string]string{
			FileStorageTypeLabelKey:          string(FileStorageTypeCache),
			constant.GreptimeDBComponentName: ResourceName(name, kind),
		}
	default:
		// Add common label: 'app.greptime.io/component: ${CLUSTER_NAME}-${COMPONENT_KIND}'.
		labels = map[string]string{
			constant.GreptimeDBComponentName: ResourceName(name, kind),
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

func GetPVCs(ctx context.Context, k8sClient client.Client, namespace, name string, kind v1alpha1.RoleKind, fsType FileStorageType) ([]corev1.PersistentVolumeClaim, error) {
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
				FileStorageTypeLabelKey:          string(FileStorageTypeWAL),
				constant.GreptimeDBComponentName: ResourceName(name, kind),
			},
		}
	case FileStorageTypeCache:
		labelSelector = &metav1.LabelSelector{
			MatchLabels: map[string]string{
				FileStorageTypeLabelKey:          string(FileStorageTypeCache),
				constant.GreptimeDBComponentName: ResourceName(name, kind),
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
	return strings.Join([]string{namespace, name, "logs"}, "-")
}

func GetMetaHTTPServiceURL(cluster *v1alpha1.GreptimeDBCluster) string {
	return fmt.Sprintf("http://%s.%s:%d", ResourceName(cluster.GetName(), v1alpha1.MetaRoleKind), cluster.GetNamespace(), cluster.Spec.Meta.RPCPort)
}

// SetMaintenanceMode requests the metasrv to set the maintenance mode.
func SetMaintenanceMode(metaHTTPServiceURL string, enabled bool) error {
	requestURL := fmt.Sprintf("%s/admin/maintenance?enable=%v", metaHTTPServiceURL, enabled)

	operation := func() error {
		rsp, err := http.Get(requestURL)
		if err != nil {
			return err
		}
		defer rsp.Body.Close()

		if rsp.StatusCode != http.StatusOK {
			return fmt.Errorf("failed to request metasrv '%s' for set maintenance mode '%v', status code: %d", requestURL, enabled, rsp.StatusCode)
		}

		return nil
	}

	// The server may not be ready to accept the request, so we retry a few times.
	if err := retry.Do(
		operation,
		retry.Attempts(5),
		retry.Delay(1*time.Second),
		retry.DelayType(retry.FixedDelay),
	); err != nil {
		klog.Errorf("Failed to set maintenance mode to '%v' by requesting metasrv '%s': %v", enabled, metaHTTPServiceURL, err)
		return err
	}

	klog.Infof("Set maintenance mode to '%v' by requesting metasrv '%s'", enabled, metaHTTPServiceURL)

	return nil
}
