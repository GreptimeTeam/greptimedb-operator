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
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

	"github.com/GreptimeTeam/greptimedb-operator/apis/v1alpha1"
)

func stubGetSecretsData(t *testing.T, data map[string]map[string]string) func() {
	t.Helper()
	original := getSecretsData
	getSecretsData = func(namespace, name string, keys []string) ([][]byte, error) {
		secret, ok := data[namespace+"/"+name]
		if !ok {
			return nil, fmt.Errorf("secret '%s/%s' not found", namespace, name)
		}

		values := make([][]byte, 0, len(keys))
		for _, key := range keys {
			value, ok := secret[key]
			if !ok {
				return nil, fmt.Errorf("secret '%s/%s' does not have key '%s'", namespace, name, key)
			}
			values = append(values, []byte(value))
		}

		return values, nil
	}

	return func() { getSecretsData = original }
}

func TestFromClusterForDatanodeConfig(t *testing.T) {
	testCluster := &v1alpha1.GreptimeDBCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-cluster",
			Namespace: "default",
		},
		Spec: v1alpha1.GreptimeDBClusterSpec{
			ObjectStorageProvider: &v1alpha1.ObjectStorageProviderSpec{
				S3: &v1alpha1.S3Storage{
					Root:     "testcluster",
					Bucket:   "testbucket",
					Endpoint: "s3.amazonaws.com",
					Region:   "us-west-2",
				},
			},
			WALProvider: &v1alpha1.WALProviderSpec{
				KafkaWAL: &v1alpha1.KafkaWAL{
					BrokerEndpoints: []string{
						"broker1:9092",
						"broker2:9092",
					},
				},
			},
		},
	}

	testConfig := `
[storage]
  bucket = "testbucket"
  endpoint = "s3.amazonaws.com"
  region = "us-west-2"
  root = "testcluster"
  type = "S3"

[wal]
  broker_endpoints = ["broker1:9092", "broker2:9092"]
  provider = "kafka"
`

	data, err := FromCluster(testCluster, testCluster.GetDatanode())
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual([]byte(testConfig), data) {
		t.Errorf("generated config is not equal to wanted config:\n, want: %s\n, got: %s\n", testConfig, string(data))
	}
}

func TestFromClusterForDatanodeConfigWithKafkaWALAuthAndTLS(t *testing.T) {
	defer stubGetSecretsData(t, map[string]map[string]string{
		"default/kafka-sasl": {
			"username": "greptime",
			"password": "secret",
		},
	})()

	testCluster := &v1alpha1.GreptimeDBCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-cluster",
			Namespace: "default",
		},
		Spec: v1alpha1.GreptimeDBClusterSpec{
			WALProvider: &v1alpha1.WALProviderSpec{
				KafkaWAL: &v1alpha1.KafkaWAL{
					BrokerEndpoints: []string{"broker1:9096"},
					SASL: &v1alpha1.KafkaSASL{
						Type: "SCRAM-SHA-512",
						SecretRef: &v1alpha1.KafkaSASLSecretRef{
							Name:        "kafka-sasl",
							UsernameKey: "username",
							PasswordKey: "password",
						},
					},
					TLS: &v1alpha1.KafkaTLS{
						ServerCACertPath: "/etc/kafka-tls/ca.crt",
						ClientCertPath:   "/etc/kafka-tls/client.crt",
						ClientKeyPath:    "/etc/kafka-tls/client.key",
					},
				},
			},
		},
	}

	testConfig := `
[wal]
  broker_endpoints = ["broker1:9096"]
  provider = "kafka"

  [wal.sasl]
    password = "secret"
    type = "SCRAM-SHA-512"
    username = "greptime"

  [wal.tls]
    client_cert_path = "/etc/kafka-tls/client.crt"
    client_key_path = "/etc/kafka-tls/client.key"
    server_ca_cert_path = "/etc/kafka-tls/ca.crt"
`

	data, err := FromCluster(testCluster, testCluster.GetDatanode())
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual([]byte(testConfig), data) {
		t.Errorf("generated config is not equal to wanted config:\n, want: %s\n, got: %s\n", testConfig, string(data))
	}
}

