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

// ComponentSpec is the common specification for all components(`frontend`/`meta`/`datanode`/`flownode`).
type ComponentSpec struct {
	// The number of replicas of the components.
	// +required
	// +kubebuilder:validation:Minimum=1
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

	// RPCPort is the gRPC port of the meta.
	// +optional
	RPCPort int32 `json:"rpcPort,omitempty"`

	// HTTPPort is the HTTP port of the meta.
	// +optional
	HTTPPort int32 `json:"httpPort,omitempty"`

	// EtcdEndpoints is the endpoints of the etcd cluster.
	// +required
	EtcdEndpoints []string `json:"etcdEndpoints"`

	// EnableCheckEtcdService indicates whether to check etcd cluster health when starting meta.
	// +optional
	EnableCheckEtcdService bool `json:"enableCheckEtcdService,omitempty"`

	// EnableRegionFailover indicates whether to enable region failover.
	// +optional
	EnableRegionFailover *bool `json:"enableRegionFailover,omitempty"`

	// StoreKeyPrefix is the prefix of the key in the etcd. We can use it to isolate the data of different clusters.
	// +optional
	StoreKeyPrefix string `json:"storeKeyPrefix,omitempty"`
}

func (in *MetaSpec) GetConfig() string {
	if in != nil {
		return in.Config
	}
	return ""
}

func (in *MetaSpec) IsEnableRegionFailover() bool {
	return in != nil && in.EnableRegionFailover != nil && *in.EnableRegionFailover
}

func (in *MetaSpec) GetStoreKeyPrefix() string {
	if in != nil {
		return in.StoreKeyPrefix
	}
	return ""
}

func (in *MetaSpec) GetEtcdEndpoints() []string {
	if in != nil {
		return in.EtcdEndpoints
	}
	return nil
}

func (in *MetaSpec) IsEnableCheckEtcdService() bool {
	return in != nil && in.EnableCheckEtcdService
}

// FrontendSpec is the specification for frontend component.
type FrontendSpec struct {
	ComponentSpec `json:",inline"`

	// Service is the service configuration of the frontend.
	// +optional
	Service *ServiceSpec `json:"service,omitempty"`

	// TLS is the TLS configuration of the frontend.
	// +optional
	TLS *TLSSpec `json:"tls,omitempty"`
}

func (in *FrontendSpec) GetTLS() *TLSSpec {
	if in != nil {
		return in.TLS
	}
	return nil
}

func (in *FrontendSpec) GetService() *ServiceSpec {
	if in != nil {
		return in.Service
	}
	return nil
}

func (in *FrontendSpec) GetConfig() string {
	if in != nil {
		return in.Config
	}
	return ""
}

// DatanodeSpec is the specification for datanode component.
type DatanodeSpec struct {
	ComponentSpec `json:",inline"`

	// RPCPort is the gRPC port of the datanode.
	// +optional
	RPCPort int32 `json:"rpcPort,omitempty"`

	// HTTPPort is the HTTP port of the datanode.
	// +optional
	HTTPPort int32 `json:"httpPort,omitempty"`

	// Storage is the default file storage of the datanode. For example, WAL, cache, index etc.
	// +optional
	Storage *DatanodeStorageSpec `json:"storage,omitempty"`
}

func (in *DatanodeSpec) GetConfig() string {
	if in != nil {
		return in.Config
	}
	return ""
}

func (in *DatanodeSpec) GetFileStorage() *FileStorage {
	if in != nil && in.Storage != nil {
		return in.Storage.FileStorage
	}
	return nil
}

func (in *DatanodeSpec) GetDataHome() string {
	if in != nil && in.Storage != nil {
		return in.Storage.DataHome
	}
	return ""
}

// FlownodeSpec is the specification for flownode component.
type FlownodeSpec struct {
	ComponentSpec `json:",inline"`

	// The gRPC port of the flownode.
	// +optional
	RPCPort int32 `json:"rpcPort,omitempty"`
}

func (in *FlownodeSpec) GetConfig() string {
	if in != nil {
		return in.Config
	}
	return ""
}

// InitializerSpec is the init container to set up components configurations before running the container.
type InitializerSpec struct {
	// The image of the initializer.
	// +optional
	Image string `json:"image,omitempty"`
}

// GreptimeDBClusterSpec defines the desired state of GreptimeDBCluster
type GreptimeDBClusterSpec struct {
	// Base is the base pod template for all components and can be overridden by template of individual component.
	// +optional
	Base *PodTemplateSpec `json:"base,omitempty"`

	// Frontend is the specification of frontend node.
	// +required
	Frontend *FrontendSpec `json:"frontend"`

	// Meta is the specification of meta node.
	// +required
	Meta *MetaSpec `json:"meta"`

	// Datanode is the specification of datanode node.
	// +required
	Datanode *DatanodeSpec `json:"datanode"`

	// Flownode is the specification of flownode node.
	// +optional
	Flownode *FlownodeSpec `json:"flownode,omitempty"`

	// HTTPPort is the HTTP port of the greptimedb cluster.
	// +optional
	HTTPPort int32 `json:"httpPort,omitempty"`

	// RPCPort is the RPC port of the greptimedb cluster.
	// +optional
	RPCPort int32 `json:"rpcPort,omitempty"`

	// MySQLPort is the MySQL port of the greptimedb cluster.
	// +optional
	MySQLPort int32 `json:"mysqlPort,omitempty"`

	// PostgreSQLPort is the PostgreSQL port of the greptimedb cluster.
	// +optional
	PostgreSQLPort int32 `json:"postgreSQLPort,omitempty"`

	// PrometheusMonitor is the specification for creating PodMonitor or ServiceMonitor.
	// +optional
	PrometheusMonitor *PrometheusMonitorSpec `json:"prometheusMonitor,omitempty"`

	// Version is the version of greptimedb.
	// +optional
	Version string `json:"version,omitempty"`

	// Initializer is the init container to set up components configurations before running the container.
	// +optional
	Initializer *InitializerSpec `json:"initializer,omitempty"`

	// ObjectStorageProvider is the storage provider for the greptimedb cluster.
	// +optional
	ObjectStorageProvider *ObjectStorageProviderSpec `json:"objectStorage,omitempty"`

	// WALProvider is the WAL provider for the greptimedb cluster.
	// +optional
	WALProvider *WALProviderSpec `json:"wal,omitempty"`
}

