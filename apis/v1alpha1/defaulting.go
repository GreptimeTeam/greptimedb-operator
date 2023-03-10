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
	"github.com/imdario/mergo"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

var (
	defaultRequestCPU    = "250m"
	defaultRequestMemory = "64Mi"
	defaultLimitCPU      = "500m"
	defaultLimitMemory   = "128Mi"

	defaultVersion = "v0.1.0"

	// The default settings for GreptimeDBClusterSpec.
	defaultHTTPServicePort     = 4000
	defaultGRPCServicePort     = 4001
	defaultMySQLServicePort    = 4002
	defaultPostgresServicePort = 4003
	defaultOpenTSDBServicePort = 4242
	defaultMetaServicePort     = 3002

	// The default storage settings for datanode.
	defaultDataNodeStorageName      = "datanode"
	defaultDataNodeStorageSize      = "10Gi"
	defaultDataNodeStorageMountPath = "/tmp/greptimedb"
	defaultStorageRetainPolicyType  = RetainStorageRetainPolicyTypeRetain

	defaultCertificateMountPath = "/etc/greptimedb-frontend-tls"
	defaultInitializer          = "greptime/greptimedb-initializer:latest"
)

func (in *GreptimeDBCluster) SetDefaults() error {
	if in == nil {
		return nil
	}

	var defaultGreptimeDBClusterSpec = &GreptimeDBClusterSpec{
		Base: &PodTemplateSpec{
			MainContainer: &MainContainerSpec{
				Resources: &corev1.ResourceRequirements{
					Requests: map[corev1.ResourceName]resource.Quantity{
						"cpu":    resource.MustParse(defaultRequestCPU),
						"memory": resource.MustParse(defaultRequestMemory),
					},
					Limits: map[corev1.ResourceName]resource.Quantity{
						"cpu":    resource.MustParse(defaultLimitCPU),
						"memory": resource.MustParse(defaultLimitMemory),
					},
				},
			},
		},
		Initializer:         &InitializerSpec{Image: defaultInitializer},
		HTTPServicePort:     int32(defaultHTTPServicePort),
		GRPCServicePort:     int32(defaultGRPCServicePort),
		MySQLServicePort:    int32(defaultMySQLServicePort),
		PostgresServicePort: int32(defaultPostgresServicePort),
		OpenTSDBServicePort: int32(defaultOpenTSDBServicePort),
		Version:             defaultVersion,
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

		if in.Spec.Frontend.TLS != nil && len(in.Spec.Frontend.TLS.CertificateMountPath) == 0 {
			in.Spec.Frontend.TLS.CertificateMountPath = defaultCertificateMountPath
		}
	}

	if in.Spec.Meta != nil {
		defaultGreptimeDBClusterSpec.Meta = &MetaSpec{
			ComponentSpec: ComponentSpec{
				Template: &PodTemplateSpec{},
			},
			ServicePort: int32(defaultMetaServicePort),
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
			},
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
