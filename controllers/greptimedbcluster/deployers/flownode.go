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
	"path"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/GreptimeTeam/greptimedb-operator/apis/v1alpha1"
	"github.com/GreptimeTeam/greptimedb-operator/controllers/common"
	"github.com/GreptimeTeam/greptimedb-operator/controllers/constant"
	"github.com/GreptimeTeam/greptimedb-operator/pkg/dbconfig"
	"github.com/GreptimeTeam/greptimedb-operator/pkg/deployer"
	"github.com/GreptimeTeam/greptimedb-operator/pkg/util"
	k8sutil "github.com/GreptimeTeam/greptimedb-operator/pkg/util/k8s"
)

// FlownodeDeployer is the deployer for flownode.
type FlownodeDeployer struct {
	*CommonDeployer
}

var _ deployer.Deployer = &FlownodeDeployer{}

func NewFlownodeDeployer(mgr ctrl.Manager) *FlownodeDeployer {
	return &FlownodeDeployer{
		CommonDeployer: NewFromManager(mgr),
	}
}

func (d *FlownodeDeployer) NewBuilder(crdObject client.Object) deployer.Builder {
	return &flownodeBuilder{
		CommonBuilder: d.NewCommonBuilder(crdObject, v1alpha1.FlownodeComponentKind),
	}
}

func (d *FlownodeDeployer) Generate(crdObject client.Object) ([]client.Object, error) {
	objects, err := d.NewBuilder(crdObject).
		BuildService().
		BuildConfigMap().
		BuildStatefulSet().
		BuildPodMonitor().
		SetControllerAndAnnotation().
		Generate()

	if err != nil {
		return nil, err
	}

	return objects, nil
}

func (d *FlownodeDeployer) CleanUp(_ context.Context, _ client.Object) error {
	return nil
}

func (d *FlownodeDeployer) CheckAndUpdateStatus(ctx context.Context, crdObject client.Object) (bool, error) {
	cluster, err := d.GetCluster(crdObject)
	if err != nil {
		return false, err
	}

	var (
		sts = new(appsv1.StatefulSet)

		objectKey = client.ObjectKey{
			Namespace: cluster.Namespace,
			Name:      common.ResourceName(cluster.Name, v1alpha1.FlownodeComponentKind),
		}
	)

	err = d.Get(ctx, objectKey, sts)
	if errors.IsNotFound(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	cluster.Status.Flownode.Replicas = *sts.Spec.Replicas
	cluster.Status.Flownode.ReadyReplicas = sts.Status.ReadyReplicas
	if err := UpdateStatus(ctx, cluster, d.Client); err != nil {
		klog.Errorf("Failed to update status: %s", err)
	}

	return k8sutil.IsStatefulSetReady(sts), nil
}

var _ deployer.Builder = &flownodeBuilder{}

type flownodeBuilder struct {
	*CommonBuilder
}

func (b *flownodeBuilder) BuildService() deployer.Builder {
	if b.Err != nil {
		return b
	}

	if b.Cluster.Spec.Flownode == nil {
		return b
	}

	svc := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: b.Cluster.Namespace,
			Name:      common.ResourceName(b.Cluster.Name, b.ComponentKind),
		},
		Spec: corev1.ServiceSpec{
			ClusterIP: corev1.ClusterIPNone,
			Selector: map[string]string{
				constant.GreptimeDBComponentName: common.ResourceName(b.Cluster.Name, b.ComponentKind),
			},
			Ports: b.servicePorts(),
		},
	}

	b.Objects = append(b.Objects, svc)

	return b
}

func (b *flownodeBuilder) BuildConfigMap() deployer.Builder {
	if b.Err != nil {
		return b
	}

	if b.Cluster.Spec.Flownode == nil {
		return b
	}

	cm, err := b.GenerateConfigMap()
	if err != nil {
		b.Err = err
		return b
	}

	b.Objects = append(b.Objects, cm)

	return b
}

func (b *flownodeBuilder) BuildStatefulSet() deployer.Builder {
	if b.Err != nil {
		return b
	}

	if b.Cluster.Spec.Flownode == nil {
		return b
	}

	sts := &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "StatefulSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      common.ResourceName(b.Cluster.Name, b.ComponentKind),
			Namespace: b.Cluster.Namespace,
		},
		Spec: appsv1.StatefulSetSpec{
			PodManagementPolicy: appsv1.ParallelPodManagement,
			ServiceName:         common.ResourceName(b.Cluster.Name, b.ComponentKind),
			Replicas:            b.Cluster.Spec.Flownode.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					constant.GreptimeDBComponentName: common.ResourceName(b.Cluster.Name, b.ComponentKind),
				},
			},
			Template: b.generatePodTemplateSpec(),
		},
	}

	configData, err := dbconfig.FromCluster(b.Cluster, b.ComponentKind)
	if err != nil {
		b.Err = err
		return b
	}

	sts.Spec.Template.Annotations = util.MergeStringMap(sts.Spec.Template.Annotations,
		map[string]string{deployer.ConfigHash: util.CalculateConfigHash(configData)})

	b.Objects = append(b.Objects, sts)

	return b
}

