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
	"net/http"
	"net/url"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"sigs.k8s.io/controller-runtime/pkg/client"

	greptimev1alpha1 "github.com/GreptimeTeam/greptimedb-operator/apis/v1alpha1"
	"github.com/GreptimeTeam/greptimedb-operator/controllers/common"
	"github.com/GreptimeTeam/greptimedb-operator/tests/e2e/helper"
)

// TestClusterFrontendGroupsIngress tests a frontend groups ingress cluster.
func TestClusterFrontendGroupsIngress(ctx context.Context, h *helper.Helper) {
	const (
		testCRFile              = "./testdata/resources/cluster/frontend-groups-ingress/cluster.yaml"
		ingressNginxNamespace   = "ingress-nginx"
		ingressNginxServiceName = "ingress-nginx-controller"
		host                    = "configure-frontend-groups-ingress.example.com"

		createTableSQL = `
	      CREATE TABLE monitor (
	      host STRING,
	      ts TIMESTAMP DEFAULT CURRENT_TIMESTAMP() TIME INDEX,
	      cpu FLOAT64 DEFAULT 0,
	      memory FLOAT64,
	      PRIMARY KEY(host)
	    );`

		insertSQL = "sql=INSERT INTO monitor VALUES (\"127.0.0.1\", 1667446797450, 0.1, 0.4), (\"127.0.0.2\", 1667446798450, 0.2, 0.3), (\"127.0.0.1\", 1667446798450, 0.5, 0.2)"
	)

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

	// Wait for GreptimeDB cluster ready.
	time.Sleep(15 * time.Second)

	err = h.Get(ctx, client.ObjectKey{Name: testCluster.Name, Namespace: testCluster.Namespace}, testCluster)
	Expect(err).NotTo(HaveOccurred(), "failed to get cluster")

	By("Run distributed HTTP test")
	ingressIP, err := h.GetIngressIP(ctx, ingressNginxNamespace, ingressNginxServiceName)
	Expect(err).NotTo(HaveOccurred(), "failed to get ingress ip")
	err = h.AddIPToHosts(ingressIP, host)
	Expect(err).NotTo(HaveOccurred(), "failed to add ip to host")

	data := fmt.Sprintf("sql=%s", url.QueryEscape(createTableSQL))
	err = h.RunHTTPTest("http://"+host+"/v1/sql", data, http.MethodPost)
	Expect(err).NotTo(HaveOccurred(), "failed to run HTTP test")

	err = h.RunHTTPTest("http://"+host+"/v1/sql", insertSQL, http.MethodPost)
	Expect(err).NotTo(HaveOccurred(), "failed to run HTTP test")

	err = h.RunHTTPTest("http://"+host+"/config", "", http.MethodGet)
	Expect(err).NotTo(HaveOccurred(), "failed to run HTTP test")

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
