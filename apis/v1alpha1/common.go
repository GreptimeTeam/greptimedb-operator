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

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type StorageRetainPolicyType string

const (
	// StorageRetainPolicyTypeRetain is the default options.
	// The storage(PVCs) will be retained when the cluster is deleted.
	StorageRetainPolicyTypeRetain StorageRetainPolicyType = "Retain"

	// StorageRetainPolicyTypeDelete specify that the storage will be deleted when the associated StatefulSet delete.
	StorageRetainPolicyTypeDelete StorageRetainPolicyType = "Delete"
)

// Phase define the phase of the cluster or standalone.
type Phase string

const (
	// PhaseStarting means the controller start to create cluster.
	PhaseStarting Phase = "Starting"

	// PhaseRunning means all the components of cluster is ready.
	PhaseRunning Phase = "Running"

	// PhaseUpdating means the cluster is updating.
	PhaseUpdating Phase = "Updating"

	// PhaseError means some kind of error happen in reconcile.
	PhaseError Phase = "Error"

	// PhaseTerminating means the cluster is terminating.
	PhaseTerminating Phase = "Terminating"
)

// ComponentKind is the kind of the component in the cluster.
type ComponentKind string

const (
	FrontendComponentKind ComponentKind = "frontend"
	DatanodeComponentKind ComponentKind = "datanode"
	MetaComponentKind     ComponentKind = "meta"
	StandaloneKind        ComponentKind = "standalone"
)

// SlimPodSpec is a slimmed down version of corev1.PodSpec.
// Most of the fields in SlimPodSpec are copied from corev1.PodSpec.
type SlimPodSpec struct {
	// NodeSelector is a selector which must be true for the pod to fit on a node.
	// Selector which must match a node's labels for the pod to be scheduled on that node.
	// More info: https://kubernetes.io/docs/concepts/configuration/assign-pod-node/
	// NodeSelector field is from 'corev1.PodSpec.NodeSelector'.
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// List of initialization containers belonging to the pod.
	// Init containers are executed in order prior to containers being started. If any
	// init container fails, the pod is considered to have failed and is handled according
	// to its restartPolicy. The name for an init container or normal container must be
	// unique among all containers.
	// Init containers may not have Lifecycle actions, Readiness probes, Liveness probes, or Startup probes.
	// The resourceRequirements of an init container are taken into account during scheduling
	// by finding the highest request/limit for each resource type, and then using the max of
	// that value or the sum of the normal containers. Limits are applied to init containers
	// in a similar fashion.
	// Init containers cannot currently be added or removed.
	// Cannot be updated.
	// More info: https://kubernetes.io/docs/concepts/workloads/pods/init-containers/
	// InitContainers field is from 'corev1.PodSpec.InitContainers'.
	// +optional
	InitContainers []corev1.Container `json:"initContainers,omitempty"`

	// Restart policy for all containers within the pod.
	// One of Always, OnFailure, Never.
	// Default to Always.
	// More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#restart-policy
	// RestartPolicy field is from 'corev1.PodSpec.RestartPolicy'.
	// +optional
	RestartPolicy corev1.RestartPolicy `json:"restartPolicy,omitempty"`

	// Optional duration in seconds the pod needs to terminate gracefully. May be decreased in delete request.
	// Value must be non-negative integer. The value zero indicates stop immediately via
	// the kill signal (no opportunity to shut down).
	// If this value is nil, the default grace period will be used instead.
	// The grace period is the duration in seconds after the processes running in the pod are sent
	// a termination signal and the time when the processes are forcibly halted with a kill signal.
	// Set this value longer than the expected cleanup time for your process.
	// Defaults to 30 seconds.
	// TerminationGracePeriodSeconds field is from 'corev1.PodSpec.TerminationGracePeriodSeconds'.
	// +optional
	TerminationGracePeriodSeconds *int64 `json:"terminationGracePeriodSeconds,omitempty"`

	// Optional duration in seconds the pod may be active on the node relative to
	// StartTime before the system will actively try to mark it failed and kill associated containers.
	// Value must be a positive integer.
	// ActiveDeadlineSeconds field is from 'corev1.PodSpec.ActiveDeadlineSeconds'.
	// +optional
	ActiveDeadlineSeconds *int64 `json:"activeDeadlineSeconds,omitempty"`

	// Set DNS policy for the pod.
	// Defaults to "ClusterFirst".
	// Valid values are 'ClusterFirstWithHostNet', 'ClusterFirst', 'Default' or 'None'.
	// DNS parameters given in DNSConfig will be merged with the policy selected with DNSPolicy.
	// To have DNS options set along with hostNetwork, you have to specify DNS policy
	// explicitly to 'ClusterFirstWithHostNet'.
	// DNSPolicy field is from 'corev1.PodSpec.DNSPolicy'.
	// +optional
	DNSPolicy corev1.DNSPolicy `json:"dnsPolicy,omitempty"`

	// ServiceAccountName is the name of the ServiceAccount to use to run this pod.
	// More info: https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/
	// ServiceAccountName field is from 'corev1.PodSpec.ServiceAccountName'.
	// +optional
	ServiceAccountName string `json:"serviceAccountName,omitempty"`

	// Host networking requested for this pod. Use the host's network namespace.
	// If this option is set, the ports that will be used must be specified.
	// Default to false.
	// HostNetwork field is from 'corev1.PodSpec.HostNetwork'.
	// +optional
	HostNetwork bool `json:"hostNetwork,omitempty"`

	// ImagePullSecrets is an optional list of references to secrets in the same namespace to use for pulling any of the images used by this PodSpec.
	// If specified, these secrets will be passed to individual puller implementations for them to use.
	// More info: https://kubernetes.io/docs/concepts/containers/images#specifying-imagepullsecrets-on-a-pod
	// ImagePullSecrets field is from 'corev1.PodSpec.ImagePullSecrets'.
	// +optional
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`

	// If specified, the pod's scheduling constraints
	// Affinity field is from 'corev1.PodSpec.Affinity'.
	// +optional
	Affinity *corev1.Affinity `json:"affinity,omitempty"`

	// If specified, the pod's tolerations.
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`

	// If specified, the pod will be dispatched by specified scheduler.
	// If not specified, the pod will be dispatched by default scheduler.
	// SchedulerName field is from 'corev1.PodSpec.SchedulerName'.
	// +optional
	SchedulerName string `json:"schedulerName,omitempty"`

	// For most time, there is one main container in a pod(frontend/meta/datanode).
	// If specified, additional containers will be added to the pod as sidecar containers.
	// +optional
	AdditionalContainers []corev1.Container `json:"additionalContainers,omitempty"`

	// List of volumes that can be mounted by containers belonging to the pod.
	// +optional
	Volumes []corev1.Volume `json:"volumes,omitempty"`
}

