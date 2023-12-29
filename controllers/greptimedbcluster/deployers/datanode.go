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
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/GreptimeTeam/greptimedb-operator/apis/v1alpha1"
	"github.com/GreptimeTeam/greptimedb-operator/pkg/deployer"
	"github.com/GreptimeTeam/greptimedb-operator/pkg/util"
	k8sutil "github.com/GreptimeTeam/greptimedb-operator/pkg/util/k8s"
)

// DatanodeDeployer is the deployer for datanode.
type DatanodeDeployer struct {
	*CommonDeployer
}

var _ deployer.Deployer = &DatanodeDeployer{}

func NewDatanodeDeployer(mgr ctrl.Manager) *DatanodeDeployer {
	return &DatanodeDeployer{
		CommonDeployer: NewFromManager(mgr),
	}
}

func (d *DatanodeDeployer) NewBuilder(crdObject client.Object) deployer.Builder {
	return &datanodeBuilder{
		CommonBuilder: d.NewCommonBuilder(crdObject, v1alpha1.DatanodeComponentKind),
	}
}

func (d *DatanodeDeployer) Generate(crdObject client.Object) ([]client.Object, error) {
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

func (d *DatanodeDeployer) CleanUp(ctx context.Context, crdObject client.Object) error {
	cluster, err := d.GetCluster(crdObject)
	if err != nil {
		return err
	}

	if cluster.Spec.Datanode != nil {
		if cluster.Spec.Datanode.Storage.StorageRetainPolicy == v1alpha1.StorageRetainPolicyTypeDelete {
			if err := d.deleteStorage(ctx, cluster); err != nil {
				return err
			}
		}
	}

	return nil
}

func (d *DatanodeDeployer) CheckAndUpdateStatus(ctx context.Context, crdObject client.Object) (bool, error) {
	cluster, err := d.GetCluster(crdObject)
	if err != nil {
		return false, err
	}

	var (
		sts = new(appsv1.StatefulSet)

		objectKey = client.ObjectKey{
			Namespace: cluster.Namespace,
			Name:      ResourceName(cluster.Name, v1alpha1.DatanodeComponentKind),
		}
	)

	err = d.Get(ctx, objectKey, sts)
	if errors.IsNotFound(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	cluster.Status.Datanode.Replicas = *sts.Spec.Replicas
	cluster.Status.Datanode.ReadyReplicas = sts.Status.ReadyReplicas
	if err := UpdateStatus(ctx, cluster, d.Client); err != nil {
		klog.Errorf("Failed to update status: %s", err)
	}

	return k8sutil.IsStatefulSetReady(sts), nil
}

func (d *DatanodeDeployer) deleteStorage(ctx context.Context, cluster *v1alpha1.GreptimeDBCluster) error {
	klog.Infof("Deleting datanode storage...")

	selector, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{
		MatchLabels: map[string]string{
			GreptimeDBComponentName: ResourceName(cluster.Name, v1alpha1.DatanodeComponentKind),
		},
	})
	if err != nil {
		return err
	}

	pvcList := new(corev1.PersistentVolumeClaimList)

	err = d.List(ctx, pvcList, client.InNamespace(cluster.Namespace), client.MatchingLabelsSelector{Selector: selector})
	if errors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}

	for _, pvc := range pvcList.Items {
		klog.Infof("Deleting datanode PVC: %s", pvc.Name)
		if err := d.Delete(ctx, &pvc); err != nil {
			return err
		}
	}

	return nil
}

var _ deployer.Builder = &datanodeBuilder{}

type datanodeBuilder struct {
	*CommonBuilder
}

func (b *datanodeBuilder) BuildService() deployer.Builder {
	if b.Err != nil {
		return b
	}

	if b.Cluster.Spec.Datanode == nil {
		return b
	}

	svc := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: b.Cluster.Namespace,
			Name:      ResourceName(b.Cluster.Name, b.ComponentKind),
		},
		Spec: corev1.ServiceSpec{
			ClusterIP: corev1.ClusterIPNone,
			Selector: map[string]string{
				GreptimeDBComponentName: ResourceName(b.Cluster.Name, b.ComponentKind),
			},
			Ports: b.servicePorts(),
		},
	}

	b.Objects = append(b.Objects, svc)

	return b
}

func (b *datanodeBuilder) BuildConfigMap() deployer.Builder {
	if b.Err != nil {
		return b
	}

	if b.Cluster.Spec.Datanode == nil {
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

func (b *datanodeBuilder) BuildStatefulSet() deployer.Builder {
	if b.Err != nil {
		return b
	}

	if b.Cluster.Spec.Datanode == nil {
		return b
	}

	sts := &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "StatefulSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      ResourceName(b.Cluster.Name, b.ComponentKind),
			Namespace: b.Cluster.Namespace,
		},
		Spec: appsv1.StatefulSetSpec{
			ServiceName: ResourceName(b.Cluster.Name, b.ComponentKind),
			Replicas:    b.Cluster.Spec.Datanode.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					GreptimeDBComponentName: ResourceName(b.Cluster.Name, b.ComponentKind),
				},
			},
			Template:             b.generatePodTemplateSpec(),
			VolumeClaimTemplates: b.generatePVC(),
		},
	}

	b.Objects = append(b.Objects, sts)

	return b
}

