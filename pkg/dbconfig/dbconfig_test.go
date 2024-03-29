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
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/GreptimeTeam/greptimedb-operator/apis/v1alpha1"
)

func TestFromClusterForDatanodeConfig(t *testing.T) {
	testCluster := &v1alpha1.GreptimeDBCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-cluster",
			Namespace: "default",
		},
		Spec: v1alpha1.GreptimeDBClusterSpec{
			ObjectStorageProvider: &v1alpha1.ObjectStorageProvider{
				S3: &v1alpha1.S3StorageProvider{
					Root:   "testcluster",
					Bucket: "testbucket",
				},
			},
		},
	}

	testConfig := `
[storage]
  bucket = "testbucket"
  root = "testcluster"
  type = "S3"
`

	data, err := FromCluster(testCluster, v1alpha1.DatanodeComponentKind)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual([]byte(testConfig), data) {
		t.Errorf("generated config is not equal to wanted config:\n, want: %s\n, got: %s\n", testConfig, string(data))
	}
}

func TestFromClusterForDatanodeConfigWithExtraConfig(t *testing.T) {
	extraConfig := `[logging]
dir = '/other/dir'
level = 'error'
`
	testCluster := &v1alpha1.GreptimeDBCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-cluster",
			Namespace: "default",
		},
		Spec: v1alpha1.GreptimeDBClusterSpec{
			ObjectStorageProvider: &v1alpha1.ObjectStorageProvider{
				S3: &v1alpha1.S3StorageProvider{
					Root:   "testcluster",
					Bucket: "testbucket",
				},
			},
			Datanode: &v1alpha1.DatanodeSpec{
				ComponentSpec: v1alpha1.ComponentSpec{
					Config: extraConfig,
				},
			},
		},
	}

	testConfig := `
[logging]
  dir = "/other/dir"
  level = "error"

[storage]
  bucket = "testbucket"
  root = "testcluster"
  type = "S3"
`

	data, err := FromCluster(testCluster, v1alpha1.DatanodeComponentKind)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual([]byte(testConfig), data) {
		t.Errorf("generated config is not equal to wanted config:\n, want: %s\n, got: %s\n", testConfig, string(data))
	}
}
