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

package internal

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"testing"

	"github.com/GreptimeTeam/greptimedb-operator/apis/v1alpha1"
	"github.com/GreptimeTeam/greptimedb-operator/pkg/dbconfig"
)

var (
	testPodIndex uint64 = 1
	testRPCPort  int32  = 4001

	testPodIP            = "192.168.0.1"
	testClusterService   = "testcluster"
	testClusterNamespace = "greptimedb"
	testPodName          = fmt.Sprintf("%s-datanode-%d", testClusterService, testPodIndex)
)

func TestConfigGenerator(t *testing.T) {

	file, err := os.CreateTemp("", "config-*.toml")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	opts := &Options{
		ConfigPath:          file.Name(),
		InitConfigPath:      "testdata/config.toml",
		Namespace:           testClusterNamespace,
		ComponentKind:       string(v1alpha1.DatanodeComponentKind),
		DatanodeRPCPort:     testRPCPort,
		DatanodeServiceName: testClusterService,
	}

	t.Setenv("POD_IP", testPodIP)
	t.Setenv("POD_NAME", testPodName)

	cg := NewConfigGenerator(opts, hostname)
	if err = cg.Generate(); err != nil {
		t.Fatal(err)
	}

	var cfg dbconfig.DatanodeConfig
	if err := dbconfig.FromFile(opts.ConfigPath, &cfg); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(testPodIndex, *cfg.NodeID) {
		t.Fatalf("nodeID is not equal, want: '%d', got: '%d'", testPodIndex, cfg.NodeID)
	}

	wantRPCAddr := fmt.Sprintf("%s:%d", testPodIP, testRPCPort)
	if !reflect.DeepEqual(wantRPCAddr, cfg.RPCAddr) {
		t.Fatalf("RPCAddr is not equal, want: '%s', got: '%s'", wantRPCAddr, cfg.RPCAddr)
	}

	wantRPCHostname := fmt.Sprintf("%s.%s.%s:%d", testPodName, testClusterService, testClusterNamespace, testRPCPort)
	if !reflect.DeepEqual(wantRPCHostname, cfg.RPCHostName) {
		t.Fatalf("RPCHostName is not equal, want: '%s', got: '%s'", wantRPCHostname, cfg.RPCHostName)
	}
}

func hostname() (name string, err error) {
	return testPodName, nil
}
