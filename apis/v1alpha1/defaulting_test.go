// Copyright 2022 Greptime Team
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

import (
	"k8s.io/apimachinery/pkg/util/intstr"
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
							Image: "greptime/greptimedb:latest",
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
					Initializer: &InitializerSpec{Image: defaultInitializer},
					Base: &PodTemplateSpec{
						MainContainer: &MainContainerSpec{
							Image: "greptime/greptimedb:latest",
							Resources: &corev1.ResourceRequirements{
								Requests: map[corev1.ResourceName]resource.Quantity{
									"cpu":    resource.MustParse(defaultCPU),
									"memory": resource.MustParse(defaultMemory),
								},
								Limits: map[corev1.ResourceName]resource.Quantity{
									"cpu":    resource.MustParse(defaultCPU),
									"memory": resource.MustParse(defaultMemory),
								},
							},
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/health",
										Port: intstr.FromInt(defaultHTTPServicePort),
									},
								},
							},
						},
					},
					Frontend: &FrontendSpec{
						ComponentSpec: ComponentSpec{
							Replicas: 3,
							Template: &PodTemplateSpec{
								MainContainer: &MainContainerSpec{
									Image: "greptime/greptimedb:latest",
									Resources: &corev1.ResourceRequirements{
										Requests: map[corev1.ResourceName]resource.Quantity{
											"cpu":    resource.MustParse(defaultCPU),
											"memory": resource.MustParse(defaultMemory),
										},
										Limits: map[corev1.ResourceName]resource.Quantity{
											"cpu":    resource.MustParse(defaultCPU),
											"memory": resource.MustParse(defaultMemory),
										},
									},
									ReadinessProbe: &corev1.Probe{
										ProbeHandler: corev1.ProbeHandler{
											HTTPGet: &corev1.HTTPGetAction{
												Path: "/health",
												Port: intstr.FromInt(defaultHTTPServicePort),
											},
										},
									},
								},
							},
						},
						Service: ServiceSpec{
							Type: corev1.ServiceTypeClusterIP,
						},
					},
					Meta: &MetaSpec{
						ComponentSpec: ComponentSpec{
							Replicas: 3,
							Template: &PodTemplateSpec{
								MainContainer: &MainContainerSpec{
									Image: "greptime/greptimedb:latest",
									Resources: &corev1.ResourceRequirements{
										Requests: map[corev1.ResourceName]resource.Quantity{
											"cpu":    resource.MustParse(defaultCPU),
											"memory": resource.MustParse(defaultMemory),
										},
										Limits: map[corev1.ResourceName]resource.Quantity{
											"cpu":    resource.MustParse(defaultCPU),
											"memory": resource.MustParse(defaultMemory),
										},
									},
									ReadinessProbe: &corev1.Probe{
										ProbeHandler: corev1.ProbeHandler{
											HTTPGet: &corev1.HTTPGetAction{
												Path: "/health",
												Port: intstr.FromInt(defaultHTTPServicePort),
											},
										},
									},
								},
							},
						},
						ServicePort: int32(defaultMetaServicePort),
					},
					Datanode: &DatanodeSpec{
						ComponentSpec: ComponentSpec{
							Replicas: 1,
							Template: &PodTemplateSpec{
								MainContainer: &MainContainerSpec{
									Image: "greptime/greptimedb:latest",
									Resources: &corev1.ResourceRequirements{
										Requests: map[corev1.ResourceName]resource.Quantity{
											"cpu":    resource.MustParse(defaultCPU),
											"memory": resource.MustParse(defaultMemory),
										},
										Limits: map[corev1.ResourceName]resource.Quantity{
											"cpu":    resource.MustParse(defaultCPU),
											"memory": resource.MustParse(defaultMemory),
										},
									},
									ReadinessProbe: &corev1.Probe{
										ProbeHandler: corev1.ProbeHandler{
											HTTPGet: &corev1.HTTPGetAction{
												Path: "/health",
												Port: intstr.FromInt(defaultHTTPServicePort),
											},
										},
									},
								},
							},
						},
						Storage: StorageSpec{
							Name:                defaultDataNodeStorageName,
							StorageSize:         defaultDataNodeStorageSize,
							MountPath:           defaultDataNodeStorageMountPath,
							StorageRetainPolicy: defaultStorageRetainPolicyType,
						},
					},
					HTTPServicePort:       int32(defaultHTTPServicePort),
					GRPCServicePort:       int32(defaultGRPCServicePort),
					MySQLServicePort:      int32(defaultMySQLServicePort),
					PostgresServicePort:   int32(defaultPostgresServicePort),
					PrometheusServicePort: int32(defaultPrometheusServicePort),
					OpenTSDBServicePort:   int32(defaultOpenTSDBServicePort),
					Version:               "latest",
				},
			},
		},

		// #1
		{
			GreptimeDBCluster{
				Spec: GreptimeDBClusterSpec{
					Base: &PodTemplateSpec{
						MainContainer: &MainContainerSpec{
							Image: "greptime/greptimedb:latest",
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
									Image: "greptime/frontend:latest",
									Args: []string{
										"--metasrv-addr",
										"meta.default:3002",
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
									Image: "greptime/meta:latest",
									Args: []string{
										"--store-addr",
										"etcd.default:2379",
									},
								},
							},
						},
						EtcdEndpoints: []string{
							"etcd.default:2379",
						},
					},
					Datanode: &DatanodeSpec{
						ComponentSpec: ComponentSpec{
							Replicas: 1,
							Template: &PodTemplateSpec{
								MainContainer: &MainContainerSpec{
									Image: "greptime/greptimedb:latest",
								},
							},
						},
					},
				},
			},
			GreptimeDBCluster{
				Spec: GreptimeDBClusterSpec{
					Initializer: &InitializerSpec{Image: defaultInitializer},
					Base: &PodTemplateSpec{
						MainContainer: &MainContainerSpec{
							Image: "greptime/greptimedb:latest",
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
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/health",
										Port: intstr.FromInt(defaultHTTPServicePort),
									},
								},
							},
						},
					},
					Frontend: &FrontendSpec{
						ComponentSpec: ComponentSpec{
							Replicas: 3,
							Template: &PodTemplateSpec{
								MainContainer: &MainContainerSpec{
									Image: "greptime/frontend:latest",
									Args: []string{
										"--metasrv-addr",
										"meta.default:3002",
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
									ReadinessProbe: &corev1.Probe{
										ProbeHandler: corev1.ProbeHandler{
											HTTPGet: &corev1.HTTPGetAction{
												Path: "/health",
												Port: intstr.FromInt(defaultHTTPServicePort),
											},
										},
									},
								},
							},
						},
						Service: ServiceSpec{
							Type: corev1.ServiceTypeClusterIP,
						},
					},
					Meta: &MetaSpec{
						ComponentSpec: ComponentSpec{
							Replicas: 3,
							Template: &PodTemplateSpec{
								MainContainer: &MainContainerSpec{
									Image: "greptime/meta:latest",
									Args: []string{
										"--store-addr",
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
									ReadinessProbe: &corev1.Probe{
										ProbeHandler: corev1.ProbeHandler{
											HTTPGet: &corev1.HTTPGetAction{
												Path: "/health",
												Port: intstr.FromInt(defaultHTTPServicePort),
											},
										},
									},
								},
							},
						},
						EtcdEndpoints: []string{
							"etcd.default:2379",
						},
						ServicePort: int32(defaultMetaServicePort),
					},
					Datanode: &DatanodeSpec{
						ComponentSpec: ComponentSpec{
							Replicas: 1,
							Template: &PodTemplateSpec{
								MainContainer: &MainContainerSpec{
									Image: "greptime/greptimedb:latest",
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
									ReadinessProbe: &corev1.Probe{
										ProbeHandler: corev1.ProbeHandler{
											HTTPGet: &corev1.HTTPGetAction{
												Path: "/health",
												Port: intstr.FromInt(defaultHTTPServicePort),
											},
										},
									},
								},
							},
						},
						Storage: StorageSpec{
							Name:                defaultDataNodeStorageName,
							StorageSize:         defaultDataNodeStorageSize,
							MountPath:           defaultDataNodeStorageMountPath,
							StorageRetainPolicy: defaultStorageRetainPolicyType,
						},
					},

					HTTPServicePort:       int32(defaultHTTPServicePort),
					GRPCServicePort:       int32(defaultGRPCServicePort),
					MySQLServicePort:      int32(defaultMySQLServicePort),
					PostgresServicePort:   int32(defaultPostgresServicePort),
					PrometheusServicePort: int32(defaultPrometheusServicePort),
					OpenTSDBServicePort:   int32(defaultOpenTSDBServicePort),
					Version:               "latest",
				},
			},
		},

		// #2
		{
			GreptimeDBCluster{
				Spec: GreptimeDBClusterSpec{
					Base: &PodTemplateSpec{
						MainContainer: &MainContainerSpec{
							Image: "greptime/greptimedb:latest",
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
					Initializer: &InitializerSpec{Image: defaultInitializer},
					Base: &PodTemplateSpec{
						MainContainer: &MainContainerSpec{
							Image: "greptime/greptimedb:latest",
							Resources: &corev1.ResourceRequirements{
								Requests: map[corev1.ResourceName]resource.Quantity{
									"cpu":    resource.MustParse(defaultCPU),
									"memory": resource.MustParse(defaultMemory),
								},
								Limits: map[corev1.ResourceName]resource.Quantity{
									"cpu":    resource.MustParse(defaultCPU),
									"memory": resource.MustParse(defaultMemory),
								},
							},
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/health",
										Port: intstr.FromInt(defaultHTTPServicePort),
									},
								},
							},
						},
					},
					Datanode: &DatanodeSpec{
						ComponentSpec: ComponentSpec{
							Replicas: 3,
							Template: &PodTemplateSpec{
								MainContainer: &MainContainerSpec{
									Image: "greptime/greptimedb:latest",
									Resources: &corev1.ResourceRequirements{
										Requests: map[corev1.ResourceName]resource.Quantity{
											"cpu":    resource.MustParse(defaultCPU),
											"memory": resource.MustParse(defaultMemory),
										},
										Limits: map[corev1.ResourceName]resource.Quantity{
											"cpu":    resource.MustParse(defaultCPU),
											"memory": resource.MustParse(defaultMemory),
										},
									},
									ReadinessProbe: &corev1.Probe{
										ProbeHandler: corev1.ProbeHandler{
											HTTPGet: &corev1.HTTPGetAction{
												Path: "/health",
												Port: intstr.FromInt(defaultHTTPServicePort),
											},
										},
									},
								},
							},
						},
						Storage: StorageSpec{
							Name:                defaultDataNodeStorageName,
							StorageSize:         defaultDataNodeStorageSize,
							MountPath:           defaultDataNodeStorageMountPath,
							StorageRetainPolicy: defaultStorageRetainPolicyType,
						},
					},

					HTTPServicePort:       int32(defaultHTTPServicePort),
					GRPCServicePort:       int32(defaultGRPCServicePort),
					MySQLServicePort:      int32(defaultMySQLServicePort),
					PostgresServicePort:   int32(defaultPostgresServicePort),
					PrometheusServicePort: int32(defaultPrometheusServicePort),
					OpenTSDBServicePort:   int32(defaultOpenTSDBServicePort),
					Version:               "latest",
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
