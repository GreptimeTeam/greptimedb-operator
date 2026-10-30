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
	"github.com/GreptimeTeam/greptimedb-operator/tests/testutil"
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
	testutil.CheckHTTPAPIServerArg(t, b.generateMainContainerArgs(), "0.0.0.0:4006")

	// Disabled.
	standalone.Spec.EnableHTTPAPIServer = ptr.To(false)
	testutil.CheckHTTPAPIServerArgAbsent(t, b.generateMainContainerArgs())
}
