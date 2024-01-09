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

	// +optional
	Service *ServiceSpec `json:"service,omitempty"`

	// The TLS configurations of the greptimedb.
	// +optional
	TLS *TLSSpec `json:"tls,omitempty"`

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
	LocalStorage *StorageSpec `json:"localStorage,omitempty"`

	// +optional
	Config string `json:"config,omitempty"`

	// +optional
	// Attention: This option need to be supported by the Reloader(https://github.com/stakater/Reloader).
	// We may consider to implement the same feature inside the operator in the future.
	ReloadWhenConfigChange bool `json:"reloadWhenConfigChange,omitempty"`
}

// GreptimeDBStandaloneStatus defines the observed state of GreptimeDBStandalone
type GreptimeDBStandaloneStatus struct {
	Version string `json:"version,omitempty"`

	// +optional
	StandalonePhase Phase `json:"standalonePhase,omitempty"`

	// +optional
	Conditions []Condition `json:"conditions,omitempty"`
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

	Spec   GreptimeDBStandaloneSpec   `json:"spec,omitempty"`
	Status GreptimeDBStandaloneStatus `json:"status,omitempty"`
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
