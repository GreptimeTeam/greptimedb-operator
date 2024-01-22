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

package dbconfig

import (
	"github.com/GreptimeTeam/greptimedb-operator/apis/v1alpha1"
	"github.com/GreptimeTeam/greptimedb-operator/pkg/util"
)

var _ Config = &FrontendConfig{}

// StandaloneConfig is the configuration for the frontend.
type StandaloneConfig struct {
	// Storage options.
	StorageType            *string `tomlmapping:"storage.type"`
	StorageDataHome        *string `tomlmapping:"storage.data_home"`
	StorageAccessKeyID     *string `tomlmapping:"storage.access_key_id"`
	StorageSecretAccessKey *string `tomlmapping:"storage.secret_access_key"`
	StorageAccessKeySecret *string `tomlmapping:"storage.access_key_secret"`
	StorageBucket          *string `tomlmapping:"storage.bucket"`
	StorageRoot            *string `tomlmapping:"storage.root"`
	StorageRegion          *string `tomlmapping:"storage.region"`
	StorageEndpoint        *string `tomlmapping:"storage.endpoint"`

	WalDir *string `tomlmapping:"wal.dir"`

	// InputConfig is from config field of cluster spec.
	InputConfig string
}

// ConfigureByCluster is not need to implement in standalone mode.
func (c *StandaloneConfig) ConfigureByCluster(cluster *v1alpha1.GreptimeDBCluster) error {
	return nil
}

// ConfigureByStandalone is not need to implement in cluster mode.
func (c *StandaloneConfig) ConfigureByStandalone(standalone *v1alpha1.GreptimeDBStandalone) error {
	// TODO(zyy17): need to refactor the following code. It's too ugly.
	if standalone.Spec.ObjectStorageProvider != nil {
		if standalone.Spec.ObjectStorageProvider.S3 != nil {
			if standalone.Spec.ObjectStorageProvider.S3.SecretName != "" {
				accessKeyID, secretAccessKey, err := getOCSCredentials(standalone.Namespace, standalone.Spec.ObjectStorageProvider.S3.SecretName)
				if err != nil {
					return err
				}
				c.StorageAccessKeyID = util.StringPtr(string(accessKeyID))
				c.StorageSecretAccessKey = util.StringPtr(string(secretAccessKey))
			}

			c.StorageType = util.StringPtr("S3")
			c.StorageBucket = util.StringPtr(standalone.Spec.ObjectStorageProvider.S3.Bucket)
			c.StorageRoot = util.StringPtr(standalone.Spec.ObjectStorageProvider.S3.Root)
			c.StorageEndpoint = util.StringPtr(standalone.Spec.ObjectStorageProvider.S3.Endpoint)
			c.StorageRegion = util.StringPtr(standalone.Spec.ObjectStorageProvider.S3.Region)

		} else if standalone.Spec.ObjectStorageProvider.OSS != nil {
			if standalone.Spec.ObjectStorageProvider.OSS.SecretName != "" {
				accessKeyID, secretAccessKey, err := getOCSCredentials(standalone.Namespace, standalone.Spec.ObjectStorageProvider.OSS.SecretName)
				if err != nil {
					return err
				}
				c.StorageAccessKeyID = util.StringPtr(string(accessKeyID))
				c.StorageAccessKeySecret = util.StringPtr(string(secretAccessKey))
			}

			c.StorageType = util.StringPtr("Oss")
			c.StorageBucket = util.StringPtr(standalone.Spec.ObjectStorageProvider.OSS.Bucket)
			c.StorageRoot = util.StringPtr(standalone.Spec.ObjectStorageProvider.OSS.Root)
			c.StorageEndpoint = util.StringPtr(standalone.Spec.ObjectStorageProvider.OSS.Endpoint)
			c.StorageRegion = util.StringPtr(standalone.Spec.ObjectStorageProvider.OSS.Region)
		}
	}

	c.WalDir = util.StringPtr(standalone.Spec.LocalStorage.WalDir)
	c.StorageDataHome = util.StringPtr(standalone.Spec.LocalStorage.DataHome)

	if len(standalone.Spec.Config) > 0 {
		if err := c.SetInputConfig(standalone.Spec.Config); err != nil {
			return err
		}
	}

	return nil
}

// Kind returns the component kind of the standalone.
func (c *StandaloneConfig) Kind() v1alpha1.ComponentKind {
	return v1alpha1.StandaloneKind
}

// GetInputConfig returns the input config of the standalone.
func (c *StandaloneConfig) GetInputConfig() string {
	return c.InputConfig
}

// SetInputConfig sets the input config of the standalone.
func (c *StandaloneConfig) SetInputConfig(inputConfig string) error {
	c.InputConfig = inputConfig
	return nil
}
