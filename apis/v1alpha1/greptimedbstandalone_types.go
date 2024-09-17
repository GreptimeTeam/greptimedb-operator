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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GreptimeDBStandaloneSpec defines the desired state of GreptimeDBStandalone
type GreptimeDBStandaloneSpec struct {
	// Base is the base pod template for all components and can be overridden by template of individual component.
	// +optional
	Base *PodTemplateSpec `json:"base,omitempty"`

	// Service is the service configuration of greptimedb.
	// +optional
	Service *ServiceSpec `json:"service,omitempty"`

	// The TLS configurations of the greptimedb.
	// +optional
	TLS *TLSSpec `json:"tls,omitempty"`

	// HTTPPort is the port of the greptimedb http service.
	// +optional
	HTTPPort int32 `json:"httpPort,omitempty"`

	// RPCPort is the port of the greptimedb rpc service.
	// +optional
	RPCPort int32 `json:"rpcPort,omitempty"`

	// MySQLPort is the port of the greptimedb mysql service.
	// +optional
	MySQLPort int32 `json:"mysqlPort,omitempty"`

	// PostgreSQLPort is the port of the greptimedb postgresql service.
	// +optional
	PostgreSQLPort int32 `json:"postgreSQLPort,omitempty"`

	// PrometheusMonitor is the specification for creating PodMonitor or ServiceMonitor.
	// +optional
	PrometheusMonitor *PrometheusMonitorSpec `json:"prometheusMonitor,omitempty"`

	// Version is the version of the greptimedb.
	// +optional
	Version string `json:"version,omitempty"`

	// Initializer is the init container to set up components configurations before running the container.
	// +optional
	Initializer *InitializerSpec `json:"initializer,omitempty"`

	// ObjectStorageProvider is the storage provider for the greptimedb cluster.
	// +optional
	ObjectStorageProvider *ObjectStorageProviderSpec `json:"objectStorage,omitempty"`

	// DatanodeStorage is the default file storage of the datanode. For example, WAL, cache, index etc.
	// +optional
	DatanodeStorage *DatanodeStorageSpec `json:"datanodeStorage,omitempty"`

	// WALProvider is the WAL provider for the greptimedb cluster.
	// +optional
	WALProvider *WALProviderSpec `json:"wal,omitempty"`

	// The content of the configuration file of the component in TOML format.
	// +optional
	Config string `json:"config,omitempty"`

	// Logging defines the logging configuration for the component.
	// +optional
	Logging *LoggingSpec `json:"logging,omitempty"`
}

