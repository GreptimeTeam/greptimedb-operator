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

	if cluster.GetDatanode().GetFileStorage().GetPolicy() == v1alpha1.StorageRetainPolicyTypeDelete {
		if err := d.deleteStorage(ctx, cluster.Namespace, cluster.Name, common.FileStorageTypeDatanode); err != nil {
			return err
		}
	}

	if cluster.GetWALProvider().GetRaftEngineWAL().GetFileStorage().GetPolicy() == v1alpha1.StorageRetainPolicyTypeDelete {
		if err := d.deleteStorage(ctx, cluster.Namespace, cluster.Name, common.FileStorageTypeWAL); err != nil {
			return err
		}
	}

	if cluster.GetObjectStorageProvider().GetCacheFileStorage().GetPolicy() == v1alpha1.StorageRetainPolicyTypeDelete {
		if err := d.deleteStorage(ctx, cluster.Namespace, cluster.Name, common.FileStorageTypeCache); err != nil {
			return err
		}
	}

	return nil
}

func (d *DatanodeDeployer) CheckAndUpdateStatus(ctx context.Context, crdObject client.Object) (bool, error) {
	cluster, err := d.GetCluster(crdObject)
	if err != nil {
		return false, err
	}

	if cluster.Spec.Datanode != nil {
		return d.checkDatanodeStatus(ctx, cluster.Namespace, common.ResourceName(cluster.Name, v1alpha1.DatanodeComponentKind), cluster)
	}

	if len(cluster.Spec.DatanodeGroups) > 0 {
		return d.checkDatanodeGroupStatus(ctx, cluster)
	}

	return true, nil
}

