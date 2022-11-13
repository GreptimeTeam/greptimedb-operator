package main

import (
	"bytes"
	goflag "flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/spf13/pflag"
	"k8s.io/klog/v2"
)

type options struct {
	configPath        string
	defaultConfigPath string
	datanodeRPCPort   int32
}

func main() {
	var opts options

	pflag.StringVar(&opts.configPath, "config-path", "/etc/datanode/datanode.toml", "default config path")
	pflag.StringVar(&opts.defaultConfigPath, "input-config-file", "/datanode/defaults/datanode.toml", "output config path")

	// FIXME(zyy17): The greptimedb should support inject configs as env.
	pflag.Int32Var(&opts.datanodeRPCPort, "datanode-rpc-port", 4001, "datanode rpc port")

	klog.InitFlags(nil)
	pflag.CommandLine.AddGoFlagSet(goflag.CommandLine)

	pflag.Parse()

	// TODO(zyy17): We should add the flag to distinguish different components, not just datanode.
	var datanodeConfig DatanodeConfig
	_, err := toml.DecodeFile(opts.defaultConfigPath, &datanodeConfig)
	if err != nil {
		klog.Fatalf("Parse input toml file failed, err '%v'", err)
	}

	nodeID, err := allocateNodeID()
	if err != nil {
		klog.Fatalf("Allocate node id failed: %v", err)
	}
	datanodeConfig.NodeID = nodeID

	podIP := os.Getenv("POD_IP")
	if len(podIP) == 0 {
		klog.Fatalf("Get empty pod ip")
	}
	datanodeConfig.RPCAddr = fmt.Sprintf("%s:%d", podIP, opts.datanodeRPCPort)

	buf := new(bytes.Buffer)
	err = toml.NewEncoder(buf).Encode(&datanodeConfig)
	if err != nil {
		klog.Fatalf("Encode data config failed, err '%v'", err)
	}

	if err := os.WriteFile(opts.configPath, buf.Bytes(), 0644); err != nil {
		klog.Fatalf("write data config failed, err '%v'", err)
	}

	klog.Infof("Allocate node-id '%d' successfully and write it into file '%s'", nodeID, opts.configPath)
}

// TODO(zyy17): the algorithm of allocating node-id will be changed in the future.
// We use the very easy way to allocate node-id: use the pod index of datanode that created by statefulset.
// If the hostname of datanode is 'basic-datanode-1', then its node-id will be '1'.
func allocateNodeID() (uint64, error) {
	name, err := os.Hostname()
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

	podIndex := token[len(token)-1]
	// Must be the valid integer type.
	nodeID, err := strconv.ParseUint(podIndex, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid hostname format '%s'", name)
	}

	return nodeID, nil
}
