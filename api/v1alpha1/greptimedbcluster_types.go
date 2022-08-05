/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SlimPodTemplateSpec is a slimmed down version of corev1.PodTemplate.
// Most of the SlimPodTemplateSpec fields are copied from corev1.PodTemplateSpec.
type SlimPodTemplateSpec struct {
	// Annotations is an unstructured key value map stored with a resource that may be
	// set by external tools to store and retrieve arbitrary metadata. They are not
	// queryable and should be preserved when modifying objects.
	// More info: http://kubernetes.io/docs/user-guide/annotations
	// Annotations field is from 'corev1.PodTemplateSpec.ObjectMeta.Annotations'.
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`

	// Map of string keys and values that can be used to organize and categorize
	// (scope and select) objects. May match selectors of replication controllers
	// and services.
	// More info: http://kubernetes.io/docs/user-guide/labels
	// Labels field is from 'corev1.PodTemplateSpec.ObjectMeta.Labels'.
	// +optional
	Labels map[string]string `json:"labels,omitempty"`

	// NodeSelector is a selector which must be true for the pod to fit on a node.
	// Selector which must match a node's labels for the pod to be scheduled on that node.
	// More info: https://kubernetes.io/docs/concepts/configuration/assign-pod-node/
	// NodeSelector field is from 'corev1.PodTemplateSpec.PodSpec.NodeSelector'.
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
	// InitContainers field is from 'corev1.PodTemplateSpec.Spec.InitContainers'.
	// +optional
	InitContainers []corev1.Container `json:"initContainers,omitempty"`

	// Restart policy for all containers within the pod.
	// One of Always, OnFailure, Never.
	// Default to Always.
	// More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#restart-policy
	// RestartPolicy field is from 'corev1.PodTemplateSpec.Spec.RestartPolicy'.
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
	// TerminationGracePeriodSeconds field is from 'corev1.PodTemplateSpec.Spec.TerminationGracePeriodSeconds'.
	// +optional
	TerminationGracePeriodSeconds *int64 `json:"terminationGracePeriodSeconds,omitempty"`

	// Optional duration in seconds the pod may be active on the node relative to
	// StartTime before the system will actively try to mark it failed and kill associated containers.
	// Value must be a positive integer.
	// ActiveDeadlineSeconds field is from 'corev1.PodTemplateSpec.Spec.ActiveDeadlineSeconds'.
	// +optional
	ActiveDeadlineSeconds *int64 `json:"activeDeadlineSeconds,omitempty"`

	// Set DNS policy for the pod.
	// Defaults to "ClusterFirst".
	// Valid values are 'ClusterFirstWithHostNet', 'ClusterFirst', 'Default' or 'None'.
	// DNS parameters given in DNSConfig will be merged with the policy selected with DNSPolicy.
	// To have DNS options set along with hostNetwork, you have to specify DNS policy
	// explicitly to 'ClusterFirstWithHostNet'.
	// DNSPolicy field is from 'corev1.PodTemplateSpec.Spec.DNSPolicy'.
	// +optional
	DNSPolicy corev1.DNSPolicy `json:"dnsPolicy,omitempty"`

	// ServiceAccountName is the name of the ServiceAccount to use to run this pod.
	// More info: https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/
	// ServiceAccountName field is from 'corev1.PodTemplateSpec.Spec.ServiceAccountName'.
	// +optional
	ServiceAccountName string `json:"serviceAccountName,omitempty"`

	// Host networking requested for this pod. Use the host's network namespace.
	// If this option is set, the ports that will be used must be specified.
	// Default to false.
	// HostNetwork field is from 'corev1.PodTemplateSpec.Spec.HostNetwork'.
	// +optional
	HostNetwork bool `json:"hostNetwork,omitempty"`

	// ImagePullSecrets is an optional list of references to secrets in the same namespace to use for pulling any of the images used by this PodSpec.
	// If specified, these secrets will be passed to individual puller implementations for them to use.
	// More info: https://kubernetes.io/docs/concepts/containers/images#specifying-imagepullsecrets-on-a-pod
	// ImagePullSecrets field is from 'corev1.PodTemplateSpec.Spec.ImagePullSecrets'.
	// +optional
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`

	// If specified, the pod's scheduling constraints
	// +optional
	Affinity *corev1.Affinity `json:"affinity,omitempty"`

	// If specified, the pod will be dispatched by specified scheduler.
	// If not specified, the pod will be dispatched by default scheduler.
	// SchedulerName field is from 'corev1.PodTemplateSpec.Spec.SchedulerName'.
	// +optional
	SchedulerName string `json:"schedulerName,omitempty"`

	// Image pull policy.
	// One of Always, Never, IfNotPresent.
	// Defaults to Always if :latest tag is specified, or IfNotPresent otherwise.
	// Cannot be updated.
	// More info: https://kubernetes.io/docs/concepts/containers/images#updating-images
	// ImagePullPolicy field is from 'corev1.PodTemplateSpec.Spec.Containers[_].ImagePullSecrets'.
	// +optional
	ImagePullPolicy *corev1.PullPolicy `json:"imagePullPolicy,omitempty"`

	// List of environment variables to set in the container.
	// Cannot be updated.
	// Env field is from 'corev1.PodTemplateSpec.Spec.Containers[_].Env'.
	// +optional
	Env []corev1.EnvVar `json:"env,omitempty"`

	// Arguments to the entrypoint of primary container.
	// +optional
	AddtionalArgs []string `json:"addtionalArgs,omitempty"`

	// For most time, there is one primary container in a pod(frontend/meta/datanode).
	// If specified, addtional containers will be added to the pod.
	// +optional
	AddtionalContainers []corev1.Container `json:"addtionalContainers,omitempty"`
}

