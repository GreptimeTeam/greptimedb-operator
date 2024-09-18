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
	corev1 "k8s.io/api/core/v1"

	"github.com/GreptimeTeam/greptimedb-operator/apis/v1alpha1"
	k8sutil "github.com/GreptimeTeam/greptimedb-operator/pkg/util/k8s"
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

	// ConfigureByStandalone configures the config by the given standalone.
	ConfigureByStandalone(standalone *v1alpha1.GreptimeDBStandalone) error

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
	case v1alpha1.FlownodeComponentKind:
		return &FlownodeConfig{}, nil
	case v1alpha1.StandaloneKind:
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

	output, err := setConfig(tree, config)
	if err != nil {
		return nil, err
	}

	return []byte(output), nil
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

// FromStandalone creates config data from the standalone CRD.
func FromStandalone(standalone *v1alpha1.GreptimeDBStandalone) ([]byte, error) {
	cfg, err := NewFromComponentKind(v1alpha1.StandaloneKind)
	if err != nil {
		return nil, err
	}

	if err := cfg.ConfigureByStandalone(standalone); err != nil {
		return nil, err
	}

	return Marshal(cfg)
}

func getServiceAccountKey(namespace, name string) (secretAccessKey []byte, err error) {
	var secret corev1.Secret
	if err = k8sutil.GetK8sResource(namespace, name, &secret); err != nil {
		return
	}

	if secret.Data == nil {
		err = fmt.Errorf("secret '%s/%s' is empty", namespace, name)
		return
	}

	secretAccessKey = secret.Data[v1alpha1.ServiceAccountKey]
	if secretAccessKey == nil {
		err = fmt.Errorf("secret '%s/%s' does not have service account key '%s'", namespace, name, v1alpha1.ServiceAccountKey)
		return
	}
	return
}

func getOCSCredentials(namespace, name string) (accessKeyID, secretAccessKey []byte, err error) {
	var ocsCredentials corev1.Secret

	if err = k8sutil.GetK8sResource(namespace, name, &ocsCredentials); err != nil {
		return
	}

	if ocsCredentials.Data == nil {
		err = fmt.Errorf("secret '%s/%s' is empty", namespace, name)
		return
	}

	accessKeyID = ocsCredentials.Data[v1alpha1.AccessKeyIDSecretKey]
	if accessKeyID == nil {
		err = fmt.Errorf("secret '%s/%s' does not have access key id '%s'", namespace, name, v1alpha1.AccessKeyIDSecretKey)
		return
	}

	secretAccessKey = ocsCredentials.Data[v1alpha1.SecretAccessKeySecretKey]
	if secretAccessKey == nil {
		err = fmt.Errorf("secret '%s/%s' does not have secret access key '%s'", namespace, name, v1alpha1.SecretAccessKeySecretKey)
		return
	}

	return
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
		case reflect.Slice:
			if field.Len() > 0 {
				input.Set(tag, field.Interface())
			}
		default:
			return "", fmt.Errorf("field %s is not a pointer", tag)
		}
	}

	return input.String(), nil
}
