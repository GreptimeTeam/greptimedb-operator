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
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ComponentSpec is the common specification for all components(`frontend`/`meta`/`datanode`/`flownode`).
type ComponentSpec struct {
	// The number of replicas of the components.
	// +optional
	// +kubebuilder:validation:Minimum=0
	Replicas *int32 `json:"replicas,omitempty"`

	// The content of the configuration file of the component in TOML format.
	// +optional
	Config string `json:"config,omitempty"`

	// Template defines the pod template for the component, if not specified, the pod template will use the default value.
	// +optional
	Template *PodTemplateSpec `json:"template,omitempty"`

	// Logging defines the logging configuration for the component.
	// +optional
	Logging *LoggingSpec `json:"logging,omitempty"`

	// Tracing defines the tracing configuration for the component.
	// +optional
	Tracing *TracingSpec `json:"tracing,omitempty"`
}

// MetaSpec is the specification for meta component.
type MetaSpec struct {
	ComponentSpec `json:",inline"`

	// RPCPort is the gRPC port of the meta.
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=65535
	// +optional
	RPCPort int32 `json:"rpcPort,omitempty"`

	// HTTPPort is the HTTP port of the meta.
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=65535
	// +optional
	HTTPPort int32 `json:"httpPort,omitempty"`

	// BackendStorage is the specification for the backend storage for meta.
	// +optional
	BackendStorage *BackendStorage `json:"backendStorage,omitempty"`

	// EnableRegionFailover indicates whether to enable region failover.
	// +optional
	EnableRegionFailover *bool `json:"enableRegionFailover,omitempty"`

	// RollingUpdate is the rolling update configuration. We always use `RollingUpdate` strategyt.
	// +optional
	RollingUpdate *appsv1.RollingUpdateDeployment `json:"rollingUpdate,omitempty"`
}

var _ RoleSpec = &MetaSpec{}

// GetRoleKind returns the role kind.
func (in *MetaSpec) GetRoleKind() RoleKind {
	return MetaRoleKind
}

func (in *MetaSpec) GetName() string {
	return ""
}

func (in *MetaSpec) GetBackendStorage() *BackendStorage {
	if in != nil {
		return in.BackendStorage
	}
	return nil
}

// BackendStorage is the specification for the backend storage for meta.
type BackendStorage struct {
	// EtcdStorage is the specification for etcd storage for meta.
	// +optional
	EtcdStorage *EtcdStorage `json:"etcd,omitempty"`

	// MySQLStorage is the specification for MySQL storage for meta.
	// +optional
	MySQLStorage *MySQLStorage `json:"mysql,omitempty"`

	// PostgreSQLStorage is the specification for PostgreSQL storage for meta.
	// +optional
	PostgreSQLStorage *PostgreSQLStorage `json:"postgresql,omitempty"`
}

func (in *BackendStorage) backendStorageCount() int {
	count := 0
	if in.EtcdStorage != nil {
		count++
	}
	if in.MySQLStorage != nil {
		count++
	}
	if in.PostgreSQLStorage != nil {
		count++
	}
	return count
}

func (in *BackendStorage) GetEtcdStorage() *EtcdStorage {
	if in != nil {
		return in.EtcdStorage
	}
	return nil
}

func (in *BackendStorage) GetMySQLStorage() *MySQLStorage {
	if in != nil {
		return in.MySQLStorage
	}
	return nil
}

func (in *BackendStorage) GetPostgreSQLStorage() *PostgreSQLStorage {
	if in != nil {
		return in.PostgreSQLStorage
	}
	return nil
}

// EtcdStorage is the specification for etcd storage for meta.
type EtcdStorage struct {
	// The endpoints of the etcd cluster.
	// +required
	Endpoints []string `json:"endpoints"`

	// EnableCheckEtcdService indicates whether to check etcd cluster health when starting meta.
	// +optional
	EnableCheckEtcdService bool `json:"enableCheckEtcdService,omitempty"`

	// StoreKeyPrefix is the prefix of the key in the etcd. We can use it to isolate the data of different clusters.
	// +optional
	StoreKeyPrefix string `json:"storeKeyPrefix,omitempty"`
}

func (in *EtcdStorage) GetEndpoints() []string {
	if in != nil {
		return in.Endpoints
	}
	return nil
}

func (in *EtcdStorage) IsEnableCheckEtcdService() bool {
	return in != nil && in.EnableCheckEtcdService
}

func (in *EtcdStorage) GetStoreKeyPrefix() string {
	if in != nil {
		return in.StoreKeyPrefix
	}
	return ""
}