// ComponentSpec is the common specification for all components(frontend/meta/datanode).
type ComponentSpec struct {
	// The primary container image name of the component.
	// +required
	Image string `json:"image,omitempty"`

	// The number of replicas of the components.
	// +required
	Replicas int `json:"replicas,omitempty"`

	// Resources describes the compute resource requirements of the primary container.
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// Template defines the pod template for the component, if not specified, the pod template will use the base pod template.
	// +optional
	Template *SlimPodTemplateSpec `json:"template,omitempty"`
}

// MetaSpec is the specification for meta component.
type MetaSpec struct {
	ComponentSpec `json:",inline"`

	// More meta settings can be added here...
}

// FrontendSpec is the specification for frontend component.
type FrontendSpec struct {
	ComponentSpec `json:",inline"`

	// More frontend settings can be added here...
}

// DatanodeSpec is the specification for datanode component.
type DatanodeSpec struct {
	ComponentSpec `json:",inline"`

	// More datanode settings can be added here...
}

// GreptimeDBClusterSpec defines the desired state of GreptimeDBCluster
type GreptimeDBClusterSpec struct {
	// Base is the base pod template for all components and can be overridden by template of invidual component.
	// +optional
	Base *SlimPodTemplateSpec `json:"base,omitempty"`

	// Frontend is the specification of frontend node.
	// +required
	Frontend FrontendSpec `json:"frontend,omitempty"`

	// Meta is the specification of meta node.
	// +required
	Meta MetaSpec `json:"meta,omitempty"`

	// Datanode is the specification of datanode node.
	// +required
	Datanode DatanodeSpec `json:"datanode,omitempty"`

	// +optional
	HTTPServicePort *int32 `json:"httpServicePort,omitempty"`

	// +optional
	GRPCServicePort *int32 `json:"grpcServicePort,omitempty"`

	// More cluster settings can be added here...
}

// GreptimeDBClusterStatus defines the observed state of GreptimeDBCluster
type GreptimeDBClusterStatus struct {
	// TODO(zyy17): add status for each component.
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// GreptimeDBCluster is the Schema for the greptimedbclusters API
type GreptimeDBCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"meta data,omitempty"`

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
