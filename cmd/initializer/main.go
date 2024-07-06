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

package main

import (
	goflag "flag"
	"os"

	"github.com/spf13/pflag"
	"k8s.io/klog/v2"

	"github.com/GreptimeTeam/greptimedb-operator/cmd/initializer/internal"
)

func main() {
	var opts internal.Options

	pflag.StringVar(&opts.ConfigPath, "config-path", "/etc/greptimedb/config.toml", "the config path")
	pflag.StringVar(&opts.InitConfigPath, "init-config-path", "/etc/greptimedb/init-config.toml", "the init config path")
	pflag.StringVar(&opts.Namespace, "namespace", "", "the namespace of greptimedb cluster")
	pflag.StringVar(&opts.ComponentKind, "component-kind", "", "the component kind")

	pflag.StringVar(&opts.ServiceName, "service-name", "", "the name of service")
	pflag.Int32Var(&opts.RPCPort, "rpc-port", 4001, "the RPC port")

	klog.InitFlags(nil)
	pflag.CommandLine.AddGoFlagSet(goflag.CommandLine)

	pflag.Parse()

	if err := internal.NewConfigGenerator(&opts, os.Hostname).Generate(); err != nil {
		klog.Fatalf("Generate config failed: %v", err)
	}
}
