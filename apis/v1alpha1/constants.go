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

type StorageReclaimPolicyType string

const (
	// The PVCs will still be retained when the cluster is deleted.
	PolicyRetain StorageReclaimPolicyType = "retain"

	// The PVCs will delete directly when the cluster is deleted.
	PolicyDelete StorageReclaimPolicyType = "delete"
)
