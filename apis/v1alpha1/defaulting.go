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
	"path"
	"strings"

	"dario.cat/mergo"
	"google.golang.org/protobuf/proto"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var (
	defaultVersion = "Unknown"

	defautlHealthEndpoint = "/health"

	// The default settings for GreptimeDBClusterSpec.
	defaultHTTPPort       int32 = 4000
	defaultRPCPort        int32 = 4001
	defaultMySQLPort      int32 = 4002
	defaultPostgreSQLPort int32 = 4003
	defaultMetaRPCPort    int32 = 3002

	// The default replicas for frontend/meta/datanode.
	defaultFrontendReplicas int32 = 1
	defaultMetaReplicas     int32 = 1
	defaultDatanodeReplicas int32 = 1
	defaultFlownodeReplicas int32 = 1

	// The default storage settings for datanode.
	defaultDataNodeStorageName      = "datanode"
	defaultStandaloneStorageName    = "standalone"
	defaultDataNodeStorageSize      = "10Gi"
	defaultDataNodeStorageMountPath = "/data/greptimedb"
	defaultStorageRetainPolicyType  = StorageRetainPolicyTypeRetain
	defaultWalDir                   = path.Join(defaultDataNodeStorageMountPath, "wal")

	defaultInitializer = "greptime/greptimedb-initializer:latest"
)

func (in *GreptimeDBCluster) SetDefaults() error {
	if in == nil {
		return nil
	}

	// Merge the default settings into the GreptimeDBClusterSpec.
	if err := mergo.Merge(&in.Spec, in.defaultClusterSpec()); err != nil {
		return err
	}

	// Merge the Base field into the frontend/meta/datanode/flownode template.
	if err := in.mergeTemplate(); err != nil {
		return err
	}

	return nil
}

func (in *GreptimeDBStandalone) SetDefaults() error {
	if in == nil {
		return nil
	}

	var defaultGreptimeDBStandaloneSpec = &GreptimeDBStandaloneSpec{
		Base: &PodTemplateSpec{
			MainContainer: &MainContainerSpec{
				// The default liveness probe for the main container of GreptimeDBStandalone.
				LivenessProbe: &corev1.Probe{
					ProbeHandler: corev1.ProbeHandler{
						HTTPGet: &corev1.HTTPGetAction{
							Path: defautlHealthEndpoint,
							Port: intstr.FromInt32(defaultHTTPPort),
						},
					},
				},
			},
		},
		HTTPPort:       defaultHTTPPort,
		RPCPort:        defaultRPCPort,
		MySQLPort:      defaultMySQLPort,
		PostgreSQLPort: defaultPostgreSQLPort,
		Version:        defaultVersion,
		LocalStorage: &StorageSpec{
			Name:                defaultStandaloneStorageName,
			StorageSize:         defaultDataNodeStorageSize,
			MountPath:           defaultDataNodeStorageMountPath,
			StorageRetainPolicy: defaultStorageRetainPolicyType,
			WalDir:              defaultWalDir,
			DataHome:            defaultDataNodeStorageMountPath,
		},
		Service: &ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
		},
	}

	if in.Spec.Version == "" &&
		in.Spec.Base != nil &&
		in.Spec.Base.MainContainer != nil &&
		in.Spec.Base.MainContainer.Image != "" {
		in.Spec.Version = getVersionFromImage(in.Spec.Base.MainContainer.Image)
	}

	if err := mergo.Merge(&in.Spec, defaultGreptimeDBStandaloneSpec); err != nil {
		return err
	}

	return nil
}

func (in *GreptimeDBCluster) defaultClusterSpec() *GreptimeDBClusterSpec {
	var defaultGreptimeDBClusterSpec = &GreptimeDBClusterSpec{
		Base: &PodTemplateSpec{
			MainContainer: &MainContainerSpec{
				// The default liveness probe for the main container of GreptimeDBCluster.
				LivenessProbe: &corev1.Probe{
					ProbeHandler: corev1.ProbeHandler{
						HTTPGet: &corev1.HTTPGetAction{
							Path: defautlHealthEndpoint,
							Port: intstr.FromInt32(defaultHTTPPort),
						},
					},
				},
			},
		},
		Initializer:    &InitializerSpec{Image: defaultInitializer},
		HTTPPort:       defaultHTTPPort,
		RPCPort:        defaultRPCPort,
		MySQLPort:      defaultMySQLPort,
		PostgreSQLPort: defaultPostgreSQLPort,
		Version:        defaultVersion,
	}

	if in.Spec.Version == "" &&
		in.Spec.Base != nil &&
		in.Spec.Base.MainContainer != nil &&
		in.Spec.Base.MainContainer.Image != "" {
		in.Spec.Version = getVersionFromImage(in.Spec.Base.MainContainer.Image)
	}

	if in.Spec.Frontend != nil {
		defaultGreptimeDBClusterSpec.Frontend = &FrontendSpec{
			ComponentSpec: ComponentSpec{
				Template: &PodTemplateSpec{},
			},
			Service: ServiceSpec{
				Type: corev1.ServiceTypeClusterIP,
			},
		}
		if in.Spec.Frontend.Replicas == nil {
			in.Spec.Frontend.Replicas = proto.Int32(defaultFrontendReplicas)
		}
	}

	if in.Spec.Meta != nil {
		enableRegionFailover := false
		if in.Spec.RemoteWalProvider != nil { // If remote wal provider is enabled, enable region failover by default.
			enableRegionFailover = true
		}
		defaultGreptimeDBClusterSpec.Meta = &MetaSpec{
			ComponentSpec: ComponentSpec{
				Template: &PodTemplateSpec{},
			},
			RPCPort:              defaultMetaRPCPort,
			HTTPPort:             defaultHTTPPort,
			EnableRegionFailover: &enableRegionFailover,
		}
		if in.Spec.Meta.Replicas == nil {
			in.Spec.Meta.Replicas = proto.Int32(defaultMetaReplicas)
		}
	}

	if in.Spec.Datanode != nil {
		defaultGreptimeDBClusterSpec.Datanode = &DatanodeSpec{
			ComponentSpec: ComponentSpec{
				Template: &PodTemplateSpec{},
			},
			Storage: StorageSpec{
				Name:                defaultDataNodeStorageName,
				StorageSize:         defaultDataNodeStorageSize,
				MountPath:           defaultDataNodeStorageMountPath,
				StorageRetainPolicy: defaultStorageRetainPolicyType,
				WalDir:              defaultWalDir,
				DataHome:            defaultDataNodeStorageMountPath,
			},
			RPCPort:  defaultRPCPort,
			HTTPPort: defaultHTTPPort,
		}
		if in.Spec.Datanode.Replicas == nil {
			in.Spec.Datanode.Replicas = proto.Int32(defaultDatanodeReplicas)
		}
	}

	if in.Spec.Flownode != nil {
		defaultGreptimeDBClusterSpec.Flownode = &FlownodeSpec{
			ComponentSpec: ComponentSpec{
				Template: &PodTemplateSpec{},
			},
			RPCPort: defaultRPCPort,
		}
		if in.Spec.Flownode.Replicas == nil {
			in.Spec.Flownode.Replicas = proto.Int32(defaultFlownodeReplicas)
		}
	}

	return defaultGreptimeDBClusterSpec
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

func getVersionFromImage(imageURL string) string {
	tokens := strings.Split(imageURL, "/")
	if len(tokens) > 0 {
		imageTag := tokens[len(tokens)-1]
		tokens = strings.Split(imageTag, ":")
		if len(tokens) == 2 {
			return tokens[1]
		}
	}
	return defaultVersion
}
