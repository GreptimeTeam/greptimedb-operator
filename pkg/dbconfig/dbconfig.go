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

// FromFile creates config from the file.
func FromFile(filename string, config interface{}) error {
	if err := checkConfigType(config); err != nil {
		return err
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	return FromRawData(data, config)
}

// FromRawData creates config from the raw input.
func FromRawData(raw []byte, config interface{}) error {
	if err := checkConfigType(config); err != nil {
		return err
	}

	if err := toml.Unmarshal(raw, config); err != nil {
		return err
	}
	return nil
}

// Marshal marshals the input config to toml.
func Marshal(config interface{}) ([]byte, error) {
	if err := checkConfigType(config); err != nil {
		return nil, err
	}

	ouput, err := toml.Marshal(config)
	if err != nil {
		return nil, err
	}
	return ouput, nil
}

// Merge merges the extra config to the input object.
func Merge(extraConfigData []byte, baseConfig interface{}) error {
	extraConfig, err := newConfig(baseConfig)
	if err != nil {
		return err
	}

	if err := FromRawData(extraConfigData, extraConfig); err != nil {
		return err
	}

	if err := mergo.Merge(baseConfig, extraConfig, mergo.WithOverride); err != nil {
		return err
	}

	return nil
}

// FromClusterCRD creates config from the cluster CRD.
func FromClusterCRD(cluster *v1alpha1.GreptimeDBCluster, config interface{}) error {
	switch config.(type) {
	case *MetasrvConfig:
		return config.(*MetasrvConfig).fromClusterCRD(cluster)
	case *FrontendConfig:
		return config.(*FrontendConfig).fromClusterCRD(cluster)
	case *DatanodeConfig:
		return config.(*DatanodeConfig).fromClusterCRD(cluster)
	default:
		return fmt.Errorf("unknown object type: %T", config)
	}
}

func newConfig(config interface{}) (interface{}, error) {
	switch config.(type) {
	case *MetasrvConfig:
		return &MetasrvConfig{}, nil
	case *FrontendConfig:
		return &FrontendConfig{}, nil
	case *DatanodeConfig:
		return &DatanodeConfig{}, nil
	default:
		return nil, fmt.Errorf("unknown object type: %T", config)
	}
}

func checkConfigType(config interface{}) error {
	switch config.(type) {
	case *MetasrvConfig, *FrontendConfig, *DatanodeConfig:
		return nil
	default:
		return fmt.Errorf("unknown object type: %T", config)
	}
}
