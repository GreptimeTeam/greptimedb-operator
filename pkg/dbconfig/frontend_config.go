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

var _ Config = &FrontendConfig{}

// FrontendConfig is the configuration for the frontend.
type (
	FrontendConfig struct {
		// Node running mode.
		Mode string `toml:"mode,omitempty"`

		NodeID string `toml:"node_id,omitempty"`

		HeartbeatOptions HeartbeatOptions `toml:"heartbeat,omitempty"`

		HTTPOptions HTTPOptions `toml:"http,omitempty"`

		GRPCOptions struct {
			Addr        string `toml:"addr,omitempty"`
			RuntimeSize int32  `toml:"runtime_size,omitempty"`
		} `toml:"grpc,omitempty"`

		// MySQL server options.
		MySQLOptions struct {
			Addr        string `toml:"addr,omitempty"`
			RuntimeSize int32  `toml:"runtime_size,omitempty"`

			// MySQL server TLS options.
			TLS struct {
				Mode     string `toml:"mode,omitempty"`
				CertPath string `toml:"cert_path,omitempty"`
				KeyPath  string `toml:"key_path,omitempty"`
			} `toml:"tls,omitempty"`
		} `toml:"mysql,omitempty"`

		// Postgres server options.
		PostgresOptions struct {
			Addr        string `toml:"addr,omitempty"`
			RuntimeSize int32  `toml:"runtime_size,omitempty"`

			// MySQL server TLS options.
			TLS struct {
				Mode     string `toml:"mode,omitempty"`
				CertPath string `toml:"cert_path,omitempty"`
				KeyPath  string `toml:"key_path,omitempty"`
			} `toml:"tls,omitempty"`
		} `toml:"postgres,omitempty"`

		OpenTSDBOptions struct {
			Addr        string `toml:"addr,omitempty"`
			RuntimeSize int32  `toml:"runtime_size,omitempty"`
		} `toml:"opentsdb,omitempty"`

		InfluxDBOptions struct {
			Enable *bool `toml:"enable,omitempty"`
		} `toml:"influxdb,omitempty"`

		PromStoreOptions struct {
			Enable *bool `toml:"enable,omitempty"`
		} `toml:"prom_store,omitempty"`

		OLTPOptions struct {
			Enable *bool `toml:"enable,omitempty"`
		} `toml:"oltp,omitempty"`

		MetaClientOptions MetaClientOptions `toml:"meta_client,omitempty"`

		LoggingOptions LoggingOptions `toml:"logging,omitempty"`

		DatanodeOptions struct {
			DatanodeClientOptions DatanodeClientOptions `toml:"client,omitempty"`
		} `toml:"datanode,omitempty"`

		UserProvider string `toml:"user_provider,omitempty"`
	}
)

// ConfigureByCluster configures the frontend configuration by the given cluster.
func (c *FrontendConfig) ConfigureByCluster(cluster *v1alpha1.GreptimeDBCluster) error {
	if cluster.Spec.Frontend != nil && len(cluster.Spec.Frontend.Config) > 0 {
		if err := Merge([]byte(cluster.Spec.Frontend.Config), c); err != nil {
			return err
		}
	}

	return nil
}

// Kind returns the component kind of the frontend.
func (c *FrontendConfig) Kind() v1alpha1.ComponentKind {
	return v1alpha1.FrontendComponentKind
}
