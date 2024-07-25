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
	"github.com/GreptimeTeam/greptimedb-operator/apis/v1alpha1"
	"github.com/GreptimeTeam/greptimedb-operator/pkg/util"
)

var _ Config = &MetasrvConfig{}

// MetasrvConfig is the configuration for the metasrv.
type MetasrvConfig struct {
	// Enable region failover.
	EnableRegionFailover *bool `tomlmapping:"enable_region_failover"`

	// If it's not empty, the metasrv will store all data with this key prefix.
	StoreKeyPrefix *string `tomlmapping:"store_key_prefix"`

	// The wal provider.
	WalProvider *string `tomlmapping:"wal.provider"`

	// The kafka broker endpoints.
	WalBrokerEndpoints []string `tomlmapping:"wal.broker_endpoints"`

	// InputConfig is from config field of cluster spec.
	InputConfig string
}

// ConfigureByCluster configures the metasrv config by the given cluster.
func (c *MetasrvConfig) ConfigureByCluster(cluster *v1alpha1.GreptimeDBCluster) error {
	if cluster.Spec.Meta != nil {
		c.EnableRegionFailover = cluster.Spec.Meta.EnableRegionFailover

		if len(cluster.Spec.Meta.StoreKeyPrefix) > 0 {
			c.StoreKeyPrefix = &cluster.Spec.Meta.StoreKeyPrefix
		}

		if len(cluster.Spec.Meta.Config) > 0 {
			if err := c.SetInputConfig(cluster.Spec.Meta.Config); err != nil {
				return err
			}
		}

		if cluster.Spec.RemoteWalProvider != nil && cluster.Spec.RemoteWalProvider.KafkaRemoteWal != nil {
			c.WalProvider = util.StringPtr("kafka")
			c.WalBrokerEndpoints = cluster.Spec.RemoteWalProvider.KafkaRemoteWal.BrokerEndpoints
		}
	}

	return nil
}

// ConfigureByStandalone is not need to implement in cluster mode.
func (c *MetasrvConfig) ConfigureByStandalone(_ *v1alpha1.GreptimeDBStandalone) error {
	return nil
}

// Kind returns the component kind of the metasrv.
func (c *MetasrvConfig) Kind() v1alpha1.ComponentKind {
	return v1alpha1.MetaComponentKind
}

// GetInputConfig returns the input config of the metasrv.
func (c *MetasrvConfig) GetInputConfig() string {
	return c.InputConfig
}

// SetInputConfig sets the input config of the metasrv.
func (c *MetasrvConfig) SetInputConfig(inputConfig string) error {
	c.InputConfig = inputConfig
	return nil
}
