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

const (
	TLSCrtSecretKey = "tls.crt"
	TLSKeySecretKey = "tls.key"
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
		deployment = new(appsv1.Deployment)

		objectKey = client.ObjectKey{
			Namespace: cluster.Namespace,
			Name:      common.ResourceName(cluster.Name, v1alpha1.FrontendComponentKind),
		}
	)

	err = d.Get(ctx, objectKey, deployment)
	if errors.IsNotFound(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	cluster.Status.Frontend.Replicas = *deployment.Spec.Replicas
	cluster.Status.Frontend.ReadyReplicas = deployment.Status.ReadyReplicas
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
			Labels:      b.Cluster.Spec.Frontend.Service.Labels,
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
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: b.Cluster.Spec.Frontend.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					constant.GreptimeDBComponentName: common.ResourceName(b.Cluster.Name, b.ComponentKind),
				},
			},
			Template: *b.generatePodTemplateSpec(),
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
		"--rpc-addr", fmt.Sprintf("0.0.0.0:%d", b.Cluster.Spec.RPCPort),
		"--metasrv-addrs", fmt.Sprintf("%s.%s:%d", common.ResourceName(b.Cluster.Name, v1alpha1.MetaComponentKind),
			b.Cluster.Namespace, b.Cluster.Spec.Meta.RPCPort),
		"--http-addr", fmt.Sprintf("0.0.0.0:%d", b.Cluster.Spec.HTTPPort),
		"--mysql-addr", fmt.Sprintf("0.0.0.0:%d", b.Cluster.Spec.MySQLPort),
		"--postgres-addr", fmt.Sprintf("0.0.0.0:%d", b.Cluster.Spec.PostgreSQLPort),
		"--config-file", path.Join(constant.GreptimeDBConfigDir, constant.GreptimeDBConfigFileName),
	}

	if b.Cluster.Spec.Frontend != nil && b.Cluster.Spec.Frontend.TLS != nil {
		args = append(args, []string{
			"--tls-mode", constant.DefaultTLSMode,
			"--tls-cert-path", path.Join(constant.GreptimeDBTLSDir, TLSCrtSecretKey),
			"--tls-key-path", path.Join(constant.GreptimeDBTLSDir, TLSKeySecretKey),
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

	b.MountConfigDir(podTemplateSpec)

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
			Name:     "rpc",
			Protocol: corev1.ProtocolTCP,
			Port:     b.Cluster.Spec.RPCPort,
		},
		{
			Name:     "http",
			Protocol: corev1.ProtocolTCP,
			Port:     b.Cluster.Spec.HTTPPort,
		},
		{
			Name:     "mysql",
			Protocol: corev1.ProtocolTCP,
			Port:     b.Cluster.Spec.MySQLPort,
		},
		{
			Name:     "pg",
			Protocol: corev1.ProtocolTCP,
			Port:     b.Cluster.Spec.PostgreSQLPort,
		},
	}
}

func (b *frontendBuilder) containerPorts() []corev1.ContainerPort {
	return []corev1.ContainerPort{
		{
			Name:          "rpc",
			Protocol:      corev1.ProtocolTCP,
			ContainerPort: b.Cluster.Spec.RPCPort,
		},
		{
			Name:          "http",
			Protocol:      corev1.ProtocolTCP,
			ContainerPort: b.Cluster.Spec.HTTPPort,
		},
		{
			Name:          "mysql",
			Protocol:      corev1.ProtocolTCP,
			ContainerPort: b.Cluster.Spec.MySQLPort,
		},
		{
			Name:          "pg",
			Protocol:      corev1.ProtocolTCP,
			ContainerPort: b.Cluster.Spec.PostgreSQLPort,
		},
	}
}
