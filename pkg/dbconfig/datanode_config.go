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
	"k8s.io/utils/pointer"

	"github.com/GreptimeTeam/greptimedb-operator/apis/v1alpha1"
)

var _ Config = &DatanodeConfig{}

// DatanodeConfig is the configuration for the datanode.
type DatanodeConfig struct {
	NodeID      *uint64 `tomlmapping:"node_id"`
	RPCAddr     *string `tomlmapping:"rpc_addr"`
	RPCHostName *string `tomlmapping:"rpc_hostname"`

	// StorageConfig is the configuration for the storage.
	StorageConfig `tomlmapping:",inline"`

	// WALConfig is the configuration for the WAL.
	WALConfig `tomlmapping:",inline"`

	// InputConfig is from config field of cluster spec.
	InputConfig string
}

// ConfigureByCluster configures the datanode config by the given cluster.
func (c *DatanodeConfig) ConfigureByCluster(cluster *v1alpha1.GreptimeDBCluster) error {
	if cluster.GetS3Storage() != nil {
		if err := c.ConfigureS3Storage(cluster.Namespace, cluster.GetS3Storage()); err != nil {
			return err
		}
	}

	if cluster.GetOSSStorage() != nil {
		if err := c.ConfigureOSSStorage(cluster.Namespace, cluster.GetOSSStorage()); err != nil {
			return err
		}
	}

	if cluster.GetGCSStorage() != nil {
		if err := c.ConfigureGCSStorage(cluster.Namespace, cluster.GetGCSStorage()); err != nil {
			return err
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
