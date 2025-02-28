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

	// Check if the .Spec.Replicas of the frontend or frontends deployment equals the .Status.ReadyReplicas to confirm it is ready.
	if cluster.GetFrontend() != nil {
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

		if !k8sutil.IsDeploymentReady(deployment) {
			return false, nil
		}
	} else if len(cluster.GetFrontends()) != 0 {
		for _, frontend := range cluster.GetFrontends() {
			objectKey := client.ObjectKey{
				Namespace: cluster.Namespace,
				Name:      common.AdditionalResourceName(cluster.Name, frontend.GetName(), v1alpha1.FrontendComponentKind),
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

			if !k8sutil.IsDeploymentReady(deployment) {
				return false, nil
			}
		}
	}

	cluster.Status.Frontend.Replicas = replicas
	cluster.Status.Frontend.ReadyReplicas = readyReplicas
	if err := UpdateStatus(ctx, cluster, d.Client); err != nil {
		klog.Errorf("Failed to update status: %s", err)
	}

	return true, nil
}

var _ deployer.Builder = &frontendBuilder{}

type frontendBuilder struct {
	*CommonBuilder
}

func (b *frontendBuilder) generateService(name string, frontendSpec *v1alpha1.FrontendSpec) {
	svc := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   b.Cluster.Namespace,
			Name:        name,
			Annotations: frontendSpec.Service.Annotations,
			Labels: util.MergeStringMap(frontendSpec.Service.Labels, map[string]string{
				constant.GreptimeDBComponentName: name,
			}),
		},
		Spec: corev1.ServiceSpec{
			Type: frontendSpec.Service.Type,
			Selector: map[string]string{
				constant.GreptimeDBComponentName: name,
			},
			Ports:             b.servicePorts(frontendSpec),
			LoadBalancerClass: frontendSpec.Service.LoadBalancerClass,
		},
	}

	b.Objects = append(b.Objects, svc)
}

func (b *frontendBuilder) BuildService() deployer.Builder {
	if b.Err != nil {
		return b
	}

	if b.Cluster.GetFrontend() == nil && len(b.Cluster.GetFrontends()) == 0 {
		return b
	}

	if b.Cluster.GetFrontend() != nil {
		b.generateService(common.ResourceName(b.Cluster.Name, b.ComponentKind), b.Cluster.Spec.Frontend)
	}

	if len(b.Cluster.GetFrontends()) != 0 {
		for _, frontend := range b.Cluster.Spec.Frontends {
			b.generateService(common.AdditionalResourceName(b.Cluster.Name, frontend.GetName(), b.ComponentKind), frontend)
		}
	}

	return b
}

func (b *frontendBuilder) generateDeployment(name string, frontendSpec *v1alpha1.FrontendSpec) {
	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: b.Cluster.Namespace,
			Labels: map[string]string{
				constant.GreptimeDBComponentName: name,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: frontendSpec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					constant.GreptimeDBComponentName: name,
				},
			},
			Template: *b.generatePodTemplateSpec(frontendSpec),
			Strategy: appsv1.DeploymentStrategy{
				Type:          appsv1.RollingUpdateDeploymentStrategyType,
				RollingUpdate: frontendSpec.RollingUpdate,
			},
		},
	}

	configData, err := dbconfig.FromFrontend(b.Frontend, b.ComponentKind)
	if err != nil {
		b.Err = err
		return
	}

	deployment.Spec.Template.Annotations = util.MergeStringMap(deployment.Spec.Template.Annotations,
		map[string]string{deployer.ConfigHash: util.CalculateConfigHash(configData)})

	b.Objects = append(b.Objects, deployment)
}

func (b *frontendBuilder) BuildDeployment() deployer.Builder {
	if b.Err != nil {
		return b
	}

	if b.Cluster.GetFrontend() == nil && len(b.Cluster.GetFrontends()) == 0 {
		return b
	}

	if b.Cluster.GetFrontend() != nil {
		b.generateDeployment(common.ResourceName(b.Cluster.Name, b.ComponentKind), b.Cluster.Spec.Frontend)
	}

	if len(b.Cluster.GetFrontends()) != 0 {
		for _, frontend := range b.Cluster.Spec.Frontends {
			b.generateDeployment(common.AdditionalResourceName(b.Cluster.Name, frontend.GetName(), b.ComponentKind), frontend)
		}
	}

	return b
}

