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

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
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

func (d *DatanodeDeployer) Render(crdObject client.Object) ([]client.Object, error) {
	var renderObjects []client.Object

	cluster, err := d.GetCluster(crdObject)
	if err != nil {
		return nil, err
	}

	if cluster.Spec.Datanode != nil {
		svc, err := d.generateSvc(cluster)
		if err != nil {
			return nil, err
		}
		renderObjects = append(renderObjects, svc)

		sts, err := d.generateSts(cluster)
		if err != nil {
			return nil, err
		}
		renderObjects = append(renderObjects, sts)

		cm, err := d.GenerateConfigMap(cluster, v1alpha1.DatanodeComponentKind)
		if err != nil {
			return nil, err
		}
		renderObjects = append(renderObjects, cm)

		if cluster.Spec.PrometheusMonitor != nil {
			if cluster.Spec.PrometheusMonitor.Enabled {
				pm, err := d.generatePodMonitor(cluster)
				if err != nil {
					return nil, err
				}
				renderObjects = append(renderObjects, pm)
			}
		}
	}

	return renderObjects, nil
}

func (d *DatanodeDeployer) CleanUp(ctx context.Context, crdObject client.Object) error {
	cluster, err := d.GetCluster(crdObject)
	if err != nil {
		return err
	}

	if cluster.Spec.Datanode != nil {
		if cluster.Spec.Datanode.Storage.StorageRetainPolicy == v1alpha1.RetainStorageRetainPolicyTypeDelete {
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
			Name:      d.ResourceName(cluster.Name, v1alpha1.DatanodeComponentKind),
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

	return deployer.IsStatefulSetReady(sts), nil
}

func (d *DatanodeDeployer) deleteStorage(ctx context.Context, cluster *v1alpha1.GreptimeDBCluster) error {
	klog.Infof("Deleting datanode storage...")

	selector, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{
		MatchLabels: map[string]string{
			GreptimeDBComponentName: d.ResourceName(cluster.Name, v1alpha1.DatanodeComponentKind),
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

func (d *DatanodeDeployer) generateSvc(cluster *v1alpha1.GreptimeDBCluster) (*corev1.Service, error) {
	svc := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: cluster.Namespace,
			Name:      d.ResourceName(cluster.Name, v1alpha1.DatanodeComponentKind),
		},
		Spec: corev1.ServiceSpec{
			ClusterIP: corev1.ClusterIPNone,
			Selector: map[string]string{
				GreptimeDBComponentName: d.ResourceName(cluster.Name, v1alpha1.DatanodeComponentKind),
			},
			Ports: []corev1.ServicePort{
				{
					Name:     "grpc",
					Protocol: corev1.ProtocolTCP,
					Port:     cluster.Spec.GRPCServicePort,
				},
				{
					Name:     "http",
					Protocol: corev1.ProtocolTCP,
					Port:     cluster.Spec.HTTPServicePort,
				},
			},
		},
	}

	if err := deployer.SetControllerAndAnnotation(cluster, svc, d.Scheme, svc.Spec); err != nil {
		return nil, err
	}

	return svc, nil
}

func (d *DatanodeDeployer) generateSts(cluster *v1alpha1.GreptimeDBCluster) (*appsv1.StatefulSet, error) {
	sts := &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "StatefulSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      d.ResourceName(cluster.Name, v1alpha1.DatanodeComponentKind),
			Namespace: cluster.Namespace,
		},
		Spec: appsv1.StatefulSetSpec{
			ServiceName: d.ResourceName(cluster.Name, v1alpha1.DatanodeComponentKind),
			Replicas:    &cluster.Spec.Datanode.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					GreptimeDBComponentName: d.ResourceName(cluster.Name, v1alpha1.DatanodeComponentKind),
				},
			},
			Template: *d.generatePodTemplateSpec(cluster),
			VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: cluster.Spec.Datanode.Storage.Name,
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						StorageClassName: cluster.Spec.Datanode.Storage.StorageClassName,
						AccessModes: []corev1.PersistentVolumeAccessMode{
							corev1.ReadWriteOnce,
						},
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceStorage: resource.MustParse(cluster.Spec.Datanode.Storage.StorageSize),
							},
						},
					},
				},
			},
		},
	}

	if err := deployer.SetControllerAndAnnotation(cluster, sts, d.Scheme, sts.Spec); err != nil {
		return nil, err
	}

	return sts, nil
}

