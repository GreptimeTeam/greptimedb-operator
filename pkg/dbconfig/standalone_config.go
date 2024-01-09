// Copyright 2024 Greptime Team
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

var _ Config = &FrontendConfig{}

// StandaloneConfig is the configuration for the frontend.
type StandaloneConfig struct {
	// InputConfig is from config field of cluster spec.
	InputConfig string
}

// ConfigureByCluster is not need to implemenet in standalone mode.
func (c *StandaloneConfig) ConfigureByCluster(cluster *v1alpha1.GreptimeDBCluster) error {
	return nil
}

// ConfigureByStandalone is not need to implemenet in cluster mode.
func (c *StandaloneConfig) ConfigureByStandalone(standalone *v1alpha1.GreptimeDBStandalone) error {
	if len(standalone.Spec.Config) > 0 {
		if err := c.SetInputConfig(standalone.Spec.Config); err != nil {
			return err
		}
	}

	return nil
}

// Kind returns the component kind of the standalone.
func (c *StandaloneConfig) Kind() v1alpha1.ComponentKind {
	return v1alpha1.StandaloneKind
}

// GetInputConfig returns the input config of the standalone.
func (c *StandaloneConfig) GetInputConfig() string {
	return c.InputConfig
}

// SetInputConfig sets the input config of the standalone.
func (c *StandaloneConfig) SetInputConfig(inputConfig string) error {
	c.InputConfig = inputConfig
	return nil
}
