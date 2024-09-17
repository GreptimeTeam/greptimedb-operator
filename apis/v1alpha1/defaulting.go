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
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/pointer"
)

func (in *GreptimeDBCluster) SetDefaults() error {
	if in == nil {
		return nil
	}

	// Set the version of the GreptimeDBClusterSpec if it is not set.
	if in.GetVersion() == "" && in.GetBaseMainContainer().GetImage() != "" {
		in.Spec.Version = getVersionFromImage(in.GetBaseMainContainer().GetImage())
	}

	// Merge the default settings into the GreptimeDBClusterSpec.
	if err := mergo.Merge(&in.Spec, in.defaultSpec()); err != nil {
		return err
	}

	// Merge the Base field into the frontend/meta/datanode/flownode template.
	if err := in.mergeTemplate(); err != nil {
		return err
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
							Path: DefautlHealthEndpoint,
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

	return defaultSpec
}

func (in *GreptimeDBCluster) defaultFrontend() *FrontendSpec {
	return &FrontendSpec{
		ComponentSpec: ComponentSpec{
			Template: &PodTemplateSpec{},
			Replicas: pointer.Int32(DefaultReplicas),
			Logging:  defaultLogging(),
		},
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
			Logging:  defaultLogging(),
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
			Logging:  defaultLogging(),
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
			Logging:  defaultLogging(),
		},
		RPCPort: DefaultRPCPort,
	}
}

func (in *GreptimeDBCluster) mergeTemplate() error {
	if err := in.mergeFrontendTemplate(); err != nil {
		return err
	}

	if err := in.mergeMetaTemplate(); err != nil {
		return err
	}

	if err := in.mergeDatanodeTemplate(); err != nil {
		return err
	}

	if err := in.mergeFlownodeTemplate(); err != nil {
		return err
	}

	return nil
}

func (in *GreptimeDBCluster) mergeFrontendTemplate() error {
	if in.Spec.Frontend != nil {
		// Use DeepCopy to avoid the same pointer.
		if err := mergo.Merge(in.Spec.Frontend.Template, in.DeepCopy().Spec.Base); err != nil {
			return err
		}

		// Reconfigure the probe settings based on the HTTP port.
		in.Spec.Frontend.Template.MainContainer.LivenessProbe.HTTPGet.Port = intstr.FromInt32(in.Spec.HTTPPort)
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

	if in.GetVersion() == "" && in.GetBaseMainContainer().GetImage() != "" {
		in.Spec.Version = getVersionFromImage(in.GetBaseMainContainer().GetImage())
	}

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
							Path: DefautlHealthEndpoint,
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
			Level:              DefaultLogingLevel,
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
		Level:              DefaultLogingLevel,
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