func (d *DatanodeDeployer) checkDatanodeStatus(ctx context.Context, namespace, resourceName string, cluster *v1alpha1.GreptimeDBCluster) (bool, error) {
	var (
		sts = new(appsv1.StatefulSet)

		objectKey = client.ObjectKey{
			Namespace: namespace,
			Name:      resourceName,
		}
	)

	err := d.Get(ctx, objectKey, sts)
	if errors.IsNotFound(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	if cluster != nil {
		cluster.Status.Datanode.Replicas = *sts.Spec.Replicas
		cluster.Status.Datanode.ReadyReplicas = sts.Status.ReadyReplicas
		if err := UpdateStatus(ctx, cluster, d.Client); err != nil {
			klog.Errorf("Failed to update status: %s", err)
		}
	}

	return k8sutils.IsStatefulSetReady(sts), nil
}

func (d *DatanodeDeployer) checkDatanodeGroupStatus(ctx context.Context, cluster *v1alpha1.GreptimeDBCluster) (bool, error) {
	readyCount := 0
	for _, spec := range cluster.Spec.DatanodeGroups {
		resourceName := common.AdditionalResourceName(cluster.Name, spec.Name, v1alpha1.DatanodeComponentKind)

		// TODO(zyy17): Update datanode status.
		ready, err := d.checkDatanodeStatus(ctx, cluster.Namespace, resourceName, nil)
		if err != nil {
			return false, err
		}
		if ready {
			readyCount++
		}
	}

	if readyCount == len(cluster.Spec.DatanodeGroups) {
		return true, nil
	}

	return false, nil
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
			specEqual, err := k8sutils.IsObjectSpecEqual(oldObject, newObject, deployer.LastAppliedResourceSpec)
			if err != nil {
				return err
			}

			labelsEqual := k8sutils.IsObjectLabelsEqual(oldObject.GetLabels(), newObject.GetLabels())

			// If the spec or labels is not equal, update the object.
			if !specEqual || !labelsEqual {
				if sts, ok := newObject.(*appsv1.StatefulSet); ok && d.shouldUseMaintenanceMode(cluster) {
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

	if d.maintenanceMode && d.shouldUseMaintenanceMode(cluster) {
		klog.Infof("Turn off maintenance mode for datanode, cluster: %s", cluster.Name)
		if err := d.requestMetasrvForMaintenance(cluster, false); err != nil {
			return err
		}
		d.maintenanceMode = false
	}

	return nil
}

func (d *DatanodeDeployer) requestMetasrvForMaintenance(cluster *v1alpha1.GreptimeDBCluster, enabled bool) error {
	requestURL := fmt.Sprintf("http://%s.%s:%d/admin/maintenance?enable=%v", common.ResourceName(cluster.GetName(), v1alpha1.MetaComponentKind), cluster.GetNamespace(), cluster.Spec.Meta.RPCPort, enabled)
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

func (d *DatanodeDeployer) deleteStorage(ctx context.Context, namespace, name string, fsType common.FileStorageType) error {
	klog.Infof("Deleting datanode storage...")

	claims, err := common.GetPVCs(ctx, d.Client, namespace, name, v1alpha1.DatanodeComponentKind, fsType)
	if err != nil {
		return err
	}

	for _, pvc := range claims {
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

func (d *DatanodeDeployer) shouldUseMaintenanceMode(cluster *v1alpha1.GreptimeDBCluster) bool {
	return cluster.GetMeta().IsEnableRegionFailover()
}

var _ deployer.Builder = &datanodeBuilder{}

type datanodeBuilder struct {
	*CommonBuilder
}

func (b *datanodeBuilder) BuildService() deployer.Builder {
	if b.Err != nil {
		return b
	}

	if b.Cluster.Spec.Datanode == nil && len(b.Cluster.Spec.DatanodeGroups) == 0 {
		return b
	}

	if b.Cluster.Spec.Datanode != nil {
		svc, err := b.buildDatanodeService(b.Cluster.Spec.Datanode)
		if err != nil {
			b.Err = err
			return b
		}
		b.Objects = append(b.Objects, svc)
		return b
	}

	for _, spec := range b.Cluster.Spec.DatanodeGroups {
		svc, err := b.buildDatanodeService(spec)
		if err != nil {
			b.Err = err
			return b
		}
		b.Objects = append(b.Objects, svc)
	}

	return b
}

func (b *datanodeBuilder) buildDatanodeService(spec *v1alpha1.DatanodeSpec) (*corev1.Service, error) {
	resourceName := common.AdditionalResourceName(b.Cluster.Name, spec.Name, b.ComponentKind)

	svc := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: b.Cluster.Namespace,
			Name:      resourceName,
			Labels: map[string]string{
				constant.GreptimeDBComponentName: resourceName,
			},
		},
		Spec: corev1.ServiceSpec{
			ClusterIP: corev1.ClusterIPNone,
			Selector: map[string]string{
				constant.GreptimeDBComponentName: resourceName,
			},
			Ports: b.servicePorts(spec),
		},
	}

	return svc, nil
}

func (b *datanodeBuilder) BuildConfigMap() deployer.Builder {
	if b.Err != nil {
		return b
	}

	if b.Cluster.Spec.Datanode == nil && len(b.Cluster.Spec.DatanodeGroups) == 0 {
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

	if b.Cluster.Spec.Datanode == nil && len(b.Cluster.Spec.DatanodeGroups) == 0 {
		return b
	}

	if b.Cluster.Spec.Datanode != nil {
		sts, err := b.buildDatanodeStatefulSet(nil, b.Cluster.Spec.Datanode)
		if err != nil {
			b.Err = err
			return b
		}
		b.Objects = append(b.Objects, sts)
		return b
	}

	for id, spec := range b.Cluster.Spec.DatanodeGroups {
		groupID := int32(id)
		sts, err := b.buildDatanodeStatefulSet(&groupID, spec)
		if err != nil {
			b.Err = err
			return b
		}
		b.Objects = append(b.Objects, sts)
	}

	return b
}

func (b *datanodeBuilder) buildDatanodeStatefulSet(groupID *int32, spec *v1alpha1.DatanodeSpec) (*appsv1.StatefulSet, error) {
	resourceName := common.AdditionalResourceName(b.Cluster.Name, spec.Name, b.ComponentKind)

	sts := &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "StatefulSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      resourceName,
			Namespace: b.Cluster.Namespace,
			Labels: map[string]string{
				constant.GreptimeDBComponentName: resourceName,
			},
		},
		Spec: appsv1.StatefulSetSpec{
			PodManagementPolicy: appsv1.ParallelPodManagement,
			ServiceName:         resourceName,
			Replicas:            spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					constant.GreptimeDBComponentName: resourceName,
				},
			},
			Template:             b.generatePodTemplateSpec(spec, groupID),
			VolumeClaimTemplates: b.generatePVCs(spec),
			UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
				Type:          appsv1.RollingUpdateStatefulSetStrategyType,
				RollingUpdate: spec.RollingUpdate,
			},
		},
	}

	configData, err := dbconfig.FromCluster(b.Cluster, b.ComponentKind)
	if err != nil {
		return nil, err
	}

	sts.Spec.Template.Annotations = util.MergeStringMap(sts.Spec.Template.Annotations,
		map[string]string{deployer.ConfigHash: util.CalculateConfigHash(configData)})

	return sts, nil
}

