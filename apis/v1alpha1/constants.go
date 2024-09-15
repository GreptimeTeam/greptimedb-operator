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

const (
	// DefaultVersion is the default version of the GreptimeDB.
	DefaultVersion = "Unknown"

	// DefautlHealthEndpoint is the default health endpoint for the liveness probe.
	DefautlHealthEndpoint = "/health"

	// DefaultHTTPPort is the default HTTP port for the GreptimeDB.
	DefaultHTTPPort int32 = 4000

	// DefaultRPCPort is the default RPC port for the GreptimeDB.
	DefaultRPCPort int32 = 4001

	// DefaultMySQLPort is the default MySQL port for the GreptimeDB.
	DefaultMySQLPort int32 = 4002

	// DefaultPostgreSQLPort is the default PostgreSQL port for the GreptimeDB.
	DefaultPostgreSQLPort int32 = 4003

	// DefaultMetaRPCPort is the default Meta RPC port for the GreptimeDB.
	DefaultMetaRPCPort int32 = 3002

	// DefaultReplicas is the default number of replicas for components of the GreptimeDB cluster.
	DefaultReplicas = 1

	// DefaultDataSize is the default size of the data when using the file storage.
	DefaultDataSize = "20Gi"

	// DefaultDataDir is the default directory for the data when using the file storage.
	DefaultDataDir = "/data/greptimedb/data"

	// DefaultWALDataSize is the default size of the WAL data when using the raft-engine wal.
	DefaultWALDataSize = "5Gi"

	// DefaultWalDir is the default directory for the WAL data when using the raft-engine wal.
	DefaultWalDir = "/data/greptimedb/wal"

	// DefaultLogsDir is the default directory for the logs.
	DefaultLogsDir = "/data/greptimedb/logs"

	// DefaultCacheDir is the default directory for the cache.
	DefaultCacheDir = "/data/greptimedb/cache"

	// DefaultStorageRetainPolicyType is the default storage retain policy type.
	DefaultStorageRetainPolicyType = StorageRetainPolicyTypeRetain

	// DefaultInitializerImage is the default image for the GreptimeDB initializer.
	DefaultInitializerImage = "greptime/greptimedb-initializer:latest"

	// DefaultLogingLevel is the default logging level for the GreptimeDB.
	DefaultLogingLevel = LoggingLevelInfo
)