func (in *GreptimeDBCluster) GetBaseMainContainer() *MainContainerSpec {
	if in != nil && in.Spec.Base != nil {
		return in.Spec.Base.MainContainer
	}
	return nil
}

func (in *GreptimeDBCluster) GetVersion() string {
	if in != nil {
		return in.Spec.Version
	}
	return ""
}

func (in *GreptimeDBCluster) GetFrontend() *FrontendSpec {
	if in != nil {
		return in.Spec.Frontend
	}
	return nil
}

func (in *GreptimeDBCluster) GetMeta() *MetaSpec {
	if in != nil {
		return in.Spec.Meta
	}
	return nil
}

func (in *GreptimeDBCluster) GetDatanode() *DatanodeSpec {
	if in != nil {
		return in.Spec.Datanode
	}
	return nil
}

func (in *GreptimeDBCluster) GetFlownode() *FlownodeSpec {
	return in.Spec.Flownode
}

func (in *GreptimeDBCluster) GetWALProvider() *WALProviderSpec {
	if in != nil {
		return in.Spec.WALProvider
	}
	return nil
}

func (in *GreptimeDBCluster) GetWALDir() string {
	if in == nil {
		return ""
	}

	if in.Spec.WALProvider != nil && in.Spec.WALProvider.RaftEngineWAL != nil {
		return in.Spec.WALProvider.RaftEngineWAL.FileStorage.MountPath
	}

	if in.Spec.Datanode != nil &&
		in.Spec.Datanode.Storage != nil &&
		in.Spec.Datanode.Storage.DataHome != "" {
		return in.Spec.Datanode.Storage.DataHome + "/wal"
	}

	return ""
}

func (in *GreptimeDBCluster) GetObjectStorageProvider() *ObjectStorageProviderSpec {
	if in != nil {
		return in.Spec.ObjectStorageProvider
	}
	return nil
}

func (in *GreptimeDBCluster) GetPrometheusMonitor() *PrometheusMonitorSpec {
	if in != nil {
		return in.Spec.PrometheusMonitor
	}
	return nil
}

// GreptimeDBClusterStatus defines the observed state of GreptimeDBCluster
type GreptimeDBClusterStatus struct {
	// Frontend is the status of frontend node.
	// +optional
	Frontend FrontendStatus `json:"frontend,omitempty"`

	// Meta is the status of meta node.
	// +optional
	Meta MetaStatus `json:"meta,omitempty"`

	// Datanode is the status of datanode node.
	// +optional
	Datanode DatanodeStatus `json:"datanode,omitempty"`

	// Flownode is the status of flownode node.
	// +optional
	Flownode FlownodeStatus `json:"flownode,omitempty"`

	// Version is the version of greptimedb.
	// +optional
	Version string `json:"version,omitempty"`

	// ClusterPhase is the phase of the greptimedb cluster.
	// +optional
	ClusterPhase Phase `json:"clusterPhase,omitempty"`

	// Conditions is an array of current conditions.
	// +optional
	Conditions []Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the last observed generation.
	// +optional
	ObservedGeneration *int64 `json:"observedGeneration,omitempty"`
}

// FrontendStatus is the status of frontend node.
type FrontendStatus struct {
	// Replicas is the number of replicas of the frontend.
	Replicas int32 `json:"replicas"`

	// ReadyReplicas is the number of ready replicas of the frontend.
	ReadyReplicas int32 `json:"readyReplicas"`
}

// MetaStatus is the status of meta node.
type MetaStatus struct {
	// Replicas is the number of replicas of the meta.
	Replicas int32 `json:"replicas"`

	// ReadyReplicas is the number of ready replicas of the meta.
	ReadyReplicas int32 `json:"readyReplicas"`

	// EtcdEndpoints is the endpoints of the etcd cluster.
	// +optional
	EtcdEndpoints []string `json:"etcdEndpoints,omitempty"`
}

// DatanodeStatus is the status of datanode node.
type DatanodeStatus struct {
	// Replicas is the number of replicas of the datanode.
	Replicas int32 `json:"replicas"`

	// ReadyReplicas is the number of ready replicas of the datanode.
	ReadyReplicas int32 `json:"readyReplicas"`
}

// FlownodeStatus is the status of flownode node.
type FlownodeStatus struct {
	// Replicas is the number of replicas of the flownode.
	Replicas int32 `json:"replicas"`

	// ReadyReplicas is the number of ready replicas of the flownode.
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
// +kubebuilder:printcolumn:name="FLOWNODE",type="integer",JSONPath=".status.flownode.readyReplicas"
// +kubebuilder:printcolumn:name="PHASE",type=string,JSONPath=".status.clusterPhase"
// +kubebuilder:printcolumn:name="VERSION",type=string,JSONPath=".status.version"
// +kubebuilder:printcolumn:name="AGE",type=date,JSONPath=".metadata.creationTimestamp"

// GreptimeDBCluster is the Schema for the greptimedbclusters API
type GreptimeDBCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec is the specification of the desired state of the GreptimeDBCluster.
	Spec GreptimeDBClusterSpec `json:"spec,omitempty"`

	// Status is the most recently observed status of the GreptimeDBCluster.
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
