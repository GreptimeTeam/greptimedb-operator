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
	"os"
	"strconv"
	"strings"

	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"

	"github.com/GreptimeTeam/greptimedb-operator/apis/v1alpha1"
	"github.com/GreptimeTeam/greptimedb-operator/pkg/dbconfig"
	"github.com/GreptimeTeam/greptimedb-operator/pkg/deployer"
)

type Options struct {
	ConfigPath     string
	InitConfigPath string
	Namespace      string
	ComponentKind  string

	// For generating config of datanode or flownode.
	RPCPort     int32
	ServiceName string

	// Note: It's Deprecated and will be removed soon. For generating config of datanode.
	DatanodeRPCPort     int32
	DatanodeServiceName string
}

type ConfigGenerator struct {
	*Options

	hostname func() (name string, err error)
}

func NewConfigGenerator(opts *Options, hostname func() (name string, err error)) *ConfigGenerator {
	return &ConfigGenerator{
		Options:  opts,
		hostname: hostname,
	}
}

// Generate generates the final config of datanode.
func (c *ConfigGenerator) Generate() error {
	initConfig, err := os.ReadFile(c.InitConfigPath)
	if err != nil {
		return err
	}

	var configData []byte
	switch c.ComponentKind {
	case string(v1alpha1.DatanodeComponentKind):
		configData, err = c.generateDatanodeConfig(initConfig)
		if err != nil {
			return err
		}
	case string(v1alpha1.FlownodeComponentKind):
		configData, err = c.generateFlownodeConfig(initConfig)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown component kind: %s", c.ComponentKind)
	}

	if err := os.WriteFile(c.ConfigPath, configData, 0644); err != nil {
		return err
	}

	klog.Infof("Generate config successfully, config path: %s", c.ConfigPath)
	klog.Infof("The config content is: \n%s", string(configData))

	return nil
}

func (c *ConfigGenerator) generateDatanodeConfig(initConfig []byte) ([]byte, error) {
	cfg, err := dbconfig.NewFromComponentKind(v1alpha1.DatanodeComponentKind)
	if err != nil {
		return nil, err
	}

	if err := cfg.SetInputConfig(string(initConfig)); err != nil {
		return nil, err
	}

	datanodeCfg, ok := cfg.(*dbconfig.DatanodeConfig)
	if !ok {
		return nil, fmt.Errorf("cfg is not datanode config")
	}

	nodeID, err := c.allocateNodeID()
	if err != nil {
		klog.Fatalf("Allocate node id failed: %v", err)
	}
	datanodeCfg.NodeID = &nodeID

	podIP := os.Getenv(deployer.EnvPodIP)
	if len(podIP) == 0 {
		return nil, fmt.Errorf("empty pod ip")
	}
	datanodeCfg.RPCBindAddr = ptr.To(fmt.Sprintf("%s:%d", podIP, c.DatanodeRPCPort))

	podName := os.Getenv(deployer.EnvPodName)
	if len(podName) == 0 {
		return nil, fmt.Errorf("empty pod name")
	}

	datanodeCfg.RPCServerAddr = ptr.To(fmt.Sprintf("%s.%s.%s:%d", podName,
		c.DatanodeServiceName, c.Namespace, c.DatanodeRPCPort))

	configData, err := dbconfig.Marshal(cfg)
	if err != nil {
		return nil, err
	}

	return configData, nil
}

func (c *ConfigGenerator) generateFlownodeConfig(initConfig []byte) ([]byte, error) {
	cfg, err := dbconfig.NewFromComponentKind(v1alpha1.FlownodeComponentKind)
	if err != nil {
		return nil, err
	}

	if err := cfg.SetInputConfig(string(initConfig)); err != nil {
		return nil, err
	}

	flownodeCfg, ok := cfg.(*dbconfig.FlownodeConfig)
	if !ok {
		return nil, fmt.Errorf("cfg is not flownode config")
	}

	nodeID, err := c.allocateNodeID()
	if err != nil {
		klog.Fatalf("Allocate node id failed: %v", err)
	}
	flownodeCfg.NodeID = &nodeID

	podIP := os.Getenv(deployer.EnvPodIP)
	if len(podIP) == 0 {
		return nil, fmt.Errorf("empty pod ip")
	}
	flownodeCfg.RPCBindAddr = ptr.To(fmt.Sprintf("%s:%d", podIP, c.RPCPort))

	podName := os.Getenv(deployer.EnvPodName)
	if len(podName) == 0 {
		return nil, fmt.Errorf("empty pod name")
	}

	flownodeCfg.RPCServerAddr = ptr.To(fmt.Sprintf("%s.%s.%s:%d", podName,
		c.ServiceName, c.Namespace, c.RPCPort))

	configData, err := dbconfig.Marshal(cfg)
	if err != nil {
		return nil, err
	}

	return configData, nil
}

// TODO(zyy17): the algorithm of allocating datanode id will be changed in the future.
// We use the very easy way to allocate node id for datanode and flownode: use the pod index of datanode that created by statefulset.
// If the hostname of datanode is 'basic-datanode-1', then its node-id will be '1'.
func (c *ConfigGenerator) allocateNodeID() (uint64, error) {
	name, err := c.hostname()
	if err != nil {
		return 0, err
	}

	if len(name) == 0 {
		return 0, fmt.Errorf("the hostname is empty")
	}

	token := strings.Split(name, "-")
	if len(token) == 0 {
		return 0, fmt.Errorf("invalid hostname format '%s'", name)
	}

	// For the pods of statefulset, the last token of datanodeHostname is the pod index.
	podIndex := token[len(token)-1]

	// Must be the valid integer type.
	nodeID, err := strconv.ParseUint(podIndex, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid hostname format '%s'", name)
	}

	return nodeID, nil
}
