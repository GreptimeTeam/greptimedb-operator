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

	"github.com/imdario/mergo"
	"google.golang.org/protobuf/proto"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var (
	defaultVersion = "Unknown"

	// The default settings for GreptimeDBClusterSpec.
	defaultHTTPServicePort     = 4000
	defaultGRPCServicePort     = 4001
	defaultMySQLServicePort    = 4002
	defaultPostgresServicePort = 4003
	defaultMetaServicePort     = 3002

	// The default replicas for frontend/meta/datanode.
	defaultFrontendReplicas int32 = 1
	defaultMetaReplicas     int32 = 1
	defaultDatanodeReplicas int32 = 3

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

	var defaultGreptimeDBClusterSpec = &GreptimeDBClusterSpec{
		Base: &PodTemplateSpec{
			MainContainer: &MainContainerSpec{
				// The default readiness probe for the main container of GreptimeDBCluster.
				ReadinessProbe: &corev1.Probe{
					ProbeHandler: corev1.ProbeHandler{
						HTTPGet: &corev1.HTTPGetAction{
							Path: "/health",
							Port: intstr.FromInt(defaultHTTPServicePort),
						},
					},
				},
			},
		},
		Initializer:         &InitializerSpec{Image: defaultInitializer},
		HTTPServicePort:     int32(defaultHTTPServicePort),
		GRPCServicePort:     int32(defaultGRPCServicePort),
		MySQLServicePort:    int32(defaultMySQLServicePort),
		PostgresServicePort: int32(defaultPostgresServicePort),
		Version:             defaultVersion,
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
		defaultGreptimeDBClusterSpec.Meta = &MetaSpec{
			ComponentSpec: ComponentSpec{
				Template: &PodTemplateSpec{},
			},
			ServicePort:          int32(defaultMetaServicePort),
			EnableRegionFailover: false,
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
		}
		if in.Spec.Datanode.Replicas == nil {
			in.Spec.Datanode.Replicas = proto.Int32(defaultDatanodeReplicas)
		}
	}

	if err := mergo.Merge(&in.Spec, defaultGreptimeDBClusterSpec); err != nil {
		return err
	}

	if in.Spec.Frontend != nil {
		if err := mergo.Merge(in.Spec.Frontend.Template, in.Spec.Base); err != nil {
			return err
		}
	}

	if in.Spec.Meta != nil {
		if err := mergo.Merge(in.Spec.Meta.Template, in.Spec.Base); err != nil {
			return err
		}
	}

	if in.Spec.Datanode != nil {
		if err := mergo.Merge(in.Spec.Datanode.Template, in.Spec.Base); err != nil {
			return err
		}
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
				// The default readiness probe for the main container of GreptimeDBCluster.
				ReadinessProbe: &corev1.Probe{
					ProbeHandler: corev1.ProbeHandler{
						HTTPGet: &corev1.HTTPGetAction{
							Path: "/health",
							Port: intstr.FromInt(defaultHTTPServicePort),
						},
					},
				},
			},
		},
		HTTPServicePort:     int32(defaultHTTPServicePort),
		GRPCServicePort:     int32(defaultGRPCServicePort),
		MySQLServicePort:    int32(defaultMySQLServicePort),
		PostgresServicePort: int32(defaultPostgresServicePort),
		Version:             defaultVersion,
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
