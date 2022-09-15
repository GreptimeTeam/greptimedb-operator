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

	// The default storage settings for datanode.
	defaultDataNodeStorageName      = "datanode"
	defaultDataNodeStorageClassName = "standard" // 'standard' is the default local storage class of kind.
	defaultDataNodeStorageSize      = "10Gi"
	defaultDataNodeStorageMountPath = "/greptimedb/data"
)