func TestFromClusterForDatanodeConfigWithKafkaWALPlaintextSASL(t *testing.T) {
	for _, saslType := range []string{"PLAIN", "SCRAM-SHA-256", "SCRAM-SHA-512"} {
		t.Run(saslType, func(t *testing.T) {
			testCluster := &v1alpha1.GreptimeDBCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cluster",
					Namespace: "default",
				},
				Spec: v1alpha1.GreptimeDBClusterSpec{
					WALProvider: &v1alpha1.WALProviderSpec{
						KafkaWAL: &v1alpha1.KafkaWAL{
							BrokerEndpoints: []string{"broker1:9096"},
							SASL: &v1alpha1.KafkaSASL{
								Type:     saslType,
								Username: "plain-user",
								Password: "plain-secret",
							},
						},
					},
				},
			}

			testConfig := fmt.Sprintf(`
[wal]
  broker_endpoints = ["broker1:9096"]
  provider = "kafka"

  [wal.sasl]
    password = "plain-secret"
    type = "%s"
    username = "plain-user"
`, saslType)

			data, err := FromCluster(testCluster, testCluster.GetDatanode())
			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual([]byte(testConfig), data) {
				t.Errorf("generated config is not equal to wanted config:\n, want: %s\n, got: %s\n", testConfig, string(data))
			}
		})
	}
}

func TestFromClusterForDatanodeConfigWithKafkaWALSecretRefOverridesPlaintextSASL(t *testing.T) {
	defer stubGetSecretsData(t, map[string]map[string]string{
		"default/kafka-sasl": {
			"username": "secret-user",
			"password": "secret-password",
		},
	})()

	testCluster := &v1alpha1.GreptimeDBCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-cluster",
			Namespace: "default",
		},
		Spec: v1alpha1.GreptimeDBClusterSpec{
			WALProvider: &v1alpha1.WALProviderSpec{
				KafkaWAL: &v1alpha1.KafkaWAL{
					BrokerEndpoints: []string{"broker1:9096"},
					SASL: &v1alpha1.KafkaSASL{
						Type:     "SCRAM-SHA-512",
						Username: "plain-user",
						Password: "plain-secret",
						SecretRef: &v1alpha1.KafkaSASLSecretRef{
							Name:        "kafka-sasl",
							UsernameKey: "username",
							PasswordKey: "password",
						},
					},
				},
			},
		},
	}

	testConfig := `
[wal]
  broker_endpoints = ["broker1:9096"]
  provider = "kafka"

  [wal.sasl]
    password = "secret-password"
    type = "SCRAM-SHA-512"
    username = "secret-user"
`

	data, err := FromCluster(testCluster, testCluster.GetDatanode())
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual([]byte(testConfig), data) {
		t.Errorf("generated config is not equal to wanted config:\n, want: %s\n, got: %s\n", testConfig, string(data))
	}
}

func TestFromClusterForMetaConfigWithKafkaWALAuthAndTLS(t *testing.T) {
	defer stubGetSecretsData(t, map[string]map[string]string{
		"default/kafka-sasl": {
			"username": "greptime",
			"password": "secret",
		},
	})()

	testCluster := &v1alpha1.GreptimeDBCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-cluster",
			Namespace: "default",
		},
		Spec: v1alpha1.GreptimeDBClusterSpec{
			WALProvider: &v1alpha1.WALProviderSpec{
				KafkaWAL: &v1alpha1.KafkaWAL{
					BrokerEndpoints: []string{"broker1:9096"},
					SASL: &v1alpha1.KafkaSASL{
						Type: "SCRAM-SHA-512",
						SecretRef: &v1alpha1.KafkaSASLSecretRef{
							Name:        "kafka-sasl",
							UsernameKey: "username",
							PasswordKey: "password",
						},
					},
					TLS: &v1alpha1.KafkaTLS{},
				},
			},
		},
	}

	testConfig := `enable_region_failover = false

[wal]
  broker_endpoints = ["broker1:9096"]
  provider = "kafka"

  [wal.sasl]
    password = "secret"
    type = "SCRAM-SHA-512"
    username = "greptime"
`

	data, err := FromCluster(testCluster, testCluster.GetMeta())
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual([]byte(testConfig), data) {
		t.Errorf("generated config is not equal to wanted config:\n, want: %s\n, got: %s\n", testConfig, string(data))
	}
}

