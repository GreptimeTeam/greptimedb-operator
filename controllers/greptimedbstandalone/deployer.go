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

package greptimedbstandalone

import (
	"context"
	"fmt"
	"path"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"
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

type StandaloneDeployer struct {
	Scheme *runtime.Scheme

	client.Client
	deployer.DefaultDeployer
}

var _ deployer.Deployer = &StandaloneDeployer{}

func NewStandaloneDeployer(mgr ctrl.Manager) *StandaloneDeployer {
	return &StandaloneDeployer{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),

		DefaultDeployer: deployer.DefaultDeployer{
			Client: mgr.GetClient(),
		},
	}
}

func (d *StandaloneDeployer) NewBuilder(crdObject client.Object) deployer.Builder {
	sb := &standaloneBuilder{
		DefaultBuilder: &deployer.DefaultBuilder{
			Scheme: d.Scheme,
			Owner:  crdObject,
		},
	}

	standalone, err := d.getStandalone(crdObject)
	if err != nil {
		sb.Err = err
	}
	sb.standalone = standalone

	return sb
}

func (d *StandaloneDeployer) Generate(crdObject client.Object) ([]client.Object, error) {
	objects, err := d.NewBuilder(crdObject).
		BuildService().
		BuildConfigMap().
		BuildStatefulSet().
		SetControllerAndAnnotation().
		Generate()

	if err != nil {
		return nil, err
	}

	return objects, nil
}

func (d *StandaloneDeployer) CleanUp(ctx context.Context, crdObject client.Object) error {
	standalone, err := d.getStandalone(crdObject)
	if err != nil {
		return err
	}

	if standalone.GetDatanodeFileStorage().GetPolicy() == v1alpha1.StorageRetainPolicyTypeDelete {
		if err := d.deleteStorage(ctx, standalone.Namespace, standalone.Name, common.FileStorageTypeDatanode); err != nil {
			return err
		}
	}

	if standalone.GetWALProvider().GetRaftEngineWAL().GetFileStorage().GetPolicy() == v1alpha1.StorageRetainPolicyTypeDelete {
		if err := d.deleteStorage(ctx, standalone.Namespace, standalone.Name, common.FileStorageTypeWAL); err != nil {
			return err
		}
	}

	if standalone.GetObjectStorageProvider().GetCacheFileStorage().GetPolicy() == v1alpha1.StorageRetainPolicyTypeDelete {
		if err := d.deleteStorage(ctx, standalone.Namespace, standalone.Name, common.FileStorageTypeCache); err != nil {
			return err
		}
	}

	return nil
}

func (d *StandaloneDeployer) CheckAndUpdateStatus(ctx context.Context, crdObject client.Object) (bool, error) {
	standalone, err := d.getStandalone(crdObject)
	if err != nil {
		return false, err
	}

	var (
		sts = new(appsv1.StatefulSet)

		objectKey = client.ObjectKey{
			Namespace: standalone.Namespace,
			Name:      common.ResourceName(standalone.Name, v1alpha1.StandaloneKind),
		}
	)

	err = d.Get(ctx, objectKey, sts)
	if errors.IsNotFound(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return k8sutil.IsStatefulSetReady(sts), nil
}

func (d *StandaloneDeployer) getStandalone(crdObject client.Object) (*v1alpha1.GreptimeDBStandalone, error) {
	standalone, ok := crdObject.(*v1alpha1.GreptimeDBStandalone)
	if !ok {
		return nil, fmt.Errorf("the object is not a GreptimeDBStandalone")
	}
	return standalone, nil
}

func (d *StandaloneDeployer) deleteStorage(ctx context.Context, namespace, name string, fsType common.FileStorageType) error {
	klog.Infof("Deleting standalone storage...")

	claims, err := common.GetPVCs(ctx, d.Client, namespace, name, v1alpha1.StandaloneKind, fsType)
	if err != nil {
		return err
	}

	for _, pvc := range claims {
		klog.Infof("Deleting standalone PVC: %s", pvc.Name)
		if err := d.Delete(ctx, &pvc); err != nil {
			return err
		}
	}

	return nil
}

var _ deployer.Builder = &standaloneBuilder{}

type standaloneBuilder struct {
	standalone *v1alpha1.GreptimeDBStandalone
	*deployer.DefaultBuilder
}

func (b *standaloneBuilder) BuildService() deployer.Builder {
	if b.Err != nil {
		return b
	}

	svc := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   b.standalone.Namespace,
			Name:        common.ResourceName(b.standalone.Name, v1alpha1.StandaloneKind),
			Annotations: b.standalone.Spec.Service.Annotations,
			Labels: util.MergeStringMap(b.standalone.Spec.Service.Labels, map[string]string{
				constant.GreptimeDBComponentName: common.ResourceName(b.standalone.Name, v1alpha1.StandaloneKind),
			}),
		},
		Spec: corev1.ServiceSpec{
			Type: b.standalone.Spec.Service.Type,
			Selector: map[string]string{
				constant.GreptimeDBComponentName: common.ResourceName(b.standalone.Name, v1alpha1.StandaloneKind),
			},
			Ports:             b.servicePorts(),
			LoadBalancerClass: b.standalone.Spec.Service.LoadBalancerClass,
		},
	}

	b.Objects = append(b.Objects, svc)

	return b
}

