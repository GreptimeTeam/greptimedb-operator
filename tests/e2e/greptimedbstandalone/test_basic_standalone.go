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

package greptimedbstandalone

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
	"github.com/GreptimeTeam/greptimedb-operator/tests/e2e/helper"
)

// TestBasicStandalone tests a basic standalone.
func TestBasicStandalone(ctx context.Context, h *helper.Helper) {
	const (
		testCRFile  = "./testdata/resources/standalone/basic/standalone.yaml"
		testSQLFile = "./testdata/sql/standalone/basic.sql"
	)

	By(fmt.Sprintf("greptimestandalone test with CR file %s and SQL file %s", testCRFile, testSQLFile))

	testStandalone := new(greptimev1alpha1.GreptimeDBStandalone)
	err := h.LoadCR(testCRFile, testStandalone)
	Expect(err).NotTo(HaveOccurred(), "failed to load greptime crd from file")

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

	By("Run SQL test")
	frontendAddr, err := h.PortForward(ctx, testStandalone.Namespace, common.ResourceName(testStandalone.Name, greptimev1alpha1.StandaloneRoleKind), int(testStandalone.Spec.PostgreSQLPort))
	Expect(err).NotTo(HaveOccurred(), "failed to port forward frontend service")
	Eventually(func() error {
		conn, err := net.Dial("tcp", frontendAddr)
		if err != nil {
			return err
		}
		conn.Close()
		return nil
	}, helper.DefaultTimeout, time.Second).ShouldNot(HaveOccurred())

	err = h.RunSQLTest(ctx, frontendAddr, testSQLFile)
	Expect(err).NotTo(HaveOccurred(), "failed to run sql test")

	By("Kill the port forwarding process")
	h.KillPortForwardProcess()

	By("Delete standalone")
	err = h.Delete(ctx, testStandalone)
	Expect(err).NotTo(HaveOccurred(), "failed to delete standalone")
	Eventually(func() error {
		// The standalone will be deleted eventually.
		return h.Get(ctx, client.ObjectKey{Name: testStandalone.Name, Namespace: testStandalone.Namespace}, testStandalone)
	}, helper.DefaultTimeout, time.Second).Should(HaveOccurred())

	By("The PVC of the database should be retained")
	dataPVCs, err := h.GetPVCs(ctx, testStandalone.Namespace, common.ResourceName(testStandalone.Name, greptimev1alpha1.StandaloneRoleKind), common.FileStorageTypeDatanode)
	Expect(err).NotTo(HaveOccurred(), "failed to get data PVCs")
	Expect(len(dataPVCs)).To(Equal(1), "the number of datanode PVCs should be equal to 1")

	By("Remove the PVC of the datanode")
	for _, pvc := range dataPVCs {
		err = h.Delete(ctx, &pvc)
		Expect(err).NotTo(HaveOccurred(), "failed to delete data PVCs")
	}
}
