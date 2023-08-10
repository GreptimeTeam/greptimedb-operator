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
	"github.com/GreptimeTeam/greptimedb-operator/pkg/deployer"
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

	t.Setenv(deployer.EnvPodIP, testPodIP)
	t.Setenv(deployer.EnvPodName, testPodName)

	cg := NewConfigGenerator(opts, hostname)
	if err = cg.Generate(); err != nil {
		t.Fatal(err)
	}

	cfg, err := dbconfig.FromFile(opts.ConfigPath, v1alpha1.DatanodeComponentKind)
	if err != nil {
		t.Fatal(err)
	}

	datanodeCfg, ok := cfg.(*dbconfig.DatanodeConfig)
	if !ok {
		t.Fatal(fmt.Errorf("invalid config type"))
	}

	if !reflect.DeepEqual(testPodIndex, *datanodeCfg.NodeID) {
		t.Fatalf("nodeID is not equal, want: '%d', got: '%d'", testPodIndex, datanodeCfg.NodeID)
	}

	wantRPCAddr := fmt.Sprintf("%s:%d", testPodIP, testRPCPort)
	if !reflect.DeepEqual(wantRPCAddr, datanodeCfg.RPCAddr) {
		t.Fatalf("RPCAddr is not equal, want: '%s', got: '%s'", wantRPCAddr, datanodeCfg.RPCAddr)
	}

	wantRPCHostname := fmt.Sprintf("%s.%s.%s:%d", testPodName, testClusterService, testClusterNamespace, testRPCPort)
	if !reflect.DeepEqual(wantRPCHostname, datanodeCfg.RPCHostName) {
		t.Fatalf("RPCHostName is not equal, want: '%s', got: '%s'", wantRPCHostname, datanodeCfg.RPCHostName)
	}
}

func hostname() (name string, err error) {
	return testPodName, nil
}
