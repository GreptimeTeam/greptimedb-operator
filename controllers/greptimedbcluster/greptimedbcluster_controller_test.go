package greptimedbcluster

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/greptime/greptimedb-operator/apis/v1alpha1"
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

		By("Check etcd resourcd")
		svc = &corev1.Service{}
		checkResource(testNamespace, testClusterName+"-etcd", svc, req)
		statefulSet = &appsv1.StatefulSet{}
		checkResource(testNamespace, testClusterName+"-etcd", statefulSet, req)

		// Move forward to sync meta.
		Expect(makeStatefulSetReady(statefulSet, 3)).NotTo(HaveOccurred(), "failed to update etcd statefulset status")

		By("Check meta resource")
		svc = &corev1.Service{}
		checkResource(testNamespace, testClusterName+"-meta", svc, req)
		deployment = &appsv1.Deployment{}
		checkResource(testNamespace, testClusterName+"-meta", deployment, req)

		// Move forward to sync datanode.
		Expect(makeDeploymentReady(deployment, cluster.Spec.Meta.Replicas)).NotTo(HaveOccurred(), "failed to update meta deploylemt status")

		By("Check datanode resource")
		svc = &corev1.Service{}
		checkResource(testNamespace, testClusterName+"-datanode", svc, req)
		deployment = &appsv1.Deployment{}
		checkResource(testNamespace, testClusterName+"-datanode", deployment, req)

		// Move forward to sync frontend.
		Expect(makeDeploymentReady(deployment, cluster.Spec.Datanode.Replicas)).NotTo(HaveOccurred(), "failed to update datanode deploylemt status")

		By("Check frontend resource")
		svc = &corev1.Service{}
		checkResource(testNamespace, testClusterName+"-frontend", svc, req)
		deployment = &appsv1.Deployment{}
		checkResource(testNamespace, testClusterName+"-frontend", deployment, req)

		// Move forward to complete status.
		Expect(makeDeploymentReady(deployment, cluster.Spec.Frontend.Replicas)).NotTo(HaveOccurred(), "failed to update frontend deploylemt status")

		By("Check status of cluster")
		Eventually(func() bool {
			var cluser v1alpha1.GreptimeDBCluster
			if err := k8sClient.Get(ctx, client.ObjectKey{Namespace: testNamespace, Name: testClusterName}, &cluser); err != nil {
				return false
			}

			for _, condition := range cluser.Status.Conditions {
				if condition.Type == v1alpha1.GreptimeDBClusterReady && condition.Status == corev1.ConditionTrue {
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
			Base: &v1alpha1.PodTemplateSpec{
				MainContainer: &v1alpha1.MainContainerSpec{
					Image: "localhost:5001/greptime/greptimedb:latest",
				},
			},
			Frontend: &v1alpha1.FrontendSpec{
				ComponentSpec: v1alpha1.ComponentSpec{
					Replicas: 1,
				},
			},
			Meta: &v1alpha1.MetaSpec{
				ComponentSpec: v1alpha1.ComponentSpec{
					Replicas: 3,
				},
			},
			Datanode: &v1alpha1.DatanodeSpec{
				ComponentSpec: v1alpha1.ComponentSpec{
					Replicas: 1,
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
			reconciler.Reconcile(context.TODO(), request)
		}
		return err
	}, 30*time.Second, time.Second).Should(BeNil())
}

func makeStatefulSetReady(statefulSet *appsv1.StatefulSet, replicas int32) error {
	statefulSet.Status.ReadyReplicas = replicas
	statefulSet.Status.Replicas = replicas
	return reconciler.Status().Update(ctx, statefulSet)
}

func makeDeploymentReady(deployment *appsv1.Deployment, replicas int32) error {
	deployment.Status.Replicas = replicas
	deployment.Status.ReadyReplicas = replicas
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