func TestFromClusterRemoteWALConfigHonorsInjectedDataFirst(t *testing.T) {
	defer stubGetSecretsData(t, map[string]map[string]string{
		"default/kafka-sasl": {
			"username": "greptime",
			"password": "secret",
		},
	})()

	testCluster := &v1alpha1.GreptimeDBCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-cluster",
			Namespace: "default",
		},
		Spec: v1alpha1.GreptimeDBClusterSpec{
			ConfigMergeStrategy: v1alpha1.ConfigMergeStrategyInjectedDataFirst,
			WALProvider: &v1alpha1.WALProviderSpec{
				KafkaWAL: &v1alpha1.KafkaWAL{
					BrokerEndpoints: []string{"broker1:9096"},
					SASL: &v1alpha1.KafkaSASL{
						Type: "SCRAM-SHA-512",
						SecretRef: &v1alpha1.KafkaSASLSecretRef{
							Name:        "kafka-sasl",
							UsernameKey: "username",
							PasswordKey: "password",
						},
					},
				},
			},
			Datanode: &v1alpha1.DatanodeSpec{
				ComponentSpec: v1alpha1.ComponentSpec{
					Config: `[wal]
provider = "raft_engine"
broker_endpoints = ["broker2:9096"]

[wal.sasl]
type = "PLAIN"
username = "other"
password = "other-secret"
`,
				},
			},
		},
	}

	testConfig := `
[wal]
  broker_endpoints = ["broker2:9096"]
  provider = "raft_engine"

  [wal.sasl]
    password = "other-secret"
    type = "PLAIN"
    username = "other"
`

	data, err := FromCluster(testCluster, testCluster.GetDatanode())
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual([]byte(testConfig), data) {
		t.Errorf("generated config is not equal to wanted config:\n, want: %s\n, got: %s\n", testConfig, string(data))
	}
}

func TestFromClusterRemoteWALConfigHonorsOperatorFirst(t *testing.T) {
	defer stubGetSecretsData(t, map[string]map[string]string{
		"default/kafka-sasl": {
			"username": "greptime",
			"password": "secret",
		},
	})()

	testCluster := &v1alpha1.GreptimeDBCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-cluster",
			Namespace: "default",
		},
		Spec: v1alpha1.GreptimeDBClusterSpec{
			ConfigMergeStrategy: v1alpha1.ConfigMergeStrategyOperatorFirst,
			WALProvider: &v1alpha1.WALProviderSpec{
				KafkaWAL: &v1alpha1.KafkaWAL{
					BrokerEndpoints: []string{"broker1:9096"},
					SASL: &v1alpha1.KafkaSASL{
						Type: "SCRAM-SHA-512",
						SecretRef: &v1alpha1.KafkaSASLSecretRef{
							Name:        "kafka-sasl",
							UsernameKey: "username",
							PasswordKey: "password",
						},
					},
				},
			},
			Datanode: &v1alpha1.DatanodeSpec{
				ComponentSpec: v1alpha1.ComponentSpec{
					Config: `[wal]
provider = "raft_engine"
broker_endpoints = ["broker2:9096"]

[wal.sasl]
type = "PLAIN"
username = "other"
password = "other-secret"
`,
				},
			},
		},
	}

	testConfig := `
[wal]
  broker_endpoints = ["broker1:9096"]
  provider = "kafka"

  [wal.sasl]
    password = "secret"
    type = "SCRAM-SHA-512"
    username = "greptime"
`

	data, err := FromCluster(testCluster, testCluster.GetDatanode())
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual([]byte(testConfig), data) {
		t.Errorf("generated config is not equal to wanted config:\n, want: %s\n, got: %s\n", testConfig, string(data))
	}
}

