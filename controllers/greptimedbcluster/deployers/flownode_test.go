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

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/GreptimeTeam/greptimedb-operator/apis/v1alpha1"
)

func newTestFlownodeBuilder() *flownodeBuilder {
	cluster := &v1alpha1.GreptimeDBCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: v1alpha1.GreptimeDBClusterSpec{
			Flownode: &v1alpha1.FlownodeSpec{
				RPCPort:  4001,
				HTTPPort: 4000,
			},
		},
	}

	return &flownodeBuilder{
		CommonBuilder: &CommonBuilder{
			Cluster: cluster,
		},
	}
}

func TestFlownodeServicePorts(t *testing.T) {
	b := newTestFlownodeBuilder()

	ports := b.servicePorts()
	if len(ports) != 2 {
		t.Fatalf("expected 2 service ports, got %d: %v", len(ports), ports)
	}

	want := map[string]int32{
		"rpc":  4001,
		"http": 4000,
	}
	for _, p := range ports {
		if p.Protocol != corev1.ProtocolTCP {
			t.Errorf("service port %q: protocol = %q, want TCP", p.Name, p.Protocol)
		}
		wantPort, ok := want[p.Name]
		if !ok {
			t.Errorf("unexpected service port %q", p.Name)
			continue
		}
		if p.Port != wantPort {
			t.Errorf("service port %q: port = %d, want %d", p.Name, p.Port, wantPort)
		}
	}
}

func TestFlownodeContainerPorts(t *testing.T) {
	b := newTestFlownodeBuilder()

	ports := b.containerPorts()
	if len(ports) != 2 {
		t.Fatalf("expected 2 container ports, got %d: %v", len(ports), ports)
	}

	want := map[string]int32{
		"grpc": 4001,
		"http": 4000,
	}
	for _, p := range ports {
		if p.Protocol != corev1.ProtocolTCP {
			t.Errorf("container port %q: protocol = %q, want TCP", p.Name, p.Protocol)
		}
		wantPort, ok := want[p.Name]
		if !ok {
			t.Errorf("unexpected container port %q", p.Name)
			continue
		}
		if p.ContainerPort != wantPort {
			t.Errorf("container port %q: containerPort = %d, want %d", p.Name, p.ContainerPort, wantPort)
		}
	}
}
