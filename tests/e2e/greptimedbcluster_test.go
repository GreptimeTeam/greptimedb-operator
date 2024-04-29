// Copyright 2022 Greptime Team
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
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/google/uuid"
	"sigs.k8s.io/controller-runtime/pkg/client"

	greptimev1alpha1 "github.com/GreptimeTeam/greptimedb-operator/apis/v1alpha1"
	"github.com/GreptimeTeam/greptimedb-operator/tests/e2e/utils"
)

var _ = Describe("Test GreptimeDBCluster", func() {
	var (
		ctx = context.Background()
	)

	AfterEach(func() {
		err := utils.CleanEtcdData(ctx, "etcd", "etcd-0")
		Expect(err).NotTo(HaveOccurred(), "failed to clean etcd data")
	})

	It("Create basic greptimedb cluster", func() {
		testCluster := new(greptimev1alpha1.GreptimeDBCluster)
		err := utils.LoadGreptimeCRDFromFile("./testdata/basic-cluster/cluster.yaml", testCluster)
		Expect(err).NotTo(HaveOccurred(), "failed to load greptimedbcluster yaml file")

		runClusterTest(ctx, testCluster)
	})

	It("Create greptimedb cluster with remote wal", func() {
		testCluster := new(greptimev1alpha1.GreptimeDBCluster)
		err := utils.LoadGreptimeCRDFromFile("./testdata/cluster-with-remote-wal/cluster.yaml", testCluster)
		Expect(err).NotTo(HaveOccurred(), "failed to load greptimedbcluster yaml file")

		runClusterTest(ctx, testCluster)
	})

	It("Create greptimedb cluster with AWS S3", func() {
		s3Region := os.Getenv("AWS_CI_TEST_BUCKET_REGION")
		if s3Region == "" {
			Fail("AWS_CI_TEST_BUCKET_REGION is not set")
		}
		s3Bucket := os.Getenv("AWS_CI_TEST_BUCKET")
		if s3Bucket == "" {
			Fail("AWS_CI_TEST_BUCKET is not set")
		}

		testCluster := new(greptimev1alpha1.GreptimeDBCluster)
		err := utils.LoadGreptimeCRDFromFile("./testdata/cluster-with-s3/cluster.yaml", testCluster)
		Expect(err).NotTo(HaveOccurred(), "failed to load greptimedbcluster yaml file")

		var (
			credentialsSecretName = "s3-credentials"
		)

		err = utils.CreateS3Credentials(ctx, k8sClient, testCluster.Namespace, credentialsSecretName)
		Expect(err).NotTo(HaveOccurred(), "failed to create s3 credentials")

		// Set the S3 storage provider for the testCluster.
		testCluster.Spec.ObjectStorageProvider = &greptimev1alpha1.ObjectStorageProvider{
			S3: &greptimev1alpha1.S3StorageProvider{
				Bucket:     s3Bucket,
				Region:     s3Region,
				SecretName: credentialsSecretName,
				Root:       fmt.Sprintf("operator-e2e-%s", uuid.New().String()),
			},
		}

		runClusterTest(ctx, testCluster)
	})
})

func runClusterTest(ctx context.Context, testCluster *greptimev1alpha1.GreptimeDBCluster) {
	err := k8sClient.Create(ctx, testCluster)
	Expect(err).NotTo(HaveOccurred(), "failed to create greptimedbcluster")

	By("Check the status of testCluster")
	Eventually(func() error {
		clusterPhase, err := utils.GetPhase(ctx, k8sClient, testCluster.Namespace, testCluster.Name, new(greptimev1alpha1.GreptimeDBCluster))
		if err != nil {
			return err
		}

		if clusterPhase != greptimev1alpha1.PhaseRunning {
			return fmt.Errorf("cluster is not running")
		}

		return nil
	}, utils.DefaultTimeout, time.Second).ShouldNot(HaveOccurred())

	By("Execute distributed SQL queries")
	var frontendIngressIP string
	Eventually(func() error {
		frontendIngressIP, err = utils.GetFrontendServiceIngressIP(ctx, k8sClient, testCluster.Namespace, fmt.Sprintf("%s-frontend", testCluster.Name))
		if err != nil {
			return err
		}
		return nil
	}, utils.DefaultTimeout, time.Second).ShouldNot(HaveOccurred())

	err = utils.RunSQLTest(ctx, frontendIngressIP, true)
	Expect(err).NotTo(HaveOccurred(), "failed to run distributed SQL queries")

	By("Delete cluster")
	err = k8sClient.Delete(ctx, testCluster)
	Expect(err).NotTo(HaveOccurred(), "failed to delete cluster")
	Eventually(func() error {
		// The cluster will be deleted eventually.
		return k8sClient.Get(ctx, client.ObjectKey{Name: testCluster.Namespace, Namespace: testCluster.Namespace}, testCluster)
	}, utils.DefaultTimeout, time.Second).Should(HaveOccurred())

	By("The PVC of the datanode should be retained")
	datanodePVCs, err := utils.GetPVCs(ctx, k8sClient, testCluster.Namespace, testCluster.Name, greptimev1alpha1.DatanodeComponentKind)
	Expect(err).NotTo(HaveOccurred(), "failed to get datanode PVCs")
	Expect(int32(len(datanodePVCs))).To(Equal(*testCluster.Spec.Datanode.Replicas), "the number of datanode PVCs should be equal to the number of datanode replicas")

	By("Remove the PVC of the datanode")
	for _, pvc := range datanodePVCs {
		err = k8sClient.Delete(ctx, &pvc)
		Expect(err).NotTo(HaveOccurred(), "failed to delete datanode PVC")
	}
}
