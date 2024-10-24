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
	"strings"

	"dario.cat/mergo"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/pointer"
)

// SetDefaults sets the default values for the GreptimeDBCluster.
func (in *GreptimeDBCluster) SetDefaults() error {
	if in == nil {
		return nil
	}

	// Set the version of the GreptimeDBClusterSpec if it is not set.
	in.Spec.Version = getVersionFromImage(in.GetBaseMainContainer().GetImage())

	// Merge the default settings into the GreptimeDBClusterSpec.
	if err := mergo.Merge(&in.Spec, in.defaultSpec()); err != nil {
		return err
	}

	return nil
}

// MergeTemplate merges the base template with the component's template.
func (in *GreptimeDBCluster) MergeTemplate() error {
	mergeFuncs := []func() error{
		in.mergeFrontendTemplate,
		in.mergeMetaTemplate,
		in.mergeDatanodeTemplate,
		in.mergeFlownodeTemplate,
	}

	for _, mergeFunc := range mergeFuncs {
		if err := mergeFunc(); err != nil {
			return err
		}
	}

	return nil
}

// MergeLogging merges the logging settings into the component's logging settings.
func (in *GreptimeDBCluster) MergeLogging() error {
	loggingSpecs := []*LoggingSpec{
		in.GetMeta().GetLogging(),
		in.GetDatanode().GetLogging(),
		in.GetFrontend().GetLogging(),
		in.GetFlownode().GetLogging(),
	}

	for _, logging := range loggingSpecs {
		if logging == nil {
			continue
		}

		if err := in.doMergeLogging(logging, in.GetLogging(), in.GetMonitoring().IsEnabled()); err != nil {
			return err
		}
	}

	return nil
}

func (in *GreptimeDBCluster) doMergeLogging(input, global *LoggingSpec, isEnableMonitoring bool) error {
	if input == nil || global == nil {
		return nil
	}

	if err := mergo.Merge(input, global.DeepCopy()); err != nil {
		return err
	}

	if isEnableMonitoring {
		// Set the default logging format to JSON if monitoring is enabled.
		input.Format = LogFormatJSON
	}

	return nil
}

func (in *GreptimeDBCluster) defaultSpec() *GreptimeDBClusterSpec {
	var defaultSpec = &GreptimeDBClusterSpec{
		Base: &PodTemplateSpec{
			MainContainer: &MainContainerSpec{
				// The default liveness probe for the main container of GreptimeDBCluster.
				LivenessProbe: &corev1.Probe{
					ProbeHandler: corev1.ProbeHandler{
						HTTPGet: &corev1.HTTPGetAction{
							Path: DefaultHealthEndpoint,
							Port: intstr.FromInt32(DefaultHTTPPort),
						},
					},
				},
			},
		},
		Initializer:    &InitializerSpec{Image: DefaultInitializerImage},
		HTTPPort:       DefaultHTTPPort,
		RPCPort:        DefaultRPCPort,
		MySQLPort:      DefaultMySQLPort,
		PostgreSQLPort: DefaultPostgreSQLPort,
		Version:        DefaultVersion,
		Frontend:       in.defaultFrontend(),
		Meta:           in.defaultMeta(),
		Datanode:       in.defaultDatanode(),
	}

	if in.GetFlownode() != nil {
		defaultSpec.Flownode = in.defaultFlownodeSpec()
	}

	defaultSpec.Logging = defaultLogging()

	if in.GetMonitoring().IsEnabled() {
		defaultSpec.Monitoring = &MonitoringSpec{
			Standalone:     in.defaultMonitoringStandaloneSpec(),
			LogsCollection: &LogsCollectionSpec{},
			Vector: &VectorSpec{
				Image: DefaultVectorImage,
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse(DefaultVectorCPURequest),
						corev1.ResourceMemory: resource.MustParse(DefaultVectorMemoryRequest),
					},
					Limits: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse(DefaultVectorCPURequest),
						corev1.ResourceMemory: resource.MustParse(DefaultVectorMemoryRequest),
					},
				},
			},
		}

		// Set the default logging format to JSON if monitoring is enabled.
		defaultSpec.Logging.Format = LogFormatJSON
	}

	return defaultSpec
}

