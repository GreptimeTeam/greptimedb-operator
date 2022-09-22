package v1alpha1

import (
	"github.com/imdario/mergo"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

var (
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
	defaultStorageRetainPolicyType  = RetainStorageRetainPolicyTypeRetain
)

func (in *GreptimeDBCluster) SetDefaults() error {
	if in == nil {
		return nil
	}

	var defaultGreptimeDBClusterSpec = &GreptimeDBClusterSpec{
		Base: &PodTemplateSpec{
			MainContainer: &MainContainerSpec{
				Resources: &corev1.ResourceRequirements{
					Requests: map[corev1.ResourceName]resource.Quantity{
						"cpu":    resource.MustParse(defaultRequestCPU),
						"memory": resource.MustParse(defaultRequestMemory),
					},
					Limits: map[corev1.ResourceName]resource.Quantity{
						"cpu":    resource.MustParse(defaultLimitCPU),
						"memory": resource.MustParse(defaultLimitMemory),
					},
				},
			},
		},
		Frontend: &FrontendSpec{
			ComponentSpec: ComponentSpec{
				Template: &PodTemplateSpec{},
			},
		},
		Meta: &MetaSpec{
			ComponentSpec: ComponentSpec{
				Template: &PodTemplateSpec{},
			},
		},
		Datanode: &DatanodeSpec{
			ComponentSpec: ComponentSpec{
				Template: &PodTemplateSpec{},
			},
			Storage: StorageSpec{
				Name:                defaultDataNodeStorageName,
				StorageClassName:    &defaultDataNodeStorageClassName,
				StorageSize:         defaultDataNodeStorageSize,
				MountPath:           defaultDataNodeStorageMountPath,
				StorageRetainPolicy: defaultStorageRetainPolicyType,
			},
		},
		HTTPServicePort:  int32(defaultHTTPServicePort),
		GRPCServicePort:  int32(defaultGRPCServicePort),
		MySQLServicePort: int32(defaultMySQLServicePort),
	}

	if err := mergo.Merge(&in.Spec, defaultGreptimeDBClusterSpec); err != nil {
		return err
	}

	if err := mergo.Merge(in.Spec.Frontend.Template, in.Spec.Base); err != nil {
		return err
	}

	if err := mergo.Merge(in.Spec.Meta.Template, in.Spec.Base); err != nil {
		return err
	}

	if err := mergo.Merge(in.Spec.Datanode.Template, in.Spec.Base); err != nil {
		return err
	}

	return nil
}
