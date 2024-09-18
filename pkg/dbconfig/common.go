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
	k8sutil "github.com/GreptimeTeam/greptimedb-operator/pkg/util/k8s"
)

// StorageConfig is the configuration for the storage.
type StorageConfig struct {
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
}

// ConfigureS3 configures the storage config with the given S3 storage.
func (c *StorageConfig) ConfigureS3(namespace string, s3 *v1alpha1.S3Storage) error {
	if s3 == nil {
		return nil
	}

	c.StorageType = pointer.String("S3")
	c.StorageBucket = pointer.String(s3.Bucket)
	c.StorageRoot = pointer.String(s3.Root)
	c.StorageEndpoint = pointer.String(s3.Endpoint)
	c.StorageRegion = pointer.String(s3.Region)

	if s3.SecretName != "" {
		data, err := k8sutil.GetSecretsData(namespace, s3.SecretName, []string{v1alpha1.AccessKeyIDSecretKey, v1alpha1.SecretAccessKeySecretKey})
		if err != nil {
			return err
		}
		c.StorageAccessKeyID = pointer.String(string(data[0]))
		c.StorageSecretAccessKey = pointer.String(string(data[1]))
	}

	return nil
}

// ConfigureOSS configures the storage config with the given OSS storage.
func (c *StorageConfig) ConfigureOSS(namespace string, oss *v1alpha1.OSSStorage) error {
	if oss == nil {
		return nil
	}

	c.StorageType = pointer.String("Oss")
	c.StorageBucket = pointer.String(oss.Bucket)
	c.StorageRoot = pointer.String(oss.Root)
	c.StorageEndpoint = pointer.String(oss.Endpoint)
	c.StorageRegion = pointer.String(oss.Region)

	if oss.SecretName != "" {
		data, err := k8sutil.GetSecretsData(namespace, oss.SecretName, []string{v1alpha1.AccessKeyIDSecretKey, v1alpha1.SecretAccessKeySecretKey})
		if err != nil {
			return err
		}
		c.StorageAccessKeyID = pointer.String(string(data[0]))
		c.StorageAccessKeySecret = pointer.String(string(data[1]))
	}

	return nil
}

// ConfigureGCS configures the storage config with the given GCS storage.
func (c *StorageConfig) ConfigureGCS(namespace string, gcs *v1alpha1.GCSStorage) error {
	if gcs == nil {
		return nil
	}

	c.StorageType = pointer.String("Gcs")
	c.StorageBucket = pointer.String(gcs.Bucket)
	c.StorageRoot = pointer.String(gcs.Root)
	c.StorageEndpoint = pointer.String(gcs.Endpoint)
	c.StorageScope = pointer.String(gcs.Scope)

	if gcs.SecretName != "" {
		data, err := k8sutil.GetSecretsData(namespace, gcs.SecretName, []string{v1alpha1.ServiceAccountKey})
		if err != nil {
			return err
		}

		serviceAccount := data[0]
		if len(serviceAccount) != 0 {
			c.StorageCredential = pointer.String(base64.StdEncoding.EncodeToString(serviceAccount))
		}
	}

	return nil
}

// WALConfig is the configuration for the WAL.
type WALConfig struct {
	// The wal file directory.
	WalDir *string `tomlmapping:"wal.dir"`

	// The wal provider.
	WalProvider *string `tomlmapping:"wal.provider"`

	// The kafka broker endpoints.
	WalBrokerEndpoints []string `tomlmapping:"wal.broker_endpoints"`
}
