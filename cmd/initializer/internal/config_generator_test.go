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

	"github.com/pelletier/go-toml"

	"github.com/GreptimeTeam/greptimedb-operator/apis/v1alpha1"
	"github.com/GreptimeTeam/greptimedb-operator/pkg/deployer"
)

var (
	testPodIndex uint64 = 1
	testRPCPort  int32  = 4001

	testPodIP            = "192.168.0.1"
	testClusterService   = "testcluster"
	testClusterNamespace = "greptimedb"
	testDatanodePodName  = fmt.Sprintf("%s-datanode-%d", testClusterService, testPodIndex)
	testFlownodePodName  = fmt.Sprintf("%s-flownode-%d", testClusterService, testPodIndex)
)

func TestDatanodeConfigGenerator(t *testing.T) {
	file, err := os.CreateTemp("", "config-*.toml")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	opts := &Options{
		ConfigPath:          file.Name(),
		InitConfigPath:      "testdata/datanode-config.toml",
		Namespace:           testClusterNamespace,
		ComponentKind:       string(v1alpha1.DatanodeComponentKind),
		DatanodeRPCPort:     testRPCPort,
		DatanodeServiceName: testClusterService,
	}

	t.Setenv(deployer.EnvPodIP, testPodIP)
	t.Setenv(deployer.EnvPodName, testDatanodePodName)

	cg := NewConfigGenerator(opts, datanodeHostname)
	if err = cg.Generate(); err != nil {
		t.Fatal(err)
	}

	tomlData, err := os.ReadFile(opts.ConfigPath)
	if err != nil {
		t.Fatal(err)
	}
	tree, err := toml.Load(string(tomlData))
	if err != nil {
		t.Fatal(err)
	}

	nodeID, ok := tree.Get("node_id").(int64)
	if !ok {
		t.Fatalf("node_id is not int64")
	}

	if !reflect.DeepEqual(testPodIndex, uint64(nodeID)) {
		t.Fatalf("nodeID is not equal, want: '%d', got: '%d'", testPodIndex, nodeID)
	}

	rpcAddr, ok := tree.Get("rpc_addr").(string)
	if !ok {
		t.Fatalf("rpc_addr is not string")
	}
	wantRPCAddr := fmt.Sprintf("%s:%d", testPodIP, testRPCPort)
	if !reflect.DeepEqual(wantRPCAddr, rpcAddr) {
		t.Fatalf("RPCAddr is not equal, want: '%s', got: '%s'", wantRPCAddr, rpcAddr)
	}

	rpcHostName, ok := tree.Get("rpc_hostname").(string)
	if !ok {
		t.Fatalf("rpc_hostname is not string")
	}
	wantRPCHostname := fmt.Sprintf("%s.%s.%s:%d", testDatanodePodName, testClusterService, testClusterNamespace, testRPCPort)
	if !reflect.DeepEqual(wantRPCHostname, rpcHostName) {
		t.Fatalf("RPCHostName is not equal, want: '%s', got: '%s'", wantRPCHostname, rpcHostName)
	}
}

func TestFlownodeConfigGenerator(t *testing.T) {
	file, err := os.CreateTemp("", "config-*.toml")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	opts := &Options{
		ConfigPath:     file.Name(),
		InitConfigPath: "testdata/flownode-config.toml",
		Namespace:      testClusterNamespace,
		ComponentKind:  string(v1alpha1.FlownodeComponentKind),
		RPCPort:        testRPCPort,
		ServiceName:    testClusterService,
	}

	t.Setenv(deployer.EnvPodIP, testPodIP)
	t.Setenv(deployer.EnvPodName, testFlownodePodName)

	cg := NewConfigGenerator(opts, flownodeHostname)
	if err = cg.Generate(); err != nil {
		t.Fatal(err)
	}

	tomlData, err := os.ReadFile(opts.ConfigPath)
	if err != nil {
		t.Fatal(err)
	}
	tree, err := toml.Load(string(tomlData))
	if err != nil {
		t.Fatal(err)
	}

	nodeID, ok := tree.Get("node_id").(int64)
	if !ok {
		t.Fatalf("node_id is not int64")
	}

	if !reflect.DeepEqual(testPodIndex, uint64(nodeID)) {
		t.Fatalf("nodeID is not equal, want: '%d', got: '%d'", testPodIndex, nodeID)
	}

	rpcAddr, ok := tree.Get("grpc.addr").(string)
	if !ok {
		t.Fatalf("rpc_addr is not string")
	}
	wantRPCAddr := fmt.Sprintf("%s:%d", testPodIP, testRPCPort)
	if !reflect.DeepEqual(wantRPCAddr, rpcAddr) {
		t.Fatalf("RPCAddr is not equal, want: '%s', got: '%s'", wantRPCAddr, rpcAddr)
	}

	rpcHostName, ok := tree.Get("grpc.hostname").(string)
	if !ok {
		t.Fatalf("rpc_hostname is not string")
	}
	wantRPCHostname := fmt.Sprintf("%s.%s.%s:%d", testFlownodePodName, testClusterService, testClusterNamespace, testRPCPort)
	if !reflect.DeepEqual(wantRPCHostname, rpcHostName) {
		t.Fatalf("RPCHostName is not equal, want: '%s', got: '%s'", wantRPCHostname, rpcHostName)
	}
}

func datanodeHostname() (name string, err error) {
	return testDatanodePodName, nil
}

func flownodeHostname() (name string, err error) {
	return testFlownodePodName, nil
}
