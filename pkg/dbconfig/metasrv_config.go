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

// MetasrvConfig is the configuration for the metasrv.
type MetasrvConfig struct {
	// The bind address of metasrv.
	BindAddr string `toml:"bind_addr,omitempty"`

	// The communication server address for frontend and datanode to connect to metasrv.
	ServerAddr string `toml:"server_addr,omitempty"`

	// Etcd server address.
	StoreAddr string `toml:"store_addr,omitempty"`

	// Datanode lease in seconds.
	DatanodeLeaseSec int32 `toml:"datanode_lease_sec,omitempty"`

	// Datanode selector type, can be "LeaseBased" or "LoadBased".
	Selector string `toml:"selector,omitempty"`

	// Store data in memory.
	UseMemoryStore *bool `toml:"use_memory_store,omitempty"`

	Logging struct {
		Dir   string `toml:"dir,omitempty"`
		Level string `toml:"level,omitempty"`
	} `toml:"logging,omitempty"`
}

func (c *MetasrvConfig) fromClusterCRD(_ *v1alpha1.GreptimeDBCluster) error {
	return nil
}