func (b *datanodeBuilder) BuildPodMonitor() deployer.Builder {
	if b.Err != nil {
		return b
	}

	if b.Cluster.Spec.Datanode == nil && len(b.Cluster.Spec.DatanodeGroups) == 0 {
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

func (b *datanodeBuilder) generateMainContainerArgs(spec *v1alpha1.DatanodeSpec) []string {
	return []string{
		"datanode", "start",
		"--metasrv-addrs", fmt.Sprintf("%s.%s:%d", common.ResourceName(b.Cluster.Name, v1alpha1.MetaComponentKind),
			b.Cluster.Namespace, b.Cluster.Spec.Meta.RPCPort),
		"--http-addr", fmt.Sprintf("0.0.0.0:%d", spec.HTTPPort),
		"--config-file", path.Join(constant.GreptimeDBConfigDir, constant.GreptimeDBConfigFileName),
	}
}

func (b *datanodeBuilder) generatePodTemplateSpec(spec *v1alpha1.DatanodeSpec, groupID *int32) corev1.PodTemplateSpec {
	podTemplateSpec := b.GeneratePodTemplateSpec(spec.Template)

	if len(spec.Template.MainContainer.Args) == 0 {
		// Setup main container args.
		podTemplateSpec.Spec.Containers[constant.MainContainerIndex].Args = b.generateMainContainerArgs(spec)
	}

	podTemplateSpec.Spec.Containers[constant.MainContainerIndex].Ports = b.containerPorts(spec)
	podTemplateSpec.Spec.Containers[constant.MainContainerIndex].Env = append(podTemplateSpec.Spec.Containers[constant.MainContainerIndex].Env, b.env(v1alpha1.DatanodeComponentKind)...)

	b.mountConfigDir(podTemplateSpec)
	b.addVolumeMounts(podTemplateSpec, spec)
	b.addInitConfigDirVolume(podTemplateSpec, common.ResourceName(b.Cluster.Name, b.ComponentKind))

	if logging := spec.GetLogging(); logging != nil &&
		!logging.IsOnlyLogToStdout() && !logging.IsPersistentWithData() {
		b.AddLogsVolume(podTemplateSpec, logging.GetLogsDir())
	}

	if b.Cluster.GetMonitoring().IsEnabled() && b.Cluster.GetMonitoring().GetVector() != nil {
		b.AddVectorConfigVolume(podTemplateSpec)
		b.AddVectorSidecar(podTemplateSpec, v1alpha1.DatanodeComponentKind)
	}

	podTemplateSpec.Spec.InitContainers = append(podTemplateSpec.Spec.InitContainers, *b.generateInitializer(spec, groupID))
	podTemplateSpec.ObjectMeta.Labels = util.MergeStringMap(podTemplateSpec.ObjectMeta.Labels, map[string]string{
		constant.GreptimeDBComponentName: common.AdditionalResourceName(b.Cluster.Name, spec.Name, b.ComponentKind),
	})

	return *podTemplateSpec
}

func (b *datanodeBuilder) generatePVCs(spec *v1alpha1.DatanodeSpec) []corev1.PersistentVolumeClaim {
	var claims []corev1.PersistentVolumeClaim

	// It's always not nil because it's the default value.
	if fs := spec.GetFileStorage(); fs != nil {
		claims = append(claims, *common.FileStorageToPVC(b.Cluster.Name, spec.Name, fs, common.FileStorageTypeDatanode, v1alpha1.DatanodeComponentKind))
	}

	// Allocate the standalone WAL storage for the raft-engine.
	if fs := b.Cluster.GetWALProvider().GetRaftEngineWAL().GetFileStorage(); fs != nil {
		claims = append(claims, *common.FileStorageToPVC(b.Cluster.Name, spec.Name, fs, common.FileStorageTypeWAL, v1alpha1.DatanodeComponentKind))
	}

	// Allocate the standalone cache file storage for the datanode.
	if fs := b.Cluster.GetObjectStorageProvider().GetCacheFileStorage(); fs != nil {
		claims = append(claims, *common.FileStorageToPVC(b.Cluster.Name, spec.Name, fs, common.FileStorageTypeCache, v1alpha1.DatanodeComponentKind))
	}

	return claims
}

func (b *datanodeBuilder) generateInitializer(spec *v1alpha1.DatanodeSpec, groupID *int32) *corev1.Container {
	initializer := &corev1.Container{
		Name:  "initializer",
		Image: b.Cluster.Spec.Initializer.Image,
		Command: []string{
			"greptimedb-initializer",
		},
		Args: []string{
			"--config-path", path.Join(constant.GreptimeDBConfigDir, constant.GreptimeDBConfigFileName),
			"--init-config-path", path.Join(constant.GreptimeDBInitConfigDir, constant.GreptimeDBConfigFileName),
			"--datanode-rpc-port", fmt.Sprintf("%d", spec.RPCPort),
			"--datanode-service-name", common.AdditionalResourceName(b.Cluster.Name, spec.Name, b.ComponentKind),
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
		Env: b.env(v1alpha1.DatanodeComponentKind),
	}

	if groupID != nil {
		initializer.Args = append(initializer.Args, "--datanode-group-id", fmt.Sprintf("%d", *groupID))
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

func (b *datanodeBuilder) addVolumeMounts(template *corev1.PodTemplateSpec, spec *v1alpha1.DatanodeSpec) {
	if fs := spec.GetFileStorage(); fs != nil {
		template.Spec.Containers[constant.MainContainerIndex].VolumeMounts =
			append(template.Spec.Containers[constant.MainContainerIndex].VolumeMounts,
				corev1.VolumeMount{
					Name:      fs.GetName(),
					MountPath: fs.GetMountPath(),
				},
			)
	}

	if fs := b.Cluster.GetWALProvider().GetRaftEngineWAL().GetFileStorage(); fs != nil {
		template.Spec.Containers[constant.MainContainerIndex].VolumeMounts =
			append(template.Spec.Containers[constant.MainContainerIndex].VolumeMounts,
				corev1.VolumeMount{
					Name:      fs.GetName(),
					MountPath: fs.GetMountPath(),
				},
			)
	}

	if fs := b.Cluster.GetObjectStorageProvider().GetCacheFileStorage(); fs != nil {
		template.Spec.Containers[constant.MainContainerIndex].VolumeMounts =
			append(template.Spec.Containers[constant.MainContainerIndex].VolumeMounts,
				corev1.VolumeMount{
					Name:      fs.GetName(),
					MountPath: fs.GetMountPath(),
				},
			)
	}
}

// The init-config volume is used for initializer.
func (b *datanodeBuilder) addInitConfigDirVolume(template *corev1.PodTemplateSpec, configMapName string) {
	template.Spec.Volumes = append(template.Spec.Volumes, corev1.Volume{
		Name: constant.InitConfigVolumeName,
		VolumeSource: corev1.VolumeSource{
			// Mount the configmap as init-config.
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: configMapName,
				},
			},
		},
	})
}

func (b *datanodeBuilder) servicePorts(spec *v1alpha1.DatanodeSpec) []corev1.ServicePort {
	return []corev1.ServicePort{
		{
			Name:     "rpc",
			Protocol: corev1.ProtocolTCP,
			Port:     spec.RPCPort,
		},
		{
			Name:     "http",
			Protocol: corev1.ProtocolTCP,
			Port:     spec.HTTPPort,
		},
	}
}

func (b *datanodeBuilder) containerPorts(spec *v1alpha1.DatanodeSpec) []corev1.ContainerPort {
	return []corev1.ContainerPort{
		{
			Name:          "rpc",
			Protocol:      corev1.ProtocolTCP,
			ContainerPort: spec.RPCPort,
		},
		{
			Name:          "http",
			Protocol:      corev1.ProtocolTCP,
			ContainerPort: spec.HTTPPort,
		},
	}
}