// MySQLStorage is the specification for MySQL storage for meta.
type MySQLStorage struct {
	// Host is the host of the MySQL database.
	// +required
	Host string `json:"host"`

	// Port is the port of the MySQL database.
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=65535
	Port int32 `json:"port"`

	// CredentialsSecretName is the name of the secret that contains the credentials for the MySQL database.
	// The secret must be in the same namespace with the greptime resource.
	// The secret must contain keys named `username` and `password`.
	// +required
	CredentialsSecretName string `json:"credentialsSecretName"`

	// Database is the name of the MySQL database.
	// +optional
	Database string `json:"database,omitempty"`

	// Table is the name of the MySQL table.
	// +optional
	Table string `json:"table,omitempty"`
}

func (in *MySQLStorage) GetCredentialsSecretName() string {
	if in != nil {
		return in.CredentialsSecretName
	}
	return ""
}

// PostgreSQLStorage is the specification for PostgreSQL storage for meta.
type PostgreSQLStorage struct {
	// Host is the host of the PostgreSQL database.
	// +required
	Host string `json:"host"`

	// Port is the port of the PostgreSQL database.
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=65535
	Port int32 `json:"port"`

	// CredentialsSecretName is the name of the secret that contains the credentials for the PostgreSQL database.
	// The secret must be in the same namespace with the greptime resource.
	// The secret must contain keys named `username` and `password`.
	// +required
	CredentialsSecretName string `json:"credentialsSecretName"`

	// Database is the name of the PostgreSQL database.
	// +optional
	Database string `json:"database,omitempty"`

	// Table is the name of the PostgreSQL table.
	// +optional
	Table string `json:"table,omitempty"`

	// ElectionLockID it the lock id in PostgreSQL for election.
	ElectionLockID uint64 `json:"electionLockID,omitempty"`
}

func (in *PostgreSQLStorage) GetCredentialsSecretName() string {
	if in != nil {
		return in.CredentialsSecretName
	}
	return ""
}

func (in *MetaSpec) GetReplicas() *int32 {
	if in != nil && in.Replicas != nil {
		return in.Replicas
	}
	return nil
}

func (in *MetaSpec) GetConfig() string {
	if in != nil {
		return in.Config
	}
	return ""
}

func (in *MetaSpec) GetLogging() *LoggingSpec {
	if in != nil {
		return in.Logging
	}
	return nil
}

func (in *MetaSpec) GetTracing() *TracingSpec {
	if in != nil {
		return in.Tracing
	}
	return nil
}

func (in *MetaSpec) IsEnableRegionFailover() bool {
	return in != nil && in.EnableRegionFailover != nil && *in.EnableRegionFailover
}

// FrontendSpec is the specification for frontend component.
type FrontendSpec struct {
	ComponentSpec `json:",inline"`

	// Name is the name of the frontend.
	// +optional
	Name string `json:"name,omitempty"`

	// RPCPort is the gRPC port of the frontend.
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=65535
	// +optional
	RPCPort int32 `json:"rpcPort,omitempty"`

	// HTTPPort is the HTTP port of the frontend.
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=65535
	// +optional
	HTTPPort int32 `json:"httpPort,omitempty"`

	// MySQLPort is the MySQL port of the frontend.
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=65535
	// +optional
	MySQLPort int32 `json:"mysqlPort,omitempty"`

	// PostgreSQLPort is the PostgreSQL port of the frontend.
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=65535
	// +optional
	PostgreSQLPort int32 `json:"postgreSQLPort,omitempty"`

	// InternalPort is the internal gRPC port of the frontend.
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=65535
	// +optional
	InternalPort *int32 `json:"internalPort,omitempty"`

	// Service is the service configuration of the frontend.
	// +optional
	Service *ServiceSpec `json:"service,omitempty"`

	// TLS is the TLS configuration of the frontend.
	// +optional
	TLS *TLSSpec `json:"tls,omitempty"`

	// RollingUpdate is the rolling update configuration. We always use `RollingUpdate` strategyt.
	// +optional
	RollingUpdate *appsv1.RollingUpdateDeployment `json:"rollingUpdate,omitempty"`

	// SlowQuery is the slow query configuration.
	// +optional
	SlowQuery *SlowQuery `json:"slowQuery,omitempty"`

	// EnableObjectStorage indicates whether to inject object storage configurations into frontend instances.
	// If true, the object storage configurations from the cluster will be injected into the frontend config.
	// Default to false.
	// +optional
	EnableObjectStorage *bool `json:"enableObjectStorage,omitempty"`
}

