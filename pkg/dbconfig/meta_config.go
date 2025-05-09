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
	"k8s.io/utils/ptr"

	"github.com/GreptimeTeam/greptimedb-operator/apis/v1alpha1"
)

var _ Config = &MetaConfig{}

// MetaConfig is the configuration for the meta.
type MetaConfig struct {
	// Enable region failover.
	EnableRegionFailover *bool `tomlmapping:"enable_region_failover"`

	// If it's not empty, the meta will store all data with this key prefix.
	StoreKeyPrefix *string `tomlmapping:"store_key_prefix"`

	// The wal provider.
	WalProvider *string `tomlmapping:"wal.provider"`

	// The kafka broker endpoints.
	WalBrokerEndpoints []string `tomlmapping:"wal.broker_endpoints"`

	// LoggingConfig is the configuration for the logging.
	LoggingConfig `tomlmapping:",inline"`

	// InputConfig is from config field of cluster spec.
	InputConfig string
}

// ConfigureByCluster configures the meta config by the given cluster.
func (c *MetaConfig) ConfigureByCluster(cluster *v1alpha1.GreptimeDBCluster) error {
	c.EnableRegionFailover = ptr.To(cluster.GetMeta().IsEnableRegionFailover())

	if prefix := cluster.GetMeta().GetStoreKeyPrefix(); prefix != "" {
		c.StoreKeyPrefix = ptr.To(prefix)
	}

	if cfg := cluster.GetMeta().GetConfig(); cfg != "" {
		if err := c.SetInputConfig(cfg); err != nil {
			return err
		}
	}

	if kafka := cluster.GetWALProvider().GetKafkaWAL(); kafka != nil {
		c.WalProvider = ptr.To("kafka")
		c.WalBrokerEndpoints = kafka.GetBrokerEndpoints()
	}

	c.ConfigureLogging(cluster.GetMeta().GetLogging())

	return nil
}

// ConfigureByStandalone is not need to implement in cluster mode.
func (c *MetaConfig) ConfigureByStandalone(_ *v1alpha1.GreptimeDBStandalone) error {
	return nil
}

func (c *MetaConfig) ConfigureByFrontend(_ *v1alpha1.FrontendSpec) error {
	return nil
}

// Kind returns the component kind of the meta.
func (c *MetaConfig) Kind() v1alpha1.ComponentKind {
	return v1alpha1.MetaComponentKind
}

// GetInputConfig returns the input config of the meta.
func (c *MetaConfig) GetInputConfig() string {
	return c.InputConfig
}

// SetInputConfig sets the input config of the meta.
func (c *MetaConfig) SetInputConfig(inputConfig string) error {
	c.InputConfig = inputConfig
	return nil
}
