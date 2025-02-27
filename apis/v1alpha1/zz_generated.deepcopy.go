//go:build !ignore_autogenerated

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

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AZBlobStorage) DeepCopyInto(out *AZBlobStorage) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AZBlobStorage.
func (in *AZBlobStorage) DeepCopy() *AZBlobStorage {
	if in == nil {
		return nil
	}
	out := new(AZBlobStorage)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CacheStorage) DeepCopyInto(out *CacheStorage) {
	*out = *in
	if in.FileStorage != nil {
		in, out := &in.FileStorage, &out.FileStorage
		*out = new(FileStorage)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CacheStorage.
func (in *CacheStorage) DeepCopy() *CacheStorage {
	if in == nil {
		return nil
	}
	out := new(CacheStorage)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ComponentSpec) DeepCopyInto(out *ComponentSpec) {
	*out = *in
	if in.Replicas != nil {
		in, out := &in.Replicas, &out.Replicas
		*out = new(int32)
		**out = **in
	}
	if in.Template != nil {
		in, out := &in.Template, &out.Template
		*out = new(PodTemplateSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.Logging != nil {
		in, out := &in.Logging, &out.Logging
		*out = new(LoggingSpec)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ComponentSpec.
func (in *ComponentSpec) DeepCopy() *ComponentSpec {
	if in == nil {
		return nil
	}
	out := new(ComponentSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Condition) DeepCopyInto(out *Condition) {
	*out = *in
	in.LastUpdateTime.DeepCopyInto(&out.LastUpdateTime)
	in.LastTransitionTime.DeepCopyInto(&out.LastTransitionTime)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Condition.
func (in *Condition) DeepCopy() *Condition {
	if in == nil {
		return nil
	}
	out := new(Condition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DatanodeSpec) DeepCopyInto(out *DatanodeSpec) {
	*out = *in
	in.ComponentSpec.DeepCopyInto(&out.ComponentSpec)
	if in.Storage != nil {
		in, out := &in.Storage, &out.Storage
		*out = new(DatanodeStorageSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.RollingUpdate != nil {
		in, out := &in.RollingUpdate, &out.RollingUpdate
		*out = new(appsv1.RollingUpdateStatefulSetStrategy)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DatanodeSpec.
func (in *DatanodeSpec) DeepCopy() *DatanodeSpec {
	if in == nil {
		return nil
	}
	out := new(DatanodeSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DatanodeStatus) DeepCopyInto(out *DatanodeStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DatanodeStatus.
func (in *DatanodeStatus) DeepCopy() *DatanodeStatus {
	if in == nil {
		return nil
	}
	out := new(DatanodeStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DatanodeStorageSpec) DeepCopyInto(out *DatanodeStorageSpec) {
	*out = *in
	if in.FileStorage != nil {
		in, out := &in.FileStorage, &out.FileStorage
		*out = new(FileStorage)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DatanodeStorageSpec.
func (in *DatanodeStorageSpec) DeepCopy() *DatanodeStorageSpec {
	if in == nil {
		return nil
	}
	out := new(DatanodeStorageSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FileStorage) DeepCopyInto(out *FileStorage) {
	*out = *in
	if in.StorageClassName != nil {
		in, out := &in.StorageClassName, &out.StorageClassName
		*out = new(string)
		**out = **in
	}
	if in.Labels != nil {
		in, out := &in.Labels, &out.Labels
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Annotations != nil {
		in, out := &in.Annotations, &out.Annotations
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FileStorage.
func (in *FileStorage) DeepCopy() *FileStorage {
	if in == nil {
		return nil
	}
	out := new(FileStorage)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FlownodeSpec) DeepCopyInto(out *FlownodeSpec) {
	*out = *in
	in.ComponentSpec.DeepCopyInto(&out.ComponentSpec)
	if in.RollingUpdate != nil {
		in, out := &in.RollingUpdate, &out.RollingUpdate
		*out = new(appsv1.RollingUpdateStatefulSetStrategy)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FlownodeSpec.
func (in *FlownodeSpec) DeepCopy() *FlownodeSpec {
	if in == nil {
		return nil
	}
	out := new(FlownodeSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FlownodeStatus) DeepCopyInto(out *FlownodeStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FlownodeStatus.
func (in *FlownodeStatus) DeepCopy() *FlownodeStatus {
	if in == nil {
		return nil
	}
	out := new(FlownodeStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FrontendSpec) DeepCopyInto(out *FrontendSpec) {
	*out = *in
	in.ComponentSpec.DeepCopyInto(&out.ComponentSpec)
	if in.Service != nil {
		in, out := &in.Service, &out.Service
		*out = new(ServiceSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.TLS != nil {
		in, out := &in.TLS, &out.TLS
		*out = new(TLSSpec)
		**out = **in
	}
	if in.RollingUpdate != nil {
		in, out := &in.RollingUpdate, &out.RollingUpdate
		*out = new(appsv1.RollingUpdateDeployment)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FrontendSpec.
func (in *FrontendSpec) DeepCopy() *FrontendSpec {
	if in == nil {
		return nil
	}
	out := new(FrontendSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FrontendStatus) DeepCopyInto(out *FrontendStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FrontendStatus.
func (in *FrontendStatus) DeepCopy() *FrontendStatus {
	if in == nil {
		return nil
	}
	out := new(FrontendStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GCSStorage) DeepCopyInto(out *GCSStorage) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GCSStorage.
func (in *GCSStorage) DeepCopy() *GCSStorage {
	if in == nil {
		return nil
	}
	out := new(GCSStorage)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GreptimeDBCluster) DeepCopyInto(out *GreptimeDBCluster) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GreptimeDBCluster.
func (in *GreptimeDBCluster) DeepCopy() *GreptimeDBCluster {
	if in == nil {
		return nil
	}
	out := new(GreptimeDBCluster)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *GreptimeDBCluster) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GreptimeDBClusterList) DeepCopyInto(out *GreptimeDBClusterList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]GreptimeDBCluster, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GreptimeDBClusterList.
func (in *GreptimeDBClusterList) DeepCopy() *GreptimeDBClusterList {
	if in == nil {
		return nil
	}
	out := new(GreptimeDBClusterList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *GreptimeDBClusterList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GreptimeDBClusterSpec) DeepCopyInto(out *GreptimeDBClusterSpec) {
	*out = *in
	if in.Base != nil {
		in, out := &in.Base, &out.Base
		*out = new(PodTemplateSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.Frontend != nil {
		in, out := &in.Frontend, &out.Frontend
		*out = new(FrontendSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.Meta != nil {
		in, out := &in.Meta, &out.Meta
		*out = new(MetaSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.Datanode != nil {
		in, out := &in.Datanode, &out.Datanode
		*out = new(DatanodeSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.Flownode != nil {
		in, out := &in.Flownode, &out.Flownode
		*out = new(FlownodeSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.FrontendGroup != nil {
		in, out := &in.FrontendGroup, &out.FrontendGroup
		*out = make([]*FrontendSpec, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(FrontendSpec)
				(*in).DeepCopyInto(*out)
			}
		}
	}
	if in.PrometheusMonitor != nil {
		in, out := &in.PrometheusMonitor, &out.PrometheusMonitor
		*out = new(PrometheusMonitorSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.Initializer != nil {
		in, out := &in.Initializer, &out.Initializer
		*out = new(InitializerSpec)
		**out = **in
	}
	if in.ObjectStorageProvider != nil {
		in, out := &in.ObjectStorageProvider, &out.ObjectStorageProvider
		*out = new(ObjectStorageProviderSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.WALProvider != nil {
		in, out := &in.WALProvider, &out.WALProvider
		*out = new(WALProviderSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.Logging != nil {
		in, out := &in.Logging, &out.Logging
		*out = new(LoggingSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.Monitoring != nil {
		in, out := &in.Monitoring, &out.Monitoring
		*out = new(MonitoringSpec)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GreptimeDBClusterSpec.
func (in *GreptimeDBClusterSpec) DeepCopy() *GreptimeDBClusterSpec {
	if in == nil {
		return nil
	}
	out := new(GreptimeDBClusterSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GreptimeDBClusterStatus) DeepCopyInto(out *GreptimeDBClusterStatus) {
	*out = *in
	out.Frontend = in.Frontend
	in.Meta.DeepCopyInto(&out.Meta)
	out.Datanode = in.Datanode
	out.Flownode = in.Flownode
	out.Monitoring = in.Monitoring
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.ObservedGeneration != nil {
		in, out := &in.ObservedGeneration, &out.ObservedGeneration
		*out = new(int64)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GreptimeDBClusterStatus.
func (in *GreptimeDBClusterStatus) DeepCopy() *GreptimeDBClusterStatus {
	if in == nil {
		return nil
	}
	out := new(GreptimeDBClusterStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GreptimeDBStandalone) DeepCopyInto(out *GreptimeDBStandalone) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GreptimeDBStandalone.
func (in *GreptimeDBStandalone) DeepCopy() *GreptimeDBStandalone {
	if in == nil {
		return nil
	}
	out := new(GreptimeDBStandalone)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *GreptimeDBStandalone) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GreptimeDBStandaloneList) DeepCopyInto(out *GreptimeDBStandaloneList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]GreptimeDBStandalone, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GreptimeDBStandaloneList.
func (in *GreptimeDBStandaloneList) DeepCopy() *GreptimeDBStandaloneList {
	if in == nil {
		return nil
	}
	out := new(GreptimeDBStandaloneList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *GreptimeDBStandaloneList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GreptimeDBStandaloneSpec) DeepCopyInto(out *GreptimeDBStandaloneSpec) {
	*out = *in
	if in.Base != nil {
		in, out := &in.Base, &out.Base
		*out = new(PodTemplateSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.Service != nil {
		in, out := &in.Service, &out.Service
		*out = new(ServiceSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.TLS != nil {
		in, out := &in.TLS, &out.TLS
		*out = new(TLSSpec)
		**out = **in
	}
	if in.PrometheusMonitor != nil {
		in, out := &in.PrometheusMonitor, &out.PrometheusMonitor
		*out = new(PrometheusMonitorSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.Initializer != nil {
		in, out := &in.Initializer, &out.Initializer
		*out = new(InitializerSpec)
		**out = **in
	}
	if in.ObjectStorageProvider != nil {
		in, out := &in.ObjectStorageProvider, &out.ObjectStorageProvider
		*out = new(ObjectStorageProviderSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.DatanodeStorage != nil {
		in, out := &in.DatanodeStorage, &out.DatanodeStorage
		*out = new(DatanodeStorageSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.WALProvider != nil {
		in, out := &in.WALProvider, &out.WALProvider
		*out = new(WALProviderSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.Logging != nil {
		in, out := &in.Logging, &out.Logging
		*out = new(LoggingSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.RollingUpdate != nil {
		in, out := &in.RollingUpdate, &out.RollingUpdate
		*out = new(appsv1.RollingUpdateStatefulSetStrategy)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GreptimeDBStandaloneSpec.
func (in *GreptimeDBStandaloneSpec) DeepCopy() *GreptimeDBStandaloneSpec {
	if in == nil {
		return nil
	}
	out := new(GreptimeDBStandaloneSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GreptimeDBStandaloneStatus) DeepCopyInto(out *GreptimeDBStandaloneStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.ObservedGeneration != nil {
		in, out := &in.ObservedGeneration, &out.ObservedGeneration
		*out = new(int64)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GreptimeDBStandaloneStatus.
func (in *GreptimeDBStandaloneStatus) DeepCopy() *GreptimeDBStandaloneStatus {
	if in == nil {
		return nil
	}
	out := new(GreptimeDBStandaloneStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *InitializerSpec) DeepCopyInto(out *InitializerSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new InitializerSpec.
func (in *InitializerSpec) DeepCopy() *InitializerSpec {
	if in == nil {
		return nil
	}
	out := new(InitializerSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KafkaWAL) DeepCopyInto(out *KafkaWAL) {
	*out = *in
	if in.BrokerEndpoints != nil {
		in, out := &in.BrokerEndpoints, &out.BrokerEndpoints
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KafkaWAL.
func (in *KafkaWAL) DeepCopy() *KafkaWAL {
	if in == nil {
		return nil
	}
	out := new(KafkaWAL)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LogPipeline) DeepCopyInto(out *LogPipeline) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LogPipeline.
func (in *LogPipeline) DeepCopy() *LogPipeline {
	if in == nil {
		return nil
	}
	out := new(LogPipeline)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LoggingSpec) DeepCopyInto(out *LoggingSpec) {
	*out = *in
	if in.Filters != nil {
		in, out := &in.Filters, &out.Filters
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.PersistentWithData != nil {
		in, out := &in.PersistentWithData, &out.PersistentWithData
		*out = new(bool)
		**out = **in
	}
	if in.OnlyLogToStdout != nil {
		in, out := &in.OnlyLogToStdout, &out.OnlyLogToStdout
		*out = new(bool)
		**out = **in
	}
	if in.SlowQuery != nil {
		in, out := &in.SlowQuery, &out.SlowQuery
		*out = new(SlowQuery)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LoggingSpec.
func (in *LoggingSpec) DeepCopy() *LoggingSpec {
	if in == nil {
		return nil
	}
	out := new(LoggingSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LogsCollectionSpec) DeepCopyInto(out *LogsCollectionSpec) {
	*out = *in
	if in.Pipeline != nil {
		in, out := &in.Pipeline, &out.Pipeline
		*out = new(LogPipeline)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LogsCollectionSpec.
func (in *LogsCollectionSpec) DeepCopy() *LogsCollectionSpec {
	if in == nil {
		return nil
	}
	out := new(LogsCollectionSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MainContainerSpec) DeepCopyInto(out *MainContainerSpec) {
	*out = *in
	in.Resources.DeepCopyInto(&out.Resources)
	if in.Command != nil {
		in, out := &in.Command, &out.Command
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Args != nil {
		in, out := &in.Args, &out.Args
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Env != nil {
		in, out := &in.Env, &out.Env
		*out = make([]v1.EnvVar, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.LivenessProbe != nil {
		in, out := &in.LivenessProbe, &out.LivenessProbe
		*out = new(v1.Probe)
		(*in).DeepCopyInto(*out)
	}
	if in.ReadinessProbe != nil {
		in, out := &in.ReadinessProbe, &out.ReadinessProbe
		*out = new(v1.Probe)
		(*in).DeepCopyInto(*out)
	}
	if in.StartupProbe != nil {
		in, out := &in.StartupProbe, &out.StartupProbe
		*out = new(v1.Probe)
		(*in).DeepCopyInto(*out)
	}
	if in.Lifecycle != nil {
		in, out := &in.Lifecycle, &out.Lifecycle
		*out = new(v1.Lifecycle)
		(*in).DeepCopyInto(*out)
	}
	if in.VolumeMounts != nil {
		in, out := &in.VolumeMounts, &out.VolumeMounts
		*out = make([]v1.VolumeMount, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.SecurityContext != nil {
		in, out := &in.SecurityContext, &out.SecurityContext
		*out = new(v1.SecurityContext)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MainContainerSpec.
func (in *MainContainerSpec) DeepCopy() *MainContainerSpec {
	if in == nil {
		return nil
	}
	out := new(MainContainerSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MetaSpec) DeepCopyInto(out *MetaSpec) {
	*out = *in
	in.ComponentSpec.DeepCopyInto(&out.ComponentSpec)
	if in.EtcdEndpoints != nil {
		in, out := &in.EtcdEndpoints, &out.EtcdEndpoints
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.EnableRegionFailover != nil {
		in, out := &in.EnableRegionFailover, &out.EnableRegionFailover
		*out = new(bool)
		**out = **in
	}
	if in.RollingUpdate != nil {
		in, out := &in.RollingUpdate, &out.RollingUpdate
		*out = new(appsv1.RollingUpdateDeployment)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MetaSpec.
func (in *MetaSpec) DeepCopy() *MetaSpec {
	if in == nil {
		return nil
	}
	out := new(MetaSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MetaStatus) DeepCopyInto(out *MetaStatus) {
	*out = *in
	if in.EtcdEndpoints != nil {
		in, out := &in.EtcdEndpoints, &out.EtcdEndpoints
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MetaStatus.
func (in *MetaStatus) DeepCopy() *MetaStatus {
	if in == nil {
		return nil
	}
	out := new(MetaStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MonitoringSpec) DeepCopyInto(out *MonitoringSpec) {
	*out = *in
	if in.Standalone != nil {
		in, out := &in.Standalone, &out.Standalone
		*out = new(GreptimeDBStandaloneSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.LogsCollection != nil {
		in, out := &in.LogsCollection, &out.LogsCollection
		*out = new(LogsCollectionSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.Vector != nil {
		in, out := &in.Vector, &out.Vector
		*out = new(VectorSpec)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MonitoringSpec.
func (in *MonitoringSpec) DeepCopy() *MonitoringSpec {
	if in == nil {
		return nil
	}
	out := new(MonitoringSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MonitoringStatus) DeepCopyInto(out *MonitoringStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MonitoringStatus.
func (in *MonitoringStatus) DeepCopy() *MonitoringStatus {
	if in == nil {
		return nil
	}
	out := new(MonitoringStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OSSStorage) DeepCopyInto(out *OSSStorage) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OSSStorage.
func (in *OSSStorage) DeepCopy() *OSSStorage {
	if in == nil {
		return nil
	}
	out := new(OSSStorage)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ObjectStorageProviderSpec) DeepCopyInto(out *ObjectStorageProviderSpec) {
	*out = *in
	if in.S3 != nil {
		in, out := &in.S3, &out.S3
		*out = new(S3Storage)
		**out = **in
	}
	if in.OSS != nil {
		in, out := &in.OSS, &out.OSS
		*out = new(OSSStorage)
		**out = **in
	}
	if in.GCS != nil {
		in, out := &in.GCS, &out.GCS
		*out = new(GCSStorage)
		**out = **in
	}
	if in.AZBlob != nil {
		in, out := &in.AZBlob, &out.AZBlob
		*out = new(AZBlobStorage)
		**out = **in
	}
	if in.Cache != nil {
		in, out := &in.Cache, &out.Cache
		*out = new(CacheStorage)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ObjectStorageProviderSpec.
func (in *ObjectStorageProviderSpec) DeepCopy() *ObjectStorageProviderSpec {
	if in == nil {
		return nil
	}
	out := new(ObjectStorageProviderSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PodTemplateSpec) DeepCopyInto(out *PodTemplateSpec) {
	*out = *in
	if in.Annotations != nil {
		in, out := &in.Annotations, &out.Annotations
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Labels != nil {
		in, out := &in.Labels, &out.Labels
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.MainContainer != nil {
		in, out := &in.MainContainer, &out.MainContainer
		*out = new(MainContainerSpec)
		(*in).DeepCopyInto(*out)
	}
	in.SlimPodSpec.DeepCopyInto(&out.SlimPodSpec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PodTemplateSpec.
func (in *PodTemplateSpec) DeepCopy() *PodTemplateSpec {
	if in == nil {
		return nil
	}
	out := new(PodTemplateSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PrometheusMonitorSpec) DeepCopyInto(out *PrometheusMonitorSpec) {
	*out = *in
	if in.Labels != nil {
		in, out := &in.Labels, &out.Labels
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PrometheusMonitorSpec.
func (in *PrometheusMonitorSpec) DeepCopy() *PrometheusMonitorSpec {
	if in == nil {
		return nil
	}
	out := new(PrometheusMonitorSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RaftEngineWAL) DeepCopyInto(out *RaftEngineWAL) {
	*out = *in
	if in.FileStorage != nil {
		in, out := &in.FileStorage, &out.FileStorage
		*out = new(FileStorage)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RaftEngineWAL.
func (in *RaftEngineWAL) DeepCopy() *RaftEngineWAL {
	if in == nil {
		return nil
	}
	out := new(RaftEngineWAL)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *S3Storage) DeepCopyInto(out *S3Storage) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new S3Storage.
func (in *S3Storage) DeepCopy() *S3Storage {
	if in == nil {
		return nil
	}
	out := new(S3Storage)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServiceSpec) DeepCopyInto(out *ServiceSpec) {
	*out = *in
	if in.Annotations != nil {
		in, out := &in.Annotations, &out.Annotations
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Labels != nil {
		in, out := &in.Labels, &out.Labels
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.LoadBalancerClass != nil {
		in, out := &in.LoadBalancerClass, &out.LoadBalancerClass
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServiceSpec.
func (in *ServiceSpec) DeepCopy() *ServiceSpec {
	if in == nil {
		return nil
	}
	out := new(ServiceSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SlimPodSpec) DeepCopyInto(out *SlimPodSpec) {
	*out = *in
	if in.NodeSelector != nil {
		in, out := &in.NodeSelector, &out.NodeSelector
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.InitContainers != nil {
		in, out := &in.InitContainers, &out.InitContainers
		*out = make([]v1.Container, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.TerminationGracePeriodSeconds != nil {
		in, out := &in.TerminationGracePeriodSeconds, &out.TerminationGracePeriodSeconds
		*out = new(int64)
		**out = **in
	}
	if in.ActiveDeadlineSeconds != nil {
		in, out := &in.ActiveDeadlineSeconds, &out.ActiveDeadlineSeconds
		*out = new(int64)
		**out = **in
	}
	if in.ImagePullSecrets != nil {
		in, out := &in.ImagePullSecrets, &out.ImagePullSecrets
		*out = make([]v1.LocalObjectReference, len(*in))
		copy(*out, *in)
	}
	if in.Affinity != nil {
		in, out := &in.Affinity, &out.Affinity
		*out = new(v1.Affinity)
		(*in).DeepCopyInto(*out)
	}
	if in.Tolerations != nil {
		in, out := &in.Tolerations, &out.Tolerations
		*out = make([]v1.Toleration, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.AdditionalContainers != nil {
		in, out := &in.AdditionalContainers, &out.AdditionalContainers
		*out = make([]v1.Container, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Volumes != nil {
		in, out := &in.Volumes, &out.Volumes
		*out = make([]v1.Volume, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.SecurityContext != nil {
		in, out := &in.SecurityContext, &out.SecurityContext
		*out = new(v1.PodSecurityContext)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SlimPodSpec.
func (in *SlimPodSpec) DeepCopy() *SlimPodSpec {
	if in == nil {
		return nil
	}
	out := new(SlimPodSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SlowQuery) DeepCopyInto(out *SlowQuery) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SlowQuery.
func (in *SlowQuery) DeepCopy() *SlowQuery {
	if in == nil {
		return nil
	}
	out := new(SlowQuery)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TLSSpec) DeepCopyInto(out *TLSSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TLSSpec.
func (in *TLSSpec) DeepCopy() *TLSSpec {
	if in == nil {
		return nil
	}
	out := new(TLSSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VectorSpec) DeepCopyInto(out *VectorSpec) {
	*out = *in
	in.Resources.DeepCopyInto(&out.Resources)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VectorSpec.
func (in *VectorSpec) DeepCopy() *VectorSpec {
	if in == nil {
		return nil
	}
	out := new(VectorSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *WALProviderSpec) DeepCopyInto(out *WALProviderSpec) {
	*out = *in
	if in.RaftEngineWAL != nil {
		in, out := &in.RaftEngineWAL, &out.RaftEngineWAL
		*out = new(RaftEngineWAL)
		(*in).DeepCopyInto(*out)
	}
	if in.KafkaWAL != nil {
		in, out := &in.KafkaWAL, &out.KafkaWAL
		*out = new(KafkaWAL)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WALProviderSpec.
func (in *WALProviderSpec) DeepCopy() *WALProviderSpec {
	if in == nil {
		return nil
	}
	out := new(WALProviderSpec)
	in.DeepCopyInto(out)
	return out
}