func (b *standaloneBuilder) BuildConfigMap() deployer.Builder {
	if b.Err != nil {
		return b
	}

	configData, err := dbconfig.FromStandalone(b.standalone)
	if err != nil {
		b.Err = err
		return b
	}

	cm, err := common.GenerateConfigMap(b.standalone.Namespace, b.standalone.Name, v1alpha1.StandaloneKind, configData, "")
	if err != nil {
		b.Err = err
		return b
	}

	b.Objects = append(b.Objects, cm)

	return b
}

func (b *standaloneBuilder) BuildStatefulSet() deployer.Builder {
	if b.Err != nil {
		return b
	}

	sts := &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "StatefulSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      common.ResourceName(b.standalone.Name, v1alpha1.StandaloneKind),
			Namespace: b.standalone.Namespace,
			Labels: map[string]string{
				constant.GreptimeDBComponentName: common.ResourceName(b.standalone.Name, v1alpha1.StandaloneKind),
			},
		},
		Spec: appsv1.StatefulSetSpec{
			// Always set replicas to 1 for standalone mode.
			Replicas: ptr.To(int32(1)),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					constant.GreptimeDBComponentName: common.ResourceName(b.standalone.Name, v1alpha1.StandaloneKind),
				},
			},
			Template:             b.generatePodTemplateSpec(),
			VolumeClaimTemplates: b.generatePVCs(),
			UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
				Type:          appsv1.RollingUpdateStatefulSetStrategyType,
				RollingUpdate: b.standalone.Spec.RollingUpdate,
			},
		},
	}

	configData, err := dbconfig.FromStandalone(b.standalone)
	if err != nil {
		b.Err = err
		return b
	}

	sts.Spec.Template.Annotations = util.MergeStringMap(sts.Spec.Template.Annotations,
		map[string]string{deployer.ConfigHash: util.CalculateConfigHash(configData)})

	b.Objects = append(b.Objects, sts)

	return b
}

func (b *standaloneBuilder) generatePodTemplateSpec() corev1.PodTemplateSpec {
	template := common.GeneratePodTemplateSpec(v1alpha1.StandaloneKind, b.standalone.Spec.Base)

	if len(b.standalone.Spec.Base.MainContainer.Args) == 0 {
		// Setup main container args.
		template.Spec.Containers[constant.MainContainerIndex].Args = b.generateMainContainerArgs()
	}

	b.addVolumeMounts(template)

	template.Spec.Containers[constant.MainContainerIndex].Ports = b.containerPorts()
	template.ObjectMeta.Labels = util.MergeStringMap(template.ObjectMeta.Labels, map[string]string{
		constant.GreptimeDBComponentName: common.ResourceName(b.standalone.Name, v1alpha1.StandaloneKind),
	})

	common.MountConfigDir(b.standalone.Name, v1alpha1.StandaloneKind, template, "")

	if b.standalone.Spec.TLS != nil {
		b.mountTLSSecret(template)
	}

	return *template
}

