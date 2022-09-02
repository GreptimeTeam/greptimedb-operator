package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func (in *GreptimeDBCluster) SetDefaults() {
	if in == nil {
		return
	}

	if in.Spec.Base != nil {
		if in.Spec.Base.MainContainer.Resources == nil {
			in.Spec.Base.MainContainer.Resources = &corev1.ResourceRequirements{
				Requests: map[corev1.ResourceName]resource.Quantity{
					"cpu":    resource.MustParse(defaultRequestCPU),
					"memory": resource.MustParse(defaultRequestMemory),
				},
				Limits: map[corev1.ResourceName]resource.Quantity{
					"cpu":    resource.MustParse(defaultLimitCPU),
					"memory": resource.MustParse(defaultLimitMemory),
				},
			}
		}
	}

	if in.Spec.Frontend != nil {
		if in.Spec.Frontend.Template == nil {
			in.Spec.Frontend.Template = in.Spec.Base
		} else {
			in.Spec.Frontend.Template.overlay(in.Spec.Base)
		}
	}

	if in.Spec.Meta != nil {
		if in.Spec.Meta.Template == nil {
			in.Spec.Meta.Template = in.Spec.Base
		} else {
			in.Spec.Meta.Template.overlay(in.Spec.Base)
		}

		in.Spec.Meta.Etcd.setDefaults()
	}

	if in.Spec.Datanode != nil {
		if in.Spec.Datanode.Template == nil {
			in.Spec.Datanode.Template = in.Spec.Base
		} else {
			in.Spec.Datanode.Template.overlay(in.Spec.Base)
		}
	}

	if in.Spec.HTTPServicePort == 0 {
		in.Spec.HTTPServicePort = int32(defaultHTTPServicePort)
	}

	if in.Spec.GRPCServicePort == 0 {
		in.Spec.GRPCServicePort = int32(defaultGRPCServicePort)
	}

	if in.Spec.MySQLServicePort == 0 {
		in.Spec.MySQLServicePort = int32(defaultMySQLServicePort)
	}
}

func (in *EtcdSpec) setDefaults() {
	if in == nil {
		return
	}

	if in.Image == "" {
		in.Image = defaultEtcdImage
	}

	if in.ClusterSize == 0 {
		in.ClusterSize = defaultClusterSize
	}

	if in.ClientPort == 0 {
		in.ClientPort = int32(defaultEtcdClientPort)
	}

	if in.PeerPort == 0 {
		in.PeerPort = int32(defaultEtcdPeerPort)
	}

	if in.Storage.Name == "" {
		in.Storage.Name = defaultEtcdStorageName
	}

	if in.Storage.StorageSize == "" {
		in.Storage.StorageSize = defaultEtcdStorageSize
	}

	if in.Storage.StorageClassName == nil {
		storageClassName := defaultEtcdStorageClassName
		in.Storage.StorageClassName = &storageClassName
	}

	if in.Storage.MountPath == "" {
		in.Storage.MountPath = defaultEtcdStorageMountPath
	}
}

func (in *PodTemplateSpec) overlay(base *PodTemplateSpec) {
	if base == nil {
		return
	}

	in.Labels = merge(base.Labels, in.Labels)
	in.Annotations = merge(base.Annotations, in.Annotations)

	if in.MainContainer == nil {
		in.MainContainer = base.MainContainer
	} else {
		in.MainContainer.overlay(base.MainContainer)
	}

	if in.NodeSelector == nil {
		in.NodeSelector = base.NodeSelector
	}

	if in.InitContainers == nil {
		in.InitContainers = base.InitContainers
	}

	if in.RestartPolicy == "" {
		in.RestartPolicy = base.RestartPolicy
	}

	if in.TerminationGracePeriodSeconds == nil {
		in.TerminationGracePeriodSeconds = base.TerminationGracePeriodSeconds
	}

	if in.ActiveDeadlineSeconds == nil {
		in.ActiveDeadlineSeconds = base.ActiveDeadlineSeconds
	}

	if in.DNSPolicy == "" {
		in.DNSPolicy = base.DNSPolicy
	}

	if in.ServiceAccountName == "" {
		in.ServiceAccountName = base.ServiceAccountName
	}

	if in.HostNetwork == nil {
		in.HostNetwork = base.HostNetwork
	}

	if in.ImagePullSecrets == nil {
		in.ImagePullSecrets = base.ImagePullSecrets
	}

	if in.Affinity == nil {
		in.Affinity = base.Affinity
	}

	if in.SchedulerName == "" {
		in.SchedulerName = base.SchedulerName
	}

	if in.AddtionalContainers == nil {
		in.AddtionalContainers = base.AddtionalContainers
	}
}

func (in *MainContainerSpec) overlay(base *MainContainerSpec) {
	if base == nil {
		return
	}

	if in.Image == "" {
		in.Image = base.Image
	}

	if in.Resources == nil {
		in.Resources = base.Resources
	}

	if in.Command == nil {
		in.Command = base.Command
	}

	if in.Args == nil {
		in.Args = base.Args
	}

	if in.WorkingDir == "" {
		in.WorkingDir = base.WorkingDir
	}

	if in.VolumeMounts == nil {
		in.VolumeMounts = base.VolumeMounts
	}

	if in.Env == nil {
		in.Env = base.Env
	}

	if in.LivenessProbe == nil {
		in.LivenessProbe = base.LivenessProbe
	}

	if in.ReadinessProbe == nil {
		in.ReadinessProbe = base.ReadinessProbe
	}

	if in.Lifecycle == nil {
		in.Lifecycle = base.Lifecycle
	}

	if in.ImagePullPolicy == nil {
		in.ImagePullPolicy = base.ImagePullPolicy
	}
}

func merge(base, overlay map[string]string) map[string]string {
	if base == nil && overlay == nil {
		return nil
	}

	if base == nil {
		return overlay
	}

	if overlay == nil {
		return base
	}

	output := make(map[string]string)
	for k, v := range base {
		output[k] = v
	}
	for k, v := range overlay {
		output[k] = v
	}

	return output
}
