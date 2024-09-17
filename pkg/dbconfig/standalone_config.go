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
	"encoding/base64"

	"k8s.io/utils/pointer"

	"github.com/GreptimeTeam/greptimedb-operator/apis/v1alpha1"
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
	StorageScope           *string `tomlmapping:"storage.scope"`
	StorageCredential      *string `tomlmapping:"storage.credential"`

	WalDir *string `tomlmapping:"wal.dir"`

	// InputConfig is from config field of cluster spec.
	InputConfig string
}

// ConfigureByCluster is not need to implement in standalone mode.
func (c *StandaloneConfig) ConfigureByCluster(_ *v1alpha1.GreptimeDBCluster) error {
	return nil
}

// ConfigureByStandalone is not need to implement in cluster mode.
func (c *StandaloneConfig) ConfigureByStandalone(standalone *v1alpha1.GreptimeDBStandalone) error {
	if standalone.GetS3Storage() != nil {
		s3 := standalone.GetS3Storage()
		c.StorageType = pointer.String("S3")
		c.StorageBucket = pointer.String(s3.Bucket)
		c.StorageRoot = pointer.String(s3.Root)
		c.StorageEndpoint = pointer.String(s3.Endpoint)
		c.StorageRegion = pointer.String(s3.Region)

		accessKeyID, secretAccessKey, err := getOCSCredentials(standalone.Namespace, s3.SecretName)
		if err != nil {
			return err
		}
		c.StorageAccessKeyID = pointer.String(string(accessKeyID))
		c.StorageSecretAccessKey = pointer.String(string(secretAccessKey))
	}

	if standalone.GetOSSStorage() != nil {
		oss := standalone.GetOSSStorage()
		c.StorageType = pointer.String("Oss")
		c.StorageBucket = pointer.String(oss.Bucket)
		c.StorageRoot = pointer.String(oss.Root)
		c.StorageEndpoint = pointer.String(oss.Endpoint)
		c.StorageRegion = pointer.String(oss.Region)

		accessKeyID, secretAccessKey, err := getOCSCredentials(standalone.Namespace, oss.SecretName)
		if err != nil {
			return err
		}
		c.StorageAccessKeyID = pointer.String(string(accessKeyID))
		c.StorageAccessKeySecret = pointer.String(string(secretAccessKey))
	}

	if standalone.GetGCSStorage() != nil {
		gcs := standalone.GetGCSStorage()
		c.StorageType = pointer.String("Gcs")
		c.StorageBucket = pointer.String(gcs.Bucket)
		c.StorageRoot = pointer.String(gcs.Root)
		c.StorageEndpoint = pointer.String(gcs.Endpoint)
		c.StorageScope = pointer.String(gcs.Scope)

		serviceAccountKey, err := getServiceAccountKey(standalone.Namespace, gcs.SecretName)
		if err != nil {
			return err
		}
		if len(serviceAccountKey) != 0 {
			c.StorageCredential = pointer.String(base64.StdEncoding.EncodeToString(serviceAccountKey))
		}
	}

	if standalone.GetWALDir() != "" {
		c.WalDir = pointer.String(standalone.GetWALDir())
	}

	if standalone.GetDataHome() != "" {
		c.StorageDataHome = pointer.String(standalone.GetDataHome())
	}

	if standalone.GetConfig() != "" {
		if err := c.SetInputConfig(standalone.GetConfig()); err != nil {
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
