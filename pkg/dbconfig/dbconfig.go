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
	"reflect"

	"github.com/pelletier/go-toml"

	"github.com/GreptimeTeam/greptimedb-operator/apis/v1alpha1"
)

const (
	TOMLMappingFieldTag = "tomlmapping"
)

// Config is the interface for the config of greptimedb.
type Config interface {
	// Kind returns the component kind of the config.
	Kind() v1alpha1.ComponentKind

	// ConfigureByCluster configures the config by the given cluster.
	ConfigureByCluster(cluster *v1alpha1.GreptimeDBCluster) error

	// GetInputConfig returns the input config.
	GetInputConfig() string

	// SetInputConfig sets the input config.
	SetInputConfig(input string) error
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

// Marshal marshals the input config to toml.
func Marshal(config Config) ([]byte, error) {
	tree, err := toml.Load(config.GetInputConfig())
	if err != nil {
		return nil, err
	}

	ouput, err := setConfig(tree, config)
	if err != nil {
		return nil, err
	}

	return []byte(ouput), nil
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

func setConfig(input *toml.Tree, config interface{}) (string, error) {
	valueOf := reflect.ValueOf(config)

	if valueOf.Kind() == reflect.Ptr {
		config = valueOf.Elem().Interface()
		valueOf = reflect.ValueOf(config)
	}

	// The 'config' will override the 'input'.
	typeOf := reflect.TypeOf(config)
	for i := 0; i < valueOf.NumField(); i++ {
		tag := typeOf.Field(i).Tag.Get(TOMLMappingFieldTag)
		if tag == "" {
			continue
		}
		field := valueOf.Field(i)
		switch field.Kind() {
		case reflect.Ptr:
			if !field.IsNil() {
				elem := field.Elem()
				switch elem.Kind() {
				case reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8, reflect.Uint:
					input.Set(tag, elem.Uint())
				case reflect.Int32, reflect.Int16, reflect.Int8, reflect.Int:
					input.Set(tag, elem.Int())
				case reflect.Float64, reflect.Float32:
					input.Set(tag, elem.Float())
				case reflect.String:
					input.Set(tag, elem.String())
				case reflect.Bool:
					input.Set(tag, elem.Bool())
				default:
					return "", fmt.Errorf("tag '%s' with type '%s is not supported", tag, elem.Kind())
				}
			}
		default:
			return "", fmt.Errorf("field %s is not a pointer", tag)
		}
	}

	return input.String(), nil
}
