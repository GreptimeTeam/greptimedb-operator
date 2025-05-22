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

// TestScaleCluster test the scaling up and down of the cluster.
func TestScaleCluster(ctx context.Context, h *helper.Helper) {
	const (
		testCRFile  = "./testdata/resources/cluster/scale/cluster.yaml"
		testSQLFile = "./testdata/sql/cluster/partition.sql"
	)

	By(fmt.Sprintf("greptimecluster test with CR file %s and SQL file %s", testCRFile, testSQLFile))

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
	frontendAddr, err := h.PortForward(ctx, testCluster.Namespace, common.ResourceName(testCluster.Name, greptimev1alpha1.FrontendComponentKind), int(testCluster.Spec.PostgreSQLPort))
	Expect(err).NotTo(HaveOccurred(), "failed to port forward frontend service")
	Eventually(func() error {
		conn, err := net.Dial("tcp", frontendAddr)
		if err != nil {
			return err
		}
		_ = conn.Close()
		return nil
	}, helper.DefaultTimeout, time.Second).ShouldNot(HaveOccurred())

	err = h.RunSQLTest(ctx, frontendAddr, testSQLFile)
	Expect(err).NotTo(HaveOccurred(), "failed to run sql test")

	By("Scale up the cluster")
	err = h.Get(ctx, client.ObjectKey{Name: testCluster.Name, Namespace: testCluster.Namespace}, testCluster)
	Expect(err).NotTo(HaveOccurred(), "failed to get cluster")
	testCluster.Spec.Datanode.Replicas = ptr.To(int32(3))
	testCluster.Spec.Frontend.Replicas = ptr.To(int32(2))
	err = h.Update(ctx, testCluster)
	Expect(err).NotTo(HaveOccurred(), "failed to update cluster")

	By("Check the replicas of testCluster after scaling up")
	Eventually(func() error {
		err = h.Get(ctx, client.ObjectKey{Name: testCluster.Name, Namespace: testCluster.Namespace}, testCluster)
		if err != nil {
			return err
		}

		if testCluster.Status.Datanode.ReadyReplicas != 3 {
			return fmt.Errorf("datanode replicas is not 3")
		}

		if testCluster.Status.Frontend.ReadyReplicas != 2 {
			return fmt.Errorf("frontend replicas is not 2")
		}

		if testCluster.Status.ClusterPhase != greptimev1alpha1.PhaseRunning {
			return fmt.Errorf("cluster is not running")
		}

		return nil
	}, helper.DefaultTimeout, time.Second).ShouldNot(HaveOccurred())

	By("Execute distributed SQL test after scaling up")
	err = h.RunSQLTest(ctx, frontendAddr, testSQLFile)
	Expect(err).NotTo(HaveOccurred(), "failed to run sql test")

	By("Scale down the cluster")
	err = h.Get(ctx, client.ObjectKey{Name: testCluster.Name, Namespace: testCluster.Namespace}, testCluster)
	Expect(err).NotTo(HaveOccurred(), "failed to get cluster")
	testCluster.Spec.Datanode.Replicas = ptr.To(int32(1))
	testCluster.Spec.Frontend.Replicas = ptr.To(int32(1))

	err = h.Update(ctx, testCluster)
	Expect(err).NotTo(HaveOccurred(), "failed to update cluster")

	By("Check the replicas of testCluster after scaling down")
	Eventually(func() error {
		err = h.Get(ctx, client.ObjectKey{Name: testCluster.Name, Namespace: testCluster.Namespace}, testCluster)
		if err != nil {
			return err
		}

		if testCluster.Status.Datanode.ReadyReplicas != 1 {
			return fmt.Errorf("datanode replicas is not 1")
		}

		if testCluster.Status.Frontend.ReadyReplicas != 1 {
			return fmt.Errorf("frontend replicas is not 1")
		}

		if testCluster.Status.ClusterPhase != greptimev1alpha1.PhaseRunning {
			return fmt.Errorf("cluster is not running")
		}

		return nil
	}, helper.DefaultTimeout, time.Second).ShouldNot(HaveOccurred())

	// FIXME(zyy17): The cluster is not stable after scaling down, maybe it's a bug of the db.
	// NOTE: Sleep for a while to make sure the cluster is stable.
	time.Sleep(10 * time.Second)

	By("Execute distributed SQL test after scaling down")
	err = h.RunSQLTest(ctx, frontendAddr, testSQLFile)
	Expect(err).NotTo(HaveOccurred(), "failed to run sql test")

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
