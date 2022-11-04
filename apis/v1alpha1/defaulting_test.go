package v1alpha1

import (
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestSetDefaults(t *testing.T) {
	tests := []struct {
		input GreptimeDBCluster
		want  GreptimeDBCluster
	}{
		// #0
		{
			GreptimeDBCluster{
				Spec: GreptimeDBClusterSpec{
					Base: &PodTemplateSpec{
						MainContainer: &MainContainerSpec{
							Image: "localhost:5001/greptime/greptimedb:latest",
						},
					},
					Frontend: &FrontendSpec{
						ComponentSpec: ComponentSpec{
							Replicas: 3,
						},
					},
					Meta: &MetaSpec{
						ComponentSpec: ComponentSpec{
							Replicas: 3,
						},
					},
					Datanode: &DatanodeSpec{
						ComponentSpec: ComponentSpec{
							Replicas: 1,
						},
					},
				},
			},
			GreptimeDBCluster{
				Spec: GreptimeDBClusterSpec{
					Base: &PodTemplateSpec{
						MainContainer: &MainContainerSpec{
							Image: "localhost:5001/greptime/greptimedb:latest",
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
							Replicas: 3,
							Template: &PodTemplateSpec{
								MainContainer: &MainContainerSpec{
									Image: "localhost:5001/greptime/greptimedb:latest",
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
						},
					},
					Meta: &MetaSpec{
						ComponentSpec: ComponentSpec{
							Replicas: 3,
							Template: &PodTemplateSpec{
								MainContainer: &MainContainerSpec{
									Image: "localhost:5001/greptime/greptimedb:latest",
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
						},
					},
					Datanode: &DatanodeSpec{
						ComponentSpec: ComponentSpec{
							Replicas: 1,
							Template: &PodTemplateSpec{
								MainContainer: &MainContainerSpec{
									Image: "localhost:5001/greptime/greptimedb:latest",
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
				},
			},
		},

		// #1
		{
			GreptimeDBCluster{
				Spec: GreptimeDBClusterSpec{
					Base: &PodTemplateSpec{
						MainContainer: &MainContainerSpec{
							Image: "localhost:5001/greptime/greptimedb:latest",
							Resources: &corev1.ResourceRequirements{
								Requests: map[corev1.ResourceName]resource.Quantity{
									"cpu":    resource.MustParse("500m"),
									"memory": resource.MustParse("256Mi"),
								},
								Limits: map[corev1.ResourceName]resource.Quantity{
									"cpu":    resource.MustParse("1000m"),
									"memory": resource.MustParse("1024Mi"),
								},
							},
						},
					},
					Frontend: &FrontendSpec{
						ComponentSpec: ComponentSpec{
							Replicas: 3,
							Template: &PodTemplateSpec{
								MainContainer: &MainContainerSpec{
									Image: "localhost:5001/greptime/frontend:latest",
									Args: []string{
										"--meta-endpoint",
										"http://mock-meta.default:3001",
									},
								},
							},
						},
					},
					Meta: &MetaSpec{
						ComponentSpec: ComponentSpec{
							Replicas: 3,
							Template: &PodTemplateSpec{
								MainContainer: &MainContainerSpec{
									Image: "localhost:5001/greptime/meta:latest",
									Args: []string{
										"--etcd-endpoint",
										"etcd.default:2379",
									},
								},
							},
						},
						EtcdEndpoints: []string{
							"127.0.0.1:2379",
						},
					},
					Datanode: &DatanodeSpec{
						ComponentSpec: ComponentSpec{
							Replicas: 1,
							Template: &PodTemplateSpec{
								MainContainer: &MainContainerSpec{
									Image: "localhost:5001/greptime/greptimedb:latest",
								},
							},
						},
					},
				},
			},
			GreptimeDBCluster{
				Spec: GreptimeDBClusterSpec{
					Base: &PodTemplateSpec{
						MainContainer: &MainContainerSpec{
							Image: "localhost:5001/greptime/greptimedb:latest",
							Resources: &corev1.ResourceRequirements{
								Requests: map[corev1.ResourceName]resource.Quantity{
									"cpu":    resource.MustParse("500m"),
									"memory": resource.MustParse("256Mi"),
								},
								Limits: map[corev1.ResourceName]resource.Quantity{
									"cpu":    resource.MustParse("1000m"),
									"memory": resource.MustParse("1024Mi"),
								},
							},
						},
					},
					Frontend: &FrontendSpec{
						ComponentSpec: ComponentSpec{
							Replicas: 3,
							Template: &PodTemplateSpec{
								MainContainer: &MainContainerSpec{
									Image: "localhost:5001/greptime/frontend:latest",
									Args: []string{
										"--meta-endpoint",
										"http://mock-meta.default:3001",
									},
									Resources: &corev1.ResourceRequirements{
										Requests: map[corev1.ResourceName]resource.Quantity{
											"cpu":    resource.MustParse("500m"),
											"memory": resource.MustParse("256Mi"),
										},
										Limits: map[corev1.ResourceName]resource.Quantity{
											"cpu":    resource.MustParse("1000m"),
											"memory": resource.MustParse("1024Mi"),
										},
									},
								},
							},
						},
					},
					Meta: &MetaSpec{
						ComponentSpec: ComponentSpec{
							Replicas: 3,
							Template: &PodTemplateSpec{
								MainContainer: &MainContainerSpec{
									Image: "localhost:5001/greptime/meta:latest",
									Args: []string{
										"--etcd-endpoint",
										"etcd.default:2379",
									},
									Resources: &corev1.ResourceRequirements{
										Requests: map[corev1.ResourceName]resource.Quantity{
											"cpu":    resource.MustParse("500m"),
											"memory": resource.MustParse("256Mi"),
										},
										Limits: map[corev1.ResourceName]resource.Quantity{
											"cpu":    resource.MustParse("1000m"),
											"memory": resource.MustParse("1024Mi"),
										},
									},
								},
							},
						},
						EtcdEndpoints: []string{
							"127.0.0.1:2379",
						},
					},
					Datanode: &DatanodeSpec{
						ComponentSpec: ComponentSpec{
							Replicas: 1,
							Template: &PodTemplateSpec{
								MainContainer: &MainContainerSpec{
									Image: "localhost:5001/greptime/greptimedb:latest",
									Resources: &corev1.ResourceRequirements{
										Requests: map[corev1.ResourceName]resource.Quantity{
											"cpu":    resource.MustParse("500m"),
											"memory": resource.MustParse("256Mi"),
										},
										Limits: map[corev1.ResourceName]resource.Quantity{
											"cpu":    resource.MustParse("1000m"),
											"memory": resource.MustParse("1024Mi"),
										},
									},
								},
							},
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
				},
			},
		},

		// #2
		{
			GreptimeDBCluster{
				Spec: GreptimeDBClusterSpec{
					Base: &PodTemplateSpec{
						MainContainer: &MainContainerSpec{
							Image: "localhost:5001/greptime/greptimedb:latest",
						},
					},
					Datanode: &DatanodeSpec{
						ComponentSpec: ComponentSpec{
							Replicas: 3,
						},
					},
				},
			},
			GreptimeDBCluster{
				Spec: GreptimeDBClusterSpec{
					Base: &PodTemplateSpec{
						MainContainer: &MainContainerSpec{
							Image: "localhost:5001/greptime/greptimedb:latest",
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
					Datanode: &DatanodeSpec{
						ComponentSpec: ComponentSpec{
							Replicas: 3,
							Template: &PodTemplateSpec{
								MainContainer: &MainContainerSpec{
									Image: "localhost:5001/greptime/greptimedb:latest",
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
				},
			},
		},
	}

	for i, tt := range tests {
		if err := tt.input.SetDefaults(); err != nil {
			t.Errorf("set default cluster failed: %v", err)
		}

		if !reflect.DeepEqual(tt.want, tt.input) {
			t.Errorf("run test [%d] failed, want %v, got %v", i, tt.want, tt.input)
		}
	}
}
