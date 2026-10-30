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
	"github.com/GreptimeTeam/greptimedb-operator/tests/testutil"
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
	testutil.CheckHTTPAPIServerArg(t, b.generateMainContainerArgs(cluster.Spec.Frontend), "0.0.0.0:4006")

	// Disabled.
	cluster.Spec.Frontend.EnableHTTPAPIServer = ptr.To(false)
	testutil.CheckHTTPAPIServerArgAbsent(t, b.generateMainContainerArgs(cluster.Spec.Frontend))
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

	testutil.CheckHTTPAPIServerArg(t, b.generateMainContainerArgs(cluster.Spec.Frontend), "[::]:4006")
}
