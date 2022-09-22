package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type StorageRetainPolicyType string

const (
	// RetainStorageRetainPolicyTypeRetain is the default options.
	// The storage(PVCs) will be retained when the cluster is deleted.
	RetainStorageRetainPolicyTypeRetain StorageRetainPolicyType = "Retain"

	// RetainStorageRetainPolicyTypeDelete specifiy that the storage will be deleted when the associated StatefulSet delete.
	RetainStorageRetainPolicyTypeDelete StorageRetainPolicyType = "Delete"
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
	HostNetwork *bool `json:"hostNetwork,omitempty"`

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

	// If specified, the pod will be dispatched by specified scheduler.
	// If not specified, the pod will be dispatched by default scheduler.
	// SchedulerName field is from 'corev1.PodSpec.SchedulerName'.
	// +optional
	SchedulerName string `json:"schedulerName,omitempty"`

	// For most time, there is one main container in a pod(frontend/meta/datanode).
	// If specified, addtional containers will be added to the pod as sidecar containers.
	// +optional
	AddtionalContainers []corev1.Container `json:"addtionalContainers,omitempty"`
}

// MainContainerSpec describes the specification of the main container of a pod.
// Most of the fields of MainContainerSpec are from 'corev1.Container'.
type MainContainerSpec struct {
	// The main container image name of the component.
	// +required
	Image string `json:"image,omitempty"`

	// The resource requirements of the main container.
	// +optional
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`

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

	// Pod volumes to mount into the container's filesystem.
	// VolumeMounts field is from 'corev1.Container.VolumeMounts'.
	// +optional
	VolumeMounts []corev1.VolumeMount `json:"volumeMounts,omitempty"`

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
	ImagePullPolicy *corev1.PullPolicy `json:"imagePullPolicy,omitempty"`
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

// ComponentSpec is the common specification for all components(frontend/meta/datanode).
type ComponentSpec struct {
	// The number of replicas of the components.
	// +reqiured
	Replicas int32 `json:"replicas"`

	// Template defines the pod template for the component, if not specified, the pod template will use the default value.
	// +optional
	Template *PodTemplateSpec `json:"template,omitempty"`
}

// MetaSpec is the specification for meta component.
type MetaSpec struct {
	ComponentSpec `json:",inline"`

	// +optional
	Service corev1.ServiceSpec `json:"service,omitempty"`

	// +optional
	EtcdEndpoints []string `json:"etcdEndpoints,omitempty"`

	// More meta settings can be added here...
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
	StorageSize string `json:"storageSize,omitempty"`

	// The mount path of the storage in datanode container.
	// +optional
	MountPath string `json:"mountPath,omitempty"`

	// The PVCs will retain or delete when the cluster is deleted, default to Retain.
	// +optional
	StorageRetainPolicy StorageRetainPolicyType `json:"storageRetainPolicy,omitempty"`
}

// FrontendSpec is the specification for frontend component.
type FrontendSpec struct {
	ComponentSpec `json:",inline"`

	Service corev1.ServiceSpec `json:"service,omitempty"`
	// More frontend settings can be added here...
}

// DatanodeSpec is the specification for datanode component.
type DatanodeSpec struct {
	ComponentSpec `json:",inline"`

	Storage StorageSpec `json:"storage,omitempty"`
	// More datanode settings can be added here...
}

// GreptimeDBClusterSpec defines the desired state of GreptimeDBCluster
type GreptimeDBClusterSpec struct {
	// Base is the base pod template for all components and can be overridden by template of individual component.
	// +optional
	Base *PodTemplateSpec `json:"base,omitempty"`

	// Frontend is the specification of frontend node.
	// +optional
	Frontend *FrontendSpec `json:"frontend"`

	// Meta is the specification of meta node.
	// +optional
	Meta *MetaSpec `json:"meta"`

	// Datanode is the specification of datanode node.
	// +optional
	Datanode *DatanodeSpec `json:"datanode"`

	// +optinal
	HTTPServicePort int32 `json:"httpServicePort,omitempty"`

	// +optional
	GRPCServicePort int32 `json:"grpcServicePort,omitempty"`

	// +optional
	MySQLServicePort int32 `json:"mysqlServicePort,omitempty"`

	// More cluster settings can be added here...
}

// GreptimeDBClusterStatus defines the observed state of GreptimeDBCluster
type GreptimeDBClusterStatus struct {
	Frontend FrontendStatus `json:"frontend,omitempty"`
	Meta     MetaStatus     `json:"meta,omitempty"`
	Datanode DatanodeStatus `json:"datanode,omitempty"`

	// +optional
	Conditions []GreptimeDBClusterCondition `json:"conditions,omitempty"`
}

// GreptimeDBClusterCondition describes the state of a deployment at a certain point.
type GreptimeDBClusterCondition struct {
	// Type of deployment condition.
	Type GreptimeDBConditionType `json:"type"`

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

	// A human readable message indicating details about the transition.
	// +optional
	Message string `json:"message,omitempty"`
}

type FrontendStatus struct {
	// +optional
	Replicas int32 `json:"replicas,omitempty"`

	// +optional
	ReadyReplicas int32 `json:"readyReplicas,omitempty"`
}

type MetaStatus struct {
	// +optional
	Replicas int32 `json:"replicas,omitempty"`

	// +optional
	ReadyReplicas int32 `json:"readyReplicas,omitempty"`
}

type DatanodeStatus struct {
	// +optional
	Replicas int32 `json:"replicas,omitempty"`

	// +optional
	ReadyReplicas int32 `json:"readyReplicas,omitempty"`
}

type GreptimeDBConditionType string

// These are valid conditions of a GreptimeDBCluster.
const (
	// GreptimeDBClusterReady indicates that the GreptimeDB cluster is ready to serve requests.
	// Every component in the cluster are all ready.
	GreptimeDBClusterReady GreptimeDBConditionType = "Ready"

	// GreptimeDBClusterProgressing indicates that the GreptimeDB cluster is progressing.
	GreptimeDBClusterProgressing GreptimeDBConditionType = "Progressing"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=gtc

// GreptimeDBCluster is the Schema for the greptimedbclusters API
type GreptimeDBCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GreptimeDBClusterSpec   `json:"spec,omitempty"`
	Status GreptimeDBClusterStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// GreptimeDBClusterList contains a list of GreptimeDBCluster
type GreptimeDBClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GreptimeDBCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GreptimeDBCluster{}, &GreptimeDBClusterList{})
}
