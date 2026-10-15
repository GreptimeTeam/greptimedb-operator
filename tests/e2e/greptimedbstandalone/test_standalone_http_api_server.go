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
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	greptimev1alpha1 "github.com/GreptimeTeam/greptimedb-operator/apis/v1alpha1"
	"github.com/GreptimeTeam/greptimedb-operator/controllers/common"
	"github.com/GreptimeTeam/greptimedb-operator/controllers/constant"
	"github.com/GreptimeTeam/greptimedb-operator/tests/e2e/helper"
)

// TestStandaloneHTTPAPIServer tests that the standalone enables the dedicated HTTP API server,
// and the http service port forwards to the configurable httpApiServerPort.
func TestStandaloneHTTPAPIServer(ctx context.Context, h *helper.Helper) {
	const testCRFile = "./testdata/resources/standalone/http-api-server/standalone.yaml"

	By(fmt.Sprintf("greptimestandalone test with CR file %s", testCRFile))

	testStandalone := new(greptimev1alpha1.GreptimeDBStandalone)
	err := h.LoadCR(testCRFile, testStandalone)
	Expect(err).NotTo(HaveOccurred(), "failed to load greptimedbstandalone yaml file")

	err = h.Create(ctx, testStandalone)
	Expect(err).NotTo(HaveOccurred(), "failed to create greptimedbstandalone")

	By("Check the status of testStandalone")
	Eventually(func() error {
		phase, err := h.GetPhase(ctx, testStandalone.Namespace, testStandalone.Name, new(greptimev1alpha1.GreptimeDBStandalone))
		if err != nil {
			return err
		}

		if phase != greptimev1alpha1.PhaseRunning {
			return fmt.Errorf("standalone is not running")
		}

		return nil
	}, helper.DefaultTimeout, time.Second).ShouldNot(HaveOccurred())

	err = h.Get(ctx, client.ObjectKey{Name: testStandalone.Name, Namespace: testStandalone.Namespace}, testStandalone)
	Expect(err).NotTo(HaveOccurred(), "failed to get standalone")

	resourceName := common.ResourceName(testStandalone.Name, greptimev1alpha1.StandaloneRoleKind)
	resourceKey := client.ObjectKey{Namespace: testStandalone.Namespace, Name: resourceName}

	By("Check the standalone service forwards http port to httpApiServerPort")
	svc := new(corev1.Service)
	err = h.Get(ctx, resourceKey, svc)
	Expect(err).NotTo(HaveOccurred(), "failed to get standalone service")

	var httpServicePort *corev1.ServicePort
	for i := range svc.Spec.Ports {
		if svc.Spec.Ports[i].Name == "http" {
			httpServicePort = &svc.Spec.Ports[i]
			break
		}
	}
	Expect(httpServicePort).NotTo(BeNil(), "standalone service should expose the http port")
	Expect(httpServicePort.Port).To(Equal(testStandalone.Spec.HTTPPort), "http service port should be the standalone httpPort")
	Expect(httpServicePort.TargetPort).To(Equal(intstr.FromInt32(testStandalone.Spec.HTTPAPIServerPort)), "http service targetPort should be httpApiServerPort")

	By("Check the standalone config enables the HTTP API server on httpApiServerPort")
	cm := new(corev1.ConfigMap)
	err = h.Get(ctx, resourceKey, cm)
	Expect(err).NotTo(HaveOccurred(), "failed to get standalone configmap")

	configData, ok := cm.Data[constant.GreptimeDBConfigFileName]
	Expect(ok).To(BeTrue(), "configmap should contain config.toml")
	Expect(configData).To(ContainSubstring("enable_api_server = true"))
	Expect(configData).To(ContainSubstring(fmt.Sprintf("api_server_addr = \"0.0.0.0:%d\"", testStandalone.Spec.HTTPAPIServerPort)))

	By("Check the standalone container keeps http port and exposes http-api port")
	sts := new(appsv1.StatefulSet)
	err = h.Get(ctx, resourceKey, sts)
	Expect(err).NotTo(HaveOccurred(), "failed to get standalone statefulset")

	var (
		httpContainerPort    *corev1.ContainerPort
		httpAPIContainerPort *corev1.ContainerPort
	)
	for i := range sts.Spec.Template.Spec.Containers[constant.MainContainerIndex].Ports {
		port := &sts.Spec.Template.Spec.Containers[constant.MainContainerIndex].Ports[i]
		if port.Name == "http" {
			httpContainerPort = port
		}
		if port.Name == "http-api" {
			httpAPIContainerPort = port
		}
	}
	Expect(httpContainerPort).NotTo(BeNil(), "standalone container should keep the http port")
	Expect(httpContainerPort.ContainerPort).To(Equal(testStandalone.Spec.HTTPPort), "http container port should be the standalone httpPort")
	Expect(httpAPIContainerPort).NotTo(BeNil(), "standalone container should expose the http-api port")
	Expect(httpAPIContainerPort.ContainerPort).To(Equal(testStandalone.Spec.HTTPAPIServerPort), "http-api container port should be httpApiServerPort")

	By("Run show tables via the HTTP API server through the http service port")
	httpAddr, err := h.PortForward(ctx, testStandalone.Namespace, resourceName, int(testStandalone.Spec.HTTPPort))
	Expect(err).NotTo(HaveOccurred(), "failed to port forward standalone http service")
	Eventually(func() error {
		conn, err := net.Dial("tcp", httpAddr)
		if err != nil {
			return err
		}
		conn.Close()
		return nil
	}, helper.DefaultTimeout, time.Second).ShouldNot(HaveOccurred())

	err = h.RunHTTPTest(fmt.Sprintf("http://%s/v1/sql", httpAddr), "sql=show tables", http.MethodPost)
	Expect(err).NotTo(HaveOccurred(), "failed to run show tables through the HTTP API server")

	By("Kill the port forwarding process")
	h.KillPortForwardProcess()

	By("Delete standalone")
	err = h.Delete(ctx, testStandalone)
	Expect(err).NotTo(HaveOccurred(), "failed to delete standalone")
	Eventually(func() error {
		// The standalone will be deleted eventually.
		return h.Get(ctx, client.ObjectKey{Name: testStandalone.Name, Namespace: testStandalone.Namespace}, testStandalone)
	}, helper.DefaultTimeout, time.Second).Should(HaveOccurred())
}