func (in *GreptimeDBCluster) defaultFrontend() *FrontendSpec {
	return &FrontendSpec{
		ComponentSpec: ComponentSpec{
			Template: &PodTemplateSpec{},
			Replicas: pointer.Int32(DefaultReplicas),
			Logging:  &LoggingSpec{},
		},
		RPCPort:        DefaultRPCPort,
		HTTPPort:       DefaultHTTPPort,
		MySQLPort:      DefaultMySQLPort,
		PostgreSQLPort: DefaultPostgreSQLPort,
		Service: &ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
		},
	}
}

func (in *GreptimeDBCluster) defaultMeta() *MetaSpec {
	enableRegionFailover := false
	if in.GetWALProvider().GetKafkaWAL() != nil { // If remote wal provider is enabled, enable region failover by default.
		enableRegionFailover = true
	}
	return &MetaSpec{
		ComponentSpec: ComponentSpec{
			Template: &PodTemplateSpec{},
			Replicas: pointer.Int32(DefaultReplicas),
			Logging:  &LoggingSpec{},
		},
		RPCPort:              DefaultMetaRPCPort,
		HTTPPort:             DefaultHTTPPort,
		EnableRegionFailover: &enableRegionFailover,
	}
}

func (in *GreptimeDBCluster) defaultDatanode() *DatanodeSpec {
	return &DatanodeSpec{
		ComponentSpec: ComponentSpec{
			Template: &PodTemplateSpec{},
			Replicas: pointer.Int32(DefaultReplicas),
			Logging:  &LoggingSpec{},
		},
		RPCPort:  DefaultRPCPort,
		HTTPPort: DefaultHTTPPort,
		Storage:  defaultDatanodeStorage(),
	}
}

func (in *GreptimeDBCluster) defaultFlownodeSpec() *FlownodeSpec {
	return &FlownodeSpec{
		ComponentSpec: ComponentSpec{
			Template: &PodTemplateSpec{},
			Replicas: pointer.Int32(DefaultReplicas),
			Logging:  &LoggingSpec{},
		},
		RPCPort: DefaultRPCPort,
	}
}

func (in *GreptimeDBCluster) defaultMonitoringStandaloneSpec() *GreptimeDBStandaloneSpec {
	standalone := new(GreptimeDBStandalone)
	standalone.Spec = *standalone.defaultSpec()

	if image := in.GetBaseMainContainer().GetImage(); image != "" {
		standalone.Spec.Base.MainContainer.Image = image
	} else {
		standalone.Spec.Base.MainContainer.Image = DefaultGreptimeDBImage
	}

	standalone.Spec.Version = getVersionFromImage(standalone.Spec.Base.MainContainer.Image)

	// For better performance and easy management, the monitoring standalone use file storage by default.
	standalone.Spec.DatanodeStorage = &DatanodeStorageSpec{
		DataHome: DefaultDataHome,
		FileStorage: &FileStorage{
			Name:                DefaultDatanodeFileStorageName,
			StorageSize:         DefaultDataSizeForMonitoring,
			MountPath:           DefaultDataHome,
			StorageRetainPolicy: DefaultStorageRetainPolicyType,
		},
	}

	return &standalone.Spec
}

func (in *GreptimeDBCluster) mergeFrontendTemplate() error {
	if in.Spec.Frontend != nil {
		// Use DeepCopy to avoid the same pointer.
		if err := mergo.Merge(in.Spec.Frontend.Template, in.DeepCopy().Spec.Base); err != nil {
			return err
		}

		// Reconfigure the probe settings based on the HTTP port.
		in.Spec.Frontend.Template.MainContainer.LivenessProbe.HTTPGet.Port = intstr.FromInt32(in.Spec.Frontend.HTTPPort)
	}

	return nil
}