// MainContainerSpec describes the specification of the main container of a pod.
// Most of the fields of MainContainerSpec are from 'corev1.Container'.
type MainContainerSpec struct {
	// The main container image name of the component.
	// +required
	Image string `json:"image,omitempty"`

	// The resource requirements of the main container.
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// Entrypoint array. Not executed within a shell.
	// The container image's ENTRYPOINT is used if this is not provided.
	// Variable references $(VAR_NAME) are expanded using the container's environment. If a variable
	// cannot be resolved, the reference in the input string will be unchanged. Double $$ are reduced
	// to a single $, which allows for escaping the $(VAR_NAME) syntax: i.e. "$$(VAR_NAME)" will
	// produce the string literal "$(VAR_NAME)". Escaped references will never be expanded, regardless
	// of whether the variable exists or not. Cannot be updated.
	// More info: https://kubernetes.io/docs/tasks/inject-data-application/define-command-argument-container/#running-a-command-in-a-shell
	// Command field is from 'corev1.Container.Command'.
	// +optional
	Command []string `json:"command,omitempty"`

	// Arguments to the entrypoint.
	// The container image's CMD is used if this is not provided.
	// Variable references $(VAR_NAME) are expanded using the container's environment. If a variable
	// cannot be resolved, the reference in the input string will be unchanged. Double $$ are reduced
	// to a single $, which allows for escaping the $(VAR_NAME) syntax: i.e. "$$(VAR_NAME)" will
	// produce the string literal "$(VAR_NAME)". Escaped references will never be expanded, regardless
	// of whether the variable exists or not. Cannot be updated.
	// More info: https://kubernetes.io/docs/tasks/inject-data-application/define-command-argument-container/#running-a-command-in-a-shell
	// Args field is from 'corev1.Container.Args'.
	// +optional
	Args []string `json:"args,omitempty"`

	// Container's working directory.
	// If not specified, the container runtime's default will be used, which
	// might be configured in the container image.
	// Cannot be updated.
	// WorkingDir field is from 'corev1.Container.WorkingDir'.
	// +optional
	WorkingDir string `json:"workingDir,omitempty"`

	// List of environment variables to set in the container.
	// Cannot be updated.
	// Env field is from 'corev1.Container.Env'.
	// +optional
	Env []corev1.EnvVar `json:"env,omitempty"`

	// Periodic probe of container liveness.
	// Container will be restarted if the probe fails.
	// More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes
	// LivenessProbe field is from 'corev1.Container.LivenessProbe'.
	// +optional
	LivenessProbe *corev1.Probe `json:"livenessProbe,omitempty"`

	// Periodic probe of container service readiness.
	// Container will be removed from service endpoints if the probe fails.
	// ReadinessProbe field is from 'corev1.Container.LivenessProbe'.
	// More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes
	// +optional
	ReadinessProbe *corev1.Probe `json:"readinessProbe,omitempty"`

	// Actions that the management system should take in response to container lifecycle events.
	// Cannot be updated.
	// Lifecycle field is from 'corev1.Container.Lifecycle'.
	// +optional
	Lifecycle *corev1.Lifecycle `json:"lifecycle,omitempty"`

	// Image pull policy.
	// One of Always, Never, IfNotPresent.
	// Defaults to Always if :latest tag is specified, or IfNotPresent otherwise.
	// Cannot be updated.
	// More info: https://kubernetes.io/docs/concepts/containers/images#updating-images
	// ImagePullPolicy field is from 'corev1.Container.ImagePullPolicy'.
	// +optional
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`

	// Pod volumes to mount into the container's filesystem.
	// Cannot be updated.
	// +optional
	VolumeMounts []corev1.VolumeMount `json:"volumeMounts,omitempty"`
}

// PodTemplateSpec defines the template for a pod of cluster.
type PodTemplateSpec struct {
	// The annotations to be created to the pod.
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`

	// The labels to be created to the pod.
	// +optional
	Labels map[string]string `json:"labels,omitempty"`

	// MainContainer defines the specification of the main container of the pod.
	// +optional
	MainContainer *MainContainerSpec `json:"main,omitempty"`

	// SlimPodSpec defines the desired behavior of the pod.
	// +optional
	SlimPodSpec `json:",inline"`
}