func (b *datanodeBuilder) BuildPodMonitor() deployer.Builder {
	if b.Err != nil {
		return b
	}

	if b.Cluster.Spec.Datanode == nil {
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

func (b *datanodeBuilder) generateMainContainerArgs() []string {
	return []string{
		"datanode", "start",
		"--metasrv-addr", fmt.Sprintf("%s.%s:%d", ResourceName(b.Cluster.Name, v1alpha1.MetaComponentKind),
			b.Cluster.Namespace, b.Cluster.Spec.Meta.ServicePort),
		// TODO(zyy17): Should we add the new field of the CRD for datanode http port?
		"--http-addr", fmt.Sprintf("0.0.0.0:%d", b.Cluster.Spec.HTTPServicePort),
		"--config-file", path.Join(GreptimeDBConfigDir, GreptimeDBConfigFileName),
	}
}

func (b *datanodeBuilder) generatePodTemplateSpec() corev1.PodTemplateSpec {
	podTemplateSpec := b.GeneratePodTemplateSpec(b.Cluster.Spec.Datanode.Template)

	if len(b.Cluster.Spec.Datanode.Template.MainContainer.Args) == 0 {
		// Setup main container args.
		podTemplateSpec.Spec.Containers[MainContainerIndex].Args = b.generateMainContainerArgs()
	}

	b.mountConfigDir(podTemplateSpec)
	b.addStorageDirMounts(podTemplateSpec)
	b.addInitConfigDirVolume(podTemplateSpec)

	podTemplateSpec.Spec.Containers[MainContainerIndex].Ports = b.containerPorts()
	podTemplateSpec.Spec.InitContainers = append(podTemplateSpec.Spec.InitContainers, *b.generateInitializer())
	podTemplateSpec.ObjectMeta.Labels = util.MergeStringMap(podTemplateSpec.ObjectMeta.Labels, map[string]string{
		GreptimeDBComponentName: ResourceName(b.Cluster.Name, b.ComponentKind),
	})

	return *podTemplateSpec
}

func (b *datanodeBuilder) generatePVC() []corev1.PersistentVolumeClaim {
	return []corev1.PersistentVolumeClaim{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: b.Cluster.Spec.Datanode.Storage.Name,
			},
			Spec: corev1.PersistentVolumeClaimSpec{
				StorageClassName: b.Cluster.Spec.Datanode.Storage.StorageClassName,
				AccessModes: []corev1.PersistentVolumeAccessMode{
					corev1.ReadWriteOnce,
				},
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceStorage: resource.MustParse(b.Cluster.Spec.Datanode.Storage.StorageSize),
					},
				},
			},
		},
	}
}

func (b *datanodeBuilder) generateInitializer() *corev1.Container {
	initializer := &corev1.Container{
		Name:  "initializer",
		Image: b.Cluster.Spec.Initializer.Image,
		Command: []string{
			"greptimedb-initializer",
		},
		Args: []string{
			"--config-path", path.Join(GreptimeDBConfigDir, GreptimeDBConfigFileName),
			"--init-config-path", path.Join(GreptimeDBInitConfigDir, GreptimeDBConfigFileName),
			"--datanode-rpc-port", fmt.Sprintf("%d", b.Cluster.Spec.GRPCServicePort),
			"--datanode-service-name", ResourceName(b.Cluster.Name, b.ComponentKind),
			"--namespace", b.Cluster.Namespace,
			"--component-kind", string(b.ComponentKind),
		},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      ConfigVolumeName,
				MountPath: GreptimeDBConfigDir,
			},
			{
				Name:      InitConfigVolumeName,
				MountPath: GreptimeDBInitConfigDir,
			},
		},

		// TODO(zyy17): the datanode don't support to accept hostname.
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

func (b *datanodeBuilder) mountConfigDir(template *corev1.PodTemplateSpec) {
	// The empty-dir will be modified by initializer.
	template.Spec.Volumes = append(template.Spec.Volumes, corev1.Volume{
		Name: ConfigVolumeName,
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		},
	})

	template.Spec.Containers[MainContainerIndex].VolumeMounts =
		append(template.Spec.Containers[MainContainerIndex].VolumeMounts,
			corev1.VolumeMount{
				Name:      ConfigVolumeName,
				MountPath: GreptimeDBConfigDir,
			},
		)
}

func (b *datanodeBuilder) addStorageDirMounts(template *corev1.PodTemplateSpec) {
	// The volume is defined in the PVC.
	template.Spec.Containers[MainContainerIndex].VolumeMounts =
		append(template.Spec.Containers[MainContainerIndex].VolumeMounts,
			corev1.VolumeMount{
				Name:      b.Cluster.Spec.Datanode.Storage.Name,
				MountPath: b.Cluster.Spec.Datanode.Storage.MountPath,
			},
		)
}

// The init-config volume is used for initializer.
func (b *datanodeBuilder) addInitConfigDirVolume(template *corev1.PodTemplateSpec) {
	template.Spec.Volumes = append(template.Spec.Volumes, corev1.Volume{
		Name: InitConfigVolumeName,
		VolumeSource: corev1.VolumeSource{
			// Mount the configmap as init-config.
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: ResourceName(b.Cluster.Name, b.ComponentKind),
				},
			},
		},
	})
}

func (b *datanodeBuilder) servicePorts() []corev1.ServicePort {
	return []corev1.ServicePort{
		{
			Name:     "grpc",
			Protocol: corev1.ProtocolTCP,
			Port:     b.Cluster.Spec.GRPCServicePort,
		},
		{
			Name:     "http",
			Protocol: corev1.ProtocolTCP,
			Port:     b.Cluster.Spec.HTTPServicePort,
		},
	}
}

func (b *datanodeBuilder) containerPorts() []corev1.ContainerPort {
	return []corev1.ContainerPort{
		{
			Name:          "grpc",
			Protocol:      corev1.ProtocolTCP,
			ContainerPort: b.Cluster.Spec.GRPCServicePort,
		},
		{
			Name:          "http",
			Protocol:      corev1.ProtocolTCP,
			ContainerPort: b.Cluster.Spec.HTTPServicePort,
		},
	}
}
