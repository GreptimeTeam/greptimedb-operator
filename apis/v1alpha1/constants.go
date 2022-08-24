/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

// TODO(zyy17): More elegant way to set default values.
const (
	defaultRequestCPU    = "250m"
	defaultRequestMemory = "64Mi"
	defaultLimitCPU      = "500m"
	defaultLimitMemory   = "128Mi"

	// The default settings for GreptimeDBClusterSpec.
	defaultHTTPServicePort  = 3000
	defaultGRPCServicePort  = 3001
	defaultMySQLServicePort = 3306

	// The default settings for EtcdSpec.
	defaultEtcdImage            = "localhost:5001/greptime/etcd:latest"
	defaultClusterSize          = 3
	defaultEtcdClientPort       = 2379
	defaultEtcdPeerPort         = 2380
	defaultEtcdStorageName      = "etcd"
	defaultEtcdStorageClassName = "standard" // 'standard' is the default local storage class of kind.
	defaultEtcdStorageSize      = "10Gi"
	defaultEtcdStorageMountPath = "/var/run/etcd"
)
