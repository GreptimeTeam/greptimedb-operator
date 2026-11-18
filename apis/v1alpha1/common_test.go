// Copyright 2024 Greptime Team
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

package v1alpha1

import (
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
)

func TestServiceSpecGetPorts(t *testing.T) {
	defaultPorts := []corev1.ServicePort{
		{Name: "rpc", Protocol: corev1.ProtocolTCP, Port: 4000},
		{Name: "http", Protocol: corev1.ProtocolTCP, Port: 4001},
	}
	customPorts := []corev1.ServicePort{
		{Name: "http", Protocol: corev1.ProtocolTCP, Port: 8080, TargetPort: intstr.FromInt32(8080), NodePort: 30080},
	}

	tests := []struct {
		name     string
		spec     *ServiceSpec
		expected []corev1.ServicePort
	}{
		{
			name:     "nil spec returns default ports",
			spec:     nil,
			expected: defaultPorts,
		},
		{
			name:     "empty ports returns default ports",
			spec:     &ServiceSpec{Type: corev1.ServiceTypeClusterIP},
			expected: defaultPorts,
		},
		{
			name:     "custom ports are used when set",
			spec:     &ServiceSpec{Type: corev1.ServiceTypeNodePort, Ports: customPorts},
			expected: customPorts,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.spec.GetPorts(defaultPorts)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Fatalf("expected %+v, got %+v", tt.expected, got)
			}
		})
	}
}

func TestDatanodeSpecIsEnableDrain(t *testing.T) {
	tests := []struct {
		name string
		spec *DatanodeSpec
		want bool
	}{
		{name: "nil spec", spec: nil, want: false},
		{name: "nil flag", spec: &DatanodeSpec{}, want: false},
		{name: "explicit false", spec: &DatanodeSpec{EnableDrain: ptr.To(false)}, want: false},
		{name: "explicit true", spec: &DatanodeSpec{EnableDrain: ptr.To(true)}, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.spec.IsEnableDrain(); got != tt.want {
				t.Fatalf("IsEnableDrain() = %v, want %v", got, tt.want)
			}
		})
	}
}
