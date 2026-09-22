package deployers

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/GreptimeTeam/greptimedb-operator/apis/v1alpha1"
)

// Regression tests for maintenance-mode targeting when ONE DatanodeDeployer
// reconciles MULTIPLE GreptimeDBClusters.
//
// With the previous single process-global bool, a disable fired from an idle
// cluster's post-sync hook (a) hit the IDLE cluster's metasrv and (b) consumed
// the flag, so the actually-rolling cluster stayed in maintenance mode (region
// failover + GC disabled) with no self-heal — see GreptimeTeam/greptimedb-operator#365.
//
// The real turnOnMaintenanceMode / turnOffMaintenanceMode run unmodified here;
// the fake k8s client provides both clusters, and both metasrv stand-ins sit
// behind one logging HTTP proxy (SetMaintenanceMode dials the real in-cluster
// style meta URL; Go's default transport honors HTTP_PROXY), so every ON/OFF
// is attributed to the metasrv it actually hit.

const (
	sharedHost = "greptimedb-meta.greptimedb"                       // cluster A metasrv
	drillHost  = "greptimedb-fork-drill-meta.greptimedb-fork-drill" // cluster B metasrv
)

var callLog []string // entries like "ON -> greptimedb-meta.greptimedb"

// One proxy for the whole package: net/http caches proxy env lookups for ~30s,
// so per-test httptest servers (fresh port each time) are unreliable.
func TestMain(m *testing.M) {
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		dir := "ON"
		if strings.HasSuffix(r.URL.Path, "/disable") {
			dir = "OFF"
		}
		callLog = append(callLog, fmt.Sprintf("%s -> %s", dir, strings.SplitN(r.Host, ":", 2)[0]))
		w.WriteHeader(http.StatusOK)
	}))
	defer proxy.Close()
	os.Setenv("HTTP_PROXY", proxy.URL)
	os.Setenv("HTTPS_PROXY", proxy.URL)
	os.Exit(m.Run())
}

func newHarness(t *testing.T) (*DatanodeDeployer, *v1alpha1.GreptimeDBCluster, *v1alpha1.GreptimeDBCluster, context.Context) {
	t.Helper()

	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = v1alpha1.AddToScheme(scheme)

	sharedCluster := makeCluster("greptimedb", "greptimedb")
	drillCluster := makeCluster("greptimedb-fork-drill", "greptimedb-fork-drill")

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(sharedCluster, drillCluster,
			makeSts("greptimedb-datanode", "greptimedb", "a"),
			makeSts("greptimedb-fork-drill-datanode", "greptimedb-fork-drill", "a")).
		Build()

	// the singleton deployer, wired exactly as the manager does
	d := &DatanodeDeployer{
		CommonDeployer: &CommonDeployer{Scheme: scheme, Client: fakeClient},
	}
	return d, sharedCluster, drillCluster, context.Background()
}

func armedFor(t *testing.T, d *DatanodeDeployer, cluster *v1alpha1.GreptimeDBCluster) bool {
	t.Helper()
	_, ok := d.maintenanceMode.Load(client.ObjectKeyFromObject(cluster))
	return ok
}

// A datanode roll on cluster A, with cluster A's own post-sync hook firing
// first (the healthy ordering): ON and OFF must pair on A and nothing may hit B.
func TestMaintenanceMode_SingleClusterRoll_PairsCorrectly(t *testing.T) {
	d, shared, _, ctx := newHarness(t)
	callLog = nil

	if err := d.turnOnMaintenanceMode(ctx, makeSts("greptimedb-datanode", "greptimedb", "b"), shared); err != nil {
		t.Fatalf("turnOn: %v", err)
	}
	if !armedFor(t, d, shared) {
		t.Fatal("flag should be armed for the rolling cluster")
	}
	if err := d.turnOffMaintenanceMode(ctx, shared); err != nil {
		t.Fatalf("turnOff: %v", err)
	}
	if armedFor(t, d, shared) {
		t.Fatal("flag should be cleared after OFF")
	}
	expectCalls(t, []string{
		"ON -> " + sharedHost,
		"OFF -> " + sharedHost,
	})
}

