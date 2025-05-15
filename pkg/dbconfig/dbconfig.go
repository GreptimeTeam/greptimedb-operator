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
	// TOMLMappingFieldTag is the special tag for the field in the config struct.
	// The value of the tag is the key in the toml.
	TOMLMappingFieldTag = "tomlmapping"

	// InlinedFieldTag indicates the field should be inlined.
	InlinedFieldTag = ",inline"
)

// Config is the interface for the config of greptimedb.
type Config interface {
	// Kind returns the component kind of the config.
	Kind() v1alpha1.RoleKind

	// ConfigureByCluster configures the config by the given cluster.
	ConfigureByCluster(cluster *v1alpha1.GreptimeDBCluster, roleSpec v1alpha1.RoleSpec) error

	// ConfigureByStandalone configures the config by the given standalone.
	ConfigureByStandalone(standalone *v1alpha1.GreptimeDBStandalone) error

	// GetInputConfig returns the input config.
	GetInputConfig() string

	// SetInputConfig sets the input config.
	SetInputConfig(input string) error
}

// NewFromRoleKind creates config from the role kind.
func NewFromRoleKind(kind v1alpha1.RoleKind) (Config, error) {
	switch kind {
	case v1alpha1.MetaRoleKind:
		return &MetaConfig{}, nil
	case v1alpha1.DatanodeRoleKind:
		return &DatanodeConfig{}, nil
	case v1alpha1.FrontendRoleKind:
		return &FrontendConfig{}, nil
	case v1alpha1.FlownodeRoleKind:
		return &FlownodeConfig{}, nil
	case v1alpha1.StandaloneRoleKind:
		return &StandaloneConfig{}, nil
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

	output, err := mergeConfig(tree, config)
	if err != nil {
		return nil, err
	}

	return []byte(output), nil
}

// FromCluster creates config data from the cluster CRD.
func FromCluster(cluster *v1alpha1.GreptimeDBCluster, roleSpec v1alpha1.RoleSpec) ([]byte, error) {
	cfg, err := NewFromRoleKind(roleSpec.GetRoleKind())
	if err != nil {
		return nil, err
	}

	if err := cfg.ConfigureByCluster(cluster, roleSpec); err != nil {
		return nil, err
	}

	return Marshal(cfg)
}

// FromStandalone creates config data from the standalone CRD.
func FromStandalone(standalone *v1alpha1.GreptimeDBStandalone) ([]byte, error) {
	cfg, err := NewFromRoleKind(v1alpha1.StandaloneRoleKind)
	if err != nil {
		return nil, err
	}

	if err := cfg.ConfigureByStandalone(standalone); err != nil {
		return nil, err
	}

	return Marshal(cfg)
}

// mergeConfig merges the input config data with the config struct.
// The config struct have higher priority than the input config data.
func mergeConfig(input *toml.Tree, config interface{}) (string, error) {
	valueOf := reflect.ValueOf(config)

	if valueOf.Kind() == reflect.Ptr {
		// Get the real value of the pointer.
		config = valueOf.Elem().Interface()
		valueOf = reflect.ValueOf(config)
	}

	if valueOf.Kind() == reflect.Struct {
		config = valueOf.Interface()
		valueOf = reflect.ValueOf(config)
	}

	// The `config` will override the `input`.
	typeOf := reflect.TypeOf(config)
	for i := 0; i < valueOf.NumField(); i++ {
		tag := typeOf.Field(i).Tag.Get(TOMLMappingFieldTag)
		if tag == "" {
			continue
		}

		// Handle the `,inline` field.
		if tag == InlinedFieldTag {
			_, err := mergeConfig(input, valueOf.Field(i).Interface())
			if err != nil {
				return "", err
			}
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
		case reflect.Slice:
			if field.Len() > 0 {
				input.Set(tag, field.Interface())
			}
		default:
			return "", fmt.Errorf("field %s is not a pointer type, got %s", tag, field.Kind())
		}
	}

	return input.String(), nil
}
