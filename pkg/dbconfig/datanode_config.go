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

	"k8s.io/utils/pointer"

	"github.com/GreptimeTeam/greptimedb-operator/apis/v1alpha1"
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
	if cluster.GetS3Storage() != nil {
		s3 := cluster.GetS3Storage()
		c.StorageType = pointer.String("S3")
		c.StorageBucket = pointer.String(s3.Bucket)
		c.StorageRoot = pointer.String(s3.Root)
		c.StorageEndpoint = pointer.String(s3.Endpoint)
		c.StorageRegion = pointer.String(s3.Region)

		if s3.SecretName != "" {
			accessKeyID, secretAccessKey, err := getOCSCredentials(cluster.Namespace, s3.SecretName)
			if err != nil {
				return err
			}
			c.StorageAccessKeyID = pointer.String(string(accessKeyID))
			c.StorageSecretAccessKey = pointer.String(string(secretAccessKey))
		}
	}

	if cluster.GetOSSStorage() != nil {
		oss := cluster.GetOSSStorage()
		c.StorageType = pointer.String("Oss")
		c.StorageBucket = pointer.String(oss.Bucket)
		c.StorageRoot = pointer.String(oss.Root)
		c.StorageEndpoint = pointer.String(oss.Endpoint)
		c.StorageRegion = pointer.String(oss.Region)

		if oss.SecretName != "" {
			accessKeyID, secretAccessKey, err := getOCSCredentials(cluster.Namespace, oss.SecretName)
			if err != nil {
				return err
			}
			c.StorageAccessKeyID = pointer.String(string(accessKeyID))
			c.StorageAccessKeySecret = pointer.String(string(secretAccessKey))
		}
	}

	if cluster.GetGCSStorage() != nil {
		gcs := cluster.GetGCSStorage()
		c.StorageType = pointer.String("Gcs")
		c.StorageBucket = pointer.String(gcs.Bucket)
		c.StorageRoot = pointer.String(gcs.Root)
		c.StorageEndpoint = pointer.String(gcs.Endpoint)
		c.StorageScope = pointer.String(gcs.Scope)

		if gcs.SecretName != "" {
			serviceAccountKey, err := getServiceAccountKey(cluster.Namespace, gcs.SecretName)
			if err != nil {
				return err
			}
			if len(serviceAccountKey) != 0 {
				c.StorageCredential = pointer.String(base64.StdEncoding.EncodeToString(serviceAccountKey))
			}
		}
	}

	// Set the wal dir if the kafka wal is not enabled.
	if cluster.GetKafkaWAL() == nil && cluster.GetWALDir() != "" {
		c.WalDir = pointer.String(cluster.GetWALDir())
	}

	if cluster.GetDataHome() != "" {
		c.StorageDataHome = pointer.String(cluster.GetDataHome())
	}

	if cluster.GetDatanodeConfig() != "" {
		if err := c.SetInputConfig(cluster.Spec.Datanode.Config); err != nil {
			return err
		}
	}

	if cluster.GetKafkaWAL() != nil {
		c.WalProvider = pointer.String("kafka")
		c.WalBrokerEndpoints = cluster.GetKafkaWAL().BrokerEndpoints
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
