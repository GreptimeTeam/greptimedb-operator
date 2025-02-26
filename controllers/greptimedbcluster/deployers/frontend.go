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
	"k8s.io/apimachinery/pkg/util/intstr"
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

type FrontendDeployer struct {
	*CommonDeployer
}

var _ deployer.Deployer = &FrontendDeployer{}

func NewFrontendDeployer(mgr ctrl.Manager) *FrontendDeployer {
	return &FrontendDeployer{
		CommonDeployer: NewFromManager(mgr),
	}
}

func (d *FrontendDeployer) NewBuilder(crdObject client.Object) deployer.Builder {
	return &frontendBuilder{
		CommonBuilder: d.NewCommonBuilder(crdObject, v1alpha1.FrontendComponentKind),
	}
}

func (d *FrontendDeployer) Generate(crdObject client.Object) ([]client.Object, error) {
	objects, err := d.NewBuilder(crdObject).
		BuildService().
		BuildConfigMap().
		BuildDeployment().
		BuildPodMonitor().
		BuildFrontendGroupService().
		BuildFrontendGroupConfigMap().
		BuildFrontendGroupDeployment().
		BuildFrontendGroupPodMonitor().
		SetControllerAndAnnotation().
		Generate()

	if err != nil {
		return nil, err
	}

	return objects, nil
}

func (d *FrontendDeployer) CheckAndUpdateStatus(ctx context.Context, crdObject client.Object) (bool, error) {
	cluster, err := d.GetCluster(crdObject)
	if err != nil {
		return false, err
	}

	var (
		deployment    = new(appsv1.Deployment)
		replicas      int32
		readyReplicas int32
	)

	if cluster.GetFrontendGroup() == nil {
		objectKey := client.ObjectKey{
			Namespace: cluster.Namespace,
			Name:      common.ResourceName(cluster.Name, v1alpha1.FrontendComponentKind),
		}

		err = d.Get(ctx, objectKey, deployment)
		if errors.IsNotFound(err) {
			return false, nil
		}
		if err != nil {
			return false, err
		}
		replicas = *deployment.Spec.Replicas
		readyReplicas = deployment.Status.ReadyReplicas
	} else {
		for _, frontend := range cluster.GetFrontendGroup() {
			objectKey := client.ObjectKey{
				Namespace: cluster.Namespace,
				Name:      common.FrontendGroupResourceName(cluster.Name, v1alpha1.FrontendComponentKind, frontend.Name),
			}

			err = d.Get(ctx, objectKey, deployment)
			if errors.IsNotFound(err) {
				return false, nil
			}
			if err != nil {
				return false, err
			}
			replicas += *deployment.Spec.Replicas
			readyReplicas += deployment.Status.ReadyReplicas
		}
	}

	cluster.Status.Frontend.Replicas = replicas
	cluster.Status.Frontend.ReadyReplicas = readyReplicas
	if err := UpdateStatus(ctx, cluster, d.Client); err != nil {
		klog.Errorf("Failed to update status: %s", err)
	}

	return k8sutil.IsDeploymentReady(deployment), nil
}

var _ deployer.Builder = &frontendBuilder{}

type frontendBuilder struct {
	*CommonBuilder
}

func (b *frontendBuilder) BuildService() deployer.Builder {
	if b.Err != nil {
		return b
	}

	if b.Cluster.Spec.Frontend == nil {
		return b
	}

	svc := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   b.Cluster.Namespace,
			Name:        common.ResourceName(b.Cluster.Name, b.ComponentKind),
			Annotations: b.Cluster.Spec.Frontend.Service.Annotations,
			Labels: util.MergeStringMap(b.Cluster.Spec.Frontend.Service.Labels, map[string]string{
				constant.GreptimeDBComponentName: common.ResourceName(b.Cluster.Name, b.ComponentKind),
			}),
		},
		Spec: corev1.ServiceSpec{
			Type: b.Cluster.Spec.Frontend.Service.Type,
			Selector: map[string]string{
				constant.GreptimeDBComponentName: common.ResourceName(b.Cluster.Name, b.ComponentKind),
			},
			Ports:             b.servicePorts(),
			LoadBalancerClass: b.Cluster.Spec.Frontend.Service.LoadBalancerClass,
		},
	}

	b.Objects = append(b.Objects, svc)

	return b
}

