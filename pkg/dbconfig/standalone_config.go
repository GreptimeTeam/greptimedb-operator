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
	"k8s.io/utils/pointer"

	"github.com/GreptimeTeam/greptimedb-operator/apis/v1alpha1"
)

var _ Config = &FrontendConfig{}

// StandaloneConfig is the configuration for the frontend.
type StandaloneConfig struct {
	// StorageConfig is the configuration for the storage.
	StorageConfig `tomlmapping:",inline"`

	// WALConfig is the configuration for the WAL.
	WALConfig `tomlmapping:",inline"`

	// LoggingConfig is the configuration for the logging.
	LoggingConfig `tomlmapping:",inline"`

	// InputConfig is from config field of cluster spec.
	InputConfig string
}

// ConfigureByCluster is not need to implement in standalone mode.
func (c *StandaloneConfig) ConfigureByCluster(_ *v1alpha1.GreptimeDBCluster) error {
	return nil
}

// ConfigureByStandalone is not need to implement in cluster mode.
func (c *StandaloneConfig) ConfigureByStandalone(standalone *v1alpha1.GreptimeDBStandalone) error {
	if standalone.GetS3Storage() != nil {
		if err := c.ConfigureS3Storage(standalone.Namespace, standalone.GetS3Storage()); err != nil {
			return err
		}
	}

	if standalone.GetOSSStorage() != nil {
		if err := c.ConfigureOSSStorage(standalone.Namespace, standalone.GetOSSStorage()); err != nil {
			return err
		}
	}

	if standalone.GetGCSStorage() != nil {
		if err := c.ConfigureGCSStorage(standalone.Namespace, standalone.GetGCSStorage()); err != nil {
			return err
		}
	}

	// Set the wal dir if the kafka wal is not enabled.
	if standalone.GetKafkaWAL() == nil && standalone.GetWALDir() != "" {
		c.WalDir = pointer.String(standalone.GetWALDir())
	}

	if standalone.GetDataHome() != "" {
		c.StorageDataHome = pointer.String(standalone.GetDataHome())
	}

	if standalone.GetConfig() != "" {
		if err := c.SetInputConfig(standalone.GetConfig()); err != nil {
			return err
		}
	}

	if standalone.GetKafkaWAL() != nil {
		c.WalProvider = pointer.String("kafka")
		c.WalBrokerEndpoints = standalone.GetKafkaWAL().BrokerEndpoints
	}

	c.ConfigureLogging(standalone.GetLogging())

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
