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

package deployers

import (
	"context"
	"errors"
	"reflect"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/GreptimeTeam/greptimedb-operator/apis/v1alpha1"
	"github.com/GreptimeTeam/greptimedb-operator/pkg/deployer"
)

func TestDrainTrackerResetKeepsOnlyCurrentNodes(t *testing.T) {
	persisted := []v1alpha1.DatanodeDrainStatus{
		{NodeID: 1, ProcedureID: "p1", Status: "Running"},
		{NodeID: 2, ProcedureID: "p2", Status: "Running"},
		{NodeID: 3, ProcedureID: "p3", Status: "Running"},
	}

	tracker := &drainTracker{proc: map[uint64]drainProc{
		// A stale entry from a previous operation.
		9: {procedureID: "stale", status: "Running"},
	}}
	tracker.reset(persisted, []uint64{2, 3})

	got := tracker.snapshot()
	want := []v1alpha1.DatanodeDrainStatus{
		{NodeID: 2, ProcedureID: "p2", Status: "Running"},
		{NodeID: 3, ProcedureID: "p3", Status: "Running"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("snapshot() = %#v, want %#v", got, want)
	}
}

func TestDrainTrackerSnapshotSortedAndEmptyAfterReset(t *testing.T) {
	tracker := &drainTracker{proc: map[uint64]drainProc{}}
	tracker.reset([]v1alpha1.DatanodeDrainStatus{
		{NodeID: 3, ProcedureID: "p3", Status: "Running"},
		{NodeID: 1, ProcedureID: "p1", Status: "Running"},
	}, []uint64{1, 3})

	got := tracker.snapshot()
	want := []v1alpha1.DatanodeDrainStatus{
		{NodeID: 1, ProcedureID: "p1", Status: "Running"},
		{NodeID: 3, ProcedureID: "p3", Status: "Running"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("snapshot() = %#v, want %#v", got, want)
	}

	// Clearing the scope drops everything.
	tracker.reset(nil, nil)
	if got := tracker.snapshot(); len(got) != 0 {
		t.Fatalf("snapshot() after reset(nil, nil) = %#v, want empty", got)
	}
}

// Regression test for the scale-safety path in Apply: CreateObjectIfNotExist
// returns the existing object as an *unstructured.Unstructured, so the old
// StatefulSet must be re-fetched as a typed object. Scaling must not fail with
// "old object for statefulset ... is not a statefulset".
func TestApplyScaleUpReFetchesTypedStatefulSet(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = v1alpha1.AddToScheme(scheme)

	cluster := &v1alpha1.GreptimeDBCluster{
		ObjectMeta: metav1.ObjectMeta{Name: "c", Namespace: "default"},
		Spec: v1alpha1.GreptimeDBClusterSpec{
			Datanode: &v1alpha1.DatanodeSpec{
				ComponentSpec: v1alpha1.ComponentSpec{Replicas: ptr.To(int32(2))},
			},
		},
	}

	oldSts := &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{Kind: "StatefulSet", APIVersion: "apps/v1"},
		ObjectMeta: metav1.ObjectMeta{
			Name:        "c-datanode",
			Namespace:   "default",
			Annotations: map[string]string{deployer.LastAppliedResourceSpec: `{"replicas":1}`},
		},
		Spec: appsv1.StatefulSetSpec{Replicas: ptr.To(int32(1))},
	}

	newSts := &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{Kind: "StatefulSet", APIVersion: "apps/v1"},
		ObjectMeta: metav1.ObjectMeta{
			Name:        "c-datanode",
			Namespace:   "default",
			Annotations: map[string]string{deployer.LastAppliedResourceSpec: `{"replicas":2}`},
		},
		Spec: appsv1.StatefulSetSpec{Replicas: ptr.To(int32(2))},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(cluster, oldSts).
		Build()
	d := &DatanodeDeployer{CommonDeployer: &CommonDeployer{Scheme: scheme, Client: fakeClient}}

	err := d.Apply(context.Background(), cluster, []client.Object{newSts})
	if !errors.Is(err, deployer.ErrSyncNotReady) {
		t.Fatalf("Apply() error = %v, want ErrSyncNotReady", err)
	}

	got := new(appsv1.StatefulSet)
	if err := fakeClient.Get(context.Background(), client.ObjectKeyFromObject(newSts), got); err != nil {
		t.Fatalf("get statefulset: %v", err)
	}
	if got.Spec.Replicas == nil || *got.Spec.Replicas != 2 {
		t.Fatalf("statefulset replicas = %v, want 2", got.Spec.Replicas)
	}
}

func TestAutoTempGroupSpec(t *testing.T) {
	withRotate := &v1alpha1.GreptimeDBCluster{
		Spec: v1alpha1.GreptimeDBClusterSpec{
			Datanode: &v1alpha1.DatanodeSpec{
				ComponentSpec: v1alpha1.ComponentSpec{Replicas: ptr.To(int32(3))},
				Drain: &v1alpha1.DatanodeDrainSpec{
					Enabled:   true,
					Policy:    v1alpha1.DrainPolicyRotate,
					TempGroup: &v1alpha1.TempDatanodeGroupSpec{Replicas: ptr.To(int32(1))},
				},
			},
		},
	}
	withScatter := &v1alpha1.GreptimeDBCluster{
		Spec: v1alpha1.GreptimeDBClusterSpec{
			Datanode: &v1alpha1.DatanodeSpec{Drain: &v1alpha1.DatanodeDrainSpec{Enabled: true}},
		},
	}

	temp := autoTempGroupSpec(withRotate)
	if temp == nil {
		t.Fatal("autoTempGroupSpec() = nil, want temp group")
	}
	if temp.GetName() != "temp" {
		t.Fatalf("temp group name = %q, want temp", temp.GetName())
	}
	if temp.GetReplicas() == nil || *temp.GetReplicas() != 1 {
		t.Fatalf("temp group replicas = %v, want 1", temp.GetReplicas())
	}

	if got := autoTempGroupSpec(withScatter); got != nil {
		t.Fatalf("autoTempGroupSpec(scatter) = %v, want nil", got)
	}
}

func TestRollingUpgradePartition(t *testing.T) {
	tests := []struct {
		name string
		ru   *v1alpha1.DatanodeRollingUpgrade
		want *int32
	}{
		{name: "no rotation", ru: nil, want: nil},
		{name: "drain highest ordinal 2", ru: &v1alpha1.DatanodeRollingUpgrade{Ordinal: 2, Phase: v1alpha1.RollingUpgradePhaseDrain}, want: ptr.To(int32(3))},
		{name: "upgrade ordinal 2", ru: &v1alpha1.DatanodeRollingUpgrade{Ordinal: 2, Phase: v1alpha1.RollingUpgradePhaseUpgrade}, want: ptr.To(int32(2))},
		{name: "drain ordinal 0", ru: &v1alpha1.DatanodeRollingUpgrade{Ordinal: 0, Phase: v1alpha1.RollingUpgradePhaseDrain}, want: ptr.To(int32(1))},
		{name: "temp phase", ru: &v1alpha1.DatanodeRollingUpgrade{Ordinal: -1, Phase: v1alpha1.RollingUpgradePhaseDrain}, want: ptr.To(int32(0))},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cluster := &v1alpha1.GreptimeDBCluster{}
			cluster.Status.Datanode.RollingUpgrade = tt.ru
			got := rollingUpgradePartition(cluster)
			if tt.want == nil {
				if got != nil {
					t.Fatalf("rollingUpgradePartition() = %v, want nil", *got)
				}
				return
			}
			if got == nil || *got != *tt.want {
				t.Fatalf("rollingUpgradePartition() = %v, want %v", got, *tt.want)
			}
		})
	}
}
