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
	k8sutil "github.com/GreptimeTeam/greptimedb-operator/pkg/util/k8s"
)

var _ Config = &DatanodeConfig{}

// DatanodeConfig is the configuration for the datanode.
type (
	DatanodeConfig struct {
		// Node running mode.
		Mode string `toml:"mode,omitempty"`

		// The datanode identifier, should be unique.
		NodeID *uint64 `toml:"node_id,omitempty"`

		// Whether to use in-memory catalog.
		EnableMemoryCatalog *bool `toml:"enable_memory_catalog,omitempty"`

		RequireLeaseBeforeStartup *bool `toml:"require_lease_before_startup,omitempty"`

		// gRPC server address.
		RPCAddr string `toml:"rpc_addr,omitempty"`

		// Hostname of this node.
		RPCHostName string `toml:"rpc_hostname,omitempty"`

		// The number of gRPC server worker threads.
		RPCRuntimeSize int32 `toml:"rpc_runtime_size,omitempty"`

		// Max gRPC receiving(decoding) message size.
		RPCMaxRecvMessageSize string `toml:"rpc_max_recv_message_size,omitempty"`

		// Max gRPC sending(encoding) message size.
		RPCMaxSendMessageSize string `toml:"rpc_max_send_message_size,omitempty"`

		HeartbeatOptions HeartbeatOptions `toml:"heartbeat,omitempty"`

		HTTPOptions HTTPOptions `toml:"http,omitempty"`

		MetaClientOptions MetaClientOptions `toml:"meta_client,omitempty"`

		Wal struct {
			Dir            string `toml:"dir,omitempty"`
			FileSize       string `toml:"file_size,omitempty"`
			PurgeThreshold string `toml:"purge_threshold,omitempty"`
			PurgeInterval  string `toml:"purge_interval,omitempty"`
			ReadBatchSize  int32  `toml:"read_batch_size,omitempty"`
			SyncWrite      *bool  `toml:"sync_write,omitempty"`
		} `toml:"wal,omitempty"`

		Storage struct {
			DataHome  string `toml:"data_home,omitempty"`
			GlobalTTL string `toml:"global_ttl,omitempty"`

			// Storage options.
			Type            string `toml:"type,omitempty"`
			Bucket          string `toml:"bucket,omitempty"`
			Root            string `toml:"root,omitempty"`
			AccessKeyID     string `toml:"access_key_id,omitempty"`
			SecretAccessKey string `toml:"secret_access_key,omitempty"`
			AccessKeySecret string `toml:"access_key_secret,omitempty"`
			Endpoint        string `toml:"endpoint,omitempty"`
			Region          string `toml:"region,omitempty"`
			CachePath       string `toml:"cache_path,omitempty"`
			CacheCapacity   string `toml:"cache_capacity,omitempty"`

			Compaction struct {
				// Max task number that can concurrently run.
				MaxInflightTasks int32 `toml:"max_inflight_tasks,omitempty"`

				// Max files in level 0 to trigger compaction.
				MaxFilesInLevel0 int32 `toml:"max_files_in_level0,omitempty"`

				// Max task number for SST purge task after compaction.
				MaxPurgeTasks int32 `toml:"max_purge_tasks,omitempty"`

				// Buffer threshold while writing SST files.
				SSTWriteBufferSize string `toml:"sst_write_buffer_size,omitempty"`
			} `toml:"compaction,omitempty"`

			// Storage manifest options.
			Manifest struct {
				// Region checkpoint actions margin.
				CheckpointMargin int32 `toml:"checkpoint_margin,omitempty"`

				// Region manifest logs and checkpoints gc execution duration.
				GCDuration string `toml:"gc_duration,omitempty"`

				// Whether to try creating a manifest checkpoint on region opening.
				Compress *bool `toml:"compress,omitempty"`
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

		RegionEngine []struct {
			MitoConfig struct {
				NumWorkers                  int    `toml:"num_workers"`
				WorkerChannelSize           int    `toml:"worker_channel_size"`
				WorkerRequestBatchSize      int    `toml:"worker_request_batch_size"`
				ManifestCheckpointDistance  int    `toml:"manifest_checkpoint_distance"`
				ManifestCompressType        string `toml:"manifest_compress_type"`
				MaxBackgroundJobs           int    `toml:"max_background_jobs"`
				AutoFlushInterval           string `toml:"auto_flush_interval"`
				GlobalWriteBufferSize       string `toml:"global_write_buffer_size"`
				GlobalWriteBufferRejectSize string `toml:"global_write_buffer_reject_size"`
				SstMetaCacheSize            string `toml:"sst_meta_cache_size"`
				VectorCacheSize             string `toml:"vector_cache_size"`
			} `toml:"mito,omitempty"`
		} `toml:"region_engine,omitempty"`

		LoggingOptions LoggingOptions `toml:"logging,omitempty"`

		EnableTelemetry *bool `toml:"enable_telemetry,omitempty"`
	}
)

// ConfigureByCluster configures the datanode config by the given cluster.
func (c *DatanodeConfig) ConfigureByCluster(cluster *v1alpha1.GreptimeDBCluster) error {
	// TODO(zyy17): need to refactor the following code. It's too ugly.
	if cluster.Spec.StorageProvider != nil {
		if cluster.Spec.StorageProvider.Local != nil {
			c.Storage.Type = "File"
			c.Storage.DataHome = cluster.Spec.StorageProvider.Local.DataHome
		} else if cluster.Spec.StorageProvider.S3 != nil {
			if cluster.Spec.StorageProvider.S3.SecretName != "" {
				accessKeyID, secretAccessKey, err := c.getOCSCredentials(cluster.Namespace, cluster.Spec.StorageProvider.S3.SecretName)
				if err != nil {
					return err
				}
				c.Storage.AccessKeyID = string(accessKeyID)
				c.Storage.SecretAccessKey = string(secretAccessKey)
			}

			c.Storage.Type = "S3"
			c.Storage.Bucket = cluster.Spec.StorageProvider.S3.Bucket
			c.Storage.Root = cluster.Spec.StorageProvider.S3.Root
			c.Storage.Endpoint = cluster.Spec.StorageProvider.S3.Endpoint
			c.Storage.Region = cluster.Spec.StorageProvider.S3.Region
			c.Storage.DataHome = cluster.Spec.StorageProvider.S3.DataHome

		} else if cluster.Spec.StorageProvider.OSS != nil {
			if cluster.Spec.StorageProvider.OSS.SecretName != "" {
				accessKeyID, secretAccessKey, err := c.getOCSCredentials(cluster.Namespace, cluster.Spec.StorageProvider.OSS.SecretName)
				if err != nil {
					return err
				}
				c.Storage.AccessKeyID = string(accessKeyID)
				c.Storage.AccessKeySecret = string(secretAccessKey)
			}

			c.Storage.Type = "Oss"
			c.Storage.Bucket = cluster.Spec.StorageProvider.OSS.Bucket
			c.Storage.Root = cluster.Spec.StorageProvider.OSS.Root
			c.Storage.Endpoint = cluster.Spec.StorageProvider.OSS.Endpoint
			c.Storage.Region = cluster.Spec.StorageProvider.OSS.Region
			c.Storage.DataHome = cluster.Spec.StorageProvider.OSS.DataHome
		}
	}

	if cluster.Spec.Datanode != nil {
		c.Wal.Dir = cluster.Spec.Datanode.Storage.WalDir

		if len(cluster.Spec.Datanode.Config) > 0 {
			if err := Merge([]byte(cluster.Spec.Datanode.Config), c); err != nil {
				return err
			}
		}
	}

	return nil
}

// Kind returns the component kind of the datanode.
func (c *DatanodeConfig) Kind() v1alpha1.ComponentKind {
	return v1alpha1.DatanodeComponentKind
}

const (
	AccessKeyIDSecretKey     = "access-key-id"
	SecretAccessKeySecretKey = "secret-access-key"
)

func (c *DatanodeConfig) getOCSCredentials(namespace, name string) (accessKeyID, secretAccessKey []byte, err error) {
	var ocsCredentials corev1.Secret

	if err = k8sutil.GetK8sResource(namespace, name, &ocsCredentials); err != nil {
		return
	}

	if ocsCredentials.Data == nil {
		err = fmt.Errorf("secret '%s/%s' is empty", namespace, name)
		return
	}

	accessKeyID = ocsCredentials.Data[AccessKeyIDSecretKey]
	if accessKeyID == nil {
		err = fmt.Errorf("secret '%s/%s' does not have access key id '%s'", namespace, name, AccessKeyIDSecretKey)
		return
	}

	secretAccessKey = ocsCredentials.Data[SecretAccessKeySecretKey]
	if secretAccessKey == nil {
		err = fmt.Errorf("secret '%s/%s' does not have secret access key '%s'", namespace, name, SecretAccessKeySecretKey)
		return
	}

	return
}
