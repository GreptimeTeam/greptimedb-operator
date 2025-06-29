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
	"fmt"

	"k8s.io/utils/ptr"

	"github.com/GreptimeTeam/greptimedb-operator/apis/v1alpha1"
)

var _ Config = &DatanodeConfig{}

// DatanodeConfig is the configuration for the datanode.
type DatanodeConfig struct {
	NodeID        *uint64 `tomlmapping:"node_id"`
	RPCBindAddr   *string `tomlmapping:"grpc.bind_addr"`
	RPCServerAddr *string `tomlmapping:"grpc.server_addr"`

	// StorageConfig is the configuration for the storage.
	StorageConfig `tomlmapping:",inline"`

	// WALConfig is the configuration for the WAL.
	WALConfig `tomlmapping:",inline"`

	// LoggingConfig is the configuration for the logging.
	LoggingConfig `tomlmapping:",inline"`

	// TracingConfig is the configuration for the tracing.
	TracingConfig `tomlmapping:",inline"`

	// InputConfig is from config field of cluster spec.
	InputConfig string
}

// ConfigureByCluster configures the datanode config by the given cluster.
func (c *DatanodeConfig) ConfigureByCluster(cluster *v1alpha1.GreptimeDBCluster, roleSpec v1alpha1.RoleSpec) error {
	if roleSpec.GetRoleKind() != v1alpha1.DatanodeRoleKind {
		return fmt.Errorf("invalid role kind: %s", roleSpec.GetRoleKind())
	}

	datanodeSpec, ok := roleSpec.(*v1alpha1.DatanodeSpec)
	if !ok {
		return fmt.Errorf("invalid role spec type: %T", roleSpec)
	}

	if objectStorage := cluster.GetObjectStorageProvider(); objectStorage != nil {
		if err := c.ConfigureObjectStorage(cluster.GetNamespace(), objectStorage); err != nil {
			return err
		}
	}

	// Set the wal dir if the kafka wal is not enabled.
	if cluster.GetWALProvider().GetKafkaWAL() == nil && cluster.GetWALDir() != "" {
		c.WalDir = ptr.To(cluster.GetWALDir())
	}

	if dataHome := datanodeSpec.GetDataHome(); dataHome != "" {
		c.StorageDataHome = ptr.To(dataHome)
	}

	if cfg := datanodeSpec.GetConfig(); cfg != "" {
		if err := c.SetInputConfig(cfg); err != nil {
			return err
		}
	}

	if kafka := cluster.GetWALProvider().GetKafkaWAL(); kafka != nil {
		c.WalProvider = ptr.To("kafka")
		c.WalBrokerEndpoints = kafka.GetBrokerEndpoints()
	}

	c.ConfigureLogging(datanodeSpec.GetLogging())
	c.ConfigureTracing(datanodeSpec.GetTracing())

	return nil
}

// ConfigureByStandalone is not need to implement in cluster mode.
func (c *DatanodeConfig) ConfigureByStandalone(_ *v1alpha1.GreptimeDBStandalone) error {
	return nil
}

// Kind returns the component kind of the datanode.
func (c *DatanodeConfig) Kind() v1alpha1.RoleKind {
	return v1alpha1.DatanodeRoleKind
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