// THE BUG of greptimedb-operator#365: cluster A rolls, but cluster B's
// (ready) reconcile runs its post-sync hook first — in the real controller A
// cannot fire its own hook yet because DefaultDeployer.Sync returns
// ErrSyncNotReady while the roll is in progress.
//
// Fixed behavior: B's hook must be a no-op (no HTTP call, B never touched);
// A's own later hook must send the OFF to A.
func TestMaintenanceMode_DisableFromIdleCluster_NoLongerMisfires(t *testing.T) {
	d, shared, drill, ctx := newHarness(t)
	callLog = nil

	// A's roll starts
	if err := d.turnOnMaintenanceMode(ctx, makeSts("greptimedb-datanode", "greptimedb", "b"), shared); err != nil {
		t.Fatalf("turnOn(shared): %v", err)
	}

	// B's reconcile runs its post-sync hook mid-roll of A
	if err := d.turnOffMaintenanceMode(ctx, drill); err != nil {
		t.Fatalf("turnOff via idle cluster's reconcile: %v", err)
	}
	if !armedFor(t, d, shared) {
		t.Fatal("precondition: A's flag must still be armed after B's idle no-op hook")
	}

	// A converges; its own hook fires
	if err := d.turnOffMaintenanceMode(ctx, shared); err != nil {
		t.Fatalf("turnOff via rolling cluster's reconcile: %v", err)
	}

	expectCalls(t, []string{
		"ON -> " + sharedHost,
		"OFF -> " + sharedHost, // the disable went to the RIGHT metasrv
	})
}

// Two clusters roll concurrently-ish with interleaved hooks: each cluster's
// flag must be tracked independently — no cross-talk in either direction.
func TestMaintenanceMode_InterleavedRolls_TrackedPerCluster(t *testing.T) {
	d, shared, drill, ctx := newHarness(t)
	callLog = nil

	// A rolls; B's idle hook fires (no-op); B then rolls; A's hook fires
	// (no-op for A's state, but B is still armed so B's own OFF must wait);
	// finally B converges.
	if err := d.turnOnMaintenanceMode(ctx, makeSts("greptimedb-datanode", "greptimedb", "b"), shared); err != nil {
		t.Fatalf("turnOn(shared): %v", err)
	}
	if err := d.turnOffMaintenanceMode(ctx, drill); err != nil { // idle no-op
		t.Fatalf("turnOff(drill idle): %v", err)
	}
	if err := d.turnOnMaintenanceMode(ctx, makeSts("greptimedb-fork-drill-datanode", "greptimedb-fork-drill", "b"), drill); err != nil {
		t.Fatalf("turnOn(drill): %v", err)
	}
	if err := d.turnOffMaintenanceMode(ctx, shared); err != nil { // A converged
		t.Fatalf("turnOff(shared): %v", err)
	}
	if err := d.turnOffMaintenanceMode(ctx, drill); err != nil { // B converged
		t.Fatalf("turnOff(drill): %v", err)
	}

	expectCalls(t, []string{
		"ON -> " + sharedHost,
		"ON -> " + drillHost,
		"OFF -> " + sharedHost,
		"OFF -> " + drillHost,
	})
}

func expectCalls(t *testing.T, want []string) {
	t.Helper()
	if fmt.Sprint(callLog) != fmt.Sprint(want) {
		t.Fatalf("\n got: %v\nwant: %v", callLog, want)
	}
}

func makeCluster(name, ns string) *v1alpha1.GreptimeDBCluster {
	enable := true
	return &v1alpha1.GreptimeDBCluster{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: ns},
		Spec: v1alpha1.GreptimeDBClusterSpec{
			Meta: &v1alpha1.MetaSpec{
				HTTPPort:             3002,
				EnableRegionFailover: &enable,
			},
		},
	}
}

// rollSeed varies the pod-template annotation: turnOnMaintenanceMode fetches
// the STORED STS as "old" and compares pod templates (isOldPodRestart) — a
// differing annotation models a Helm-driven roll.
func makeSts(name, ns, rollSeed string) *appsv1.StatefulSet {
	return &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: ns},
		Spec: appsv1.StatefulSetSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{"metricgator.io/roll": rollSeed}},
				Spec: corev1.PodSpec{
					InitContainers: []corev1.Container{{Name: "init", Image: "busybox:1.36"}},
					Containers:     []corev1.Container{{Name: "datanode", Image: "greptime:img"}},
				},
			},
		},
	}
}
