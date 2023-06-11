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
	"fmt"
	"os"

	"github.com/imdario/mergo"
	"github.com/pelletier/go-toml/v2"

	"github.com/GreptimeTeam/greptimedb-operator/apis/v1alpha1"
)

// Config is the interface for the config of greptimedb.
type Config interface {
	// Kind returns the component kind of the config.
	Kind() v1alpha1.ComponentKind

	// ConfigureByCluster configures the config by the given cluster.
	ConfigureByCluster(cluster *v1alpha1.GreptimeDBCluster) error
}

// NewFromComponentKind creates config from the component kind.
func NewFromComponentKind(kind v1alpha1.ComponentKind) (Config, error) {
	switch kind {
	case v1alpha1.MetaComponentKind:
		return &MetasrvConfig{}, nil
	case v1alpha1.DatanodeComponentKind:
		return &DatanodeConfig{}, nil
	case v1alpha1.FrontendComponentKind:
		return &FrontendConfig{}, nil
	default:
		return nil, fmt.Errorf("unknown component kind: %s", kind)
	}
}

// FromFile creates config from the file.
func FromFile(filename string, kind v1alpha1.ComponentKind) (Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return FromRawData(data, kind)
}

// FromRawData creates config from the raw input.
func FromRawData(raw []byte, kind v1alpha1.ComponentKind) (Config, error) {
	cfg, err := NewFromComponentKind(kind)
	if err != nil {
		return nil, err
	}

	if err := toml.Unmarshal(raw, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Marshal marshals the input config to toml.
func Marshal(config Config) ([]byte, error) {
	ouput, err := toml.Marshal(config)
	if err != nil {
		return nil, err
	}
	return ouput, nil
}

// Merge merges the extra config to the input object.
func Merge(extraConfigData []byte, baseConfig Config) error {
	extraConfig, err := FromRawData(extraConfigData, baseConfig.Kind())
	if err != nil {
		return err
	}

	if err := mergo.Merge(baseConfig, extraConfig, mergo.WithOverride); err != nil {
		return err
	}

	return nil
}

// FromCluster creates config data from the cluster CRD.
func FromCluster(cluster *v1alpha1.GreptimeDBCluster, componentKind v1alpha1.ComponentKind) ([]byte, error) {
	cfg, err := NewFromComponentKind(componentKind)
	if err != nil {
		return nil, err
	}

	if err := cfg.ConfigureByCluster(cluster); err != nil {
		return nil, err
	}

	return Marshal(cfg)
}