func (b *frontendBuilder) BuildConfigMap() deployer.Builder {
	if b.Err != nil {
		return b
	}

	if b.Cluster.GetFrontend() == nil && len(b.Cluster.GetFrontends()) == 0 {
		return b
	}

	if b.Cluster.GetFrontend() != nil {
		b.Frontend = b.Cluster.GetFrontend()
		cm, err := b.GenerateConfigMap()
		if err != nil {
			b.Err = err
			return b
		}
		b.Objects = append(b.Objects, cm)
	}

	if len(b.Cluster.GetFrontends()) != 0 {
		for _, frontend := range b.Cluster.Spec.Frontends {
			b.Frontend = frontend
			cm, err := b.GenerateConfigMap()
			if err != nil {
				b.Err = err
				return b
			}
			b.Objects = append(b.Objects, cm)
		}
	}

	return b
}

func (b *frontendBuilder) BuildPodMonitor() deployer.Builder {
	if b.Err != nil {
		return b
	}

	if b.Cluster.GetFrontend() == nil && len(b.Cluster.GetFrontends()) == 0 {
		return b
	}

	if b.Cluster.Spec.PrometheusMonitor == nil || !b.Cluster.Spec.PrometheusMonitor.Enabled {
		return b
	}

	if b.Cluster.GetFrontend() != nil {
		b.Frontend = b.Cluster.GetFrontend()
		pm, err := b.GeneratePodMonitor()
		if err != nil {
			b.Err = err
			return b
		}

		b.Objects = append(b.Objects, pm)
	}

	if len(b.Cluster.GetFrontends()) != 0 {
		for _, frontend := range b.Cluster.Spec.Frontends {
			b.Frontend = frontend
			pm, err := b.GeneratePodMonitor()
			if err != nil {
				b.Err = err
				return b
			}
			b.Objects = append(b.Objects, pm)
		}
	}

	return b
}

func (b *frontendBuilder) Generate() ([]client.Object, error) {
	return b.Objects, b.Err
}

func (b *frontendBuilder) generateMainContainerArgs(frontendSpec *v1alpha1.FrontendSpec) []string {
	var args = []string{
		"frontend", "start",
		"--rpc-bind-addr", fmt.Sprintf("0.0.0.0:%d", frontendSpec.RPCPort),
		"--rpc-server-addr", fmt.Sprintf("$(%s):%d", deployer.EnvPodIP, frontendSpec.RPCPort),
		"--metasrv-addrs", fmt.Sprintf("%s.%s:%d", common.ResourceName(b.Cluster.Name, v1alpha1.MetaComponentKind),
			b.Cluster.Namespace, b.Cluster.Spec.Meta.RPCPort),
		"--http-addr", fmt.Sprintf("0.0.0.0:%d", frontendSpec.HTTPPort),
		"--mysql-addr", fmt.Sprintf("0.0.0.0:%d", frontendSpec.MySQLPort),
		"--postgres-addr", fmt.Sprintf("0.0.0.0:%d", frontendSpec.PostgreSQLPort),
		"--config-file", path.Join(constant.GreptimeDBConfigDir, constant.GreptimeDBConfigFileName),
	}

	if frontendSpec.TLS != nil {
		args = append(args, []string{
			"--tls-mode", constant.DefaultTLSMode,
			"--tls-cert-path", path.Join(constant.GreptimeDBTLSDir, v1alpha1.TLSCrtSecretKey),
			"--tls-key-path", path.Join(constant.GreptimeDBTLSDir, v1alpha1.TLSKeySecretKey),
		}...)
	}

	return args
}

