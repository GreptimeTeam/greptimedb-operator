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
	"net/http"
	"path"
	"reflect"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
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
	k8sutils "github.com/GreptimeTeam/greptimedb-operator/pkg/util/k8s"
)

// DatanodeDeployer is the deployer for datanode.
type DatanodeDeployer struct {
	*CommonDeployer
	maintenanceMode bool
}

var _ deployer.Deployer = &DatanodeDeployer{}

func NewDatanodeDeployer(mgr ctrl.Manager) *DatanodeDeployer {
	return &DatanodeDeployer{
		CommonDeployer:  NewFromManager(mgr),
		maintenanceMode: false,
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
			Name:      common.ResourceName(cluster.Name, v1alpha1.DatanodeComponentKind),
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

	return k8sutils.IsStatefulSetReady(sts), nil
}

// Apply is re-implemented for datanode to handle the maintenance mode.
func (d *DatanodeDeployer) Apply(ctx context.Context, crdObject client.Object, objects []client.Object) error {
	updateObject := false

	cluster, err := d.GetCluster(crdObject)
	if err != nil {
		return err
	}

	for _, newObject := range objects {
		oldObject, err := k8sutils.CreateObjectIfNotExist(ctx, d.Client, k8sutils.SourceObject(newObject), newObject)
		if err != nil {
			return err
		}

		if oldObject != nil {
			equal, err := k8sutils.IsObjectSpecEqual(oldObject, newObject, deployer.LastAppliedResourceSpec)
			if err != nil {
				return err
			}

			// If the spec is not equal, update the object.
			if !equal {
				if sts, ok := newObject.(*appsv1.StatefulSet); ok && d.shouldUserMaintenanceMode(cluster) {
					if err := d.turnOnMaintenanceMode(ctx, sts, cluster); err != nil {
						return err
					}
				}
				if err := d.Client.Patch(ctx, newObject, client.MergeFrom(oldObject)); err != nil {
					return err
				}
				updateObject = true
			}
		}
	}

	if updateObject {
		// If the object is updated, we need to wait for the object to be ready.
		// When the related object is ready, we will receive the event and enter the next reconcile loop.
		return deployer.ErrSyncNotReady
	}

	return nil
}

func (d *DatanodeDeployer) PostSyncHooks() []deployer.Hook {
	return []deployer.Hook{
		d.turnOffMaintenanceMode,
	}
}

func (d *DatanodeDeployer) turnOnMaintenanceMode(ctx context.Context, newSts *appsv1.StatefulSet, cluster *v1alpha1.GreptimeDBCluster) error {
	oldSts := new(appsv1.StatefulSet)
	// The oldSts must exist since we have checked it before.
	if err := d.Get(ctx, client.ObjectKeyFromObject(newSts), oldSts); err != nil {
		return err
	}

	if !d.maintenanceMode && d.isOldPodRestart(*newSts, *oldSts) {
		klog.Infof("Turn on maintenance mode for datanode, statefulset: %s", newSts.Name)
		if err := d.requestMetasrvForMaintenance(cluster, true); err != nil {
			return err
		}
		d.maintenanceMode = true
	}

	return nil
}

func (d *DatanodeDeployer) turnOffMaintenanceMode(ctx context.Context, crdObject client.Object) error {
	cluster, err := d.GetCluster(crdObject)
	if err != nil {
		return err
	}

	if d.maintenanceMode && d.shouldUserMaintenanceMode(cluster) {
		klog.Infof("Turn off maintenance mode for datanode, cluster: %s", cluster.Name)
		if err := d.requestMetasrvForMaintenance(cluster, false); err != nil {
			return err
		}
		d.maintenanceMode = false
	}

	return nil
}

func (d *DatanodeDeployer) requestMetasrvForMaintenance(cluster *v1alpha1.GreptimeDBCluster, enabled bool) error {
	requestURL := fmt.Sprintf("http://%s.%s:%d/admin/maintenance?enable=%v", common.ResourceName(cluster.GetName(), v1alpha1.MetaComponentKind), cluster.GetNamespace(), cluster.Spec.Meta.ServicePort, enabled)
	rsp, err := http.Get(requestURL)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()

	if rsp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to turn off maintenance mode for datanode, status code: %d", rsp.StatusCode)
	}
	return nil
}

