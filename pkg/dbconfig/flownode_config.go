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
)

var _ Config = &FlownodeConfig{}

// FlownodeConfig is the configuration for the datanode.
type FlownodeConfig struct {
	NodeID   *uint64 `tomlmapping:"node_id"`
	Addr     *string `tomlmapping:"grpc.addr"`
	Hostname *string `tomlmapping:"grpc.hostname"`

	// LoggingConfig is the configuration for the logging.
	LoggingConfig `tomlmapping:",inline"`

	// InputConfig is from config field of cluster spec.
	InputConfig string
}

// ConfigureByCluster configures the datanode config by the given cluster.
func (c *FlownodeConfig) ConfigureByCluster(cluster *v1alpha1.GreptimeDBCluster) error {
	if cfg := cluster.GetFlownode().GetConfig(); cfg != "" {
		if err := c.SetInputConfig(cfg); err != nil {
			return err
		}
	}

	c.ConfigureLogging(cluster.GetLogging(), cluster.GetFlownode().GetLogging())

	return nil
}

// ConfigureByStandalone is not need to implement in cluster mode.
func (c *FlownodeConfig) ConfigureByStandalone(_ *v1alpha1.GreptimeDBStandalone) error {
	return nil
}

// Kind returns the component kind of the datanode.
func (c *FlownodeConfig) Kind() v1alpha1.ComponentKind {
	return v1alpha1.FlownodeComponentKind
}

// GetInputConfig returns the input config.
func (c *FlownodeConfig) GetInputConfig() string {
	return c.InputConfig
}

// SetInputConfig sets the input config.
func (c *FlownodeConfig) SetInputConfig(input string) error {
	c.InputConfig = input
	return nil
}
