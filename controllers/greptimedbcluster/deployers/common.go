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

	"github.com/GreptimeTeam/greptimedb-operator/controllers/constant"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/GreptimeTeam/greptimedb-operator/apis/v1alpha1"
	"github.com/GreptimeTeam/greptimedb-operator/controllers/common"
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

func (c *CommonBuilder) GeneratePodTemplateSpec(template *v1alpha1.PodTemplateSpec) *corev1.PodTemplateSpec {
	return common.GeneratePodTemplateSpec(c.ComponentKind, template)
}

func (c *CommonBuilder) GeneratePodMonitor() (*monitoringv1.PodMonitor, error) {
	return common.GeneratePodMonitor(c.Cluster.Namespace, c.Cluster.Name, c.ComponentKind, c.Cluster.Spec.PrometheusMonitor)
}

// MountConfigDir mounts the configmap to the main container as '/etc/greptimedb/config.toml'.
func (c *CommonBuilder) MountConfigDir(template *corev1.PodTemplateSpec) {
	common.MountConfigDir(c.Cluster.Name, c.ComponentKind, template)
}

// AddLogsVolume will create a shared volume for logs and mount it to the main container and sidecar container.
func (c *CommonBuilder) AddLogsVolume(template *corev1.PodTemplateSpec, mountPath string) {
	template.Spec.Volumes = append(template.Spec.Volumes, corev1.Volume{
		Name: "logs",
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		},
	})

	template.Spec.Containers[constant.MainContainerIndex].VolumeMounts = append(template.Spec.Containers[constant.MainContainerIndex].VolumeMounts, corev1.VolumeMount{
		Name:      "logs",
		MountPath: mountPath,
	})
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
