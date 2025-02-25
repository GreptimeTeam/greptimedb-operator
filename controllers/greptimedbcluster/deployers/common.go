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
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"text/template"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/GreptimeTeam/greptimedb-operator/apis/v1alpha1"
	"github.com/GreptimeTeam/greptimedb-operator/controllers/common"
	"github.com/GreptimeTeam/greptimedb-operator/controllers/constant"
	"github.com/GreptimeTeam/greptimedb-operator/controllers/greptimedbcluster/deployers/config"
	"github.com/GreptimeTeam/greptimedb-operator/pkg/dbconfig"
	"github.com/GreptimeTeam/greptimedb-operator/pkg/deployer"
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

	return common.GenerateConfigMap(c.Cluster.Namespace, c.Cluster.Name, c.ComponentKind, configData)
}

func (c *CommonBuilder) GenerateFrontendGroupConfigMap(frontend *v1alpha1.FrontendSpec) (*corev1.ConfigMap, error) {
	configData, err := dbconfig.FromFrontendGroup(frontend, c.ComponentKind)
	if err != nil {
		return nil, err
	}

	return common.GenerateFrontendGroupConfigMap(c.Cluster.Namespace, c.Cluster.Name, c.ComponentKind, configData, frontend.Name)
}

func (c *CommonBuilder) GeneratePodTemplateSpec(template *v1alpha1.PodTemplateSpec) *corev1.PodTemplateSpec {
	return common.GeneratePodTemplateSpec(c.ComponentKind, template)
}

func (c *CommonBuilder) GeneratePodMonitor() (*monitoringv1.PodMonitor, error) {
	return common.GeneratePodMonitor(c.Cluster.Namespace, c.Cluster.Name, c.ComponentKind, c.Cluster.Spec.PrometheusMonitor)
}

func (c *CommonBuilder) GenerateFrontendGroupPodMonitor(frontend *v1alpha1.FrontendSpec) (*monitoringv1.PodMonitor, error) {
	return common.GenerateFrontendGroupPodMonitor(c.Cluster.Namespace, c.Cluster.Name, c.ComponentKind, c.Cluster.Spec.PrometheusMonitor, frontend.Name)
}

// MountConfigDir mounts the configmap to the main container as '/etc/greptimedb/config.toml'.
func (c *CommonBuilder) MountConfigDir(template *corev1.PodTemplateSpec) {
	common.MountConfigDir(c.Cluster.Name, c.ComponentKind, template)
}

// AddLogsVolume will create a shared volume for logs and mount it to the main container and sidecar container.
func (c *CommonBuilder) AddLogsVolume(template *corev1.PodTemplateSpec, mountPath string) {
	template.Spec.Volumes = append(template.Spec.Volumes, corev1.Volume{
		Name: constant.DefaultLogsVolumeName,
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		},
	})

	template.Spec.Containers[constant.MainContainerIndex].VolumeMounts = append(template.Spec.Containers[constant.MainContainerIndex].VolumeMounts, corev1.VolumeMount{
		Name:      constant.DefaultLogsVolumeName,
		MountPath: mountPath,
	})
}

func (c *CommonBuilder) AddVectorConfigVolume(template *corev1.PodTemplateSpec) {
	template.Spec.Volumes = append(template.Spec.Volumes, corev1.Volume{
		Name: constant.DefaultVectorConfigName,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: constant.DefaultVectorConfigName,
				},
			},
		},
	})
}

func (c *CommonBuilder) AddVectorSidecar(template *corev1.PodTemplateSpec, kind v1alpha1.ComponentKind) {
	template.Spec.Containers = append(template.Spec.Containers, corev1.Container{
		Name:  "vector",
		Image: c.Cluster.Spec.Monitoring.Vector.Image,
		Args: []string{
			"--config", "/etc/vector/vector.yaml",
		},
		Env:       c.env(kind),
		Resources: c.Cluster.Spec.Monitoring.Vector.Resources,
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      constant.DefaultLogsVolumeName,
				MountPath: "/logs",
			},
			{
				Name:      constant.DefaultVectorConfigName,
				MountPath: "/etc/vector",
			},
		},
	})
}

func (c *CommonBuilder) GenerateVectorConfigMap() (*corev1.ConfigMap, error) {
	standaloneName := common.ResourceName(c.Cluster.Name+"-monitor", v1alpha1.StandaloneKind)
	svc := fmt.Sprintf("%s.%s.svc.cluster.local", standaloneName, c.Cluster.Namespace)
	vars := map[string]string{
		"ClusterName":          c.Cluster.Name,
		"LogsTableName":        constant.LogsTableName,
		"SlowQueriesTableName": constant.SlowQueriesTableName,
		"PipelineName":         common.LogsPipelineName(c.Cluster.Namespace, c.Cluster.Name),
		"LoggingService":       fmt.Sprintf("http://%s:%d", svc, v1alpha1.DefaultHTTPPort),
		"MetricService":        fmt.Sprintf("http://%s:%d/v1/prometheus/write?db=public", svc, v1alpha1.DefaultHTTPPort),
	}

	vectorConfigTemplate, err := c.vectorConfigTemplate()
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	tmpl, err := template.New("vector-config").Parse(vectorConfigTemplate)
	if err != nil {
		return nil, err
	}

	if err := tmpl.Execute(&buf, vars); err != nil {
		return nil, err
	}

	configmap := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      constant.DefaultVectorConfigName,
			Namespace: c.Cluster.Namespace,
		},
		Data: map[string]string{
			"vector.yaml": buf.String(),
		},
	}

	return configmap, nil
}

func (c *CommonBuilder) vectorConfigTemplate() (string, error) {
	data, err := fs.ReadFile(config.VectorConfigTemplate, "vector-config-template.yaml")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (c *CommonBuilder) env(kind v1alpha1.ComponentKind) []corev1.EnvVar {
	return []corev1.EnvVar{
		{
			Name: deployer.EnvPodIP,
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "status.podIP",
				},
			},
		},
		{
			Name: deployer.EnvPodName,
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "metadata.name",
				},
			},
		},
		{
			Name: deployer.EnvPodNamespace,
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "metadata.namespace",
				},
			},
		},
		{
			Name:  deployer.EnvRole,
			Value: string(kind),
		},
	}
}

func UpdateStatus(ctx context.Context, input *v1alpha1.GreptimeDBCluster, kc client.Client, opts ...client.SubResourceUpdateOption) error {
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
