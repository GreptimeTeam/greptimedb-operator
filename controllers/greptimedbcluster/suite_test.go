package greptimedbcluster

import (
	"context"
	"io"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"go.etcd.io/etcd/client/v3"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/GreptimeTeam/greptimedb-operator/apis/v1alpha1"
	// +kubebuilder:scaffold:imports
)

var (
	cfg        *rest.Config
	k8sClient  client.Client
	reconciler *Reconciler
	testEnv    *envtest.Environment

	ctx    context.Context
	cancel context.CancelFunc
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "greptimedbcluster controller suite")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	By("bootstrapping test environment")
	useExistingCluster := false
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "..", "config", "crd", "bases")},
		ErrorIfCRDPathMissing: true,
		UseExistingCluster:    &useExistingCluster,
	}

	var err error
	// cfg is defined in this file globally.
	cfg, err = testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	err = v1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	// +kubebuilder:scaffold:scheme

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	manager, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme: scheme.Scheme,
	})
	Expect(err).ToNot(HaveOccurred())

	ctx, cancel = context.WithCancel(context.TODO())
	reconciler = &Reconciler{
		Client:                 manager.GetClient(),
		Scheme:                 manager.GetScheme(),
		Recorder:               manager.GetEventRecorderFor("greptimedbcluster-controller"),
		etcdMaintenanceBuilder: buildMockEtcdMaintenance,
	}
	err = reconciler.SetupWithManager(manager)
	Expect(err).ToNot(HaveOccurred())

	go func() {
		defer GinkgoRecover()
		err = manager.Start(ctx)
		Expect(err).ToNot(HaveOccurred(), "failed to run manager")
	}()
})

var _ = AfterSuite(func() {
	cancel()
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})

func buildMockEtcdMaintenance(etcdEndpoints []string) (clientv3.Maintenance, error) {
	return &mockEtcdMaintenance{}, nil
}

// TODO(zyy17): Maybe can use testify.
var _ clientv3.Maintenance = &mockEtcdMaintenance{}

type mockEtcdMaintenance struct{}

func (_ *mockEtcdMaintenance) AlarmList(ctx context.Context) (*clientv3.AlarmResponse, error) {
	return &clientv3.AlarmResponse{}, nil
}

func (_ *mockEtcdMaintenance) AlarmDisarm(ctx context.Context, m *clientv3.AlarmMember) (*clientv3.AlarmResponse, error) {
	return &clientv3.AlarmResponse{}, nil
}

func (_ *mockEtcdMaintenance) Defragment(ctx context.Context, endpoint string) (*clientv3.DefragmentResponse, error) {
	return &clientv3.DefragmentResponse{}, nil
}

func (_ *mockEtcdMaintenance) Status(ctx context.Context, endpoint string) (*clientv3.StatusResponse, error) {
	return &clientv3.StatusResponse{}, nil
}

func (_ *mockEtcdMaintenance) HashKV(ctx context.Context, endpoint string, rev int64) (*clientv3.HashKVResponse, error) {
	return &clientv3.HashKVResponse{}, nil
}

func (_ *mockEtcdMaintenance) Snapshot(ctx context.Context) (io.ReadCloser, error) {
	return nil, nil
}

func (_ *mockEtcdMaintenance) MoveLeader(ctx context.Context, transfereeID uint64) (*clientv3.MoveLeaderResponse, error) {
	return &clientv3.MoveLeaderResponse{}, nil
}
