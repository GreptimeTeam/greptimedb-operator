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
	"os"
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/GreptimeTeam/greptimedb-operator/apis/v1alpha1"
)

func TestMetasrvConfigFromFile(t *testing.T) {
	testFromFile("testdata/metasrv.toml", v1alpha1.MetaComponentKind, t)
}

func TestFrontendConfigFromFile(t *testing.T) {
	testFromFile("testdata/frontend.toml", v1alpha1.FrontendComponentKind, t)
}

func TestDatanodeConfigFromFile(t *testing.T) {
	testFromFile("testdata/datanode.toml", v1alpha1.DatanodeComponentKind, t)
}

func TestFromClusterForDatanodeConfig(t *testing.T) {
	testCluster := &v1alpha1.GreptimeDBCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-cluster",
			Namespace: "default",
		},
		Spec: v1alpha1.GreptimeDBClusterSpec{
			StorageProvider: &v1alpha1.StorageProvider{
				S3: &v1alpha1.S3StorageProvider{
					Root:   "testcluster",
					Bucket: "testbucket",
				},
			},
		},
	}

	testConfig := `[storage]
type = 'S3'
bucket = 'testbucket'
root = 'testcluster'
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
			StorageProvider: &v1alpha1.StorageProvider{
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

	testConfig := `[storage]
type = 'S3'
bucket = 'testbucket'
root = 'testcluster'

[logging]
dir = '/other/dir'
level = 'error'
`

	data, err := FromCluster(testCluster, v1alpha1.DatanodeComponentKind)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual([]byte(testConfig), data) {
		t.Errorf("generated config is not equal to wanted config:\n, want: %s\n, got: %s\n", testConfig, string(data))
	}
}

func TestMerge(t *testing.T) {
	extraInput := `
[logging]
dir = '/other/dir'
level = 'error'
`
	extraCfg, err := FromRawData([]byte(extraInput), v1alpha1.FrontendComponentKind)
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := FromFile("testdata/frontend.toml", v1alpha1.FrontendComponentKind)
	if err != nil {
		t.Fatal(err)
	}

	if err := Merge([]byte(extraInput), cfg); err != nil {
		t.Fatal(err)
	}

	frontendCfg := cfg.(*FrontendConfig)
	frontendExtraCfg := extraCfg.(*FrontendConfig)
	if !reflect.DeepEqual(frontendCfg.Logging.Level, frontendExtraCfg.Logging.Level) {
		t.Errorf("logging.level is not equal: want %s, got %s", frontendExtraCfg.Logging.Level, frontendCfg.Logging.Level)
	}

	if !reflect.DeepEqual(frontendCfg.Logging.Dir, frontendExtraCfg.Logging.Dir) {
		t.Errorf("logging.dir is not equal: want %s, got %s", frontendExtraCfg.Logging.Dir, frontendCfg.Logging.Dir)
	}
}

func testFromFile(filename string, kind v1alpha1.ComponentKind, t *testing.T) {
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := FromFile(filename, kind)
	if err != nil {
		t.Fatal(err)
	}

	output, err := Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(data, output) {
		t.Errorf("generated config is not equal to original config:\n, want: %s\n, got: %s\n",
			string(data), string(output))
	}
}
