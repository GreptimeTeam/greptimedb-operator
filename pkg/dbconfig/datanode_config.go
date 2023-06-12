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

	corev1 "k8s.io/api/core/v1"

	"github.com/GreptimeTeam/greptimedb-operator/apis/v1alpha1"
	"github.com/GreptimeTeam/greptimedb-operator/pkg/utils"
)

var _ Config = &DatanodeConfig{}

// DatanodeConfig is the configuration for the datanode.
type (
	DatanodeConfig struct {
		// Node running mode.
		Mode string `toml:"mode,omitempty"`

		// Whether to use in-memory catalog.
		EnableMemoryCatalog *bool `toml:"enable_memory_catalog,omitempty"`

		// The datanode identifier, should be unique.
		NodeID *uint64 `toml:"node_id,omitempty"`

		// gRPC server address.
		RPCAddr string `toml:"rpc_addr,omitempty"`

		// Hostname of this node.
		RPCHostName string `toml:"rpc_hostname,omitempty"`

		// The number of gRPC server worker threads.
		RPCRuntimeSize int32 `toml:"rpc_runtime_size,omitempty"`

		MetaClientOptions struct {
			// Metasrv address list.
			MetaSrvAddrs []string `toml:"metasrv_addrs,omitempty"`

			// Operation timeout in milliseconds.
			TimeoutMillis int32 `toml:"timeout_millis,omitempty"`

			// Connect server timeout in milliseconds.
			ConnectTimeoutMillis int32 `toml:"connect_timeout_millis,omitempty"`

			// `TCP_NODELAY` option for accepted connections.
			TCPNoDelay *bool `toml:"tcp_nodelay,omitempty"`
		} `toml:"meta_client_options,omitempty"`

		Wal struct {
			FileSize       string `toml:"file_size,omitempty"`
			PurgeThreshold string `toml:"purge_threshold,omitempty"`
			PurgeInterval  string `toml:"purge_interval,omitempty"`
			ReadBatchSize  int32  `toml:"read_batch_size,omitempty"`
			SyncWrite      *bool  `toml:"sync_write,omitempty"`
		} `toml:"wal,omitempty"`

		Storage struct {
			// Storage options.
			Type            string `toml:"type,omitempty"`
			DataHome        string `toml:"data_home,omitempty"`
			Bucket          string `toml:"bucket,omitempty"`
			Root            string `toml:"root,omitempty"`
			AccessKeyID     string `toml:"access_key_id,omitempty"`
			SecretAccessKey string `toml:"secret_access_key,omitempty"`
			Endpoint        string `toml:"endpoint,omitempty"`
			Region          string `toml:"region,omitempty"`
			CachePath       string `toml:"cache_path,omitempty"`
			CacheCapacity   string `toml:"cache_capacity,omitempty"`

			// Storage manifest options.
			Manifest struct {
				// Region checkpoint actions margin.
				CheckpointMargin int32 `toml:"checkpoint_margin,omitempty"`

				// Region manifest logs and checkpoints gc execution duration.
				GCDuration string `toml:"gc_duration,omitempty"`

				// Whether to try creating a manifest checkpoint on region opening.
				CheckpointOnStartup *bool `toml:"checkpoint_on_startup,omitempty"`
			} `toml:"manifest,omitempty"`

			// Storage flush options.
			Flush struct {
				// Max inflight flush tasks.
				MaxFlushTasks int32 `toml:"max_flush_tasks,omitempty"`

				// Default write buffer size for a region.
				RegionWriteBufferSize string `toml:"region_write_buffer_size,omitempty"`

				// Interval to check whether a region needs flush.
				PickerScheduleInterval string `toml:"picker_schedule_interval,omitempty"`

				// Interval to auto flush a region if it has not flushed yet.
				AutoFlushInterval string `toml:"auto_flush_interval,omitempty"`

				// Global write buffer size for all regions.
				GlobalWriteBufferSize string `toml:"global_write_buffer_size,omitempty"`
			} `toml:"flush,omitempty"`
		} `toml:"storage,omitempty"`

		Procedure struct {
			MaxRetryTimes int32  `toml:"max_retry_times,omitempty"`
			RetryDelay    string `toml:"retry_delay,omitempty"`
		} `toml:"procedure,omitempty"`

		Logging struct {
			Dir   string `toml:"dir,omitempty"`
			Level string `toml:"level,omitempty"`
		} `toml:"logging,omitempty"`
	}
)

// ConfigureByCluster configures the datanode config by the given cluster.
func (c *DatanodeConfig) ConfigureByCluster(cluster *v1alpha1.GreptimeDBCluster) error {
	if cluster.Spec.StorageProvider != nil {
		if cluster.Spec.StorageProvider.Local != nil {
			c.Storage.Type = "File"
			c.Storage.DataHome = cluster.Spec.StorageProvider.Local.Directory
		} else if cluster.Spec.StorageProvider.S3 != nil {
			if cluster.Spec.StorageProvider.S3.SecretName != "" {
				accessKeyID, secretAccessKey, err := c.getS3Credentials(cluster.Namespace, cluster.Spec.StorageProvider.S3.SecretName)
				if err != nil {
					return err
				}
				c.Storage.AccessKeyID = string(accessKeyID)
				c.Storage.SecretAccessKey = string(secretAccessKey)
			}

			c.Storage.Type = "S3"
			c.Storage.Bucket = cluster.Spec.StorageProvider.S3.Bucket
			c.Storage.Root = cluster.Spec.StorageProvider.S3.Prefix
			c.Storage.Endpoint = cluster.Spec.StorageProvider.S3.Endpoint
			c.Storage.Region = cluster.Spec.StorageProvider.S3.Region
		}
	}

	if cluster.Spec.Datanode != nil && len(cluster.Spec.Datanode.Config) > 0 {
		if err := Merge([]byte(cluster.Spec.Datanode.Config), c); err != nil {
			return err
		}
	}

	return nil
}

// Kind returns the component kind of the frontend.
func (c *DatanodeConfig) Kind() v1alpha1.ComponentKind {
	return v1alpha1.DatanodeComponentKind
}

const (
	AccessKeyIDSecretKey     = "access-key-id"
	SecretAccessKeySecretKey = "secret-access-key"
)

func (c *DatanodeConfig) getS3Credentials(namespace, name string) (accessKeyID, secretAccessKey []byte, err error) {
	var s3Credentials corev1.Secret

	if err = utils.GetK8sResource(namespace, name, &s3Credentials); err != nil {
		return
	}

	if s3Credentials.Data == nil {
		err = fmt.Errorf("secret %s/%s is empty", namespace, name)
		return
	}

	accessKeyID = s3Credentials.Data[AccessKeyIDSecretKey]
	if accessKeyID == nil {
		err = fmt.Errorf("secret '%s/%s' does not have key '%s'", namespace, name, AccessKeyIDSecretKey)
		return
	}

	secretAccessKey = s3Credentials.Data[SecretAccessKeySecretKey]
	if secretAccessKey == nil {
		err = fmt.Errorf("secret '%s/%s' does not have key '%s'", namespace, name, SecretAccessKeySecretKey)
		return
	}

	return
}
