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
	testFromFile("testdata/metasrv.toml", &MetasrvConfig{}, t)
}

func TestFrontendConfigFromFile(t *testing.T) {
	testFromFile("testdata/frontend.toml", &FrontendConfig{}, t)
}

func TestDatanodeConfigFromFile(t *testing.T) {
	testFromFile("testdata/datanode.toml", &DatanodeConfig{}, t)
}

func TestDatanodeFromClusterCRD(t *testing.T) {
	testCluster := &v1alpha1.GreptimeDBCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-cluster",
			Namespace: "default",
		},
		Spec: v1alpha1.GreptimeDBClusterSpec{
			StorageProvider: &v1alpha1.StorageProvider{
				S3: &v1alpha1.S3StorageProvider{
					Prefix: "testcluster",
					Bucket: "testbucket",
				},
			},
		},
	}

	testConfig := `
	[storage]
	type = 'S3'
	bucket = 'testbucket'
	root = 'testcluster'`

	cfg := &DatanodeConfig{}
	if err := FromClusterCRD(testCluster, cfg); err != nil {
		t.Fatal(err)
	}

	if _, err := Marshal(cfg); err != nil {
		t.Fatal(err)
	}

	wantedCfg := &DatanodeConfig{}
	if err := FromRawData([]byte(testConfig), wantedCfg); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(wantedCfg, cfg) {
		t.Errorf("generated config is not equal to wanted config:\n, want: %v\n, got: %v\n", wantedCfg, cfg)
	}
}

func TestMetasrvConfigMerge(t *testing.T) {
	var extraInput = `
[logging]
dir = '/other/dir'
level = 'error'
`

	extraCfg := &MetasrvConfig{}
	if err := FromRawData([]byte(extraInput), extraCfg); err != nil {
		t.Fatal(err)
	}

	baseCfg := &MetasrvConfig{}
	if err := FromFile("testdata/metasrv.toml", baseCfg); err != nil {
		t.Fatal(err)
	}

	if err := Merge([]byte(extraInput), baseCfg); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(baseCfg.Logging.Level, extraCfg.Logging.Level) {
		t.Errorf("logging.level is not equal: want %s, got %s", extraCfg.Logging.Level, baseCfg.Logging.Level)
	}

	if !reflect.DeepEqual(baseCfg.Logging.Dir, extraCfg.Logging.Dir) {
		t.Errorf("logging.dir is not equal: want %s, got %s", extraCfg.Logging.Dir, baseCfg.Logging.Dir)
	}
}

func TestFrontendConfigMerge(t *testing.T) {
	var extraInput = `
[logging]
dir = '/other/dir'
level = 'error'
`

	extraCfg := &FrontendConfig{}
	if err := FromRawData([]byte(extraInput), extraCfg); err != nil {
		t.Fatal(err)
	}

	baseCfg := &FrontendConfig{}
	if err := FromFile("testdata/frontend.toml", baseCfg); err != nil {
		t.Fatal(err)
	}

	if err := Merge([]byte(extraInput), baseCfg); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(baseCfg.Logging.Level, extraCfg.Logging.Level) {
		t.Errorf("logging.level is not equal: want %s, got %s", extraCfg.Logging.Level, baseCfg.Logging.Level)
	}

	if !reflect.DeepEqual(baseCfg.Logging.Dir, extraCfg.Logging.Dir) {
		t.Errorf("logging.dir is not equal: want %s, got %s", extraCfg.Logging.Dir, baseCfg.Logging.Dir)
	}
}

func TestDatanodeConfigMerge(t *testing.T) {
	var extraInput = `
[logging]
dir = '/other/dir'
level = 'error'
`

	extraCfg := &DatanodeConfig{}
	if err := FromRawData([]byte(extraInput), extraCfg); err != nil {
		t.Fatal(err)
	}

	baseCfg := &DatanodeConfig{}
	if err := FromFile("testdata/datanode.toml", baseCfg); err != nil {
		t.Fatal(err)
	}

	if err := Merge([]byte(extraInput), baseCfg); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(baseCfg.Logging.Level, extraCfg.Logging.Level) {
		t.Errorf("logging.level is not equal: want %s, got %s", extraCfg.Logging.Level, baseCfg.Logging.Level)
	}

	if !reflect.DeepEqual(baseCfg.Logging.Dir, extraCfg.Logging.Dir) {
		t.Errorf("logging.dir is not equal: want %s, got %s", extraCfg.Logging.Dir, baseCfg.Logging.Dir)
	}
}

func testFromFile(filename string, config interface{}, t *testing.T) {
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}

	if err := FromFile(filename, config); err != nil {
		t.Fatal(err)
	}

	output, err := Marshal(config)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(data, output) {
		t.Errorf("generated config is not equal to original config:\n, want: %s\n, got: %s\n",
			string(data), string(output))
	}
}
