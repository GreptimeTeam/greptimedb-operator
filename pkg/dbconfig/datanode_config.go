// Copyright 2023 Greptime Team
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

package dbconfig

import (
	"encoding/base64"

	"github.com/GreptimeTeam/greptimedb-operator/apis/v1alpha1"
	"github.com/GreptimeTeam/greptimedb-operator/pkg/util"
)

var _ Config = &DatanodeConfig{}

// DatanodeConfig is the configuration for the datanode.
type DatanodeConfig struct {
	NodeID      *uint64 `tomlmapping:"node_id"`
	RPCAddr     *string `tomlmapping:"rpc_addr"`
	RPCHostName *string `tomlmapping:"rpc_hostname"`

	// Storage options.
	StorageType            *string `tomlmapping:"storage.type"`
	StorageDataHome        *string `tomlmapping:"storage.data_home"`
	StorageAccessKeyID     *string `tomlmapping:"storage.access_key_id"`
	StorageSecretAccessKey *string `tomlmapping:"storage.secret_access_key"`
	StorageAccessKeySecret *string `tomlmapping:"storage.access_key_secret"`
	StorageBucket          *string `tomlmapping:"storage.bucket"`
	StorageRoot            *string `tomlmapping:"storage.root"`
	StorageRegion          *string `tomlmapping:"storage.region"`
	StorageEndpoint        *string `tomlmapping:"storage.endpoint"`
	StorageScope           *string `tomlmapping:"storage.scope"`
	StorageCredential      *string `tomlmapping:"storage.credential"`

	// The wal file directory.
	WalDir *string `tomlmapping:"wal.dir"`

	// The wal provider.
	WalProvider *string `tomlmapping:"wal.provider"`

	// The kafka broker endpoints.
	WalBrokerEndpoints []string `tomlmapping:"wal.broker_endpoints"`

	// InputConfig is from config field of cluster spec.
	InputConfig string
}

// ConfigureByCluster configures the datanode config by the given cluster.
func (c *DatanodeConfig) ConfigureByCluster(cluster *v1alpha1.GreptimeDBCluster) error {
	// TODO(zyy17): need to refactor the following code. It's too ugly.
	if cluster.Spec.ObjectStorageProvider != nil {
		switch {
		case cluster.Spec.ObjectStorageProvider.S3 != nil:
			if cluster.Spec.ObjectStorageProvider.S3.SecretName != "" {
				accessKeyID, secretAccessKey, err := getOCSCredentials(cluster.Namespace, cluster.Spec.ObjectStorageProvider.S3.SecretName)
				if err != nil {
					return err
				}
				c.StorageAccessKeyID = util.StringPtr(string(accessKeyID))
				c.StorageSecretAccessKey = util.StringPtr(string(secretAccessKey))
			}
			c.StorageType = util.StringPtr("S3")
			c.StorageBucket = util.StringPtr(cluster.Spec.ObjectStorageProvider.S3.Bucket)
			c.StorageRoot = util.StringPtr(cluster.Spec.ObjectStorageProvider.S3.Root)
			c.StorageEndpoint = util.StringPtr(cluster.Spec.ObjectStorageProvider.S3.Endpoint)
			c.StorageRegion = util.StringPtr(cluster.Spec.ObjectStorageProvider.S3.Region)

		case cluster.Spec.ObjectStorageProvider.OSS != nil:
			if cluster.Spec.ObjectStorageProvider.OSS.SecretName != "" {
				accessKeyID, secretAccessKey, err := getOCSCredentials(cluster.Namespace, cluster.Spec.ObjectStorageProvider.OSS.SecretName)
				if err != nil {
					return err
				}
				c.StorageAccessKeyID = util.StringPtr(string(accessKeyID))
				c.StorageAccessKeySecret = util.StringPtr(string(secretAccessKey))
			}
			c.StorageType = util.StringPtr("Oss")
			c.StorageBucket = util.StringPtr(cluster.Spec.ObjectStorageProvider.OSS.Bucket)
			c.StorageRoot = util.StringPtr(cluster.Spec.ObjectStorageProvider.OSS.Root)
			c.StorageEndpoint = util.StringPtr(cluster.Spec.ObjectStorageProvider.OSS.Endpoint)
			c.StorageRegion = util.StringPtr(cluster.Spec.ObjectStorageProvider.OSS.Region)

		case cluster.Spec.ObjectStorageProvider.GCS != nil:
			if cluster.Spec.ObjectStorageProvider.GCS.SecretName != "" {
				serviceAccountKey, err := getServiceAccountKey(cluster.Namespace, cluster.Spec.ObjectStorageProvider.GCS.SecretName)
				if err != nil {
					return err
				}
				if len(serviceAccountKey) != 0 {
					c.StorageCredential = util.StringPtr(base64.StdEncoding.EncodeToString(serviceAccountKey))
				}
			}
			c.StorageType = util.StringPtr("Gcs")
			c.StorageBucket = util.StringPtr(cluster.Spec.ObjectStorageProvider.GCS.Bucket)
			c.StorageRoot = util.StringPtr(cluster.Spec.ObjectStorageProvider.GCS.Root)
			c.StorageEndpoint = util.StringPtr(cluster.Spec.ObjectStorageProvider.GCS.Endpoint)
			c.StorageScope = util.StringPtr(cluster.Spec.ObjectStorageProvider.GCS.Scope)
		}
	}

	// Set the wal dir if the kafka wal is not enabled.
	if cluster.GetWALProvider().GetKafkaWAL() == nil && cluster.GetWALDir() != "" {
		c.WalDir = util.StringPtr(cluster.GetWALDir())
	}

	if cluster.GetDatanode().GetDataHome() != "" {
		c.StorageDataHome = util.StringPtr(cluster.GetDatanode().GetDataHome())
	}

	if cluster.GetDatanode().GetConfig() != "" {
		if err := c.SetInputConfig(cluster.GetDatanode().GetConfig()); err != nil {
			return err
		}
	}

	if cluster.GetWALProvider().GetKafkaWAL() != nil {
		c.WalProvider = util.StringPtr("kafka")
		c.WalBrokerEndpoints = cluster.GetWALProvider().GetKafkaWAL().GetBrokerEndpoints()
	}

	return nil
}

// ConfigureByStandalone is not need to implement in cluster mode.
func (c *DatanodeConfig) ConfigureByStandalone(_ *v1alpha1.GreptimeDBStandalone) error {
	return nil
}

// Kind returns the component kind of the datanode.
func (c *DatanodeConfig) Kind() v1alpha1.ComponentKind {
	return v1alpha1.DatanodeComponentKind
}

// GetInputConfig returns the input config.
func (c *DatanodeConfig) GetInputConfig() string {
	return c.InputConfig
}

// SetInputConfig sets the input config.
func (c *DatanodeConfig) SetInputConfig(input string) error {
	c.InputConfig = input
	return nil
}