func TestFromClusterRemoteWALConfigMergeStrategyResolvesConflictingConfigData(t *testing.T) {
	defer stubGetSecretsData(t, map[string]map[string]string{
		"default/kafka-sasl": {
			"username": "greptime",
			"password": "secret",
		},
	})()

	newTestCluster := func(strategy v1alpha1.ConfigMergeStrategy) *v1alpha1.GreptimeDBCluster {
		return &v1alpha1.GreptimeDBCluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-cluster",
				Namespace: "default",
			},
			Spec: v1alpha1.GreptimeDBClusterSpec{
				ConfigMergeStrategy: strategy,
				WALProvider: &v1alpha1.WALProviderSpec{
					KafkaWAL: &v1alpha1.KafkaWAL{
						BrokerEndpoints: []string{"spec-broker:9096"},
						SASL: &v1alpha1.KafkaSASL{
							Type: "SCRAM-SHA-512",
							SecretRef: &v1alpha1.KafkaSASLSecretRef{
								Name:        "kafka-sasl",
								UsernameKey: "username",
								PasswordKey: "password",
							},
						},
					},
				},
				Datanode: &v1alpha1.DatanodeSpec{
					ComponentSpec: v1alpha1.ComponentSpec{
						Config: `[wal]
provider = "raft_engine"
broker_endpoints = ["config-broker:9096"]

[wal.sasl]
type = "PLAIN"
username = "config-user"
password = "config-secret"
`,
					},
				},
			},
		}
	}

	tests := []struct {
		name     string
		strategy v1alpha1.ConfigMergeStrategy
		want     string
	}{
		{
			name:     "injected data first keeps configData wal",
			strategy: v1alpha1.ConfigMergeStrategyInjectedDataFirst,
			want: `
[wal]
  broker_endpoints = ["config-broker:9096"]
  provider = "raft_engine"

  [wal.sasl]
    password = "config-secret"
    type = "PLAIN"
    username = "config-user"
`,
		},
		{
			name:     "operator first keeps remote wal spec",
			strategy: v1alpha1.ConfigMergeStrategyOperatorFirst,
			want: `
[wal]
  broker_endpoints = ["spec-broker:9096"]
  provider = "kafka"

  [wal.sasl]
    password = "secret"
    type = "SCRAM-SHA-512"
    username = "greptime"
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testCluster := newTestCluster(tt.strategy)

			data, err := FromCluster(testCluster, testCluster.GetDatanode())
			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual([]byte(tt.want), data) {
				t.Errorf("generated config is not equal to wanted config:\n, want: %s\n, got: %s\n", tt.want, string(data))
			}
		})
	}
}

func TestFromClusterForDatanodeConfigWithExtraConfig(t *testing.T) {
	extraConfig := `[extra]
key1 = 'value1'
key2 = 'value2'
`
	testCluster := &v1alpha1.GreptimeDBCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-cluster",
			Namespace: "default",
		},
		Spec: v1alpha1.GreptimeDBClusterSpec{
			ObjectStorageProvider: &v1alpha1.ObjectStorageProviderSpec{
				S3: &v1alpha1.S3Storage{
					Root:     "testcluster",
					Bucket:   "testbucket",
					Endpoint: "s3.amazonaws.com",
					Region:   "us-west-2",
				},
			},
			Datanode: &v1alpha1.DatanodeSpec{
				ComponentSpec: v1alpha1.ComponentSpec{
					Config: extraConfig,
					Logging: &v1alpha1.LoggingSpec{
						Level:   v1alpha1.LoggingLevelInfo,
						LogsDir: "/data/greptimedb/logs",
						Format:  v1alpha1.LogFormatJSON,
					},
				},
			},
		},
	}

	testConfig := `
[extra]
  key1 = "value1"
  key2 = "value2"

[logging]
  dir = "/data/greptimedb/logs"
  level = "info"
  log_format = "json"

[storage]
  bucket = "testbucket"
  endpoint = "s3.amazonaws.com"
  region = "us-west-2"
  root = "testcluster"
  type = "S3"
`

	data, err := FromCluster(testCluster, testCluster.GetDatanode())
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual([]byte(testConfig), data) {
		t.Errorf("generated config is not equal to wanted config:\n, want: %s\n, got: %s\n", testConfig, string(data))
	}
}

func TestMergeConfigWithStrategy(t *testing.T) {
	inputConfig := `
enable_region_failover = true
`

	// Use ConfigMergeStrategyInjectedDataFirst to test the merge config logic.
	testCluster := &v1alpha1.GreptimeDBCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-cluster",
			Namespace: "default",
		},
		Spec: v1alpha1.GreptimeDBClusterSpec{
			ConfigMergeStrategy: v1alpha1.ConfigMergeStrategyInjectedDataFirst,
			Meta: &v1alpha1.MetaSpec{
				EnableRegionFailover: ptr.To(false),
				ComponentSpec: v1alpha1.ComponentSpec{
					Config: inputConfig,
				},
			},
		},
	}

	data, err := FromCluster(testCluster, testCluster.GetMeta())
	if err != nil {
		t.Fatal(err)
	}

	// The input config will override the config that generated by the operator.
	expectedConfig := `enable_region_failover = true
`
	if !reflect.DeepEqual([]byte(expectedConfig), data) {
		t.Errorf("generated config is not equal to wanted config:\n, want: '%s'\n, got: '%s'\n", expectedConfig, string(data))
	}

	// Use ConfigMergeStrategyOperatorFirst to test the merge config logic.
	testCluster.Spec.ConfigMergeStrategy = v1alpha1.ConfigMergeStrategyOperatorFirst
	data, err = FromCluster(testCluster, testCluster.GetMeta())
	if err != nil {
		t.Fatal(err)
	}

	// The config that generated by the operator will override the input config.
	expectedConfig = `enable_region_failover = false
`
	if !reflect.DeepEqual([]byte(expectedConfig), data) {
		t.Errorf("generated config is not equal to wanted config:\n, want: '%s'\n, got: '%s'\n", expectedConfig, string(data))
	}
}