func (b *frontendBuilder) BuildDeployment() deployer.Builder {
	if b.Err != nil {
		return b
	}

	if b.Cluster.Spec.Frontend == nil {
		return b
	}

	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      common.ResourceName(b.Cluster.Name, b.ComponentKind),
			Namespace: b.Cluster.Namespace,
			Labels: map[string]string{
				constant.GreptimeDBComponentName: common.ResourceName(b.Cluster.Name, b.ComponentKind),
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: b.Cluster.Spec.Frontend.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					constant.GreptimeDBComponentName: common.ResourceName(b.Cluster.Name, b.ComponentKind),
				},
			},
			Template: *b.generatePodTemplateSpec(),
			Strategy: appsv1.DeploymentStrategy{
				Type:          appsv1.RollingUpdateDeploymentStrategyType,
				RollingUpdate: b.Cluster.Spec.Frontend.RollingUpdate,
			},
		},
	}

	configData, err := dbconfig.FromCluster(b.Cluster, b.ComponentKind)
	if err != nil {
		b.Err = err
		return b
	}

	deployment.Spec.Template.Annotations = util.MergeStringMap(deployment.Spec.Template.Annotations,
		map[string]string{deployer.ConfigHash: util.CalculateConfigHash(configData)})

	b.Objects = append(b.Objects, deployment)

	return b
}