// StorageSpec will generate PVC.
type StorageSpec struct {
	// The name of the storage.
	// +optional
	Name string `json:"name,omitempty"`

	// The name of the storage class to use for the volume.
	// +optional
	StorageClassName *string `json:"storageClassName,omitempty"`

	// The size of the storage.
	// +optional
	// +kubebuilder:validation:Pattern=(^([+-]?[0-9.]+)([eEinumkKMGTP]*[-+]?[0-9]*)$)
	StorageSize string `json:"storageSize,omitempty"`

	// The mount path of the storage in datanode container.
	// +optional
	MountPath string `json:"mountPath,omitempty"`

	// The PVCs will retain or delete when the cluster is deleted, default to Retain.
	// +optional
	// +kubebuilder:validation:Enum:={"Retain", "Delete"}
	StorageRetainPolicy StorageRetainPolicyType `json:"storageRetainPolicy,omitempty"`

	// The wal directory of the storage.
	WalDir string `json:"walDir,omitempty"`

	// The datahome directory.
	// +optional
	DataHome string `json:"dataHome,omitempty"`
}

// RemoteWalProvider defines the remote wal provider for the cluster.
type RemoteWalProvider struct {
	// +optional
	KafkaRemoteWal *KafkaRemoteWal `json:"kafka,omitempty"`
}

// KafkaRemoteWal is the specification for remote WAL that uses Kafka.
type KafkaRemoteWal struct {
	// +optional
	BrokerEndpoints []string `json:"brokerEndpoints,omitempty"`
}