func (in *GreptimeDBCluster) mergeMetaTemplate() error {
	if in.Spec.Meta != nil {
		// Use DeepCopy to avoid the same pointer.
		if err := mergo.Merge(in.Spec.Meta.Template, in.DeepCopy().Spec.Base); err != nil {
			return err
		}

		// Reconfigure the probe settings based on the HTTP port.
		in.Spec.Meta.Template.MainContainer.LivenessProbe.HTTPGet.Port = intstr.FromInt32(in.Spec.Meta.HTTPPort)
	}

	return nil
}

func (in *GreptimeDBCluster) mergeDatanodeTemplate() error {
	if in.Spec.Datanode != nil {
		// Use DeepCopy to avoid the same pointer.
		if err := mergo.Merge(in.Spec.Datanode.Template, in.DeepCopy().Spec.Base); err != nil {
			return err
		}

		// Reconfigure the probe settings based on the HTTP port.
		in.Spec.Datanode.Template.MainContainer.LivenessProbe.HTTPGet.Port = intstr.FromInt32(in.Spec.Datanode.HTTPPort)
	}

	return nil
}

func (in *GreptimeDBCluster) mergeFlownodeTemplate() error {
	if in.Spec.Flownode != nil {
		// Use DeepCopy to avoid the same pointer.
		if err := mergo.Merge(in.Spec.Flownode.Template, in.DeepCopy().Spec.Base); err != nil {
			return err
		}

		// TODO(zyy17): The flownode does not need liveness probe and will be added in the future.
		in.Spec.Flownode.Template.MainContainer.LivenessProbe = nil
	}

	return nil
}

func (in *GreptimeDBStandalone) SetDefaults() error {
	if in == nil {
		return nil
	}

	in.Spec.Version = getVersionFromImage(in.GetBaseMainContainer().GetImage())

	if err := mergo.Merge(&in.Spec, in.defaultSpec()); err != nil {
		return err
	}

	return nil
}

func (in *GreptimeDBStandalone) defaultSpec() *GreptimeDBStandaloneSpec {
	var defaultSpec = &GreptimeDBStandaloneSpec{
		Base: &PodTemplateSpec{
			MainContainer: &MainContainerSpec{
				// The default liveness probe for the main container of GreptimeDBStandalone.
				LivenessProbe: &corev1.Probe{
					ProbeHandler: corev1.ProbeHandler{
						HTTPGet: &corev1.HTTPGetAction{
							Path: DefaultHealthEndpoint,
							Port: intstr.FromInt32(DefaultHTTPPort),
						},
					},
				},
			},
		},
		HTTPPort:       DefaultHTTPPort,
		RPCPort:        DefaultRPCPort,
		MySQLPort:      DefaultMySQLPort,
		PostgreSQLPort: DefaultPostgreSQLPort,
		Version:        DefaultVersion,
		Service: &ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
		},
		Logging: &LoggingSpec{
			Level:              DefaultLoggingLevel,
			LogsDir:            DefaultLogsDir,
			Format:             LogFormatText,
			PersistentWithData: pointer.Bool(false),
			OnlyLogToStdout:    pointer.Bool(false),
		},
		DatanodeStorage: defaultDatanodeStorage(),
	}

	return defaultSpec
}

func defaultDatanodeStorage() *DatanodeStorageSpec {
	return &DatanodeStorageSpec{
		DataHome: DefaultDataHome,
		FileStorage: &FileStorage{
			Name:                DefaultDatanodeFileStorageName,
			StorageSize:         DefaultDataSize,
			MountPath:           DefaultDataHome,
			StorageRetainPolicy: DefaultStorageRetainPolicyType,
		},
	}
}

func defaultLogging() *LoggingSpec {
	return &LoggingSpec{
		Level:              DefaultLoggingLevel,
		LogsDir:            DefaultLogsDir,
		Format:             LogFormatText,
		PersistentWithData: pointer.Bool(false),
		OnlyLogToStdout:    pointer.Bool(false),
	}
}

func getVersionFromImage(imageURL string) string {
	tokens := strings.Split(imageURL, "/")
	if len(tokens) > 0 {
		imageTag := tokens[len(tokens)-1]
		tokens = strings.Split(imageTag, ":")
		if len(tokens) == 2 {
			return tokens[1]
		}
	}
	return DefaultVersion
}