func (b *frontendBuilder) BuildConfigMap() deployer.Builder {
	if b.Err != nil {
		return b
	}

	if b.Cluster.Spec.Frontend == nil {
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

func (b *frontendBuilder) BuildPodMonitor() deployer.Builder {
	if b.Err != nil {
		return b
	}

	if b.Cluster.Spec.Frontend == nil {
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

func (b *frontendBuilder) Generate() ([]client.Object, error) {
	return b.Objects, b.Err
}

func (b *frontendBuilder) generateMainContainerArgs() []string {
	var args = []string{
		"frontend", "start",
		"--rpc-bind-addr", fmt.Sprintf("0.0.0.0:%d", b.Cluster.Spec.Frontend.RPCPort),
		"--rpc-server-addr", fmt.Sprintf("$(%s):%d", deployer.EnvPodIP, b.Cluster.Spec.Frontend.RPCPort),
		"--metasrv-addrs", fmt.Sprintf("%s.%s:%d", common.ResourceName(b.Cluster.Name, v1alpha1.MetaComponentKind),
			b.Cluster.Namespace, b.Cluster.Spec.Meta.RPCPort),
		"--http-addr", fmt.Sprintf("0.0.0.0:%d", b.Cluster.Spec.Frontend.HTTPPort),
		"--mysql-addr", fmt.Sprintf("0.0.0.0:%d", b.Cluster.Spec.Frontend.MySQLPort),
		"--postgres-addr", fmt.Sprintf("0.0.0.0:%d", b.Cluster.Spec.Frontend.PostgreSQLPort),
		"--config-file", path.Join(constant.GreptimeDBConfigDir, constant.GreptimeDBConfigFileName),
	}

	if b.Cluster.Spec.Frontend != nil && b.Cluster.Spec.Frontend.TLS != nil {
		args = append(args, []string{
			"--tls-mode", constant.DefaultTLSMode,
			"--tls-cert-path", path.Join(constant.GreptimeDBTLSDir, v1alpha1.TLSCrtSecretKey),
			"--tls-key-path", path.Join(constant.GreptimeDBTLSDir, v1alpha1.TLSKeySecretKey),
		}...)
	}

	return args
}

func (b *frontendBuilder) generatePodTemplateSpec() *corev1.PodTemplateSpec {
	podTemplateSpec := b.GeneratePodTemplateSpec(b.Cluster.Spec.Frontend.Template)

	if len(b.Cluster.Spec.Frontend.Template.MainContainer.Args) == 0 {
		// Setup main container args.
		podTemplateSpec.Spec.Containers[constant.MainContainerIndex].Args = b.generateMainContainerArgs()
	}

	podTemplateSpec.ObjectMeta.Labels = util.MergeStringMap(podTemplateSpec.ObjectMeta.Labels, map[string]string{
		constant.GreptimeDBComponentName: common.ResourceName(b.Cluster.Name, b.ComponentKind),
	})

	podTemplateSpec.Spec.Containers[constant.MainContainerIndex].Ports = b.containerPorts()
	podTemplateSpec.Spec.Containers[constant.MainContainerIndex].Env = append(podTemplateSpec.Spec.Containers[constant.MainContainerIndex].Env, b.env(v1alpha1.FrontendComponentKind)...)

	b.MountConfigDir(podTemplateSpec)

	if logging := b.Cluster.GetFrontend().GetLogging(); logging != nil && !logging.IsOnlyLogToStdout() {
		b.AddLogsVolume(podTemplateSpec, logging.GetLogsDir())
	}

	if b.Cluster.GetMonitoring().IsEnabled() && b.Cluster.GetMonitoring().GetVector() != nil {
		b.AddVectorConfigVolume(podTemplateSpec)
		b.AddVectorSidecar(podTemplateSpec, v1alpha1.FrontendComponentKind)
	}

	if b.Cluster.Spec.Frontend.TLS != nil {
		b.mountTLSSecret(podTemplateSpec)
	}

	return podTemplateSpec
}

func (b *frontendBuilder) mountTLSSecret(template *corev1.PodTemplateSpec) {
	template.Spec.Volumes = append(template.Spec.Volumes, corev1.Volume{
		Name: constant.TLSVolumeName,
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: b.Cluster.Spec.Frontend.TLS.SecretName,
			},
		},
	})

	template.Spec.Containers[constant.MainContainerIndex].VolumeMounts =
		append(template.Spec.Containers[constant.MainContainerIndex].VolumeMounts,
			corev1.VolumeMount{
				Name:      constant.TLSVolumeName,
				MountPath: constant.GreptimeDBTLSDir,
				ReadOnly:  true,
			},
		)
}

func (b *frontendBuilder) servicePorts() []corev1.ServicePort {
	return []corev1.ServicePort{
		{
			Name:       "rpc",
			Protocol:   corev1.ProtocolTCP,
			Port:       b.Cluster.Spec.RPCPort,
			TargetPort: intstr.FromInt32(b.Cluster.Spec.Frontend.RPCPort),
		},
		{
			Name:       "http",
			Protocol:   corev1.ProtocolTCP,
			Port:       b.Cluster.Spec.HTTPPort,
			TargetPort: intstr.FromInt32(b.Cluster.Spec.Frontend.HTTPPort),
		},
		{
			Name:       "mysql",
			Protocol:   corev1.ProtocolTCP,
			Port:       b.Cluster.Spec.MySQLPort,
			TargetPort: intstr.FromInt32(b.Cluster.Spec.Frontend.MySQLPort),
		},
		{
			Name:       "pg",
			Protocol:   corev1.ProtocolTCP,
			Port:       b.Cluster.Spec.PostgreSQLPort,
			TargetPort: intstr.FromInt32(b.Cluster.Spec.Frontend.PostgreSQLPort),
		},
	}
}

func (b *frontendBuilder) containerPorts() []corev1.ContainerPort {
	return []corev1.ContainerPort{
		{
			Name:          "rpc",
			Protocol:      corev1.ProtocolTCP,
			ContainerPort: b.Cluster.Spec.Frontend.RPCPort,
		},
		{
			Name:          "http",
			Protocol:      corev1.ProtocolTCP,
			ContainerPort: b.Cluster.Spec.Frontend.HTTPPort,
		},
		{
			Name:          "mysql",
			Protocol:      corev1.ProtocolTCP,
			ContainerPort: b.Cluster.Spec.Frontend.MySQLPort,
		},
		{
			Name:          "pg",
			Protocol:      corev1.ProtocolTCP,
			ContainerPort: b.Cluster.Spec.Frontend.PostgreSQLPort,
		},
	}
}

func (b *frontendBuilder) BuildFrontendGroupService() deployer.Builder {
	if b.Err != nil {
		return b
	}

	if b.Cluster.Spec.FrontendGroup == nil {
		return b
	}

	for _, frontend := range b.Cluster.Spec.FrontendGroup {
		svc := &corev1.Service{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Service",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Namespace:   b.Cluster.Namespace,
				Name:        common.FrontendGroupResourceName(b.Cluster.Name, b.ComponentKind, frontend.Name),
				Annotations: frontend.Service.Annotations,
				Labels: util.MergeStringMap(frontend.Service.Labels, map[string]string{
					constant.GreptimeDBComponentName: common.FrontendGroupResourceName(b.Cluster.Name, b.ComponentKind, frontend.Name),
				}),
			},
			Spec: corev1.ServiceSpec{
				Type: frontend.Service.Type,
				Selector: map[string]string{
					constant.GreptimeDBComponentName: common.FrontendGroupResourceName(b.Cluster.Name, b.ComponentKind, frontend.Name),
				},
				Ports:             b.frontendGroupServicePorts(frontend),
				LoadBalancerClass: frontend.Service.LoadBalancerClass,
			},
		}

		b.Objects = append(b.Objects, svc)
	}

	return b
}

func (b *frontendBuilder) BuildFrontendGroupDeployment() deployer.Builder {
	if b.Err != nil {
		return b
	}

	if b.Cluster.Spec.FrontendGroup == nil {
		return b
	}

	for _, frontend := range b.Cluster.Spec.FrontendGroup {
		deployment := &appsv1.Deployment{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Deployment",
				APIVersion: "apps/v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      common.FrontendGroupResourceName(b.Cluster.Name, b.ComponentKind, frontend.Name),
				Namespace: b.Cluster.Namespace,
				Labels: map[string]string{
					constant.GreptimeDBComponentName: common.FrontendGroupResourceName(b.Cluster.Name, b.ComponentKind, frontend.Name),
				},
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: frontend.Replicas,
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						constant.GreptimeDBComponentName: common.FrontendGroupResourceName(b.Cluster.Name, b.ComponentKind, frontend.Name),
					},
				},
				Template: *b.generateFrontendGroupPodTemplateSpec(frontend),
				Strategy: appsv1.DeploymentStrategy{
					Type:          appsv1.RollingUpdateDeploymentStrategyType,
					RollingUpdate: frontend.RollingUpdate,
				},
			},
		}

		configData, err := dbconfig.FromFrontendGroup(frontend, b.ComponentKind)
		if err != nil {
			b.Err = err
			return b
		}

		deployment.Spec.Template.Annotations = util.MergeStringMap(deployment.Spec.Template.Annotations,
			map[string]string{deployer.ConfigHash: util.CalculateConfigHash(configData)})

		b.Objects = append(b.Objects, deployment)
	}

	return b
}

