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

package greptimedbcluster

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

// TestClusterHTTPAPIServer tests that the frontend enables the dedicated HTTP API server,
// and the http service port forwards to the configurable httpApiServerPort.
func TestClusterHTTPAPIServer(ctx context.Context, h *helper.Helper) {
	const testCRFile = "./testdata/resources/cluster/http-api-server/cluster.yaml"

	By(fmt.Sprintf("greptimecluster test with CR file %s", testCRFile))

	testCluster := new(greptimev1alpha1.GreptimeDBCluster)
	err := h.LoadCR(testCRFile, testCluster)
	Expect(err).NotTo(HaveOccurred(), "failed to load greptimedbcluster yaml file")

	err = h.Create(ctx, testCluster)
	Expect(err).NotTo(HaveOccurred(), "failed to create greptimedbcluster")

	By("Check the status of testCluster")
	Eventually(func() error {
		clusterPhase, err := h.GetPhase(ctx, testCluster.Namespace, testCluster.Name, new(greptimev1alpha1.GreptimeDBCluster))
		if err != nil {
			return err
		}

		if clusterPhase != greptimev1alpha1.PhaseRunning {
			return fmt.Errorf("cluster is not running")
		}

		return nil
	}, helper.DefaultTimeout, time.Second).ShouldNot(HaveOccurred())

	err = h.Get(ctx, client.ObjectKey{Name: testCluster.Name, Namespace: testCluster.Namespace}, testCluster)
	Expect(err).NotTo(HaveOccurred(), "failed to get cluster")

	frontendName := common.ResourceName(testCluster.Name, greptimev1alpha1.FrontendRoleKind)
	frontendKey := client.ObjectKey{Namespace: testCluster.Namespace, Name: frontendName}

	By("Check the frontend service forwards http port to httpApiServerPort")
	svc := new(corev1.Service)
	err = h.Get(ctx, frontendKey, svc)
	Expect(err).NotTo(HaveOccurred(), "failed to get frontend service")

	var httpServicePort *corev1.ServicePort
	for i := range svc.Spec.Ports {
		if svc.Spec.Ports[i].Name == "http" {
			httpServicePort = &svc.Spec.Ports[i]
			break
		}
	}
	Expect(httpServicePort).NotTo(BeNil(), "frontend service should expose the http port")
	Expect(httpServicePort.Port).To(Equal(testCluster.Spec.HTTPPort), "http service port should be the cluster httpPort")
	Expect(httpServicePort.TargetPort).To(Equal(intstr.FromInt32(testCluster.Spec.Frontend.HTTPAPIServerPort)), "http service targetPort should be httpApiServerPort")

	By("Check the frontend config enables the HTTP API server on httpApiServerPort")
	cm := new(corev1.ConfigMap)
	err = h.Get(ctx, frontendKey, cm)
	Expect(err).NotTo(HaveOccurred(), "failed to get frontend configmap")

	configData, ok := cm.Data[constant.GreptimeDBConfigFileName]
	Expect(ok).To(BeTrue(), "configmap should contain config.toml")
	Expect(configData).To(ContainSubstring("enable_api_server = true"))
	Expect(configData).To(ContainSubstring(fmt.Sprintf("api_server_addr = \"0.0.0.0:%d\"", testCluster.Spec.Frontend.HTTPAPIServerPort)))

	By("Check the frontend container keeps http port and exposes http-api port")
	deploy := new(appsv1.Deployment)
	err = h.Get(ctx, frontendKey, deploy)
	Expect(err).NotTo(HaveOccurred(), "failed to get frontend deployment")

	var (
		httpContainerPort    *corev1.ContainerPort
		httpAPIContainerPort *corev1.ContainerPort
	)
	for i := range deploy.Spec.Template.Spec.Containers[constant.MainContainerIndex].Ports {
		port := &deploy.Spec.Template.Spec.Containers[constant.MainContainerIndex].Ports[i]
		if port.Name == "http" {
			httpContainerPort = port
		}
		if port.Name == "http-api" {
			httpAPIContainerPort = port
		}
	}
	Expect(httpContainerPort).NotTo(BeNil(), "frontend container should keep the http port")
	Expect(httpContainerPort.ContainerPort).To(Equal(testCluster.Spec.Frontend.HTTPPort), "http container port should be the frontend httpPort")
	Expect(httpAPIContainerPort).NotTo(BeNil(), "frontend container should expose the http-api port")
	Expect(httpAPIContainerPort.ContainerPort).To(Equal(testCluster.Spec.Frontend.HTTPAPIServerPort), "http-api container port should be httpApiServerPort")

	By("Run show tables via the HTTP API server through the http service port")
	httpAddr, err := h.PortForward(ctx, testCluster.Namespace, frontendName, int(testCluster.Spec.HTTPPort))
	Expect(err).NotTo(HaveOccurred(), "failed to port forward frontend http service")
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

	By("Delete cluster")
	err = h.Delete(ctx, testCluster)
	Expect(err).NotTo(HaveOccurred(), "failed to delete cluster")
	Eventually(func() error {
		// The cluster will be deleted eventually.
		return h.Get(ctx, client.ObjectKey{Name: testCluster.Name, Namespace: testCluster.Namespace}, testCluster)
	}, helper.DefaultTimeout, time.Second).Should(HaveOccurred())
}
