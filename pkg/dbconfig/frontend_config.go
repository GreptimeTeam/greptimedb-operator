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

// FrontendConfig is the configuration for the frontend.
type (
	FrontendConfig struct {
		// Node running mode.
		Mode string `toml:"mode,omitempty"`

		HTTPOptions struct {
			Addr    string `toml:"addr,omitempty"`
			Timeout string `toml:"timeout,omitempty"`
		} `toml:"http_options,omitempty"`

		GRPCOptions struct {
			Addr        string `toml:"addr,omitempty"`
			RuntimeSize int32  `toml:"runtime_size,omitempty"`
		} `toml:"grpc_options,omitempty"`

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
		} `toml:"mysql_options,omitempty"`

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
		} `toml:"postgres_options,omitempty"`

		OpenTSDBOptions struct {
			Addr        string `toml:"addr,omitempty"`
			RuntimeSize int32  `toml:"runtime_size,omitempty"`
		} `toml:"opentsdb_options,omitempty"`

		InfluxDBOptions struct {
			Enable *bool `toml:"enable,omitempty"`
		} `toml:"influxdb_options,omitempty"`

		PrometheusOptions struct {
			Enable *bool `toml:"enable,omitempty"`
		} `toml:"prometheus_options,omitempty"`

		PromOptions struct {
			Addr string `toml:"addr,omitempty"`
		} `toml:"prom_options,omitempty"`

		MetaClientOptions struct {
			// Metasrv address list.
			MetaSrvAddrs []string `toml:"metasrv_addrs,omitempty"`

			// Operation timeout in milliseconds.
			TimeoutMillis int32 `toml:"timeout_millis,omitempty"`

			// Connect server timeout in milliseconds.
			ConnectTimeoutMillis int32 `toml:"connect_timeout_millis,omitempty"`

			// `TCP_NODELAY` option for accepted connections.
			TCPNoDelay *bool `toml:"tcp_nodelay,omitempty"`
		} `toml:"meta_client_options,omitempty"`

		Logging struct {
			Dir   string `toml:"dir,omitempty"`
			Level string `toml:"level,omitempty"`
		} `toml:"logging,omitempty"`
	}
)

func (c *FrontendConfig) fromClusterCRD(_ *v1alpha1.GreptimeDBCluster) error {
	return nil
}
