// Copyright 2022 Greptime Team
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
	configPath          string
	defaultConfigPath   string
	datanodeRPCPort     int32
	datanodeServiceName string
	namespace           string

	// The storage options.
	storageType     string
	localStorageDir string
	bucket          string
	prefix          string
	accessKeyID     string
	secretAccessKey string

	endpoint string
	region   string
}

func main() {
	var opts options

	pflag.StringVar(&opts.configPath, "config-path", "/etc/datanode/datanode.toml", "default config path")
	pflag.StringVar(&opts.defaultConfigPath, "input-config-file", "/datanode/defaults/datanode.toml", "output config path")
	pflag.StringVar(&opts.datanodeServiceName, "datanode-service-name", "", "the name of datanode service")
	pflag.StringVar(&opts.namespace, "namespace", "", "the cluster namespace")

	// Handle storage options.
	pflag.StringVar(&opts.storageType, "storage-type", "", "the storage type")
	pflag.StringVar(&opts.localStorageDir, "local-storage-dir", "", "the directory of local storage")
	pflag.StringVar(&opts.bucket, "bucket", "", "the bucket of s3 storage")
	pflag.StringVar(&opts.prefix, "prefix", "", "the prefix of s3 storage")
	pflag.StringVar(&opts.accessKeyID, "access-key-id", "", "the access key id of s3 storage")
	pflag.StringVar(&opts.secretAccessKey, "secret-access-key", "", "the secret access key of s3 storage")
	pflag.StringVar(&opts.endpoint, "endpoint", "", "the endpoint of s3 storage")
	pflag.StringVar(&opts.region, "region", "", "the region of s3 storage")

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

	if opts.storageType == "S3" {
		datanodeConfig.StorageConfig.Type = "S3"
		datanodeConfig.StorageConfig.Bucket = opts.bucket
		datanodeConfig.StorageConfig.AccessKeyID = opts.accessKeyID
		datanodeConfig.StorageConfig.SecretAccessKey = opts.secretAccessKey
		datanodeConfig.StorageConfig.Endpoint = opts.endpoint
		datanodeConfig.StorageConfig.Region = opts.region
		datanodeConfig.StorageConfig.Root = opts.prefix
		datanodeConfig.StorageConfig.DataDir = ""
	}

	if opts.storageType == "Local" {
		datanodeConfig.StorageConfig.DataDir = opts.localStorageDir
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

	podName := os.Getenv("POD_NAME")
	if len(podName) == 0 {
		klog.Fatalf("Get empty pod name")
	}
	datanodeConfig.RPCHostName = fmt.Sprintf("%s.%s.%s:%d", podName, opts.datanodeServiceName, opts.namespace, opts.datanodeRPCPort)
	klog.Infof("rpc hostname: %s", datanodeConfig.RPCHostName)

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