// GreptimeDBStandaloneStatus defines the observed state of GreptimeDBStandalone
type GreptimeDBStandaloneStatus struct {
	// Version is the version of the greptimedb.
	// +optional
	Version string `json:"version,omitempty"`

	// StandalonePhase is the phase of the greptimedb standalone.
	// +optional
	StandalonePhase Phase `json:"standalonePhase,omitempty"`

	// Conditions represent the latest available observations of an object's current state.
	// +optional
	Conditions []Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the most recent generation observed for this GreptimeDBStandalone.
	// +optional
	ObservedGeneration *int64 `json:"observedGeneration,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=gts
// +kubebuilder:printcolumn:name="PHASE",type=string,JSONPath=".status.standalonePhase"
// +kubebuilder:printcolumn:name="VERSION",type=string,JSONPath=".status.version"
// +kubebuilder:printcolumn:name="AGE",type=date,JSONPath=".metadata.creationTimestamp"

// GreptimeDBStandalone is the Schema for the greptimedbstandalones API
type GreptimeDBStandalone struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec is the specification of the desired state of the GreptimeDBStandalone.
	Spec GreptimeDBStandaloneSpec `json:"spec,omitempty"`

	// Status is the most recently observed status of the GreptimeDBStandalone.
	Status GreptimeDBStandaloneStatus `json:"status,omitempty"`
}

func (in *GreptimeDBStandalone) GetConfig() string {
	return in.Spec.Config
}

func (in *GreptimeDBStandalone) GetRaftEngineWAL() *RaftEngineWAL {
	if in.Spec.WALProvider != nil {
		return in.Spec.WALProvider.RaftEngineWAL
	}
	return nil
}

func (in *GreptimeDBStandalone) GetRaftEngineWALFileStorage() *FileStorage {
	if in.Spec.WALProvider != nil && in.Spec.WALProvider.RaftEngineWAL != nil {
		return in.Spec.WALProvider.RaftEngineWAL.FileStorage
	}
	return nil
}

func (in *GreptimeDBStandalone) GetKafkaWAL() *KafkaWAL {
	if in.Spec.WALProvider != nil {
		return in.Spec.WALProvider.KafkaWAL
	}
	return nil
}

func (in *GreptimeDBStandalone) GetMainContainer() *MainContainerSpec {
	if in.Spec.Base != nil {
		return in.Spec.Base.MainContainer
	}
	return nil
}

func (in *GreptimeDBStandalone) GetMainContainerImage() string {
	if in.GetMainContainer() != nil {
		return in.GetMainContainer().Image
	}
	return ""
}

func (in *GreptimeDBStandalone) GetVersion() string {
	return in.Spec.Version
}

func (in *GreptimeDBStandalone) GetWALProvider() *WALProviderSpec {
	return in.Spec.WALProvider
}

func (in *GreptimeDBStandalone) GetStorageProvider() *ObjectStorageProviderSpec {
	return in.Spec.ObjectStorageProvider
}

func (in *GreptimeDBStandalone) GetS3Storage() *S3Storage {
	if in.Spec.ObjectStorageProvider != nil {
		return in.Spec.ObjectStorageProvider.S3
	}
	return nil
}

func (in *GreptimeDBStandalone) GetGCSStorage() *GCSStorage {
	if in.Spec.ObjectStorageProvider != nil {
		return in.Spec.ObjectStorageProvider.GCS
	}
	return nil
}

func (in *GreptimeDBStandalone) GetOSSStorage() *OSSStorage {
	if in.Spec.ObjectStorageProvider != nil {
		return in.Spec.ObjectStorageProvider.OSS
	}
	return nil
}

func (in *GreptimeDBStandalone) GetCacheFileStorage() *FileStorage {
	if in.Spec.ObjectStorageProvider != nil && in.Spec.ObjectStorageProvider.Cache != nil {
		return in.Spec.ObjectStorageProvider.Cache.FileStorage
	}
	return nil
}

func (in *GreptimeDBStandalone) GetFrontendTLSSecretName() string {
	if in.Spec.TLS != nil {
		return in.Spec.TLS.SecretName
	}
	return ""
}

func (in *GreptimeDBStandalone) EnablePrometheusMonitor() bool {
	return in.Spec.PrometheusMonitor != nil && in.Spec.PrometheusMonitor.Enabled
}

func (in *GreptimeDBStandalone) GetWALDir() string {
	if in.Spec.WALProvider != nil && in.Spec.WALProvider.RaftEngineWAL != nil {
		return in.Spec.WALProvider.RaftEngineWAL.FileStorage.MountPath
	}
	if in.Spec.DatanodeStorage != nil && in.Spec.DatanodeStorage.DataHome != "" {
		return in.Spec.DatanodeStorage.DataHome + "/wal"
	}
	return ""
}

func (in *GreptimeDBStandalone) GetDatanodeFileStorage() *FileStorage {
	if in.Spec.DatanodeStorage != nil {
		return in.Spec.DatanodeStorage.FileStorage
	}
	return nil
}

func (in *GreptimeDBStandalone) GetDataHome() string {
	if in.Spec.DatanodeStorage != nil {
		return in.Spec.DatanodeStorage.DataHome
	}
	return ""
}

func (in *GreptimeDBStandaloneStatus) GetCondition(conditionType ConditionType) *Condition {
	return GetCondition(in.Conditions, conditionType)
}

func (in *GreptimeDBStandaloneStatus) SetCondition(condition Condition) {
	in.Conditions = SetCondition(in.Conditions, condition)
}

// +kubebuilder:object:root=true

// GreptimeDBStandaloneList contains a list of GreptimeDBStandalone
type GreptimeDBStandaloneList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GreptimeDBStandalone `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GreptimeDBStandalone{}, &GreptimeDBStandaloneList{})
}
