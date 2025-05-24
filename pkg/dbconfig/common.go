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
	"fmt"
	"strings"

	"k8s.io/utils/ptr"

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
	Container              *string `tomlmapping:"storage.container"`
	AccountName            *string `tomlmapping:"storage.account_name"`
	AccountKey             *string `tomlmapping:"storage.account_key"`
	CacheCapacity          *string `tomlmapping:"storage.cache_capacity"`
	EnableVirtualHostStyle *bool   `tomlmapping:"storage.enable_virtual_host_style"`
}

// ConfigureObjectStorage configures the storage config by the given object storage provider accessor.
func (c *StorageConfig) ConfigureObjectStorage(namespace string, accessor v1alpha1.ObjectStorageProviderAccessor) error {
	if s3 := accessor.GetS3Storage(); s3 != nil {
		if err := c.configureS3(namespace, s3); err != nil {
			return err
		}
	} else if oss := accessor.GetOSSStorage(); oss != nil {
		if err := c.configureOSS(namespace, oss); err != nil {
			return err
		}
	} else if gcs := accessor.GetGCSStorage(); gcs != nil {
		if err := c.configureGCS(namespace, gcs); err != nil {
			return err
		}
	} else if blob := accessor.GetAZBlobStorage(); blob != nil {
		if err := c.configureAZBlob(namespace, blob); err != nil {
			return err
		}
	}

	if cacheStorage := accessor.GetCacheStorage(); cacheStorage != nil {
		c.configureCacheStorage(cacheStorage)
	}

	return nil
}

func (c *StorageConfig) configureCacheStorage(cacheStorage *v1alpha1.CacheStorage) {
	if len(cacheStorage.CacheCapacity) != 0 {
		c.CacheCapacity = ptr.To(cacheStorage.CacheCapacity)
	}
}

func (c *StorageConfig) configureS3(namespace string, s3 *v1alpha1.S3Storage) error {
	c.StorageType = ptr.To("S3")
	c.StorageBucket = ptr.To(s3.Bucket)
	c.StorageRoot = ptr.To(s3.Root)
	c.StorageEndpoint = ptr.To(s3.Endpoint)
	c.StorageRegion = ptr.To(s3.Region)

	if s3.SecretName != "" {
		data, err := k8sutil.GetSecretsData(namespace, s3.SecretName, []string{v1alpha1.AccessKeyIDSecretKey, v1alpha1.SecretAccessKeySecretKey})
		if err != nil {
			return err
		}
		c.StorageAccessKeyID = ptr.To(string(data[0]))
		c.StorageSecretAccessKey = ptr.To(string(data[1]))
	}

	if s3.EnableVirtualHostStyle {
		c.EnableVirtualHostStyle = ptr.To(s3.EnableVirtualHostStyle)
	}

	return nil
}

func (c *StorageConfig) configureOSS(namespace string, oss *v1alpha1.OSSStorage) error {
	c.StorageType = ptr.To("Oss")
	c.StorageBucket = ptr.To(oss.Bucket)
	c.StorageRoot = ptr.To(oss.Root)
	c.StorageEndpoint = ptr.To(oss.Endpoint)
	c.StorageRegion = ptr.To(oss.Region)

	if oss.SecretName != "" {
		data, err := k8sutil.GetSecretsData(namespace, oss.SecretName, []string{v1alpha1.AccessKeyIDSecretKey, v1alpha1.AccessKeySecretSecretKey})
		if err != nil {
			return err
		}
		c.StorageAccessKeyID = ptr.To(string(data[0]))
		c.StorageAccessKeySecret = ptr.To(string(data[1]))
	}

	return nil
}

func (c *StorageConfig) configureGCS(namespace string, gcs *v1alpha1.GCSStorage) error {
	c.StorageType = ptr.To("Gcs")
	c.StorageBucket = ptr.To(gcs.Bucket)
	c.StorageRoot = ptr.To(gcs.Root)
	c.StorageEndpoint = ptr.To(gcs.Endpoint)
	c.StorageScope = ptr.To(gcs.Scope)

	if gcs.SecretName != "" {
		data, err := k8sutil.GetSecretsData(namespace, gcs.SecretName, []string{v1alpha1.ServiceAccountKey})
		if err != nil {
			return err
		}

		serviceAccountKey := data[0]
		if len(serviceAccountKey) != 0 {
			c.StorageCredential = ptr.To(base64.StdEncoding.EncodeToString(serviceAccountKey))
		}
	}

	return nil
}

func (c *StorageConfig) configureAZBlob(namespace string, azblob *v1alpha1.AZBlobStorage) error {
	c.StorageType = ptr.To("Azblob")
	c.Container = ptr.To(azblob.Container)
	c.StorageRoot = ptr.To(azblob.Root)
	c.StorageEndpoint = ptr.To(azblob.Endpoint)

	if azblob.SecretName != "" {
		data, err := k8sutil.GetSecretsData(namespace, azblob.SecretName, []string{v1alpha1.AccountName, v1alpha1.AccountKey})
		if err != nil {
			return err
		}
		c.AccountName = ptr.To(string(data[0]))
		c.AccountKey = ptr.To(string(data[1]))
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

// LoggingConfig is the configuration for the logging.
type LoggingConfig struct {
	// The directory to store the log files. If set to empty, logs will not be written to files.
	Dir *string `tomlmapping:"logging.dir"`

	// The log level. Can be `info`/`debug`/`warn`/`error`.
	Level *string `tomlmapping:"logging.level"`

	// The log format. Can be `text`/`json`.
	LogFormat *string `tomlmapping:"logging.log_format"`
}

// ConfigureLogging configures the logging config with the given logging spec.
func (c *LoggingConfig) ConfigureLogging(spec *v1alpha1.LoggingSpec) {
	if spec == nil {
		return
	}

	// Default to empty string.
	c.Dir = ptr.To("")

	// If logsDir is set, use it as the log directory.
	if len(spec.LogsDir) > 0 {
		c.Dir = ptr.To(spec.LogsDir)
	}

	// If only log to stdout, disable log to file even if logsDir is set.
	if spec.IsOnlyLogToStdout() {
		c.Dir = ptr.To("")
	}

	c.Level = ptr.To(c.levelWithFilters(string(spec.Level), spec.Filters))
	c.LogFormat = ptr.To(string(spec.Format))
}

// levelWithFilters returns the level with filters. For example, it will output "info,mito2=debug" if the level is "info" and the filters are ["mito2=debug"].
func (c *LoggingConfig) levelWithFilters(level string, filters []string) string {
	if len(filters) > 0 {
		return fmt.Sprintf("%s,%s", level, strings.Join(filters, ","))
	}
	return level
}