type ServiceSpec struct {
	// type determines how the Service is exposed.
	// +optional
	Type corev1.ServiceType `json:"type,omitempty"`

	// Additional annotations for the service
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`

	// Additional labels for the service
	// +optional
	Labels map[string]string `json:"labels,omitempty"`

	// loadBalancerClass is the class of the load balancer implementation this Service belongs to.
	// +optional
	LoadBalancerClass *string `json:"loadBalancerClass,omitempty"`
}

type TLSSpec struct {
	// The secret name of the TLS certificate, and it must be in the same namespace of the cluster.
	// The secret must contain keys named ca.crt, tls.crt and tls.key.
	// +optional
	SecretName string `json:"secretName,omitempty"`
}

// ObjectStorageProvider defines the storage provider for the cluster. The data will be stored in the storage.
type ObjectStorageProvider struct {
	S3            *S3StorageProvider  `json:"s3,omitempty"`
	OSS           *OSSStorageProvider `json:"oss,omitempty"`
	CachePath     string              `json:"cachePath,omitempty"`
	CacheCapacity string              `json:"cacheCapacity,omitempty"`
}

type S3StorageProvider struct {
	// The data will be stored in the bucket.
	// +optional
	Bucket string `json:"bucket,omitempty"`

	// The region of the bucket.
	// +optional
	Region string `json:"region,omitempty"`

	// The endpoint of the bucket.
	// +optional
	Endpoint string `json:"endpoint,omitempty"`

	// The secret of storing the credentials of access key id and secret access key.
	// The secret must be the same namespace with the GreptimeDBCluster resource.
	// +optional
	SecretName string `json:"secretName,omitempty"`

	// The S3 directory path.
	// +optional
	Root string `json:"root,omitempty"`
}

type OSSStorageProvider struct {
	// The data will be stored in the bucket.
	// +optional
	Bucket string `json:"bucket,omitempty"`

	// The region of the bucket.
	// +optional
	Region string `json:"region,omitempty"`

	// The endpoint of the bucket.
	// +optional
	Endpoint string `json:"endpoint,omitempty"`

	// The secret of storing the credentials of access key id and secret access key.
	// The secret must be the same namespace with the GreptimeDBCluster resource.
	// +optional
	SecretName string `json:"secretName,omitempty"`

	// The OSS directory path.
	// +optional
	Root string `json:"root,omitempty"`
}

// PrometheusMonitorSpec defines the PodMonitor configuration.
type PrometheusMonitorSpec struct {
	// Enable a Prometheus PodMonitor
	// +optional
	Enabled bool `json:"enabled,omitempty"`

	// Prometheus PodMonitor labels.
	// +optional
	Labels map[string]string `json:"labels,omitempty"`

	// Interval at which metrics should be scraped
	// +optional
	Interval string `json:"interval,omitempty"`
}

type ConditionType string

// These are valid conditions of a GreptimeDBCluster and GreptimeDBStandalone.
const (
	// ConditionTypeReady indicates that the GreptimeDB cluster is ready to serve requests.
	// Every component in the cluster are all ready.
	ConditionTypeReady ConditionType = "Ready"

	// ConditionTypeProgressing indicates that the GreptimeDB cluster is progressing.
	ConditionTypeProgressing ConditionType = "Progressing"
)

// Condition describes the state of a deployment at a certain point.
type Condition struct {
	// Type of deployment condition.
	Type ConditionType `json:"type"`

	// Status of the condition, one of True, False, Unknown.
	Status corev1.ConditionStatus `json:"status"`

	// The last time this condition was updated.
	LastUpdateTime metav1.Time `json:"lastUpdateTime,omitempty"`

	// Last time the condition transitioned from one status to another.
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`

	// The reason for the condition's last transition.
	// +optional
	Reason string `json:"reason,omitempty"`

	// A human-readable message indicating details about the transition.
	// +optional
	Message string `json:"message,omitempty"`
}

func NewCondition(conditionType ConditionType, conditionStatus corev1.ConditionStatus,
	reason, message string) *Condition {
	condition := Condition{
		Type:               conditionType,
		Status:             conditionStatus,
		LastTransitionTime: metav1.Now(),
		LastUpdateTime:     metav1.Now(),
		Reason:             reason,
		Message:            message,
	}
	return &condition
}

func GetCondition(conditions []Condition, conditionType ConditionType) *Condition {
	for i := range conditions {
		c := conditions[i]
		if conditions[i].Type == conditionType {
			return &c
		}
	}
	return nil
}

func SetCondition(conditions []Condition, condition Condition) []Condition {
	currentCondition := GetCondition(conditions, condition.Type)
	if currentCondition != nil &&
		currentCondition.Status == condition.Status &&
		currentCondition.Reason == condition.Reason {
		currentCondition.LastUpdateTime = condition.LastUpdateTime
		return conditions
	}

	if currentCondition != nil && currentCondition.Status == condition.Status {
		condition.LastTransitionTime = currentCondition.LastTransitionTime
	}

	return append(filterOutCondition(conditions, condition.Type), condition)
}

func filterOutCondition(conditions []Condition, conditionType ConditionType) []Condition {
	var newConditions []Condition
	for _, c := range conditions {
		if c.Type == conditionType {
			continue
		}
		newConditions = append(newConditions, c)
	}
	return newConditions
}
