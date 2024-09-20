// Copyright 2024 Greptime Team
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

package v1alpha1

import (
	"context"
	"fmt"

	"github.com/pelletier/go-toml"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Validate checks the GreptimeDBCluster and returns an error if it is invalid.
func (in *GreptimeDBCluster) Validate() error {
	if in == nil {
		return nil
	}

	if err := in.validateFrontend(); err != nil {
		return err
	}

	if err := in.validateMeta(); err != nil {
		return err
	}

	if err := in.validateDatanode(); err != nil {
		return err
	}

	if in.GetFlownode() != nil {
		if err := in.validateFlownode(); err != nil {
			return err
		}
	}

	if wal := in.GetWALProvider(); wal != nil {
		if err := validateWALProvider(wal); err != nil {
			return err
		}
	}

	if osp := in.GetObjectStorageProvider(); osp != nil {
		if err := valiateStorageProvider(osp); err != nil {
			return err
		}
	}

	return nil
}

// Check checks the GreptimeDBCluster with other resources and returns an error if it is invalid.
func (in *GreptimeDBCluster) Check(ctx context.Context, client client.Client) error {
	// Check if the TLS secret exists and contains the required keys.
	if secretName := in.GetFrontend().GetTLS().GetSecretName(); secretName != "" {
		if err := checkTLSSecret(ctx, client, in.GetNamespace(), secretName); err != nil {
			return err
		}
	}

	// Check if the PodMonitor CRD exists.
	if in.GetPrometheusMonitor().IsEnablePrometheusMonitor() {
		if err := checkPodMonitorExists(ctx, client); err != nil {
			return err
		}
	}

	if secretName := in.GetObjectStorageProvider().GetS3Storage().GetSecretName(); secretName != "" {
		if err := checkS3CredentialsSecret(ctx, client, in.GetNamespace(), secretName); err != nil {
			return err
		}
	}

	if secretName := in.GetObjectStorageProvider().GetOSSStorage().GetSecretName(); secretName != "" {
		if err := checkOSSCredentialsSecret(ctx, client, in.GetNamespace(), secretName); err != nil {
			return err
		}
	}

	if secretName := in.GetObjectStorageProvider().GetGCSStorage().GetSecretName(); secretName != "" {
		if err := checkGCSCredentialsSecret(ctx, client, in.GetNamespace(), secretName); err != nil {
			return err
		}
	}

	return nil
}

func (in *GreptimeDBCluster) validateFrontend() error {
	if err := validateTomlConfig(in.GetFrontend().GetConfig()); err != nil {
		return fmt.Errorf("invalid frontend toml config: '%v'", err)
	}
	return nil
}

func (in *GreptimeDBCluster) validateMeta() error {
	if err := validateTomlConfig(in.GetMeta().GetConfig()); err != nil {
		return fmt.Errorf("invalid meta toml config: '%v'", err)
	}

	if in.GetMeta().IsEnableRegionFailover() {
		if in.GetWALProvider().GetKafkaWAL() == nil {
			return fmt.Errorf("meta enable region failover requires kafka WAL")
		}
	}

	return nil
}

func (in *GreptimeDBCluster) validateDatanode() error {
	if err := validateTomlConfig(in.GetDatanode().GetConfig()); err != nil {
		return fmt.Errorf("invalid datanode toml config: '%v'", err)
	}
	return nil
}

func (in *GreptimeDBCluster) validateFlownode() error {
	if err := validateTomlConfig(in.GetFlownode().GetConfig()); err != nil {
		return fmt.Errorf("invalid flownode toml config: '%v'", err)
	}
	return nil
}

// Validate checks the GreptimeDBStandalone and returns an error if it is invalid.
func (in *GreptimeDBStandalone) Validate() error {
	if in == nil {
		return nil
	}

	if err := validateTomlConfig(in.GetConfig()); err != nil {
		return fmt.Errorf("invalid toml config: '%v'", err)
	}

	if wal := in.GetWALProvider(); wal != nil {
		if err := validateWALProvider(wal); err != nil {
			return err
		}
	}

	if osp := in.GetObjectStorageProvider(); osp != nil {
		if err := valiateStorageProvider(osp); err != nil {
			return err
		}
	}

	return nil
}

// Check checks the GreptimeDBStandalone with other resources and returns an error if it is invalid.
func (in *GreptimeDBStandalone) Check(ctx context.Context, client client.Client) error {
	// Check if the TLS secret exists and contains the required keys.
	if secretName := in.GetTLS().GetSecretName(); secretName != "" {
		if err := checkTLSSecret(ctx, client, in.GetNamespace(), secretName); err != nil {
			return err
		}
	}

	// Check if the PodMonitor CRD exists.
	if in.GetPrometheusMonitor().IsEnablePrometheusMonitor() {
		if err := checkPodMonitorExists(ctx, client); err != nil {
			return err
		}
	}

	if secretName := in.GetObjectStorageProvider().GetS3Storage().GetSecretName(); secretName != "" {
		if err := checkS3CredentialsSecret(ctx, client, in.GetNamespace(), secretName); err != nil {
			return err
		}
	}

	if secretName := in.GetObjectStorageProvider().GetOSSStorage().GetSecretName(); secretName != "" {
		if err := checkOSSCredentialsSecret(ctx, client, in.GetNamespace(), secretName); err != nil {
			return err
		}
	}

	if secretName := in.GetObjectStorageProvider().GetGCSStorage().GetSecretName(); secretName != "" {
		if err := checkGCSCredentialsSecret(ctx, client, in.GetNamespace(), secretName); err != nil {
			return err
		}
	}

	return nil
}

func validateTomlConfig(input string) error {
	if len(input) > 0 {
		data := make(map[string]interface{})
		err := toml.Unmarshal([]byte(input), &data)
		if err != nil {
			return err
		}
	}
	return nil
}

func validateWALProvider(input *WALProviderSpec) error {
	if input == nil {
		return nil
	}

	if input.GetRaftEngineWAL() != nil && input.GetKafkaWAL() != nil {
		return fmt.Errorf("only one of 'raftEngine' or 'kafka' can be set")
	}

	if fs := input.GetRaftEngineWAL().GetFileStorage(); fs != nil {
		if err := validateFileStorage(fs); err != nil {
			return err
		}
	}

	return nil
}

func valiateStorageProvider(input *ObjectStorageProviderSpec) error {
	if input == nil {
		return nil
	}

	if input.getSetObjectStorageCount() > 1 {
		return fmt.Errorf("only one storage provider can be set")
	}

	if fs := input.GetCacheFileStorage(); fs != nil {
		if err := validateFileStorage(fs); err != nil {
			return err
		}
	}

	return nil
}

func validateFileStorage(input *FileStorage) error {
	if input == nil {
		return nil
	}

	if input.GetName() == "" {
		return fmt.Errorf("name is required in file storage")
	}

	if input.GetMountPath() == "" {
		return fmt.Errorf("mountPath is required in file storage")
	}

	if input.GetSize() == "" {
		return fmt.Errorf("storageSize is required in file storage")
	}

	return nil
}

// checkTLSSecret checks if the secret exists and contains the required keys.
func checkTLSSecret(ctx context.Context, client client.Client, namespace, name string) error {
	return checkSecretData(ctx, client, namespace, name, []string{TLSCrtSecretKey, TLSKeySecretKey})
}

func checkGCSCredentialsSecret(ctx context.Context, client client.Client, namespace, name string) error {
	return checkSecretData(ctx, client, namespace, name, []string{ServiceAccountKey})
}

func checkOSSCredentialsSecret(ctx context.Context, client client.Client, namespace, name string) error {
	return checkSecretData(ctx, client, namespace, name, []string{AccessKeyIDSecretKey, SecretAccessKeySecretKey})
}

func checkS3CredentialsSecret(ctx context.Context, client client.Client, namespace, name string) error {
	return checkSecretData(ctx, client, namespace, name, []string{AccessKeyIDSecretKey, SecretAccessKeySecretKey})
}

// checkPodMonitorExists checks if the PodMonitor CRD exists.
func checkPodMonitorExists(ctx context.Context, client client.Client) error {
	const (
		kind  = "podmonitors"
		group = "monitoring.coreos.com"
	)

	var crd apiextensionsv1.CustomResourceDefinition
	if err := client.Get(ctx, types.NamespacedName{Name: fmt.Sprintf("%s.%s", kind, group)}, &crd); err != nil {
		return err
	}

	return nil
}

// checkSecretData checks if the secret exists and contains the required keys.
func checkSecretData(ctx context.Context, client client.Client, namespace, name string, keys []string) error {
	var secret corev1.Secret
	if err := client.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, &secret); err != nil {
		return err
	}

	if secret.Data == nil {
		return fmt.Errorf("the data of secret '%s/%s' is empty", namespace, name)
	}

	for _, key := range keys {
		if _, ok := secret.Data[key]; !ok {
			return fmt.Errorf("secret '%s/%s' does not have key '%s'", namespace, name, key)
		}
	}

	return nil
}