func (b *frontendBuilder) generatePodTemplateSpec(frontendSpec *v1alpha1.FrontendSpec) *corev1.PodTemplateSpec {
	podTemplateSpec := b.GeneratePodTemplateSpec(frontendSpec.Template)

	if len(frontendSpec.Template.MainContainer.Args) == 0 {
		// Setup main container args.
		podTemplateSpec.Spec.Containers[constant.MainContainerIndex].Args = b.generateMainContainerArgs(frontendSpec)
	}

	resourceName := common.ResourceName(b.Cluster.Name, b.ComponentKind)
	if len(frontendSpec.GetName()) != 0 {
		resourceName = common.AdditionalResourceName(b.Cluster.Name, frontendSpec.GetName(), b.ComponentKind)
	}
	podTemplateSpec.ObjectMeta.Labels = util.MergeStringMap(podTemplateSpec.ObjectMeta.Labels, map[string]string{
		constant.GreptimeDBComponentName: resourceName,
	})

	podTemplateSpec.Spec.Containers[constant.MainContainerIndex].Ports = b.containerPorts(frontendSpec)
	podTemplateSpec.Spec.Containers[constant.MainContainerIndex].Env = append(podTemplateSpec.Spec.Containers[constant.MainContainerIndex].Env, b.env(v1alpha1.FrontendComponentKind)...)

	b.Frontend = frontendSpec
	b.MountConfigDir(podTemplateSpec)

	if logging := frontendSpec.GetLogging(); logging != nil && !logging.IsOnlyLogToStdout() {
		b.AddLogsVolume(podTemplateSpec, logging.GetLogsDir())
	}

	if b.Cluster.GetMonitoring().IsEnabled() && b.Cluster.GetMonitoring().GetVector() != nil {
		b.AddVectorConfigVolume(podTemplateSpec)
		b.AddVectorSidecar(podTemplateSpec, v1alpha1.FrontendComponentKind)
	}

	if frontendSpec.TLS != nil {
		b.mountTLSSecret(frontendSpec.TLS.SecretName, podTemplateSpec)
	}

	return podTemplateSpec
}

func (b *frontendBuilder) mountTLSSecret(secretName string, template *corev1.PodTemplateSpec) {
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

func (b *frontendBuilder) servicePorts(frontendSpec *v1alpha1.FrontendSpec) []corev1.ServicePort {
	return []corev1.ServicePort{
		{
			Name:       "rpc",
			Protocol:   corev1.ProtocolTCP,
			Port:       b.Cluster.Spec.RPCPort,
			TargetPort: intstr.FromInt32(frontendSpec.RPCPort),
		},
		{
			Name:       "http",
			Protocol:   corev1.ProtocolTCP,
			Port:       b.Cluster.Spec.HTTPPort,
			TargetPort: intstr.FromInt32(frontendSpec.HTTPPort),
		},
		{
			Name:       "mysql",
			Protocol:   corev1.ProtocolTCP,
			Port:       b.Cluster.Spec.MySQLPort,
			TargetPort: intstr.FromInt32(frontendSpec.MySQLPort),
		},
		{
			Name:       "pg",
			Protocol:   corev1.ProtocolTCP,
			Port:       b.Cluster.Spec.PostgreSQLPort,
			TargetPort: intstr.FromInt32(frontendSpec.PostgreSQLPort),
		},
	}
}

func (b *frontendBuilder) containerPorts(frontendSpec *v1alpha1.FrontendSpec) []corev1.ContainerPort {
	return []corev1.ContainerPort{
		{
			Name:          "rpc",
			Protocol:      corev1.ProtocolTCP,
			ContainerPort: frontendSpec.RPCPort,
		},
		{
			Name:          "http",
			Protocol:      corev1.ProtocolTCP,
			ContainerPort: frontendSpec.HTTPPort,
		},
		{
			Name:          "mysql",
			Protocol:      corev1.ProtocolTCP,
			ContainerPort: frontendSpec.MySQLPort,
		},
		{
			Name:          "pg",
			Protocol:      corev1.ProtocolTCP,
			ContainerPort: frontendSpec.PostgreSQLPort,
		},
	}
}
