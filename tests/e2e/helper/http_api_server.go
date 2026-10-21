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

package helper

import (
	"fmt"

	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/GreptimeTeam/greptimedb-operator/controllers/constant"
)

// CheckHTTPAPIServerServicePort checks that the http service forwards to httpApiServerPort.
func CheckHTTPAPIServerServicePort(svc *corev1.Service, httpPort, httpAPIServerPort int32) {
	var httpServicePort *corev1.ServicePort
	for i := range svc.Spec.Ports {
		if svc.Spec.Ports[i].Name == "http" {
			httpServicePort = &svc.Spec.Ports[i]
			break
		}
	}
	Expect(httpServicePort).NotTo(BeNil(), "service should expose the http port")
	Expect(httpServicePort.Port).To(Equal(httpPort), "http service port should be the httpPort")
	Expect(httpServicePort.TargetPort).To(Equal(intstr.FromInt32(httpAPIServerPort)), "http service targetPort should be httpApiServerPort")
}

// CheckHTTPAPIServerConfig checks that the config enables the HTTP API server on httpApiServerPort.
func CheckHTTPAPIServerConfig(cm *corev1.ConfigMap, httpAPIServerPort int32) {
	configData, ok := cm.Data[constant.GreptimeDBConfigFileName]
	Expect(ok).To(BeTrue(), "configmap should contain config.toml")
	Expect(configData).To(ContainSubstring("enable_api_server = true"))
	Expect(configData).To(ContainSubstring(fmt.Sprintf("api_server_addr = \"0.0.0.0:%d\"", httpAPIServerPort)))
}

// CheckHTTPAPIServerContainerPorts checks that the container keeps the http port and exposes the http-api port.
func CheckHTTPAPIServerContainerPorts(ports []corev1.ContainerPort, httpPort, httpAPIServerPort int32) {
	var (
		httpContainerPort    *corev1.ContainerPort
		httpAPIContainerPort *corev1.ContainerPort
	)
	for i := range ports {
		port := &ports[i]
		if port.Name == "http" {
			httpContainerPort = port
		}
		if port.Name == "http-api" {
			httpAPIContainerPort = port
		}
	}
	Expect(httpContainerPort).NotTo(BeNil(), "container should keep the http port")
	Expect(httpContainerPort.ContainerPort).To(Equal(httpPort), "http container port should be the httpPort")
	Expect(httpAPIContainerPort).NotTo(BeNil(), "container should expose the http-api port")
	Expect(httpAPIContainerPort.ContainerPort).To(Equal(httpAPIServerPort), "http-api container port should be httpApiServerPort")
}
