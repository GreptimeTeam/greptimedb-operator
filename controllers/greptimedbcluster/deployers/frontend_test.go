// Copyright 2026 Greptime Team
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

package deployers

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

	"github.com/GreptimeTeam/greptimedb-operator/apis/v1alpha1"
)

func TestFrontendGenerateMainContainerArgsHTTPAPIServer(t *testing.T) {
	cluster := &v1alpha1.GreptimeDBCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: v1alpha1.GreptimeDBClusterSpec{
			Meta: &v1alpha1.MetaSpec{
				RPCPort: 3002,
			},
			Frontend: &v1alpha1.FrontendSpec{
				RPCPort:           4001,
				HTTPPort:          4000,
				MySQLPort:         4002,
				PostgreSQLPort:    4003,
				InternalPort:      4010,
				HTTPAPIServerPort: v1alpha1.DefaultHTTPAPIServerPort,
			},
		},
	}

	b := &frontendBuilder{
		CommonBuilder: &CommonBuilder{
			Cluster: cluster,
		},
	}

	// Enabled by default.
	checkHTTPAPIServerArg(t, b.generateMainContainerArgs(cluster.Spec.Frontend), "0.0.0.0:4006")

	// Disabled.
	cluster.Spec.Frontend.EnableHTTPAPIServer = ptr.To(false)
	checkHTTPAPIServerArgAbsent(t, b.generateMainContainerArgs(cluster.Spec.Frontend))
}

func TestFrontendGenerateMainContainerArgsHTTPAPIServerIPv6(t *testing.T) {
	cluster := &v1alpha1.GreptimeDBCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: v1alpha1.GreptimeDBClusterSpec{
			EnableIPv6: true,
			Meta: &v1alpha1.MetaSpec{
				RPCPort: 3002,
			},
			Frontend: &v1alpha1.FrontendSpec{
				RPCPort:           4001,
				HTTPPort:          4000,
				MySQLPort:         4002,
				PostgreSQLPort:    4003,
				InternalPort:      4010,
				HTTPAPIServerPort: v1alpha1.DefaultHTTPAPIServerPort,
			},
		},
	}

	b := &frontendBuilder{
		CommonBuilder: &CommonBuilder{
			Cluster: cluster,
		},
	}

	checkHTTPAPIServerArg(t, b.generateMainContainerArgs(cluster.Spec.Frontend), "[::]:4006")
}

func checkHTTPAPIServerArg(t *testing.T, args []string, wantAddr string) {
	t.Helper()
	for i := 0; i < len(args)-1; i++ {
		if args[i] == "--http-api-server-addr" {
			if args[i+1] != wantAddr {
				t.Fatalf("--http-api-server-addr = %q, want %q (args: %v)", args[i+1], wantAddr, args)
			}
			return
		}
	}
	t.Fatalf("--http-api-server-addr not found in args: %v", args)
}

func checkHTTPAPIServerArgAbsent(t *testing.T, args []string) {
	t.Helper()
	for i := range args {
		if args[i] == "--http-api-server-addr" {
			t.Fatalf("--http-api-server-addr should not be set when the HTTP API server is disabled, args: %v", args)
		}
	}
}
