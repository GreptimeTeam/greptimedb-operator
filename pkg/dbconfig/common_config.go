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

type HTTPOptions struct {
	Addr             string `toml:"addr,omitempty"`
	Timeout          string `toml:"timeout,omitempty"`
	DisableDashboard *bool  `toml:"disable_dashboard,omitempty"`
	BodyLimit        string `toml:"body_limit,omitempty"`
}

type HeartbeatOptions struct {
	Interval      string `toml:"interval,omitempty"`
	RetryInterval string `toml:"retry_interval,omitempty"`
}

type MetaClientOptions struct {
	// Metasrv address list.
	MetaSrvAddrs []string `toml:"metasrv_addrs,omitempty"`

	// Operation timeout in milliseconds.
	Timeout string `toml:"timeout,omitempty"`

	// Connect server timeout in milliseconds.
	ConnectTimeout string `toml:"connect_timeout,omitempty"`

	// DDL operation timeout.
	DDLTimeout string `toml:"ddl_timeout,omitempty"`

	// `TCP_NODELAY` option for accepted connections.
	TCPNoDelay *bool `toml:"tcp_nodelay,omitempty"`
}

type LoggingOptions struct {
	Dir                 string `toml:"dir,omitempty"`
	Level               string `toml:"level,omitempty"`
	EnableJaegerTracing bool   `toml:"enable_jaeger_tracing,omitempty"`
}

type ProcedureConfig struct {
	// Max retry times of procedure.
	MaxRetryTimes int32 `toml:"max_retry_times,omitempty"`

	// Initial retry delay of procedures, increases exponentially.
	RetryDelay string `toml:"retry_delay,omitempty"`
}

type DatanodeClientOptions struct {
	Timeout        string `toml:"timeout,omitempty"`
	ConnectTimeout string `toml:"connect_timeout,omitempty"`
	TcpNoDelay     bool   `toml:"tcp_nodelay,omitempty"`
}
