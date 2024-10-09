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
	"sort"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/google/go-cmp/cmp"
	"github.com/jackc/pgx/v5"
	"sigs.k8s.io/controller-runtime/pkg/client"

	greptimev1alpha1 "github.com/GreptimeTeam/greptimedb-operator/apis/v1alpha1"
	"github.com/GreptimeTeam/greptimedb-operator/controllers/common"
	"github.com/GreptimeTeam/greptimedb-operator/tests/e2e/helper"
)

// TestClusterEnableMonitoring tests a cluster that enables monitoring.
func TestClusterEnableMonitoring(ctx context.Context, h *helper.Helper) {
	const (
		testCRFile  = "./testdata/resources/cluster/enable-monitoring/cluster.yaml"
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
		conn.Close()
		return nil
	}, helper.DefaultTimeout, time.Second).ShouldNot(HaveOccurred())

	err = h.RunSQLTest(ctx, frontendAddr, testSQLFile)
	Expect(err).NotTo(HaveOccurred(), "failed to run sql test")

	monitoringAddr, err := h.PortForward(ctx, testCluster.Namespace, common.ResourceName(common.MonitoringServiceName(testCluster.Name), greptimev1alpha1.StandaloneKind), int(testCluster.Spec.PostgreSQLPort))
	Expect(err).NotTo(HaveOccurred(), "failed to port forward monitoring service")
	Eventually(func() error {
		conn, err := net.Dial("tcp", monitoringAddr)
		if err != nil {
			return err
		}
		conn.Close()
		return nil
	}, helper.DefaultTimeout, time.Second).ShouldNot(HaveOccurred())
	err = testMonitoringStandalone(ctx, monitoringAddr)
	Expect(err).NotTo(HaveOccurred(), "failed to test monitoring")

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
	datanodePVCs, err := h.GetPVCs(ctx, testCluster.Namespace, testCluster.Name, greptimev1alpha1.DatanodeComponentKind, common.FileStorageTypeDatanode)
	Expect(err).NotTo(HaveOccurred(), "failed to get datanode PVCs")
	Expect(int32(len(datanodePVCs))).To(Equal(*testCluster.Spec.Datanode.Replicas), "the number of datanode PVCs should be equal to the number of datanode replicas")

	By("Remove the PVC of the datanode")
	for _, pvc := range datanodePVCs {
		err = h.Delete(ctx, &pvc)
		Expect(err).NotTo(HaveOccurred(), "failed to delete datanode PVC")
	}
}

func testMonitoringStandalone(ctx context.Context, addr string) error {
	// connect public database by default.
	url := fmt.Sprintf("postgres://postgres@%s/public?sslmode=disable", addr)

	fmt.Printf("Connecting to %s\n", url)
	conn, err := pgx.Connect(ctx, url)
	if err != nil {
		return err
	}
	defer conn.Close(ctx)

	// Check metrics table.
	if err := checkCollectedRoles(ctx, conn, "SELECT DISTINCT app FROM greptime_app_version", []string{"greptime-datanode", "greptime-frontend", "greptime-metasrv"}); err != nil {
		return err
	}

	// Check logs table.
	if err := checkCollectedRoles(ctx, conn, "SELECT DISTINCT role FROM gtlogs", []string{"datanode", "frontend", "meta"}); err != nil {
		return err
	}

	var count int
	if err = conn.QueryRow(context.Background(), "SELECT COUNT(*) FROM gtlogs").Scan(&count); err != nil {
		return err
	}
	// The number of logs should be greater than 0.
	if count == 0 {
		return fmt.Errorf("no logs found")
	}

	return nil
}

func checkCollectedRoles(ctx context.Context, conn *pgx.Conn, query string, expected []string) error {
	rows, err := conn.Query(ctx, query)
	if err != nil {
		return err
	}
	defer rows.Close()

	var result []string
	for rows.Next() {
		var role string
		err = rows.Scan(&role)
		if err != nil {
			return err
		}
		result = append(result, role)
	}
	sort.Strings(result)
	sort.Strings(expected)

	if !cmp.Equal(result, expected) {
		return fmt.Errorf("results mismatch, got %v, expect %v", result, expected)
	}

	return nil
}
