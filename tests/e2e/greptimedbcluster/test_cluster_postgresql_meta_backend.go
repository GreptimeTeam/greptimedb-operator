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

	"sigs.k8s.io/controller-runtime/pkg/client"

	greptimev1alpha1 "github.com/GreptimeTeam/greptimedb-operator/apis/v1alpha1"
	"github.com/GreptimeTeam/greptimedb-operator/controllers/common"
	"github.com/GreptimeTeam/greptimedb-operator/tests/e2e/helper"
)

// TestPostgreSQLMetaBackend tests a cluster with postgresql as the meta backend.
func TestPostgreSQLMetaBackend(ctx context.Context, h *helper.Helper) {
	const (
		testCRFile  = "./testdata/resources/cluster/postgresql-meta-backend/cluster.yaml"
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

	By("Kill the port forwarding process")
	h.KillPortForwardProcess()

	By("Delete cluster")
	err = h.Delete(ctx, testCluster)
	Expect(err).NotTo(HaveOccurred(), "failed to delete cluster")
	Eventually(func() error {
		// The cluster will be deleted eventually.
		return h.Get(ctx, client.ObjectKey{Name: testCluster.Name, Namespace: testCluster.Namespace}, testCluster)
	}, helper.DefaultTimeout, time.Second).Should(HaveOccurred())

	By("The PVC of the datanode should be retained")
	datanodePVCs, err := h.GetPVCs(ctx, testCluster.Namespace, common.ResourceName(testCluster.Name, greptimev1alpha1.DatanodeRoleKind), common.FileStorageTypeDatanode)
	Expect(err).NotTo(HaveOccurred(), "failed to get datanode PVCs")
	Expect(int32(len(datanodePVCs))).To(Equal(*testCluster.Spec.Datanode.Replicas), "the number of datanode PVCs should be equal to the number of datanode replicas")

	By("Remove the PVC of the datanode")
	for _, pvc := range datanodePVCs {
		err = h.Delete(ctx, &pvc)
		Expect(err).NotTo(HaveOccurred(), "failed to delete datanode PVC")
	}
}