func (d *DatanodeDeployer) generatePodMonitor(cluster *v1alpha1.GreptimeDBCluster) (*monitoringv1.PodMonitor, error) {
	pm := &monitoringv1.PodMonitor{
		TypeMeta: metav1.TypeMeta{
			Kind:       monitoringv1.PodMonitorsKind,
			APIVersion: monitoringv1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      d.ResourceName(cluster.Name, v1alpha1.DatanodeComponentKind),
			Namespace: cluster.Namespace,
			Labels:    cluster.Spec.PrometheusMonitor.LabelsSelector,
		},
		Spec: monitoringv1.PodMonitorSpec{
			PodMetricsEndpoints: []monitoringv1.PodMetricsEndpoint{
				{
					Path:        cluster.Spec.PrometheusMonitor.Path,
					Port:        cluster.Spec.PrometheusMonitor.Port,
					Interval:    cluster.Spec.PrometheusMonitor.Interval,
					HonorLabels: cluster.Spec.PrometheusMonitor.HonorLabels,
				},
			},
			Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					GreptimeDBComponentName: d.ResourceName(cluster.Name, v1alpha1.DatanodeComponentKind),
				},
			},
			NamespaceSelector: monitoringv1.NamespaceSelector{
				MatchNames: []string{
					cluster.Namespace,
				},
			},
		},
	}

	if err := deployer.SetControllerAndAnnotation(cluster, pm, d.Scheme, pm.Spec); err != nil {
		return nil, err
	}

	return pm, nil
}

func (d *DatanodeDeployer) buildDatanodeArgs(cluster *v1alpha1.GreptimeDBCluster) []string {
	return []string{
		"datanode", "start",
		"--metasrv-addr", fmt.Sprintf("%s.%s:%d", d.ResourceName(cluster.Name, v1alpha1.MetaComponentKind), cluster.Namespace, cluster.Spec.Meta.ServicePort),
		"--config-file", path.Join(GreptimeDBConfigDir, GreptimeDBConfigFileName),
	}
}

func (d *DatanodeDeployer) generatePodTemplateSpec(cluster *v1alpha1.GreptimeDBCluster) *corev1.PodTemplateSpec {
	podTemplateSpec := deployer.GeneratePodTemplateSpec(cluster.Spec.Datanode.Template, string(v1alpha1.DatanodeComponentKind))

	if len(cluster.Spec.Datanode.Template.MainContainer.Args) == 0 {
		// Setup main container args.
		podTemplateSpec.Spec.Containers[0].Args = d.buildDatanodeArgs(cluster)
	}

	podTemplateSpec.Spec.Containers[0].Env = append(podTemplateSpec.Spec.Containers[0].Env, corev1.EnvVar{
		Name: "POD_IP",
		ValueFrom: &corev1.EnvVarSource{
			FieldRef: &corev1.ObjectFieldSelector{
				FieldPath: "status.podIP",
			},
		},
	})

	podTemplateSpec.Spec.Containers[0].VolumeMounts = []corev1.VolumeMount{
		{
			Name:      cluster.Spec.Datanode.Storage.Name,
			MountPath: cluster.Spec.Datanode.Storage.MountPath,
		},
		{
			Name:      "config",
			MountPath: GreptimeDBConfigDir,
		},
	}

	podTemplateSpec.Spec.Containers[0].Ports = []corev1.ContainerPort{
		{
			Name:          "grpc",
			Protocol:      corev1.ProtocolTCP,
			ContainerPort: cluster.Spec.GRPCServicePort,
		},
		{
			Name:          "http",
			Protocol:      corev1.ProtocolTCP,
			ContainerPort: cluster.Spec.HTTPServicePort,
		},
	}

	podTemplateSpec.Spec.Volumes = []corev1.Volume{
		{
			Name: "config",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
		{
			Name: "init-config",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: d.ResourceName(cluster.Name, v1alpha1.DatanodeComponentKind),
					},
				},
			},
		},
	}

	podTemplateSpec.Spec.InitContainers = append(podTemplateSpec.Spec.InitContainers, *d.generateInitializer(cluster))

	podTemplateSpec.ObjectMeta.Labels = deployer.MergeStringMap(podTemplateSpec.ObjectMeta.Labels, map[string]string{
		GreptimeDBComponentName: d.ResourceName(cluster.Name, v1alpha1.DatanodeComponentKind),
	})

	return podTemplateSpec
}

func (d *DatanodeDeployer) generateInitializer(cluster *v1alpha1.GreptimeDBCluster) *corev1.Container {
	initializer := &corev1.Container{
		Name:  "initializer",
		Image: cluster.Spec.Initializer.Image,
		Command: []string{
			"greptimedb-initializer",
		},
		Args: []string{
			"--config-path", path.Join(GreptimeDBConfigDir, GreptimeDBConfigFileName),
			"--init-config-path", path.Join(GreptimeDBInitConfigDir, GreptimeDBConfigFileName),
			"--datanode-rpc-port", fmt.Sprintf("%d", cluster.Spec.GRPCServicePort),
			"--datanode-service-name", d.ResourceName(cluster.Name, v1alpha1.DatanodeComponentKind),
			"--namespace", cluster.Namespace,
			"--component-kind", string(v1alpha1.DatanodeComponentKind),
		},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      "config",
				MountPath: GreptimeDBConfigDir,
			},
			{
				Name:      "init-config",
				MountPath: GreptimeDBInitConfigDir,
			},
		},
		// TODO(zyy17): the datanode don't support to accept hostname.
		Env: []corev1.EnvVar{
			{
				Name: "POD_IP",
				ValueFrom: &corev1.EnvVarSource{
					FieldRef: &corev1.ObjectFieldSelector{
						FieldPath: "status.podIP",
					},
				},
			},
			{
				Name: "POD_NAME",
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
