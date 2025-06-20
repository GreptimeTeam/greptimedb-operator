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

// The following constants are the default values for the GreptimeDBCluster and GreptimeDBStandalone.
const (
	// DefaultVersion is the default version of the GreptimeDB.
	DefaultVersion = "Unknown"

	// DefaultHealthEndpoint is the default health endpoint for the liveness probe.
	DefaultHealthEndpoint = "/health"

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
	DefaultDataSize = "10Gi"

	// DefaultDataSizeForMonitoring is the default size of the data for monitoring.
	DefaultDataSizeForMonitoring = "30Gi"

	// DefaultDataHome is the default directory for the data.
	DefaultDataHome = "/data/greptimedb"

	// DefaultDatanodeFileStorageName is the default file storage name for the datanode.
	DefaultDatanodeFileStorageName = "datanode"

	// DefaultLogsDir is the default directory for the logs.
	DefaultLogsDir = DefaultDataHome + "/logs"

	// DefaultStorageRetainPolicyType is the default storage retain policy type.
	DefaultStorageRetainPolicyType = StorageRetainPolicyTypeRetain

	// DefaultInitializerImage is the default image for the GreptimeDB initializer.
	DefaultInitializerImage = "greptime/greptimedb-initializer:latest"

	// DefaultGreptimeDBImage is the default image for the GreptimeDB.
	DefaultGreptimeDBImage = "greptime/greptimedb:latest"

	// DefaultLoggingLevel is the default logging level for the GreptimeDB.
	DefaultLoggingLevel = LoggingLevelInfo

	// DefaultVectorImage is the default image for the vector.
	DefaultVectorImage = "timberio/vector:nightly-alpine"

	// DefaultVectorCPURequest is the default CPU request for the vector.
	DefaultVectorCPURequest = "50m"

	// DefaultVectorMemoryRequest is the default memory request for the vector.
	DefaultVectorMemoryRequest = "64Mi"

	// DefaultVectorCPULimit is the default CPU limit for the vector.
	DefaultVectorCPULimit = "200m"

	// DefaultVectorMemoryLimit is the default memory limit for the vector.
	DefaultVectorMemoryLimit = "256Mi"
)

// The following constants are the constant configuration for the GreptimeDBCluster and GreptimeDBStandalone.
const (
	// TLSCrtSecretKey is the key for the TLS certificate in the secret.
	TLSCrtSecretKey = "tls.crt"

	// TLSKeySecretKey is the key for the TLS key in the secret.
	TLSKeySecretKey = "tls.key"

	// AccessKeyIDSecretKey is the key for the access key ID in the secret.
	AccessKeyIDSecretKey = "access-key-id"

	// SecretAccessKeySecretKey is the key for the secret access key in the secret.
	SecretAccessKeySecretKey = "secret-access-key"

	// AccessKeySecretSecretKey is the key for the access key secret in the secret.
	AccessKeySecretSecretKey = "access-key-secret"

	// ServiceAccountKey is the key for the service account in the secret.
	ServiceAccountKey = "service-account-key"

	// AccountName is the name for the account in the secret.
	AccountName = "account-name"

	// AccountKey is the key for the account in the secret.
	AccountKey = "account-key"

	// MetaDatabaseUsernameKey is the key for the username in the secret when using MySQL or PostgreSQL as the backend storage.
	MetaDatabaseUsernameKey = "username"

	// MetaDatabasePasswordKey is the key for the password in the secret when using MySQL or PostgreSQL as the backend storage.
	MetaDatabasePasswordKey = "password"
)