func (b *standaloneBuilder) mountTLSSecret(template *corev1.PodTemplateSpec) {
	template.Spec.Volumes = append(template.Spec.Volumes, corev1.Volume{
		Name: constant.TLSVolumeName,
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: b.standalone.Spec.TLS.SecretName,
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

func (b *standaloneBuilder) generatePVCs() []corev1.PersistentVolumeClaim {
	var claims []corev1.PersistentVolumeClaim

	// It's always not nil because it's the default value.
	if fs := b.standalone.GetDatanodeFileStorage(); fs != nil {
		claims = append(claims, *common.FileStorageToPVC(b.standalone.Name, fs, common.FileStorageTypeDatanode, v1alpha1.StandaloneKind))
	}

	// Allocate the standalone WAL storage for the raft-engine.
	if fs := b.standalone.GetWALProvider().GetRaftEngineWAL().GetFileStorage(); fs != nil {
		claims = append(claims, *common.FileStorageToPVC(b.standalone.Name, fs, common.FileStorageTypeWAL, v1alpha1.StandaloneKind))
	}

	// Allocate the standalone cache file storage for the datanode.
	if fs := b.standalone.GetObjectStorageProvider().GetCacheFileStorage(); fs != nil {
		claims = append(claims, *common.FileStorageToPVC(b.standalone.Name, fs, common.FileStorageTypeCache, v1alpha1.StandaloneKind))
	}

	return claims
}

func (b *standaloneBuilder) servicePorts() []corev1.ServicePort {
	return []corev1.ServicePort{
		{
			Name:     "grpc",
			Protocol: corev1.ProtocolTCP,
			Port:     b.standalone.Spec.RPCPort,
		},
		{
			Name:     "http",
			Protocol: corev1.ProtocolTCP,
			Port:     b.standalone.Spec.HTTPPort,
		},
		{
			Name:     "mysql",
			Protocol: corev1.ProtocolTCP,
			Port:     b.standalone.Spec.MySQLPort,
		},
		{
			Name:     "postgres",
			Protocol: corev1.ProtocolTCP,
			Port:     b.standalone.Spec.PostgreSQLPort,
		},
	}
}

func (b *standaloneBuilder) containerPorts() []corev1.ContainerPort {
	return []corev1.ContainerPort{
		{
			Name:          "grpc",
			Protocol:      corev1.ProtocolTCP,
			ContainerPort: b.standalone.Spec.RPCPort,
		},
		{
			Name:          "http",
			Protocol:      corev1.ProtocolTCP,
			ContainerPort: b.standalone.Spec.HTTPPort,
		},
		{
			Name:          "mysql",
			Protocol:      corev1.ProtocolTCP,
			ContainerPort: b.standalone.Spec.MySQLPort,
		},
		{
			Name:          "postgres",
			Protocol:      corev1.ProtocolTCP,
			ContainerPort: b.standalone.Spec.PostgreSQLPort,
		},
	}
}

func (b *standaloneBuilder) generateMainContainerArgs() []string {
	var args = []string{
		"standalone", "start",
		"--rpc-bind-addr", fmt.Sprintf("0.0.0.0:%d", b.standalone.Spec.RPCPort),
		"--mysql-addr", fmt.Sprintf("0.0.0.0:%d", b.standalone.Spec.MySQLPort),
		"--http-addr", fmt.Sprintf("0.0.0.0:%d", b.standalone.Spec.HTTPPort),
		"--postgres-addr", fmt.Sprintf("0.0.0.0:%d", b.standalone.Spec.PostgreSQLPort),
		"--config-file", path.Join(constant.GreptimeDBConfigDir, constant.GreptimeDBConfigFileName),
	}

	if b.standalone.Spec.TLS != nil {
		args = append(args, []string{
			"--tls-mode", "require",
			"--tls-cert-path", path.Join(constant.GreptimeDBTLSDir, v1alpha1.TLSCrtSecretKey),
			"--tls-key-path", path.Join(constant.GreptimeDBTLSDir, v1alpha1.TLSKeySecretKey),
		}...)
	}

	return args
}

func (b *standaloneBuilder) addVolumeMounts(template *corev1.PodTemplateSpec) {
	if fs := b.standalone.GetDatanodeFileStorage(); fs != nil {
		template.Spec.Containers[constant.MainContainerIndex].VolumeMounts =
			append(template.Spec.Containers[constant.MainContainerIndex].VolumeMounts,
				corev1.VolumeMount{
					Name:      fs.GetName(),
					MountPath: fs.GetMountPath(),
				},
			)
	}

	if fs := b.standalone.GetWALProvider().GetRaftEngineWAL().GetFileStorage(); fs != nil {
		template.Spec.Containers[constant.MainContainerIndex].VolumeMounts =
			append(template.Spec.Containers[constant.MainContainerIndex].VolumeMounts,
				corev1.VolumeMount{
					Name:      fs.GetName(),
					MountPath: fs.GetMountPath(),
				},
			)
	}

	if fs := b.standalone.GetObjectStorageProvider().GetCacheFileStorage(); fs != nil {
		template.Spec.Containers[constant.MainContainerIndex].VolumeMounts =
			append(template.Spec.Containers[constant.MainContainerIndex].VolumeMounts,
				corev1.VolumeMount{
					Name:      fs.GetName(),
					MountPath: fs.GetMountPath(),
				},
			)
	}

	if logging := b.standalone.GetLogging(); logging != nil && !logging.IsOnlyLogToStdout() && !logging.IsPersistentWithData() {
		template.Spec.Volumes = append(template.Spec.Volumes, corev1.Volume{
			Name: "logs",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		})

		template.Spec.Containers[constant.MainContainerIndex].VolumeMounts = append(template.Spec.Containers[constant.MainContainerIndex].VolumeMounts, corev1.VolumeMount{
			Name:      "logs",
			MountPath: logging.GetLogsDir(),
		})
	}
}
