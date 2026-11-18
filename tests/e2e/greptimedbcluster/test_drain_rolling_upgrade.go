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

	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	greptimev1alpha1 "github.com/GreptimeTeam/greptimedb-operator/apis/v1alpha1"
	"github.com/GreptimeTeam/greptimedb-operator/controllers/common"
	"github.com/GreptimeTeam/greptimedb-operator/tests/e2e/helper"
)

// runDrainRollingUpgrade creates a cluster with datanode draining enabled, writes
// a partitioned table, triggers a rolling update by changing the pod template
// annotation, and verifies the data survives. When verifyTempGroup is true, it
// also checks that the auto-generated temp group exists and the Rotate rolling
// upgrade completes.
func runDrainRollingUpgrade(ctx context.Context, h *helper.Helper, crFile string, verifyTempGroup bool) {
	const (
		testSQLFile = "./testdata/sql/cluster/drain.sql"
		verifySQL   = "./testdata/sql/cluster/drain_verify.sql"
	)

	By(fmt.Sprintf("greptimecluster drain rolling upgrade test with CR file %s", crFile))

	testCluster := new(greptimev1alpha1.GreptimeDBCluster)
	err := h.LoadCR(crFile, testCluster)
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

	if verifyTempGroup {
		By("Check that the temp datanode group is auto-generated")
		tempSts := new(appsv1.StatefulSet)
		tempName := common.ResourceName(testCluster.Name, greptimev1alpha1.DatanodeRoleKind, "temp")
		err = h.Get(ctx, client.ObjectKey{Name: tempName, Namespace: testCluster.Namespace}, tempSts)
		Expect(err).NotTo(HaveOccurred(), "failed to get the temp datanode statefulset")
	}

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

	By("Trigger a rolling update by changing the pod template annotation")
	err = h.Get(ctx, client.ObjectKey{Name: testCluster.Name, Namespace: testCluster.Namespace}, testCluster)
	Expect(err).NotTo(HaveOccurred(), "failed to get cluster")
	if testCluster.Spec.Datanode.Template == nil {
		testCluster.Spec.Datanode.Template = &greptimev1alpha1.PodTemplateSpec{}
	}
	testCluster.Spec.Datanode.Template.Annotations = map[string]string{"e2e.greptime.io/rolling": "v2"}
	err = h.Update(ctx, testCluster)
	Expect(err).NotTo(HaveOccurred(), "failed to update cluster")

	By("Wait for the rolling upgrade to complete")
	Eventually(func() error {
		clusterPhase, err := h.GetPhase(ctx, testCluster.Namespace, testCluster.Name, new(greptimev1alpha1.GreptimeDBCluster))
		if err != nil {
			return err
		}
		if clusterPhase != greptimev1alpha1.PhaseRunning {
			return fmt.Errorf("cluster is not running")
		}
		return nil
	}, 10*time.Minute, time.Second).ShouldNot(HaveOccurred())

	if verifyTempGroup {
		By("Check that the Rotate rolling upgrade has finished")
		err = h.Get(ctx, client.ObjectKey{Name: testCluster.Name, Namespace: testCluster.Namespace}, testCluster)
		Expect(err).NotTo(HaveOccurred(), "failed to get cluster")
		Expect(testCluster.Status.Datanode.RollingUpgrade).To(BeNil(), "rolling upgrade should have completed")
	}

	By("Verify the data is still readable and writable after the rolling upgrade")
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

// TestDrainScatterRollingUpgrade tests the Scatter drain policy (Plan A) during a
// rolling upgrade.
func TestDrainScatterRollingUpgrade(ctx context.Context, h *helper.Helper) {
	runDrainRollingUpgrade(ctx, h, "./testdata/resources/cluster/drain-scatter-roll/cluster.yaml", false)
}

// TestDrainRotateRollingUpgrade tests the Rotate drain policy (Plan C) during a
// rolling upgrade.
func TestDrainRotateRollingUpgrade(ctx context.Context, h *helper.Helper) {
	runDrainRollingUpgrade(ctx, h, "./testdata/resources/cluster/drain-rotate-roll/cluster.yaml", true)
}