func (b *frontendBuilder) frontendGroupServicePorts(frontend *v1alpha1.FrontendSpec) []corev1.ServicePort {
	return []corev1.ServicePort{
		{
			Name:       "rpc",
			Protocol:   corev1.ProtocolTCP,
			Port:       b.Cluster.Spec.RPCPort,
			TargetPort: intstr.FromInt32(frontend.RPCPort),
		},
		{
			Name:       "http",
			Protocol:   corev1.ProtocolTCP,
			Port:       b.Cluster.Spec.HTTPPort,
			TargetPort: intstr.FromInt32(frontend.HTTPPort),
		},
		{
			Name:       "mysql",
			Protocol:   corev1.ProtocolTCP,
			Port:       b.Cluster.Spec.MySQLPort,
			TargetPort: intstr.FromInt32(frontend.MySQLPort),
		},
		{
			Name:       "pg",
			Protocol:   corev1.ProtocolTCP,
			Port:       b.Cluster.Spec.PostgreSQLPort,
			TargetPort: intstr.FromInt32(frontend.PostgreSQLPort),
		},
	}
}

func (b *frontendBuilder) frontendGroupContainerPorts(frontend *v1alpha1.FrontendSpec) []corev1.ContainerPort {
	return []corev1.ContainerPort{
		{
			Name:          "rpc",
			Protocol:      corev1.ProtocolTCP,
			ContainerPort: frontend.RPCPort,
		},
		{
			Name:          "http",
			Protocol:      corev1.ProtocolTCP,
			ContainerPort: frontend.HTTPPort,
		},
		{
			Name:          "mysql",
			Protocol:      corev1.ProtocolTCP,
			ContainerPort: frontend.MySQLPort,
		},
		{
			Name:          "pg",
			Protocol:      corev1.ProtocolTCP,
			ContainerPort: frontend.PostgreSQLPort,
		},
	}
}

func (b *frontendBuilder) generateFrontendGroupMainContainerArgs(frontend *v1alpha1.FrontendSpec) []string {
	var args = []string{
		"frontend", "start",
		"--rpc-bind-addr", fmt.Sprintf("0.0.0.0:%d", frontend.RPCPort),
		"--rpc-server-addr", fmt.Sprintf("$(%s):%d", deployer.EnvPodIP, frontend.RPCPort),
		"--metasrv-addrs", fmt.Sprintf("%s.%s:%d", common.ResourceName(b.Cluster.Name, v1alpha1.MetaComponentKind),
			b.Cluster.Namespace, b.Cluster.Spec.Meta.RPCPort),
		"--http-addr", fmt.Sprintf("0.0.0.0:%d", frontend.HTTPPort),
		"--mysql-addr", fmt.Sprintf("0.0.0.0:%d", frontend.MySQLPort),
		"--postgres-addr", fmt.Sprintf("0.0.0.0:%d", frontend.PostgreSQLPort),
		"--config-file", path.Join(constant.GreptimeDBConfigDir, constant.GreptimeDBConfigFileName),
	}

	if frontend != nil && frontend.TLS != nil {
		args = append(args, []string{
			"--tls-mode", constant.DefaultTLSMode,
			"--tls-cert-path", path.Join(constant.GreptimeDBTLSDir, v1alpha1.TLSCrtSecretKey),
			"--tls-key-path", path.Join(constant.GreptimeDBTLSDir, v1alpha1.TLSKeySecretKey),
		}...)
	}

	return args
}

