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

package greptimedbcluster

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/GreptimeTeam/greptimedb-operator/apis/v1alpha1"
	"github.com/GreptimeTeam/greptimedb-operator/controllers/common"
)

var _ = Describe("Test greptimedbcluster controller", func() {
	It("Test reconcile", func() {
		var (
			testClusterName = "testcluster"
			testNamespace   = "default"

			req = reconcile.Request{NamespacedName: client.ObjectKey{
				Name:      testClusterName,
				Namespace: testNamespace,
			}}

			cluster     *v1alpha1.GreptimeDBCluster
			svc         *corev1.Service
			statefulSet *appsv1.StatefulSet
			deployment  *appsv1.Deployment
		)

		By("Create GreptimeDBCluster")
		testCluster := createCluster(testClusterName, testNamespace)
		err := k8sClient.Create(ctx, testCluster)
		Expect(err).NotTo(HaveOccurred(), "failed to create greptimedbcluster")

		By("Check GreptimeDBCluster resource")
		cluster = &v1alpha1.GreptimeDBCluster{}
		err = k8sClient.Get(ctx, client.ObjectKey{Name: testClusterName, Namespace: testNamespace}, cluster)
		Expect(err).NotTo(HaveOccurred(), "failed to get cluster")

		By("Check meta resource")
		svc = &corev1.Service{}
		checkResource(testNamespace, common.ResourceName(testClusterName, v1alpha1.MetaRoleKind), svc, req)
		deployment = &appsv1.Deployment{}
		checkResource(testNamespace, common.ResourceName(testClusterName, v1alpha1.MetaRoleKind), deployment, req)

		// Move forward to sync datanode.
		Expect(makeDeploymentReady(deployment, cluster.Spec.Meta.Replicas)).NotTo(HaveOccurred(), "failed to update meta deployment status")

		By("Check datanode resource")
		svc = &corev1.Service{}
		checkResource(testNamespace, common.ResourceName(testClusterName, v1alpha1.DatanodeRoleKind), svc, req)
		statefulSet = &appsv1.StatefulSet{}
		checkResource(testNamespace, common.ResourceName(testClusterName, v1alpha1.DatanodeRoleKind), statefulSet, req)

		// Move forward to sync frontend.
		Expect(makeStatefulSetReady(statefulSet, cluster.Spec.Datanode.Replicas)).NotTo(HaveOccurred(), "failed to update datanode statefulset status")

		By("Check frontend resource")
		svc = &corev1.Service{}
		checkResource(testNamespace, common.ResourceName(testClusterName, v1alpha1.FrontendRoleKind), svc, req)
		deployment = &appsv1.Deployment{}
		checkResource(testNamespace, common.ResourceName(testClusterName, v1alpha1.FrontendRoleKind), deployment, req)

		// Move forward to complete status.
		Expect(makeDeploymentReady(deployment, cluster.Spec.Frontend.Replicas)).NotTo(HaveOccurred(), "failed to update frontend deployment status")

		By("Check status of cluster")
		Eventually(func() bool {
			var cluster v1alpha1.GreptimeDBCluster
			if err := k8sClient.Get(ctx, client.ObjectKey{Namespace: testNamespace, Name: testClusterName}, &cluster); err != nil {
				return false
			}

			for _, condition := range cluster.Status.Conditions {
				if condition.Type == v1alpha1.ConditionTypeReady && condition.Status == corev1.ConditionTrue {
					return true
				}
			}
			return false
		}, 30*time.Second, time.Second).Should(BeTrue())

		By("Delete cluster")
		err = k8sClient.Delete(ctx, testCluster)
		Expect(err).NotTo(HaveOccurred(), "failed to delete cluster")
		Eventually(func() error {
			// The cluster will be deleted eventually.
			return k8sClient.Get(ctx, client.ObjectKey{Name: testClusterName, Namespace: testNamespace}, cluster)
		}, 30*time.Second, time.Second).Should(HaveOccurred())
	})
})

func createCluster(name, namespace string) *v1alpha1.GreptimeDBCluster {
	return &v1alpha1.GreptimeDBCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1alpha1.GreptimeDBClusterSpec{
			Initializer: &v1alpha1.InitializerSpec{Image: "greptime/greptimedb-initializer:latest"},
			Base: &v1alpha1.PodTemplateSpec{
				MainContainer: &v1alpha1.MainContainerSpec{
					Image: "greptime/greptimedb:latest",
				},
			},
			Frontend: &v1alpha1.FrontendSpec{
				ComponentSpec: v1alpha1.ComponentSpec{
					Replicas: ptr.To(int32(1)),
				},
			},
			Meta: &v1alpha1.MetaSpec{
				ComponentSpec: v1alpha1.ComponentSpec{
					Replicas: ptr.To(int32(1)),
				},
				BackendStorage: &v1alpha1.BackendStorage{
					EtcdStorage: &v1alpha1.EtcdStorage{
						Endpoints: []string{
							"etcd.default:2379",
						},
					},
				},
			},
			Datanode: &v1alpha1.DatanodeSpec{
				ComponentSpec: v1alpha1.ComponentSpec{
					Replicas: ptr.To(int32(3)),
				},
			},
		},
	}
}

func checkResource(namespace, name string, object client.Object, request reconcile.Request) {
	Eventually(func() error {
		err := k8sClient.Get(ctx, client.ObjectKey{Namespace: namespace, Name: name}, object)
		if err != nil {
			// Try 3 (= 1s/300ms) times
			_, err := reconciler.Reconcile(context.TODO(), request)
			if err != nil {
				return err
			}
		}
		return err
	}, 30*time.Second, time.Second).Should(BeNil())
}

func makeStatefulSetReady(statefulSet *appsv1.StatefulSet, replicas *int32) error {
	statefulSet.Status.ReadyReplicas = *replicas
	statefulSet.Status.Replicas = *replicas
	statefulSet.Status.CurrentReplicas = *replicas
	statefulSet.Status.ObservedGeneration = statefulSet.Generation
	return reconciler.Status().Update(ctx, statefulSet)
}

func makeDeploymentReady(deployment *appsv1.Deployment, replicas *int32) error {
	deployment.Status.Replicas = *replicas
	deployment.Status.ReadyReplicas = *replicas
	deployment.Status.ObservedGeneration = deployment.Generation
	deployment.Status.Conditions = append(deployment.Status.Conditions, appsv1.DeploymentCondition{
		Type:               appsv1.DeploymentProgressing,
		Status:             corev1.ConditionTrue,
		LastUpdateTime:     metav1.Now(),
		LastTransitionTime: metav1.Now(),
		Reason:             "NewReplicaSetAvailable",
		Message:            "",
	})
	return reconciler.Status().Update(ctx, deployment)
}
