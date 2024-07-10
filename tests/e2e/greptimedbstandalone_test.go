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

package e2e

import (
	"context"
	"fmt"
	"net"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"sigs.k8s.io/controller-runtime/pkg/client"

	greptimev1alpha1 "github.com/GreptimeTeam/greptimedb-operator/apis/v1alpha1"
	"github.com/GreptimeTeam/greptimedb-operator/controllers/common"
	"github.com/GreptimeTeam/greptimedb-operator/tests/e2e/utils"
)

var _ = Describe("Test GreptimeDBStandalone", func() {
	var (
		ctx = context.Background()
	)

	It("Create basic greptimedb standalone", func() {
		testStandalone := new(greptimev1alpha1.GreptimeDBStandalone)
		err := utils.LoadGreptimeCRDFromFile("./testdata/basic-standalone/standalone.yaml", testStandalone)
		Expect(err).NotTo(HaveOccurred(), "failed to load greptime crd from file")

		err = k8sClient.Create(ctx, testStandalone)
		Expect(err).NotTo(HaveOccurred(), "failed to create greptimedbstandalone")

		By("Check the status of testStandalone")
		Eventually(func() error {
			phase, err := utils.GetPhase(ctx, k8sClient, testStandalone.Namespace, testStandalone.Name, new(greptimev1alpha1.GreptimeDBStandalone))
			if err != nil {
				return err
			}

			if phase != greptimev1alpha1.PhaseRunning {
				return fmt.Errorf("standalone is not running")
			}

			return nil
		}, utils.DefaultTimeout, time.Second).ShouldNot(HaveOccurred())

		By("Run SQL test")
		frontendAddr, err := utils.PortForward(ctx, testStandalone.Namespace, common.ResourceName(testStandalone.Name, greptimev1alpha1.StandaloneKind), int(testStandalone.Spec.MySQLServicePort))
		Expect(err).NotTo(HaveOccurred(), "failed to port forward frontend service")
		Eventually(func() error {
			conn, err := net.Dial("tcp", frontendAddr)
			if err != nil {
				return err
			}
			conn.Close()
			return nil
		}, utils.DefaultTimeout, time.Second).ShouldNot(HaveOccurred())

		err = utils.RunSQLTest(ctx, frontendAddr, false)
		Expect(err).NotTo(HaveOccurred(), "failed to run SQL test")

		By("Kill the port forwarding process")
		utils.KillPortForwardProcess()

		By("Delete standalone")
		err = k8sClient.Delete(ctx, testStandalone)
		Expect(err).NotTo(HaveOccurred(), "failed to delete standalone")
		Eventually(func() error {
			// The standalone will be deleted eventually.
			return k8sClient.Get(ctx, client.ObjectKey{Name: testStandalone.Namespace, Namespace: testStandalone.Namespace}, testStandalone)
		}, utils.DefaultTimeout, time.Second).Should(HaveOccurred())

		By("The PVC of the database should be retained")
		dataPVCs, err := utils.GetPVCs(ctx, k8sClient, testStandalone.Namespace, testStandalone.Name, greptimev1alpha1.StandaloneKind)
		Expect(err).NotTo(HaveOccurred(), "failed to get data PVCs")
		Expect(len(dataPVCs)).To(Equal(1), "the number of datanode PVCs should be equal to 1")

		By("Remove the PVC of the datanode")
		for _, pvc := range dataPVCs {
			err = k8sClient.Delete(ctx, &pvc)
			Expect(err).NotTo(HaveOccurred(), "failed to delete data PVCs")
		}
	})
})
