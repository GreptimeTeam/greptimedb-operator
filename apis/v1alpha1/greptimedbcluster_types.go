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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ComponentSpec is the common specification for all components(frontend/meta/datanode).
type ComponentSpec struct {
	// The number of replicas of the components.
	// +optional
	// +kubebuilder:validation:Minimum=0
	Replicas *int32 `json:"replicas"`

	// The content of the configuration file of the component in TOML format.
	// +optional
	Config string `json:"config,omitempty"`

	// Template defines the pod template for the component, if not specified, the pod template will use the default value.
	// +optional
	Template *PodTemplateSpec `json:"template,omitempty"`
}

// MetaSpec is the specification for meta component.
type MetaSpec struct {
	ComponentSpec `json:",inline"`

	// +optional
	ServicePort int32 `json:"servicePort,omitempty"`

	// +optional
	EtcdEndpoints []string `json:"etcdEndpoints,omitempty"`

	// EnableCheckEtcdService indicates whether to check etcd cluster health when starting meta.
	// +optional
	EnableCheckEtcdService bool `json:"enableCheckEtcdService,omitempty"`

	// EnableRegionFailover indicates whether to enable region failover.
	// +optional
	EnableRegionFailover bool `json:"enableRegionFailover,omitempty"`

	// The meta will store data with this key prefix.
	// +optional
	StoreKeyPrefix string `json:"storeKeyPrefix,omitempty"`

	// More meta settings can be added here...
}

// FrontendSpec is the specification for frontend component.
type FrontendSpec struct {
	ComponentSpec `json:",inline"`

	// +optional
	Service ServiceSpec `json:"service,omitempty"`

	// The TLS configurations of the frontend.
	// +optional
	TLS *TLSSpec `json:"tls,omitempty"`

	// More frontend settings can be added here...
}

// DatanodeSpec is the specification for datanode component.
type DatanodeSpec struct {
	ComponentSpec `json:",inline"`

	Storage StorageSpec `json:"storage,omitempty"`
	// More datanode settings can be added here...
}

// InitializerSpec is the init container to set up components configurations before running the container.
type InitializerSpec struct {
	// +optional
	Image string `json:"image,omitempty"`
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

	// +optional
	HTTPServicePort int32 `json:"httpServicePort,omitempty"`

	// +optional
	GRPCServicePort int32 `json:"grpcServicePort,omitempty"`

	// +optional
	MySQLServicePort int32 `json:"mysqlServicePort,omitempty"`

	// +optional
	PostgresServicePort int32 `json:"postgresServicePort,omitempty"`

	// +optional
	OpenTSDBServicePort int32 `json:"openTSDBServicePort,omitempty"`

	// +optional
	PrometheusServicePort int32 `json:"prometheusServicePort,omitempty"`

	// +optional
	EnableInfluxDBProtocol bool `json:"enableInfluxDBProtocol,omitempty"`

	// +optional
	PrometheusMonitor *PrometheusMonitorSpec `json:"prometheusMonitor,omitempty"`

	// +optional
	// The version of greptimedb.
	Version string `json:"version,omitempty"`

	// +optional
	Initializer *InitializerSpec `json:"initializer,omitempty"`

	// +optional
	ObjectStorageProvider *ObjectStorageProvider `json:"objectStorage,omitempty"`

	// +optional
	// Attention: This option need to be supported by the Reloader(https://github.com/stakater/Reloader).
	// We may consider to implement the same feature inside the operator in the future.
	ReloadWhenConfigChange bool `json:"reloadWhenConfigChange,omitempty"`

	// More cluster settings can be added here.
}

// GreptimeDBClusterStatus defines the observed state of GreptimeDBCluster
type GreptimeDBClusterStatus struct {
	// +optional
	Frontend FrontendStatus `json:"frontend,omitempty"`

	// +optional
	Meta MetaStatus `json:"meta,omitempty"`

	// +optional
	Datanode DatanodeStatus `json:"datanode,omitempty"`

	// +optional
	Version string `json:"version,omitempty"`

	// +optional
	ClusterPhase Phase `json:"clusterPhase,omitempty"`

	// +optional
	Conditions []Condition `json:"conditions,omitempty"`

	// +optional
	ObservedGeneration *int64 `json:"observedGeneration,omitempty"`
}

type FrontendStatus struct {
	Replicas      int32 `json:"replicas"`
	ReadyReplicas int32 `json:"readyReplicas"`
}

type MetaStatus struct {
	Replicas      int32 `json:"replicas"`
	ReadyReplicas int32 `json:"readyReplicas"`

	// +optional
	EtcdEndpoints []string `json:"etcdEndpoints,omitempty"`
}

type DatanodeStatus struct {
	Replicas      int32 `json:"replicas"`
	ReadyReplicas int32 `json:"readyReplicas"`
}

func (in *GreptimeDBClusterStatus) GetCondition(conditionType ConditionType) *Condition {
	return GetCondition(in.Conditions, conditionType)
}

func (in *GreptimeDBClusterStatus) SetCondition(condition Condition) {
	in.Conditions = SetCondition(in.Conditions, condition)
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=gtc
// +kubebuilder:printcolumn:name="FRONTEND",type="integer",JSONPath=".status.frontend.readyReplicas"
// +kubebuilder:printcolumn:name="DATANODE",type="integer",JSONPath=".status.datanode.readyReplicas"
// +kubebuilder:printcolumn:name="META",type="integer",JSONPath=".status.meta.readyReplicas"
// +kubebuilder:printcolumn:name="PHASE",type=string,JSONPath=".status.clusterPhase"
// +kubebuilder:printcolumn:name="VERSION",type=string,JSONPath=".status.version"
// +kubebuilder:printcolumn:name="AGE",type=date,JSONPath=".metadata.creationTimestamp"

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