func (b *flownodeBuilder) BuildPodMonitor() deployer.Builder {
	if b.Err != nil {
		return b
	}

	if b.Cluster.Spec.Flownode == nil {
		return b
	}

	if b.Cluster.Spec.PrometheusMonitor == nil || !b.Cluster.Spec.PrometheusMonitor.Enabled {
		return b
	}

	pm, err := b.GeneratePodMonitor()
	if err != nil {
		b.Err = err
		return b
	}

	b.Objects = append(b.Objects, pm)

	return b
}

func (b *flownodeBuilder) generateMainContainerArgs() []string {
	return []string{
		"flownode", "start",
		"--metasrv-addrs", fmt.Sprintf("%s.%s:%d", common.ResourceName(b.Cluster.Name, v1alpha1.MetaComponentKind), b.Cluster.Namespace, b.Cluster.Spec.Meta.RPCPort),
		"--config-file", path.Join(constant.GreptimeDBConfigDir, constant.GreptimeDBConfigFileName),
	}
}

func (b *flownodeBuilder) generatePodTemplateSpec() corev1.PodTemplateSpec {
	podTemplateSpec := b.GeneratePodTemplateSpec(b.Cluster.Spec.Flownode.Template)

	if len(b.Cluster.Spec.Flownode.Template.MainContainer.Args) == 0 {
		// Setup main container args.
		podTemplateSpec.Spec.Containers[constant.MainContainerIndex].Args = b.generateMainContainerArgs()
	}

	b.mountConfigDir(podTemplateSpec)
	b.addInitConfigDirVolume(podTemplateSpec)

	podTemplateSpec.Spec.Containers[constant.MainContainerIndex].Ports = b.containerPorts()
	podTemplateSpec.Spec.InitContainers = append(podTemplateSpec.Spec.InitContainers, *b.generateInitializer())
	podTemplateSpec.ObjectMeta.Labels = util.MergeStringMap(podTemplateSpec.ObjectMeta.Labels, map[string]string{
		constant.GreptimeDBComponentName: common.ResourceName(b.Cluster.Name, b.ComponentKind),
	})

	return *podTemplateSpec
}

func (b *flownodeBuilder) generateInitializer() *corev1.Container {
	initializer := &corev1.Container{
		Name:  "initializer",
		Image: b.Cluster.Spec.Initializer.Image,
		Command: []string{
			"greptimedb-initializer",
		},
		Args: []string{
			"--config-path", path.Join(constant.GreptimeDBConfigDir, constant.GreptimeDBConfigFileName),
			"--init-config-path", path.Join(constant.GreptimeDBInitConfigDir, constant.GreptimeDBConfigFileName),
			"--rpc-port", fmt.Sprintf("%d", b.Cluster.Spec.Flownode.RPCPort),
			"--service-name", common.ResourceName(b.Cluster.Name, b.ComponentKind),
			"--namespace", b.Cluster.Namespace,
			"--component-kind", string(b.ComponentKind),
		},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      constant.ConfigVolumeName,
				MountPath: constant.GreptimeDBConfigDir,
			},
			{
				Name:      constant.InitConfigVolumeName,
				MountPath: constant.GreptimeDBInitConfigDir,
			},
		},

		// TODO(zyy17): the flownode don't support to accept hostname.
		Env: []corev1.EnvVar{
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
		},
	}

	return initializer
}

func (b *flownodeBuilder) mountConfigDir(template *corev1.PodTemplateSpec) {
	// The empty-dir will be modified by initializer.
	template.Spec.Volumes = append(template.Spec.Volumes, corev1.Volume{
		Name: constant.ConfigVolumeName,
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
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

// The init-config volume is used for initializer.
func (b *flownodeBuilder) addInitConfigDirVolume(template *corev1.PodTemplateSpec) {
	template.Spec.Volumes = append(template.Spec.Volumes, corev1.Volume{
		Name: constant.InitConfigVolumeName,
		VolumeSource: corev1.VolumeSource{
			// Mount the configmap as init-config.
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: common.ResourceName(b.Cluster.Name, b.ComponentKind),
				},
			},
		},
	})
}

func (b *flownodeBuilder) servicePorts() []corev1.ServicePort {
	return []corev1.ServicePort{
		{
			Name:     "rpc",
			Protocol: corev1.ProtocolTCP,
			Port:     b.Cluster.Spec.Flownode.RPCPort,
		},
	}
}

func (b *flownodeBuilder) containerPorts() []corev1.ContainerPort {
	return []corev1.ContainerPort{
		{
			Name:          "grpc",
			Protocol:      corev1.ProtocolTCP,
			ContainerPort: b.Cluster.Spec.Flownode.RPCPort,
		},
	}
}