func (d *DatanodeDeployer) deleteStorage(ctx context.Context, cluster *v1alpha1.GreptimeDBCluster) error {
	klog.Infof("Deleting datanode storage...")

	selector, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{
		MatchLabels: map[string]string{
			constant.GreptimeDBComponentName: common.ResourceName(cluster.Name, v1alpha1.DatanodeComponentKind),
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

// isOldPodRestart checks if the existed pod needs to be restarted. For convenience, we only compare the necessary fields.
// TODO(zyy17): Do we have a easy way to implement this?
func (d *DatanodeDeployer) isOldPodRestart(new, old appsv1.StatefulSet) bool {
	var (
		newPodTemplate = new.Spec.Template
		oldPodTemplate = old.Spec.Template
	)

	if !reflect.DeepEqual(newPodTemplate.GetObjectMeta().GetAnnotations(), oldPodTemplate.GetObjectMeta().GetAnnotations()) {
		return true
	}

	if newPodTemplate.Spec.InitContainers[0].Image != oldPodTemplate.Spec.InitContainers[0].Image {
		return true
	}

	// If the tolerations, affinity, nodeSelector are changed, the original Pod may need to be restarted for re-scheduling.
	if !reflect.DeepEqual(newPodTemplate.Spec.Tolerations, oldPodTemplate.Spec.Tolerations) ||
		!reflect.DeepEqual(newPodTemplate.Spec.Affinity, oldPodTemplate.Spec.Affinity) ||
		!reflect.DeepEqual(newPodTemplate.Spec.NodeSelector, oldPodTemplate.Spec.NodeSelector) {
		return true
	}

	// Compare the main container settings.
	newMainContainer := newPodTemplate.Spec.Containers[constant.MainContainerIndex]
	oldMainContainer := oldPodTemplate.Spec.Containers[constant.MainContainerIndex]
	if newMainContainer.Image != oldMainContainer.Image {
		return true
	}

	if !reflect.DeepEqual(newMainContainer.Command, oldMainContainer.Command) ||
		!reflect.DeepEqual(newMainContainer.Args, oldMainContainer.Args) ||
		!reflect.DeepEqual(newMainContainer.Env, oldMainContainer.Env) ||
		!reflect.DeepEqual(newMainContainer.VolumeMounts, oldMainContainer.VolumeMounts) ||
		!reflect.DeepEqual(newMainContainer.Ports, oldMainContainer.Ports) ||
		!reflect.DeepEqual(newMainContainer.Resources, oldMainContainer.Resources) {
		return true
	}

	return false
}

func (d *DatanodeDeployer) shouldUserMaintenanceMode(cluster *v1alpha1.GreptimeDBCluster) bool {
	if cluster.Spec.RemoteWalProvider != nil &&
		cluster.Spec.Meta.EnableRegionFailover != nil &&
		*cluster.Spec.Meta.EnableRegionFailover {
		return true
	}
	return false
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
			Name:      common.ResourceName(b.Cluster.Name, b.ComponentKind),
			Namespace: b.Cluster.Namespace,
		},
		Spec: appsv1.StatefulSetSpec{
			PodManagementPolicy: appsv1.ParallelPodManagement,
			ServiceName:         common.ResourceName(b.Cluster.Name, b.ComponentKind),
			Replicas:            b.Cluster.Spec.Datanode.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					constant.GreptimeDBComponentName: common.ResourceName(b.Cluster.Name, b.ComponentKind),
				},
			},
			Template:             b.generatePodTemplateSpec(),
			VolumeClaimTemplates: b.generatePVC(),
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
		"--metasrv-addrs", fmt.Sprintf("%s.%s:%d", common.ResourceName(b.Cluster.Name, v1alpha1.MetaComponentKind),
			b.Cluster.Namespace, b.Cluster.Spec.Meta.ServicePort),
		// TODO(zyy17): Should we add the new field of the CRD for datanode http port?
		"--http-addr", fmt.Sprintf("0.0.0.0:%d", b.Cluster.Spec.HTTPServicePort),
		"--config-file", path.Join(constant.GreptimeDBConfigDir, constant.GreptimeDBConfigFileName),
	}
}

func (b *datanodeBuilder) generatePodTemplateSpec() corev1.PodTemplateSpec {
	podTemplateSpec := b.GeneratePodTemplateSpec(b.Cluster.Spec.Datanode.Template)

	if len(b.Cluster.Spec.Datanode.Template.MainContainer.Args) == 0 {
		// Setup main container args.
		podTemplateSpec.Spec.Containers[constant.MainContainerIndex].Args = b.generateMainContainerArgs()
	}

	b.mountConfigDir(podTemplateSpec)
	b.addStorageDirMounts(podTemplateSpec)
	b.addInitConfigDirVolume(podTemplateSpec)

	podTemplateSpec.Spec.Containers[constant.MainContainerIndex].Ports = b.containerPorts()
	podTemplateSpec.Spec.InitContainers = append(podTemplateSpec.Spec.InitContainers, *b.generateInitializer())
	podTemplateSpec.ObjectMeta.Labels = util.MergeStringMap(podTemplateSpec.ObjectMeta.Labels, map[string]string{
		constant.GreptimeDBComponentName: common.ResourceName(b.Cluster.Name, b.ComponentKind),
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
			"--config-path", path.Join(constant.GreptimeDBConfigDir, constant.GreptimeDBConfigFileName),
			"--init-config-path", path.Join(constant.GreptimeDBInitConfigDir, constant.GreptimeDBConfigFileName),
			"--datanode-rpc-port", fmt.Sprintf("%d", b.Cluster.Spec.GRPCServicePort),
			"--datanode-service-name", common.ResourceName(b.Cluster.Name, b.ComponentKind),
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

func (b *datanodeBuilder) addStorageDirMounts(template *corev1.PodTemplateSpec) {
	// The volume is defined in the PVC.
	template.Spec.Containers[constant.MainContainerIndex].VolumeMounts =
		append(template.Spec.Containers[constant.MainContainerIndex].VolumeMounts,
			corev1.VolumeMount{
				Name:      b.Cluster.Spec.Datanode.Storage.Name,
				MountPath: b.Cluster.Spec.Datanode.Storage.MountPath,
			},
		)
}

// The init-config volume is used for initializer.
func (b *datanodeBuilder) addInitConfigDirVolume(template *corev1.PodTemplateSpec) {
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
