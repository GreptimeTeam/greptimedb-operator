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
	"reflect"
	"testing"

	"github.com/GreptimeTeam/greptimedb-operator/apis/v1alpha1"
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
