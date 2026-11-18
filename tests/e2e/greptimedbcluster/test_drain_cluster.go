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
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	greptimev1alpha1 "github.com/GreptimeTeam/greptimedb-operator/apis/v1alpha1"
	"github.com/GreptimeTeam/greptimedb-operator/controllers/common"
	"github.com/GreptimeTeam/greptimedb-operator/tests/e2e/helper"
)

// TestDrainCluster verifies that scaling a datanode down drains the datanode
// before the pod is removed and that data remains readable and writable.
//
// On GreptimeDB < v1.4.0 the drain API is unavailable and the operator falls
// back to the previous scale-down behavior; the test still passes there.
func TestDrainCluster(ctx context.Context, h *helper.Helper) {
	const (
		testCRFile  = "./testdata/resources/cluster/drain/cluster.yaml"
		testSQLFile = "./testdata/sql/cluster/drain.sql"
		verifySQL   = "./testdata/sql/cluster/drain_verify.sql"
	)

	By(fmt.Sprintf("greptimecluster drain test with CR file %s", testCRFile))

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

	By("Execute distributed SQL test")
	frontendAddr, err := h.PortForward(ctx, testCluster.Namespace, common.ResourceName(testCluster.Name, greptimev1alpha1.FrontendRoleKind), int(testCluster.Spec.PostgreSQLPort))
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

	By("Scale down the datanodes from 3 to 2")
	err = h.Get(ctx, client.ObjectKey{Name: testCluster.Name, Namespace: testCluster.Namespace}, testCluster)
	Expect(err).NotTo(HaveOccurred(), "failed to get cluster")
	testCluster.Spec.Datanode.Replicas = ptr.To(int32(2))
	err = h.Update(ctx, testCluster)
	Expect(err).NotTo(HaveOccurred(), "failed to update cluster")

	By("Check the datanode replicas after scaling down")
	Eventually(func() error {
		err = h.Get(ctx, client.ObjectKey{Name: testCluster.Name, Namespace: testCluster.Namespace}, testCluster)
		if err != nil {
			return err
		}

		if testCluster.Status.Datanode.ReadyReplicas != 2 {
			return fmt.Errorf("datanode ready replicas is not 2, got %d", testCluster.Status.Datanode.ReadyReplicas)
		}

		if testCluster.Status.ClusterPhase != greptimev1alpha1.PhaseRunning {
			return fmt.Errorf("cluster is not running, phase: %s", testCluster.Status.ClusterPhase)
		}

		return nil
	}, helper.DefaultTimeout, time.Second).ShouldNot(HaveOccurred())

	By("Check that no datanode drain is left in progress")
	Expect(testCluster.Status.Datanode.Draining).To(BeEmpty(), "datanode drain should have completed")
	Expect(testCluster.Status.Datanode.PendingSchedulable).To(BeEmpty(), "no datanode should be pending schedulable after scale-down")

	// FIXME(zyy17): The cluster is not stable after scaling down; wait for it to settle.
	time.Sleep(10 * time.Second)

	By("Verify the data is still readable and writable after scaling down")
	err = h.RunSQLTest(ctx, frontendAddr, verifySQL)
	Expect(err).NotTo(HaveOccurred(), "failed to run verify sql test")

	By("Kill the port forwarding process")
	h.KillPortForwardProcess()

	By("Delete cluster")
	err = h.Delete(ctx, testCluster)
	Expect(err).NotTo(HaveOccurred(), "failed to delete cluster")
	Eventually(func() error {
		return h.Get(ctx, client.ObjectKey{Name: testCluster.Name, Namespace: testCluster.Namespace}, testCluster)
	}, helper.DefaultTimeout, time.Second).Should(HaveOccurred())
}