func (b *frontendBuilder) generateFrontendGroupPodTemplateSpec(frontend *v1alpha1.FrontendSpec) *corev1.PodTemplateSpec {
	podTemplateSpec := b.GeneratePodTemplateSpec(frontend.Template)

	if len(frontend.Template.MainContainer.Args) == 0 {
		podTemplateSpec.Spec.Containers[constant.MainContainerIndex].Args = b.generateFrontendGroupMainContainerArgs(frontend)
	}

	podTemplateSpec.ObjectMeta.Labels = util.MergeStringMap(podTemplateSpec.ObjectMeta.Labels, map[string]string{
		constant.GreptimeDBComponentName: common.FrontendGroupResourceName(b.Cluster.Name, b.ComponentKind, frontend.Name),
	})

	podTemplateSpec.Spec.Containers[constant.MainContainerIndex].Ports = b.frontendGroupContainerPorts(frontend)
	podTemplateSpec.Spec.Containers[constant.MainContainerIndex].Env = append(podTemplateSpec.Spec.Containers[constant.MainContainerIndex].Env, b.env(v1alpha1.FrontendComponentKind)...)

	b.MountFrontendGroupConfigDir(podTemplateSpec, frontend.Name)

	if logging := frontend.GetLogging(); logging != nil && !logging.IsOnlyLogToStdout() {
		b.AddLogsVolume(podTemplateSpec, logging.GetLogsDir())
	}

	if b.Cluster.GetMonitoring().IsEnabled() && b.Cluster.GetMonitoring().GetVector() != nil {
		b.AddVectorConfigVolume(podTemplateSpec)
		b.AddVectorSidecar(podTemplateSpec, v1alpha1.FrontendComponentKind)
	}

	if frontend.TLS != nil && len(frontend.TLS.SecretName) != 0 {
		b.mountFrontendGroupTLSSecret(podTemplateSpec, frontend.TLS.SecretName)
	}

	return podTemplateSpec
}

func (b *frontendBuilder) mountFrontendGroupTLSSecret(template *corev1.PodTemplateSpec, secretName string) {
	template.Spec.Volumes = append(template.Spec.Volumes, corev1.Volume{
		Name: constant.TLSVolumeName,
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: secretName,
			},
		},
	})

	template.Spec.Containers[constant.MainContainerIndex].VolumeMounts =
		append(template.Spec.Containers[constant.MainContainerIndex].VolumeMounts,
			corev1.VolumeMount{
				Name:      constant.TLSVolumeName,
				MountPath: constant.GreptimeDBTLSDir,
				ReadOnly:  true,
			},
		)
}

func (b *frontendBuilder) BuildFrontendGroupConfigMap() deployer.Builder {
	if b.Err != nil {
		return b
	}

	if b.Cluster.Spec.FrontendGroup == nil {
		return b
	}

	for _, frontend := range b.Cluster.Spec.FrontendGroup {
		cm, err := b.GenerateFrontendGroupConfigMap(frontend)
		if err != nil {
			b.Err = err
			return b
		}

		b.Objects = append(b.Objects, cm)
	}

	return b
}

func (b *frontendBuilder) BuildFrontendGroupPodMonitor() deployer.Builder {
	if b.Err != nil {
		return b
	}

	if b.Cluster.Spec.FrontendGroup == nil {
		return b
	}

	if b.Cluster.Spec.PrometheusMonitor == nil || !b.Cluster.Spec.PrometheusMonitor.Enabled {
		return b
	}

	for _, frontend := range b.Cluster.Spec.FrontendGroup {
		pm, err := b.GenerateFrontendGroupPodMonitor(frontend)
		if err != nil {
			b.Err = err
			return b
		}

		b.Objects = append(b.Objects, pm)
	}

	return b
}
