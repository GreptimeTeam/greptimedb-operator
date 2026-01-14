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
		ConfigPath:     file.Name(),
		InitConfigPath: "testdata/datanode-config.toml",
		Namespace:      testClusterNamespace,
		RoleKind:       string(v1alpha1.DatanodeRoleKind),
		RPCPort:        testRPCPort,
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

	rpcBindAddr, ok := tree.Get("grpc.bind_addr").(string)
	if !ok {
		t.Fatalf("grpc bind_addr is not string")
	}
	wantRPCAddr := fmt.Sprintf("%s:%d", testPodIP, testRPCPort)
	if !reflect.DeepEqual(wantRPCAddr, rpcBindAddr) {
		t.Fatalf("RPCBindAddr is not equal, want: '%s', got: '%s'", wantRPCAddr, rpcBindAddr)
	}

	rpcServerAddr, ok := tree.Get("grpc.server_addr").(string)
	if !ok {
		t.Fatalf("grpc server_addr is not string")
	}
	wantRPCServerAddr := fmt.Sprintf("%s:%d", testPodIP, testRPCPort)
	if !reflect.DeepEqual(wantRPCServerAddr, rpcServerAddr) {
		t.Fatalf("RPCServerAddr is not equal, want: '%s', got: '%s'", wantRPCServerAddr, rpcServerAddr)
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
		RoleKind:       string(v1alpha1.FlownodeRoleKind),
		RPCPort:        testRPCPort,
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

	rpcBindAddr, ok := tree.Get("grpc.bind_addr").(string)
	if !ok {
		t.Fatalf("grpc bind_addr is not string")
	}
	wantRPCAddr := fmt.Sprintf("%s:%d", testPodIP, testRPCPort)
	if !reflect.DeepEqual(wantRPCAddr, rpcBindAddr) {
		t.Fatalf("RPCBindAddr is not equal, want: '%s', got: '%s'", wantRPCAddr, rpcBindAddr)
	}

	rpcServerAddr, ok := tree.Get("grpc.server_addr").(string)
	if !ok {
		t.Fatalf("grpc server_addr is not string")
	}
	wantRPCServerAddr := fmt.Sprintf("%s:%d", testPodIP, testRPCPort)
	if !reflect.DeepEqual(wantRPCServerAddr, rpcServerAddr) {
		t.Fatalf("RPCServerAddr is not equal, want: '%s', got: '%s'", wantRPCServerAddr, rpcServerAddr)
	}
}

func TestDatanodeConfigGeneratorWithDatanodeGroupID(t *testing.T) {
	var testDatanodeGroupID int32 = 42

	tmpConfigFile, err := os.CreateTemp("", "config-*.toml")
	if err != nil {
		log.Fatal(err)
	}
	defer tmpConfigFile.Close()

	opts := &Options{
		ConfigPath:      tmpConfigFile.Name(),
		InitConfigPath:  "testdata/datanode-config.toml",
		Namespace:       testClusterNamespace,
		RoleKind:        string(v1alpha1.DatanodeRoleKind),
		RPCPort:         testRPCPort,
		DatanodeGroupID: testDatanodeGroupID,
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

	expectedNodeID := uint64(testDatanodeGroupID)<<32 | uint64(testPodIndex)

	if !reflect.DeepEqual(expectedNodeID, uint64(nodeID)) {
		t.Fatalf("nodeID is not equal, want: '%d', got: '%d'", expectedNodeID, nodeID)
	}
}

func datanodeHostname() (name string, err error) {
	return testDatanodePodName, nil
}

func flownodeHostname() (name string, err error) {
	return testFlownodePodName, nil
}

func TestDatanodeConfigGeneratorWithStartNodeID(t *testing.T) {
	var testStartNodeID int32 = 10

	tmpConfigFile, err := os.CreateTemp("", "config-*.toml")
	if err != nil {
		log.Fatal(err)
	}
	defer tmpConfigFile.Close()

	opts := &Options{
		ConfigPath:     tmpConfigFile.Name(),
		InitConfigPath: "testdata/datanode-config.toml",
		Namespace:      testClusterNamespace,
		RoleKind:       string(v1alpha1.DatanodeRoleKind),
		RPCPort:        testRPCPort,
		StartNodeID:    testStartNodeID,
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

	// expectedNodeID = StartNodeID + PodIndex = 10 + 1 = 11
	expectedNodeID := uint64(testStartNodeID) + testPodIndex

	if !reflect.DeepEqual(expectedNodeID, uint64(nodeID)) {
		t.Fatalf("nodeID is not equal, want: '%d', got: '%d'", expectedNodeID, nodeID)
	}
}

func TestDatanodeConfigGeneratorWithNotStartNodeID(t *testing.T) {
	var testStartNodeID int32 = 0

	tmpConfigFile, err := os.CreateTemp("", "config-*.toml")
	if err != nil {
		log.Fatal(err)
	}
	defer tmpConfigFile.Close()

	opts := &Options{
		ConfigPath:     tmpConfigFile.Name(),
		InitConfigPath: "testdata/datanode-config.toml",
		Namespace:      testClusterNamespace,
		RoleKind:       string(v1alpha1.DatanodeRoleKind),
		RPCPort:        testRPCPort,
		StartNodeID:    testStartNodeID,
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

	// expectedNodeID = StartNodeID + PodIndex = 0 + 1 = 1
	expectedNodeID := uint64(testStartNodeID) + testPodIndex

	if !reflect.DeepEqual(expectedNodeID, uint64(nodeID)) {
		t.Fatalf("nodeID is not equal, want: '%d', got: '%d'", expectedNodeID, nodeID)
	}
}

func TestFlownodeConfigGeneratorWithStartNodeID(t *testing.T) {
	var testStartNodeID int32 = 99

	tmpConfigFile, err := os.CreateTemp("", "config-*.toml")
	if err != nil {
		log.Fatal(err)
	}
	defer tmpConfigFile.Close()

	opts := &Options{
		ConfigPath:     tmpConfigFile.Name(),
		InitConfigPath: "testdata/flownode-config.toml",
		Namespace:      testClusterNamespace,
		RoleKind:       string(v1alpha1.FlownodeRoleKind),
		RPCPort:        testRPCPort,
		StartNodeID:    testStartNodeID,
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

	// expectedNodeID = StartNodeID + PodIndex = 99 + 1 = 100
	expectedNodeID := uint64(testStartNodeID) + testPodIndex

	if !reflect.DeepEqual(expectedNodeID, uint64(nodeID)) {
		t.Fatalf("nodeID is not equal, want: '%d', got: '%d'", expectedNodeID, nodeID)
	}
}