var _ RoleSpec = &FrontendSpec{}

// GetRoleKind returns the role kind.
func (in *FrontendSpec) GetRoleKind() RoleKind {
	return FrontendRoleKind
}

func (in *FrontendSpec) GetReplicas() *int32 {
	if in != nil && in.Replicas != nil {
		return in.Replicas
	}
	return nil
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

// ShouldInjectObjectStorage returns whether object storage configurations should be injected into frontend instances.
func (in *FrontendSpec) ShouldInjectObjectStorage() bool {
	return in != nil && in.EnableObjectStorage != nil && *in.EnableObjectStorage
}

func (in *FrontendSpec) GetName() string {
	if in != nil {
		return in.Name
	}
	return ""
}

func (in *FrontendSpec) GetLogging() *LoggingSpec {
	if in != nil {
		return in.Logging
	}
	return nil
}

func (in *FrontendSpec) GetTracing() *TracingSpec {
	if in != nil {
		return in.Tracing
	}
	return nil
}

func (in *FrontendSpec) GetSlowQuery() *SlowQuery {
	if in != nil {
		return in.SlowQuery
	}
	return nil
}

// DatanodeSpec is the specification for datanode component.
type DatanodeSpec struct {
	ComponentSpec `json:",inline"`

	// Name is the name of the datanode.
	// +optional
	Name string `json:"name,omitempty"`

	// RPCPort is the gRPC port of the datanode.
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=65535
	// +optional
	RPCPort int32 `json:"rpcPort,omitempty"`

	// HTTPPort is the HTTP port of the datanode.
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=65535
	// +optional
	HTTPPort int32 `json:"httpPort,omitempty"`

	// Storage is the default file storage of the datanode. For example, WAL, cache, index etc.
	// +optional
	Storage *DatanodeStorageSpec `json:"storage,omitempty"`

	// RollingUpdate is the rolling update configuration. We always use `RollingUpdate` strategy.
	// +optional
	RollingUpdate *appsv1.RollingUpdateStatefulSetStrategy `json:"rollingUpdate,omitempty"`

	// StartNodeID is the start node id of the datanode.
	// +optional
	StartNodeID *int32 `json:"startNodeID,omitempty"`
}

var _ RoleSpec = &DatanodeSpec{}

// GetRoleKind returns the role kind.
func (in *DatanodeSpec) GetRoleKind() RoleKind {
	return DatanodeRoleKind
}

func (in *DatanodeSpec) GetName() string {
	if in != nil {
		return in.Name
	}
	return ""
}

func (in *DatanodeSpec) GetReplicas() *int32 {
	if in != nil && in.Replicas != nil {
		return in.Replicas
	}
	return nil
}

func (in *DatanodeSpec) GetConfig() string {
	if in != nil {
		return in.Config
	}
	return ""
}

func (in *DatanodeSpec) GetLogging() *LoggingSpec {
	if in != nil {
		return in.Logging
	}
	return nil
}

func (in *DatanodeSpec) GetTracing() *TracingSpec {
	if in != nil {
		return in.Tracing
	}
	return nil
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

func (in *DatanodeSpec) GetStartNodeID() *int32 {
	if in != nil {
		return in.StartNodeID
	}
	return nil
}

// FlownodeSpec is the specification for flownode component.
type FlownodeSpec struct {
	ComponentSpec `json:",inline"`

	// The gRPC port of the flownode.
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=65535
	// +optional
	RPCPort int32 `json:"rpcPort,omitempty"`

	// The HTTP port of the flownode.
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=65535
	// +optional
	HTTPPort int32 `json:"httpPort,omitempty"`

	// RollingUpdate is the rolling update configuration. We always use `RollingUpdate` strategy.
	// +optional
	RollingUpdate *appsv1.RollingUpdateStatefulSetStrategy `json:"rollingUpdate,omitempty"`

	// StartNodeID is the start node id of the flownode.
	// +optional
	StartNodeID *int32 `json:"startNodeID,omitempty"`
}

var _ RoleSpec = &FlownodeSpec{}

// GetRoleKind returns the role kind.
func (in *FlownodeSpec) GetRoleKind() RoleKind {
	return FlownodeRoleKind
}

func (in *FlownodeSpec) GetName() string {
	return ""
}

func (in *FlownodeSpec) GetReplicas() *int32 {
	if in != nil && in.Replicas != nil {
		return in.Replicas
	}
	return nil
}

func (in *FlownodeSpec) GetConfig() string {
	if in != nil {
		return in.Config
	}
	return ""
}

func (in *FlownodeSpec) GetLogging() *LoggingSpec {
	if in != nil {
		return in.Logging
	}
	return nil
}

func (in *FlownodeSpec) GetTracing() *TracingSpec {
	if in != nil {
		return in.Tracing
	}
	return nil
}

func (in *FlownodeSpec) GetStartNodeID() *int32 {
	if in != nil {
		return in.StartNodeID
	}
	return nil
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
	// +optional
	Frontend *FrontendSpec `json:"frontend,omitempty"`

	// Meta is the specification of meta node.
	// +optional
	Meta *MetaSpec `json:"meta,omitempty"`

	// Datanode is the specification of datanode node.
	// +optional
	Datanode *DatanodeSpec `json:"datanode,omitempty"`

	// DatanodeGroups is a group of datanode statefulsets.
	// +optional
	DatanodeGroups []*DatanodeSpec `json:"datanodeGroups,omitempty"`

	// Flownode is the specification of flownode node.
	// +optional
	Flownode *FlownodeSpec `json:"flownode,omitempty"`

	// FrontendGroups is groups of frontend node.
	// +optional
	FrontendGroups []*FrontendSpec `json:"frontendGroups,omitempty"`

	// HTTPPort is the HTTP port of the greptimedb cluster.
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=65535
	// +optional
	HTTPPort int32 `json:"httpPort,omitempty"`

	// RPCPort is the RPC port of the greptimedb cluster.
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=65535
	// +optional
	RPCPort int32 `json:"rpcPort,omitempty"`

	// MySQLPort is the MySQL port of the greptimedb cluster.
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=65535
	// +optional
	MySQLPort int32 `json:"mysqlPort,omitempty"`

	// PostgreSQLPort is the PostgreSQL port of the greptimedb cluster.
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=65535
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

	// The global logging configuration for all components. It can be overridden by the logging configuration of individual component.
	// +optional
	Logging *LoggingSpec `json:"logging,omitempty"`

	// Monitoring is the specification for monitor bootstrapping. It will create a standalone greptimedb instance to monitor the cluster.
	// +optional
	Monitoring *MonitoringSpec `json:"monitoring,omitempty"`

	// Ingress is the Ingress configuration of the frontend.
	// +optional
	Ingress *IngressSpec `json:"ingress,omitempty"`

	// The global tracing configuration for all components. It can be overridden by the tracing configuration of individual component.
	// +optional
	Tracing *TracingSpec `json:"tracing,omitempty"`

	// ConfigMergeStrategy is the strategy for merging the input config with the config that generated by the operator.
	// +optional
	ConfigMergeStrategy ConfigMergeStrategy `json:"configMergeStrategy,omitempty"`
}

// MonitoringSpec is the specification for monitor bootstrapping. It will create a standalone greptimedb instance to monitor the cluster.
type MonitoringSpec struct {
	// Enabled indicates whether to enable the monitoring service.
	// +required
	Enabled bool `json:"enabled"`

	// The specification of the standalone greptimedb instance.
	// +optional
	Standalone *GreptimeDBStandaloneSpec `json:"standalone,omitempty"`

	// The specification of cluster logs collection.
	// +optional
	LogsCollection *LogsCollectionSpec `json:"logsCollection,omitempty"`

	// The specification of the vector instance.
	// +optional
	Vector *VectorSpec `json:"vector,omitempty"`
}

// LogsCollectionSpec is the specification for cluster logs collection.
type LogsCollectionSpec struct {
	// The specification of the log pipeline.
	// +optional
	Pipeline *LogPipeline `json:"pipeline,omitempty"`
}

func (in *LogsCollectionSpec) GetPipeline() *LogPipeline {
	if in != nil {
		return in.Pipeline
	}
	return nil
}

// LogPipeline is the specification for log pipeline.
type LogPipeline struct {
	// The content of the pipeline configuration file in YAML format.
	// +optional
	Data string `json:"data,omitempty"`
}

func (in *LogPipeline) GetData() string {
	if in != nil {
		return in.Data
	}
	return ""
}

// VectorSpec is the specification for vector instance.
type VectorSpec struct {
	// The image of the vector instance.
	// +optional
	Image string `json:"image,omitempty"`

	// The resources of the vector instance.
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
}

func (in *MonitoringSpec) IsEnabled() bool {
	return in != nil && in.Enabled
}

func (in *MonitoringSpec) GetStandalone() *GreptimeDBStandaloneSpec {
	if in != nil {
		return in.Standalone
	}
	return nil
}

func (in *MonitoringSpec) GetLogsCollection() *LogsCollectionSpec {
	if in != nil {
		return in.LogsCollection
	}
	return nil
}

func (in *MonitoringSpec) GetVector() *VectorSpec {
	if in != nil {
		return in.Vector
	}
	return nil
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

func (in *GreptimeDBCluster) GetLogging() *LoggingSpec {
	if in != nil {
		return in.Spec.Logging
	}
	return nil
}

func (in *GreptimeDBCluster) GetTracing() *TracingSpec {
	if in != nil {
		return in.Spec.Tracing
	}
	return nil
}

func (in *GreptimeDBCluster) GetFrontend() *FrontendSpec {
	if in != nil {
		return in.Spec.Frontend
	}
	return nil
}

func (in *GreptimeDBCluster) GetFrontendGroups() []*FrontendSpec {
	if in != nil {
		return in.Spec.FrontendGroups
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

func (in *GreptimeDBCluster) GetDatanodeGroups() []*DatanodeSpec {
	if in != nil {
		return in.Spec.DatanodeGroups
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

func (in *GreptimeDBCluster) GetMonitoring() *MonitoringSpec {
	if in != nil {
		return in.Spec.Monitoring
	}
	return nil
}

func (in *GreptimeDBCluster) GetIngress() *IngressSpec {
	if in != nil {
		return in.Spec.Ingress
	}
	return nil
}

func (in *GreptimeDBCluster) getAllLoggingSpecs() []*LoggingSpec {
	var specs []*LoggingSpec

	if logging := in.GetMeta().GetLogging(); logging != nil {
		specs = append(specs, logging)
	}

	if logging := in.GetFlownode().GetLogging(); logging != nil {
		specs = append(specs, logging)
	}

	if logging := in.GetDatanode().GetLogging(); logging != nil {
		specs = append(specs, logging)
	}
	for _, datanode := range in.GetDatanodeGroups() {
		if logging := datanode.GetLogging(); logging != nil {
			specs = append(specs, logging)
		}
	}

	if logging := in.GetFrontend().GetLogging(); logging != nil {
		specs = append(specs, logging)
	}
	for _, frontend := range in.GetFrontendGroups() {
		if logging := frontend.GetLogging(); logging != nil {
			specs = append(specs, logging)
		}
	}

	return specs
}

func (in *GreptimeDBCluster) getAllTracingSpecs() []*TracingSpec {
	var specs []*TracingSpec

	if tracing := in.GetMeta().GetTracing(); tracing != nil {
		specs = append(specs, tracing)
	}

	if tracing := in.GetFlownode().GetTracing(); tracing != nil {
		specs = append(specs, tracing)
	}

	if tracing := in.GetDatanode().GetTracing(); tracing != nil {
		specs = append(specs, tracing)
	}

	for _, datanode := range in.GetDatanodeGroups() {
		if tracing := datanode.GetTracing(); tracing != nil {
			specs = append(specs, tracing)
		}
	}

	if tracing := in.GetFrontend().GetTracing(); tracing != nil {
		specs = append(specs, tracing)
	}

	for _, frontend := range in.GetFrontendGroups() {
		if tracing := frontend.GetTracing(); tracing != nil {
			specs = append(specs, tracing)
		}
	}

	return specs
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

	// The status of the monitoring service.
	// +optional
	Monitoring MonitoringStatus `json:"monitoring,omitempty"`

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
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
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

	// MaintenanceMode is the maintenance mode of the meta.
	MaintenanceMode bool `json:"maintenanceMode"`
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

// MonitoringStatus is the status of the monitoring service.
type MonitoringStatus struct {
	// InternalDNSName is the internal DNS name of the monitoring service. For example, 'mycluster-standalone-monitor.default.svc.cluster.local'.
	// +optional
	InternalDNSName string `json:"internalDNSName,omitempty"`
}

func (in *GreptimeDBClusterStatus) GetCondition(conditionType ConditionType) *Condition {
	return GetCondition(in.Conditions, conditionType)
}

func (in *GreptimeDBClusterStatus) SetCondition(condition Condition) {
	in.Conditions = SetCondition(in.Conditions, condition)
}

// +genclient
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

func (in *GreptimeDBCluster) GetDatanodeReplicas() int32 {
	var count int32

	if replicas := in.GetDatanode().GetReplicas(); replicas != nil {
		return *replicas
	}

	for _, datanodeGroup := range in.GetDatanodeGroups() {
		if replicas := datanodeGroup.GetReplicas(); replicas != nil {
			count += *replicas
		}
	}

	return count
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
