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
	networkingv1 "k8s.io/api/networking/v1"
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
		BuildIngress().
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
	}
	if cluster.GetFrontends() != nil {
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

func (b *frontendBuilder) generateService(frontend *v1alpha1.FrontendSpec) {
	name := common.ResourceName(b.Cluster.Name, b.ComponentKind)
	if len(frontend.GetName()) != 0 {
		name = common.AdditionalResourceName(b.Cluster.Name, frontend.GetName(), b.ComponentKind)
	}

	svc := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   b.Cluster.Namespace,
			Name:        name,
			Annotations: frontend.Service.Annotations,
			Labels: util.MergeStringMap(frontend.Service.Labels, map[string]string{
				constant.GreptimeDBComponentName: name,
			}),
		},
		Spec: corev1.ServiceSpec{
			Type: frontend.Service.Type,
			Selector: map[string]string{
				constant.GreptimeDBComponentName: name,
			},
			Ports:             b.servicePorts(frontend),
			LoadBalancerClass: frontend.Service.LoadBalancerClass,
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
		b.generateService(b.Cluster.Spec.Frontend)
	}

	if len(b.Cluster.GetFrontends()) != 0 {
		for _, frontend := range b.Cluster.Spec.Frontends {
			b.generateService(frontend)
		}
	}

	return b
}

func (b *frontendBuilder) generateDeployment(frontend *v1alpha1.FrontendSpec) {
	name := common.ResourceName(b.Cluster.Name, b.ComponentKind)
	if len(frontend.GetName()) != 0 {
		name = common.AdditionalResourceName(b.Cluster.Name, frontend.GetName(), b.ComponentKind)
	}

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
			Replicas: frontend.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					constant.GreptimeDBComponentName: name,
				},
			},
			Template: *b.generatePodTemplateSpec(frontend),
			Strategy: appsv1.DeploymentStrategy{
				Type:          appsv1.RollingUpdateDeploymentStrategyType,
				RollingUpdate: frontend.RollingUpdate,
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
		b.generateDeployment(b.Cluster.Spec.Frontend)
	}

	if len(b.Cluster.GetFrontends()) != 0 {
		for _, frontend := range b.Cluster.Spec.Frontends {
			b.generateDeployment(frontend)
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

func (b *frontendBuilder) generateIngress(frontend *v1alpha1.FrontendSpec) {
	name := common.ResourceName(b.Cluster.Name, b.ComponentKind)
	if len(frontend.GetName()) != 0 {
		name = common.AdditionalResourceName(b.Cluster.Name, frontend.GetName(), b.ComponentKind)
	}

	ing := &networkingv1.Ingress{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Ingress",
			APIVersion: "networking.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   b.Cluster.Namespace,
			Name:        name,
			Annotations: frontend.Ingress.Annotations,
			Labels: util.MergeStringMap(frontend.Ingress.Labels, map[string]string{
				constant.GreptimeDBComponentName: name,
			}),
		},
		Spec: frontend.Ingress.IngressSpec,
	}

	// Forward to frontend service when backend is not set.
	const (
		defaultPath = "/"
	)
	if len(ing.Spec.Rules) != 0 {
		for i := range ing.Spec.Rules {
			rule := &ing.Spec.Rules[i]
			if rule.HTTP == nil {
				rule.HTTP = &networkingv1.HTTPIngressRuleValue{
					Paths: []networkingv1.HTTPIngressPath{
						{
							Backend: networkingv1.IngressBackend{
								Service: &networkingv1.IngressServiceBackend{
									Name: name,
									Port: networkingv1.ServiceBackendPort{
										Number: b.Cluster.Spec.HTTPPort,
									},
								},
							},
							Path: defaultPath,
							PathType: func() *networkingv1.PathType {
								pathType := networkingv1.PathTypePrefix
								return &pathType
							}(),
						},
					},
				}
			}
		}
	}

	b.Objects = append(b.Objects, ing)
}

func (b *frontendBuilder) BuildIngress() deployer.Builder {
	if b.Err != nil {
		return b
	}

	if b.Cluster.GetFrontend() == nil && len(b.Cluster.GetFrontends()) == 0 {
		return b
	}

	if b.Cluster.GetFrontend() != nil && b.Cluster.GetFrontend().GetIngress() != nil {
		b.generateIngress(b.Cluster.GetFrontend())
	}

	if len(b.Cluster.GetFrontends()) != 0 {
		for _, frontend := range b.Cluster.Spec.Frontends {
			if frontend.GetIngress() != nil {
				b.generateIngress(frontend)
			}
		}
	}

	return b
}

func (b *frontendBuilder) Generate() ([]client.Object, error) {
	return b.Objects, b.Err
}

func (b *frontendBuilder) generateMainContainerArgs(frontend *v1alpha1.FrontendSpec) []string {
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

	if frontend.TLS != nil {
		args = append(args, []string{
			"--tls-mode", constant.DefaultTLSMode,
			"--tls-cert-path", path.Join(constant.GreptimeDBTLSDir, v1alpha1.TLSCrtSecretKey),
			"--tls-key-path", path.Join(constant.GreptimeDBTLSDir, v1alpha1.TLSKeySecretKey),
		}...)
	}

	return args
}

func (b *frontendBuilder) generatePodTemplateSpec(frontend *v1alpha1.FrontendSpec) *corev1.PodTemplateSpec {
	podTemplateSpec := b.GeneratePodTemplateSpec(frontend.Template)

	if len(frontend.Template.MainContainer.Args) == 0 {
		// Setup main container args.
		podTemplateSpec.Spec.Containers[constant.MainContainerIndex].Args = b.generateMainContainerArgs(frontend)
	}

	resourceName := common.ResourceName(b.Cluster.Name, b.ComponentKind)
	if len(frontend.GetName()) != 0 {
		resourceName = common.AdditionalResourceName(b.Cluster.Name, frontend.GetName(), b.ComponentKind)
	}
	podTemplateSpec.ObjectMeta.Labels = util.MergeStringMap(podTemplateSpec.ObjectMeta.Labels, map[string]string{
		constant.GreptimeDBComponentName: resourceName,
	})

	podTemplateSpec.Spec.Containers[constant.MainContainerIndex].Ports = b.containerPorts(frontend)
	podTemplateSpec.Spec.Containers[constant.MainContainerIndex].Env = append(podTemplateSpec.Spec.Containers[constant.MainContainerIndex].Env, b.env(v1alpha1.FrontendComponentKind)...)

	b.Frontend = frontend
	b.MountConfigDir(podTemplateSpec)

	if logging := frontend.GetLogging(); logging != nil && !logging.IsOnlyLogToStdout() {
		b.AddLogsVolume(podTemplateSpec, logging.GetLogsDir())
	}

	if b.Cluster.GetMonitoring().IsEnabled() && b.Cluster.GetMonitoring().GetVector() != nil {
		b.AddVectorConfigVolume(podTemplateSpec)
		b.AddVectorSidecar(podTemplateSpec, v1alpha1.FrontendComponentKind)
	}

	if frontend.TLS != nil {
		b.mountTLSSecret(frontend.TLS.SecretName, podTemplateSpec)
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

func (b *frontendBuilder) servicePorts(frontend *v1alpha1.FrontendSpec) []corev1.ServicePort {
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

func (b *frontendBuilder) containerPorts(frontend *v1alpha1.FrontendSpec) []corev1.ContainerPort {
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
