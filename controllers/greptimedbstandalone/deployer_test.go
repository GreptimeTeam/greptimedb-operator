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

package greptimedbstandalone

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

	"github.com/GreptimeTeam/greptimedb-operator/apis/v1alpha1"
)

func TestStandaloneGenerateMainContainerArgsHTTPAPIServer(t *testing.T) {
	standalone := &v1alpha1.GreptimeDBStandalone{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: v1alpha1.GreptimeDBStandaloneSpec{
			RPCPort:           4001,
			HTTPPort:          4000,
			MySQLPort:         4002,
			PostgreSQLPort:    4003,
			HTTPAPIServerPort: v1alpha1.DefaultHTTPAPIServerPort,
		},
	}

	b := &standaloneBuilder{
		standalone: standalone,
	}

	// Enabled by default.
	checkHTTPAPIServerArg(t, b.generateMainContainerArgs(), "0.0.0.0:4006")

	// Disabled.
	standalone.Spec.EnableHTTPAPIServer = ptr.To(false)
	checkHTTPAPIServerArgAbsent(t, b.generateMainContainerArgs())
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
