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
