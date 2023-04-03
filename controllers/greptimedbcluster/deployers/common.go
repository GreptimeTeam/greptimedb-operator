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

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/GreptimeTeam/greptimedb-operator/apis/v1alpha1"
	"github.com/GreptimeTeam/greptimedb-operator/pkg/deployer"
)

const (
	GreptimeComponentName = "app.greptime.io/component"
)

var (
	DefaultConfigPath = "/etc/greptimedb"
)

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

// CommonDeployer is the common deployer for all components of GreptimeDBCluster.
type CommonDeployer struct {
	client.Client
	Scheme *runtime.Scheme

	deployer.DefaultDeployer
}

func NewFromManager(mgr ctrl.Manager) *CommonDeployer {
	return &CommonDeployer{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),

		DefaultDeployer: deployer.DefaultDeployer{
			Client: mgr.GetClient(),
		},
	}
}

func (c *CommonDeployer) ResourceName(clusterName string, componentKind v1alpha1.ComponentKind) string {
	return clusterName + "-" + string(componentKind)
}

func (c *CommonDeployer) GetCluster(crdObject client.Object) (*v1alpha1.GreptimeDBCluster, error) {
	cluster, ok := crdObject.(*v1alpha1.GreptimeDBCluster)
	if !ok {
		return nil, fmt.Errorf("the object is not GreptimeDBCluster")
	}
	return cluster, nil
}

func (c *CommonDeployer) GenerateConfigMap(cluster *v1alpha1.GreptimeDBCluster, componentKind v1alpha1.ComponentKind) (*corev1.ConfigMap, error) {
	var config string

	switch componentKind {
	case v1alpha1.MetaComponentKind:
		config = cluster.Spec.Meta.Config
	case v1alpha1.FrontendComponentKind:
		config = cluster.Spec.Frontend.Config
	case v1alpha1.DatanodeComponentKind:
		config = cluster.Spec.Datanode.Config
	}

	configmap := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.ResourceName(cluster.Name, componentKind),
			Namespace: cluster.Namespace,
		},
		Data: map[string]string{
			"config.toml": config,
		},
	}

	if err := deployer.SetControllerAndAnnotation(cluster, configmap, c.Scheme, configmap.Data); err != nil {
		return nil, err
	}

	return configmap, nil
}
