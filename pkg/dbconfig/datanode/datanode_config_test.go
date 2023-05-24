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

package datanode

import (
	"io/ioutil"
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/GreptimeTeam/greptimedb-operator/apis/v1alpha1"
)

func TestDatanodeConfigFromRawData(t *testing.T) {
	input, err := ioutil.ReadFile("../testdata/datanode.toml")
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := FromRawData(input)
	if err != nil {
		t.Fatal(err)
	}

	ouput, err := cfg.Marshal()
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(input, ouput) {
		t.Errorf("generated config is not equal to original config:\n, want: %s\n, got: %s\n",
			string(input), string(ouput))
	}
}

func TestDatanodeConfigFromClusterCRD(t *testing.T) {
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

	cfg, err := FromClusterCRD(testCluster)
	if err != nil {
		t.Fatal(err)
	}

	_, err = cfg.Marshal()
	if err != nil {
		t.Fatal(err)
	}

	wantedCfg, err := FromRawData([]byte(testConfig))
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(wantedCfg, cfg) {
		t.Errorf("generated config is not equal to wanted config:\n, want: %v\n, got: %v\n", wantedCfg, cfg)
	}
}
