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

package metasrv

import (
	"github.com/pelletier/go-toml/v2"

	"github.com/GreptimeTeam/greptimedb-operator/apis/v1alpha1"
)

type Config struct {
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

// FromRawData creates metasrv config from the given input.
func FromRawData(input []byte) (*Config, error) {
	var cfg Config
	if err := toml.Unmarshal(input, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// FromClusterCRD creates metasrv config from the cluster CRD.
func FromClusterCRD(cluster *v1alpha1.GreptimeDBCluster) (*Config, error) {
	return &Config{}, nil
}

// Marshal marshals the metasrv config to string in TOML format.
func (c *Config) Marshal() ([]byte, error) {
	ouput, err := toml.Marshal(c)
	if err != nil {
		return nil, err
	}

	return ouput, nil
}
