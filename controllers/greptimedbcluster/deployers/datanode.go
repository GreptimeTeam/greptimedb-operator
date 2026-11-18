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

package deployers

import (
	"context"
	stderrors "errors"
	"fmt"
	"path"
	"reflect"
	"sort"
	"sync"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/GreptimeTeam/greptimedb-operator/apis/v1alpha1"
	"github.com/GreptimeTeam/greptimedb-operator/controllers/common"
	"github.com/GreptimeTeam/greptimedb-operator/controllers/constant"
	"github.com/GreptimeTeam/greptimedb-operator/pkg/dbconfig"
	"github.com/GreptimeTeam/greptimedb-operator/pkg/deployer"
	"github.com/GreptimeTeam/greptimedb-operator/pkg/util"
	k8sutils "github.com/GreptimeTeam/greptimedb-operator/pkg/util/k8s"
)

// DatanodeDeployer is the deployer for datanode.
type DatanodeDeployer struct {
	*CommonDeployer

	// maintenanceMode tracks whether WE turned on datanode maintenance mode,
	// PER CLUSTER (keyed by the cluster's NamespacedName). A single operator
	// deployment typically reconciles several GreptimeDBClusters; with the
	// previous single bool, the disable fired against whichever cluster's
	// reconcile happened to run the post-sync hook while a datanode roll was
	// in progress, leaving the actually-rolling cluster in maintenance mode
	// (region failover + GC disabled) with no self-heal.
	maintenanceMode sync.Map

	// drainTrackers tracks in-flight drain procedure IDs per cluster and per
	// datanode ID. It is keyed by the cluster's NamespacedName and maps to the
	// cluster's drainTracker. The state is intentionally in-memory: a drain
	// submission is idempotent, so losing it on an operator restart only causes
	// a re-submit (which metasrv either rejects while a drain is running or
	// turns into a quick no-op once the node has no leader regions left).
	drainTrackers sync.Map

	// drainUnsupported remembers, per cluster, that the metasrv does not expose
	// the drain API (GreptimeDB < v1.4.0). Once observed, the drain orchestration
	// is skipped for that cluster instead of repeatedly hitting 404 endpoints.
	drainUnsupported sync.Map
}

var _ deployer.Deployer = &DatanodeDeployer{}

func NewDatanodeDeployer(mgr ctrl.Manager) *DatanodeDeployer {
	return &DatanodeDeployer{
		CommonDeployer: NewFromManager(mgr),
	}
}

func (d *DatanodeDeployer) NewBuilder(crdObject client.Object) deployer.Builder {
	return &datanodeBuilder{
		CommonBuilder: d.NewCommonBuilder(crdObject, v1alpha1.DatanodeRoleKind),
	}
}

func (d *DatanodeDeployer) Generate(crdObject client.Object) ([]client.Object, error) {
	objects, err := d.NewBuilder(crdObject).
		BuildService().
		BuildConfigMap().
		BuildStatefulSet().
		BuildPodMonitor().
		SetControllerAndAnnotation().
		Generate()

	if err != nil {
		return nil, err
	}

	return objects, nil
}

func (d *DatanodeDeployer) CleanUp(ctx context.Context, crdObject client.Object) error {
	cluster, err := d.GetCluster(crdObject)
	if err != nil {
		return err
	}

	if fs := cluster.GetDatanode().GetFileStorage(); fs != nil {
		if fs.IsUseEmptyDir() {
			return nil
		}

		if fs.GetPolicy() == v1alpha1.StorageRetainPolicyTypeDelete {
			if err := d.deleteStorage(ctx, cluster.Namespace, common.ResourceName(cluster.Name, v1alpha1.DatanodeRoleKind), common.FileStorageTypeDatanode); err != nil {
				return err
			}
		}
	}

	for _, datanodeGroup := range cluster.GetDatanodeGroups() {
		if fs := datanodeGroup.GetFileStorage(); fs != nil {
			if fs.IsUseEmptyDir() {
				continue
			}

			if fs.GetPolicy() == v1alpha1.StorageRetainPolicyTypeDelete {
				if err := d.deleteStorage(ctx, cluster.Namespace, common.ResourceName(cluster.Name, v1alpha1.DatanodeRoleKind, datanodeGroup.GetName()), common.FileStorageTypeDatanode); err != nil {
					return err
				}
			}
		}
	}

	if cluster.GetWALProvider().GetRaftEngineWAL().GetFileStorage().GetPolicy() == v1alpha1.StorageRetainPolicyTypeDelete {
		if err := d.deleteStorage(ctx, cluster.Namespace, common.ResourceName(cluster.Name, v1alpha1.DatanodeRoleKind), common.FileStorageTypeWAL); err != nil {
			return err
		}
	}

	if cluster.GetObjectStorageProvider().GetCacheFileStorage().GetPolicy() == v1alpha1.StorageRetainPolicyTypeDelete {
		if err := d.deleteStorage(ctx, cluster.Namespace, common.ResourceName(cluster.Name, v1alpha1.DatanodeRoleKind), common.FileStorageTypeCache); err != nil {
			return err
		}
	}

	return nil
}

func (d *DatanodeDeployer) CheckAndUpdateStatus(ctx context.Context, crdObject client.Object) (bool, error) {
	cluster, err := d.GetCluster(crdObject)
	if err != nil {
		return false, err
	}

	if cluster.GetDatanode() != nil {
		return d.checkDatanodeStatus(ctx, cluster.Namespace, common.ResourceName(cluster.Name, v1alpha1.DatanodeRoleKind), cluster)
	}

	if len(cluster.GetDatanodeGroups()) > 0 {
		return d.checkDatanodeGroupsStatus(ctx, cluster)
	}

	return true, nil
}

// Apply is re-implemented for datanode to handle the maintenance mode.
func (d *DatanodeDeployer) Apply(ctx context.Context, crdObject client.Object, objects []client.Object) error {
	updateObject := false

	cluster, err := d.GetCluster(crdObject)
	if err != nil {
		return err
	}

	for _, newObject := range objects {
		oldObject, err := k8sutils.CreateObjectIfNotExist(ctx, d.Client, k8sutils.SourceObject(newObject), newObject)
		if err != nil {
			return err
		}

		if oldObject != nil {
			specEqual, err := k8sutils.IsObjectSpecEqual(oldObject, newObject, deployer.LastAppliedResourceSpec)
			if err != nil {
				return err
			}

			labelsEqual := k8sutils.IsObjectLabelsEqual(oldObject.GetLabels(), newObject.GetLabels())

			// If the spec or labels is not equal, update the object.
			if !specEqual || !labelsEqual {
				if newSts, ok := newObject.(*appsv1.StatefulSet); ok {
					// oldObject is an *unstructured.Unstructured (see SourceObject), so
					// re-fetch the existing StatefulSet as a typed object for the
					// scale-safety checks below.
					oldSts := new(appsv1.StatefulSet)
					if err := d.Get(ctx, client.ObjectKeyFromObject(newSts), oldSts); err != nil {
						return err
					}
					if err := d.ensureDatanodeScaleSafety(ctx, cluster, newSts, oldSts); err != nil {
						return err
					}
					if err := d.ensureRollingUpdateDrain(ctx, cluster, newSts, oldSts); err != nil {
						return err
					}
				}

				if sts, ok := newObject.(*appsv1.StatefulSet); ok && d.shouldUseMaintenanceMode(cluster) {
					if err := d.turnOnMaintenanceMode(ctx, sts, cluster); err != nil {
						return err
					}
				}
				if err := d.Patch(ctx, newObject, client.MergeFrom(oldObject)); err != nil {
					return err
				}
				updateObject = true
			}
		}
	}

	if updateObject {
		// If the object is updated, we need to wait for the object to be ready.
		// When the related object is ready, we will receive the event and enter the next reconcile loop.
		return deployer.ErrSyncNotReady
	}

	return nil
}

// drainTracker tracks in-flight drain procedures per datanode node ID for a
// single cluster. The state is mirrored to the cluster's DatanodeStatus.Draining
// so a restarted operator can resume polling the same procedure IDs.
type drainTracker struct {
	mu   sync.Mutex
	proc map[uint64]drainProc
}

type drainProc struct {
	procedureID string
	status      string
}

func (t *drainTracker) get(nodeID uint64) (drainProc, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	p, ok := t.proc[nodeID]
	return p, ok
}

func (t *drainTracker) put(nodeID uint64, p drainProc) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.proc[nodeID] = p
}

func (t *drainTracker) delete(nodeID uint64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.proc, nodeID)
}

// snapshot returns a copy of the tracked drains, sorted by node ID for a stable
// status output.
func (t *drainTracker) snapshot() []v1alpha1.DatanodeDrainStatus {
	t.mu.Lock()
	defer t.mu.Unlock()

	out := make([]v1alpha1.DatanodeDrainStatus, 0, len(t.proc))
	for nodeID, p := range t.proc {
		out = append(out, v1alpha1.DatanodeDrainStatus{
			NodeID:      nodeID,
			ProcedureID: p.procedureID,
			Status:      p.status,
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].NodeID < out[j].NodeID })
	return out
}

// reset rebuilds the tracker from persisted drain state, keeping only the given
// node IDs. This scopes the tracker to the current drain operation and drops
// stale entries left by an aborted previous operation.
func (t *drainTracker) reset(draining []v1alpha1.DatanodeDrainStatus, keep []uint64) {
	t.mu.Lock()
	defer t.mu.Unlock()

	keepSet := make(map[uint64]struct{}, len(keep))
	for _, id := range keep {
		keepSet[id] = struct{}{}
	}

	t.proc = make(map[uint64]drainProc)
	for _, s := range draining {
		if s.ProcedureID == "" {
			continue
		}
		if _, ok := keepSet[s.NodeID]; ok {
			t.proc[s.NodeID] = drainProc{procedureID: s.ProcedureID, status: s.Status}
		}
	}
}

func (d *DatanodeDeployer) getDrainTracker(cluster *v1alpha1.GreptimeDBCluster) *drainTracker {
	key := client.ObjectKeyFromObject(cluster)
	value, _ := d.drainTrackers.LoadOrStore(key, &drainTracker{proc: map[uint64]drainProc{}})
	return value.(*drainTracker)
}

func (d *DatanodeDeployer) isDrainUnsupported(cluster *v1alpha1.GreptimeDBCluster) bool {
	_, ok := d.drainUnsupported.Load(client.ObjectKeyFromObject(cluster))
	return ok
}

func (d *DatanodeDeployer) markDrainUnsupported(cluster *v1alpha1.GreptimeDBCluster) {
	d.drainUnsupported.Store(client.ObjectKeyFromObject(cluster), struct{}{})
}

// ensureDatanodeScaleSafety drains datanodes before they are removed by a
// scale-down and restores scheduling for datanodes added by a scale-up.
func (d *DatanodeDeployer) ensureDatanodeScaleSafety(ctx context.Context, cluster *v1alpha1.GreptimeDBCluster, newSts, oldSts *appsv1.StatefulSet) error {
	if newSts.Spec.Replicas == nil || oldSts.Spec.Replicas == nil {
		return nil
	}

	spec, _, err := datanodeSpecForStatefulSet(cluster, newSts.Name)
	if err != nil {
		return err
	}
	if !spec.IsEnableDrain() {
		return nil
	}

	newReplicas := *newSts.Spec.Replicas
	oldReplicas := *oldSts.Spec.Replicas

	switch {
	case newReplicas < oldReplicas:
		drained, err := d.drainScaleDownPods(ctx, cluster, newSts, newReplicas, oldReplicas)
		if err != nil {
			return err
		}
		if !drained {
			return deployer.ErrDrainInProgress
		}
	case newReplicas > oldReplicas:
		return d.markScaleUpPodsSchedulable(ctx, cluster, newSts, oldReplicas, newReplicas)
	}

	return nil
}

// drainScaleDownPods drains the datanodes whose StatefulSet ordinals fall in
// [newReplicas, oldReplicas). It returns true only when every node is fully drained.
func (d *DatanodeDeployer) drainScaleDownPods(ctx context.Context, cluster *v1alpha1.GreptimeDBCluster, sts *appsv1.StatefulSet, newReplicas, oldReplicas int32) (bool, error) {
	groupID, startNodeID, err := d.datanodeNodeIDParams(cluster, sts.Name)
	if err != nil {
		return false, err
	}

	metaURL := common.GetMetaHTTPServiceURL(cluster)

	if d.isDrainUnsupported(cluster) {
		// Older metasrv without the drain API: fall back to the previous
		// scale-down behavior (rely on region failover).
		return true, nil
	}

	tracker := d.getDrainTracker(cluster)

	nodeIDs := make([]uint64, 0, oldReplicas-newReplicas)
	for ordinal := newReplicas; ordinal < oldReplicas; ordinal++ {
		nodeIDs = append(nodeIDs, common.DatanodeNodeID(ordinal, groupID, startNodeID))
	}
	tracker.reset(cluster.Status.Datanode.Draining, nodeIDs)

	allDrained := true
	for _, nodeID := range nodeIDs {
		drained, err := d.drainNode(ctx, metaURL, tracker, nodeID)
		if stderrors.Is(err, common.ErrDatanodeDrainUnsupported) {
			d.markDrainUnsupported(cluster)
			return true, nil
		}
		if err != nil {
			return false, err
		}
		if !drained {
			allDrained = false
		}
	}

	if err := d.persistDrainStatus(ctx, cluster, tracker); err != nil {
		return false, err
	}
	return allDrained, nil
}

// drainNode ensures the given datanode has no leader regions left. It returns:
//   - (true, nil) when the node is drained,
//   - (false, nil) when a drain is still in progress,
//   - (false, err) on failure.
func (d *DatanodeDeployer) drainNode(ctx context.Context, metaURL string, tracker *drainTracker, nodeID uint64) (bool, error) {
	proc, ok := tracker.get(nodeID)
	if !ok || proc.procedureID == "" {
		id, err := common.SubmitDatanodeDrain(metaURL, nodeID)
		if err != nil {
			if stderrors.Is(err, common.ErrDatanodeDrainUnsupported) {
				return false, fmt.Errorf("%w: node %d", common.ErrDatanodeDrainUnsupported, nodeID)
			}
			if stderrors.Is(err, common.ErrDatanodeDrainRunning) {
				// A drain submitted by a previous operator instance or a concurrent
				// reconcile is still running; wait and resubmit once it settles.
				klog.V(2).Infof("datanode %d is already being drained; waiting for it to finish", nodeID)
				return false, nil
			}
			return false, err
		}
		proc = drainProc{procedureID: id, status: common.ProcedureStateRunning}
		tracker.put(nodeID, proc)
	}

	status, err := common.ProcedureStatus(metaURL, proc.procedureID)
	if err != nil {
		return false, err
	}
	proc.status = status

	switch status {
	case common.ProcedureStateDone:
		tracker.delete(nodeID)
		klog.Infof("Drained datanode %d", nodeID)
		return true, nil
	case common.ProcedureStateFailed, common.ProcedureStatePoisoned:
		tracker.delete(nodeID)
		return false, fmt.Errorf("drain procedure %s for datanode %d ended in state %s", proc.procedureID, nodeID, status)
	default:
		// Running, Retrying, PrepareRollback, or RollingBack.
		tracker.put(nodeID, proc)
		return false, nil
	}
}

// persistDrainStatus mirrors the tracker into the cluster's DatanodeStatus.Draining.
func (d *DatanodeDeployer) persistDrainStatus(ctx context.Context, cluster *v1alpha1.GreptimeDBCluster, tracker *drainTracker) error {
	draining := tracker.snapshot()
	if reflect.DeepEqual(cluster.Status.Datanode.Draining, draining) {
		return nil
	}

	cluster.Status.Datanode.Draining = draining
	if err := UpdateStatus(ctx, cluster, d.Client); err != nil {
		klog.Errorf("Failed to update drain status: %s", err)
		return err
	}
	return nil
}

// ensureRollingUpdateDrain drains every datanode before a pod-restarting update
// (e.g. an image or config change) so leader regions are migrated away before the
// StatefulSet controller rolls the pods. The drained nodes are recorded in
// DatanodeStatus.PendingSchedulable and marked schedulable again once the rollout
// completes. It only runs when region failover is enabled, matching the existing
// maintenance-mode behavior.
func (d *DatanodeDeployer) ensureRollingUpdateDrain(ctx context.Context, cluster *v1alpha1.GreptimeDBCluster, newSts, oldSts *appsv1.StatefulSet) error {
	if newSts.Spec.Replicas == nil || oldSts.Spec.Replicas == nil {
		return nil
	}
	if *newSts.Spec.Replicas != *oldSts.Spec.Replicas {
		// Scale changes are handled by ensureDatanodeScaleSafety.
		return nil
	}

	spec, _, err := datanodeSpecForStatefulSet(cluster, newSts.Name)
	if err != nil {
		return err
	}
	if !spec.IsEnableDrain() {
		return nil
	}

	if !d.shouldUseMaintenanceMode(cluster) {
		return nil
	}
	if !d.isOldPodRestart(*newSts, *oldSts) {
		return nil
	}
	if d.isDrainUnsupported(cluster) {
		return nil
	}

	groupID, startNodeID, err := d.datanodeNodeIDParams(cluster, newSts.Name)
	if err != nil {
		return err
	}

	metaURL := common.GetMetaHTTPServiceURL(cluster)
	tracker := d.getDrainTracker(cluster)

	replicas := *oldSts.Spec.Replicas
	nodeIDs := make([]uint64, 0, replicas)
	for ordinal := replicas - 1; ordinal >= 0; ordinal-- {
		nodeIDs = append(nodeIDs, common.DatanodeNodeID(ordinal, groupID, startNodeID))
	}
	tracker.reset(cluster.Status.Datanode.Draining, nodeIDs)

	allDrained := true
	for _, nodeID := range nodeIDs {
		drained, err := d.drainNode(ctx, metaURL, tracker, nodeID)
		if stderrors.Is(err, common.ErrDatanodeDrainUnsupported) {
			d.markDrainUnsupported(cluster)
			return nil
		}
		if err != nil {
			return err
		}
		if !drained {
			allDrained = false
		}
	}

	if err := d.persistDrainStatus(ctx, cluster, tracker); err != nil {
		return err
	}

	if !allDrained {
		return deployer.ErrDrainInProgress
	}

	return d.persistPendingSchedulable(ctx, cluster, nodeIDs)
}

// persistPendingSchedulable records the node IDs that must be marked schedulable
// after the rollout completes.
func (d *DatanodeDeployer) persistPendingSchedulable(ctx context.Context, cluster *v1alpha1.GreptimeDBCluster, pending []uint64) error {
	sort.Slice(pending, func(i, j int) bool { return pending[i] < pending[j] })
	if reflect.DeepEqual(cluster.Status.Datanode.PendingSchedulable, pending) {
		return nil
	}

	cluster.Status.Datanode.PendingSchedulable = pending
	if err := UpdateStatus(ctx, cluster, d.Client); err != nil {
		klog.Errorf("Failed to update pending schedulable status: %s", err)
		return err
	}
	return nil
}

// markPendingSchedulable restores scheduling for datanodes that were drained for
// a rolling update. It runs as a post-sync hook, i.e. after the rolled pods are
// ready again.
func (d *DatanodeDeployer) markPendingSchedulable(ctx context.Context, crdObject client.Object) error {
	cluster, err := d.GetCluster(crdObject)
	if err != nil {
		return err
	}

	pending := cluster.Status.Datanode.PendingSchedulable
	if len(pending) == 0 {
		return nil
	}

	if d.isDrainUnsupported(cluster) {
		cluster.Status.Datanode.PendingSchedulable = nil
		if err := UpdateStatus(ctx, cluster, d.Client); err != nil {
			klog.Errorf("Failed to clear pending schedulable status: %s", err)
			return err
		}
		return nil
	}

	metaURL := common.GetMetaHTTPServiceURL(cluster)
	for _, nodeID := range pending {
		if err := common.MarkDatanodeSchedulable(metaURL, nodeID); err != nil {
			if stderrors.Is(err, common.ErrDatanodeDrainUnsupported) {
				d.markDrainUnsupported(cluster)
				cluster.Status.Datanode.PendingSchedulable = nil
				if err := UpdateStatus(ctx, cluster, d.Client); err != nil {
					klog.Errorf("Failed to clear pending schedulable status: %s", err)
					return err
				}
				return nil
			}
			return err
		}
		klog.Infof("Marked datanode %d schedulable after rolling update", nodeID)
	}

	cluster.Status.Datanode.PendingSchedulable = nil
	if err := UpdateStatus(ctx, cluster, d.Client); err != nil {
		klog.Errorf("Failed to clear pending schedulable status: %s", err)
		return err
	}
	return nil
}

// markScaleUpPodsSchedulable restores scheduling for datanodes whose ordinals fall
// in [oldReplicas, newReplicas), removing any fence left by a previous drain.
func (d *DatanodeDeployer) markScaleUpPodsSchedulable(ctx context.Context, cluster *v1alpha1.GreptimeDBCluster, sts *appsv1.StatefulSet, oldReplicas, newReplicas int32) error {
	if d.isDrainUnsupported(cluster) {
		return nil
	}

	groupID, startNodeID, err := d.datanodeNodeIDParams(cluster, sts.Name)
	if err != nil {
		return err
	}

	metaURL := common.GetMetaHTTPServiceURL(cluster)
	for ordinal := oldReplicas; ordinal < newReplicas; ordinal++ {
		nodeID := common.DatanodeNodeID(ordinal, groupID, startNodeID)
		if err := common.MarkDatanodeSchedulable(metaURL, nodeID); err != nil {
			if stderrors.Is(err, common.ErrDatanodeDrainUnsupported) {
				d.markDrainUnsupported(cluster)
				return nil
			}
			return err
		}
		klog.Infof("Marked datanode %d schedulable after scale-up", nodeID)
	}

	// Scaling up aborts any in-flight scale-down drain: the nodes being added
	// back must not remain fenced. Clear the stale drain state.
	return d.clearDrainState(ctx, cluster)
}

// clearDrainState drops the in-memory tracker and persisted Draining status for
// the cluster.
func (d *DatanodeDeployer) clearDrainState(ctx context.Context, cluster *v1alpha1.GreptimeDBCluster) error {
	tracker := d.getDrainTracker(cluster)
	tracker.reset(nil, nil)

	if len(cluster.Status.Datanode.Draining) == 0 {
		return nil
	}
	cluster.Status.Datanode.Draining = nil
	if err := UpdateStatus(ctx, cluster, d.Client); err != nil {
		klog.Errorf("Failed to clear drain status: %s", err)
		return err
	}
	return nil
}

// datanodeNodeIDParams resolves the group ID and start node ID used to compute
// datanode node IDs for the given StatefulSet.
func (d *DatanodeDeployer) datanodeNodeIDParams(cluster *v1alpha1.GreptimeDBCluster, stsName string) (groupID, startNodeID int32, err error) {
	spec, groupID, err := datanodeSpecForStatefulSet(cluster, stsName)
	if err != nil {
		return 0, 0, err
	}
	if snid := spec.GetStartNodeID(); snid != nil {
		startNodeID = *snid
	}
	return groupID, startNodeID, nil
}

// datanodeSpecForStatefulSet resolves the datanode spec and group ID that a
// StatefulSet was generated from. The single datanode has group ID -1; each
// datanode group uses its index in DatanodeGroups.
func datanodeSpecForStatefulSet(cluster *v1alpha1.GreptimeDBCluster, stsName string) (*v1alpha1.DatanodeSpec, int32, error) {
	if cluster.GetDatanode() != nil {
		return cluster.GetDatanode(), -1, nil
	}

	for i, group := range cluster.GetDatanodeGroups() {
		if common.ResourceName(cluster.Name, v1alpha1.DatanodeRoleKind, group.GetName()) == stsName {
			return group, int32(i), nil
		}
	}

	return nil, 0, fmt.Errorf("no datanode spec matches statefulset %s", stsName)
}

func (d *DatanodeDeployer) PostSyncHooks() []deployer.Hook {
	return []deployer.Hook{
		d.markPendingSchedulable,
		d.turnOffMaintenanceMode,
	}
}

func (d *DatanodeDeployer) checkDatanodeGroupsStatus(ctx context.Context, cluster *v1alpha1.GreptimeDBCluster) (bool, error) {
	var (
		readyCount         int32
		totalReadyReplicas int32
	)

	for _, spec := range cluster.GetDatanodeGroups() {
		resourceName := common.ResourceName(cluster.Name, v1alpha1.DatanodeRoleKind, spec.GetName())

		ready, readyReplicas, err := d.getStatefulSetStatus(ctx, cluster.Namespace, resourceName)
		if err != nil {
			return false, err
		}

		totalReadyReplicas += readyReplicas
		if ready {
			readyCount++
		}
	}

	// Update the status if the cluster is not nil.
	if cluster != nil {
		cluster.Status.Datanode.Replicas = cluster.GetDatanodeReplicas()
		cluster.Status.Datanode.ReadyReplicas = totalReadyReplicas
		if err := UpdateStatus(ctx, cluster, d.Client); err != nil {
			klog.Errorf("Failed to update status: %s", err)
		}
	}

	if readyCount == int32(len(cluster.GetDatanodeGroups())) {
		return true, nil
	}

	return false, nil
}

func (d *DatanodeDeployer) checkDatanodeStatus(ctx context.Context, namespace, resourceName string, cluster *v1alpha1.GreptimeDBCluster) (bool, error) {
	ready, readyReplicas, err := d.getStatefulSetStatus(ctx, namespace, resourceName)
	if err != nil {
		return false, err
	}

	// Update the status if the cluster is not nil.
	if cluster != nil {
		cluster.Status.Datanode.Replicas = cluster.GetDatanodeReplicas()
		cluster.Status.Datanode.ReadyReplicas = readyReplicas
		if err := UpdateStatus(ctx, cluster, d.Client); err != nil {
			klog.Errorf("Failed to update status: %s", err)
		}
	}

	return ready, nil
}

func (d *DatanodeDeployer) getStatefulSetStatus(ctx context.Context, namespace, resourceName string) (bool, int32, error) {
	var (
		sts = new(appsv1.StatefulSet)

		objectKey = client.ObjectKey{
			Namespace: namespace,
			Name:      resourceName,
		}
	)

	err := d.Get(ctx, objectKey, sts)
	if errors.IsNotFound(err) {
		return false, 0, nil
	}
	if err != nil {
		return false, 0, err
	}

	return k8sutils.IsStatefulSetReady(sts), sts.Status.ReadyReplicas, nil
}

func (d *DatanodeDeployer) turnOnMaintenanceMode(ctx context.Context, newSts *appsv1.StatefulSet, cluster *v1alpha1.GreptimeDBCluster) error {
	oldSts := new(appsv1.StatefulSet)
	// The oldSts must exist since we have checked it before.
	if err := d.Get(ctx, client.ObjectKeyFromObject(newSts), oldSts); err != nil {
		return err
	}

	key := client.ObjectKeyFromObject(cluster)
	if _, armed := d.maintenanceMode.Load(key); !armed && d.isOldPodRestart(*newSts, *oldSts) {
		klog.Infof("Turn on maintenance mode for datanode, cluster: %s, statefulset: %s", cluster.Name, newSts.Name)
		// FIXME(zyy17): Should record the maintenance mode in the status.
		if err := common.SetMaintenanceMode(common.GetMetaHTTPServiceURL(cluster), true); err != nil {
			return err
		}
		d.maintenanceMode.Store(key, true)
	}

	return nil
}

func (d *DatanodeDeployer) turnOffMaintenanceMode(ctx context.Context, crdObject client.Object) error {
	cluster, err := d.GetCluster(crdObject)
	if err != nil {
		return err
	}

	key := client.ObjectKeyFromObject(cluster)
	if _, armed := d.maintenanceMode.Load(key); armed && d.shouldUseMaintenanceMode(cluster) {
		klog.Infof("Turn off maintenance mode for datanode, cluster: %s", cluster.Name)
		// FIXME(zyy17): Should record the maintenance mode in the status.
		if err := common.SetMaintenanceMode(common.GetMetaHTTPServiceURL(cluster), false); err != nil {
			return err
		}
		d.maintenanceMode.Delete(key)
	}

	return nil
}

func (d *DatanodeDeployer) deleteStorage(ctx context.Context, namespace, resourceName string, fsType common.FileStorageType) error {
	klog.Infof("Deleting datanode storage...")

	claims, err := common.GetPVCs(ctx, d.Client, namespace, resourceName, fsType)
	if err != nil {
		return err
	}

	for _, pvc := range claims {
		klog.Infof("Deleting datanode PVC: %s", pvc.Name)
		if err := d.Delete(ctx, &pvc); err != nil {
			return err
		}
	}

	return nil
}

// isOldPodRestart checks if the existed pod needs to be restarted. For convenience, we only compare the necessary fields.
// TODO(zyy17): Do we have a easy way to implement this?
func (d *DatanodeDeployer) isOldPodRestart(new, old appsv1.StatefulSet) bool {
	var (
		newPodTemplate = new.Spec.Template
		oldPodTemplate = old.Spec.Template
	)

	if !reflect.DeepEqual(newPodTemplate.GetObjectMeta().GetAnnotations(), oldPodTemplate.GetObjectMeta().GetAnnotations()) {
		return true
	}

	if newPodTemplate.Spec.InitContainers[0].Image != oldPodTemplate.Spec.InitContainers[0].Image {
		return true
	}

	// If the tolerations, affinity, nodeSelector are changed, the original Pod may need to be restarted for re-scheduling.
	if !reflect.DeepEqual(newPodTemplate.Spec.Tolerations, oldPodTemplate.Spec.Tolerations) ||
		!reflect.DeepEqual(newPodTemplate.Spec.Affinity, oldPodTemplate.Spec.Affinity) ||
		!reflect.DeepEqual(newPodTemplate.Spec.NodeSelector, oldPodTemplate.Spec.NodeSelector) {
		return true
	}

	// Compare the main container settings.
	newMainContainer := newPodTemplate.Spec.Containers[constant.MainContainerIndex]
	oldMainContainer := oldPodTemplate.Spec.Containers[constant.MainContainerIndex]
	if newMainContainer.Image != oldMainContainer.Image {
		return true
	}

	if !reflect.DeepEqual(newMainContainer.Command, oldMainContainer.Command) ||
		!reflect.DeepEqual(newMainContainer.Args, oldMainContainer.Args) ||
		!reflect.DeepEqual(newMainContainer.Env, oldMainContainer.Env) ||
		!reflect.DeepEqual(newMainContainer.VolumeMounts, oldMainContainer.VolumeMounts) ||
		!reflect.DeepEqual(newMainContainer.Ports, oldMainContainer.Ports) ||
		!reflect.DeepEqual(newMainContainer.Resources, oldMainContainer.Resources) {
		return true
	}

	return false
}

func (d *DatanodeDeployer) shouldUseMaintenanceMode(cluster *v1alpha1.GreptimeDBCluster) bool {
	return cluster.GetMeta().IsEnableRegionFailover()
}

var _ deployer.Builder = &datanodeBuilder{}

type datanodeBuilder struct {
	*CommonBuilder
}

func (b *datanodeBuilder) BuildService() deployer.Builder {
	if b.Err != nil {
		return b
	}

	if b.Cluster.GetDatanode() == nil && len(b.Cluster.GetDatanodeGroups()) == 0 {
		return b
	}

	if b.Cluster.GetDatanode() != nil {
		svc, err := b.generateDatanodeService(b.Cluster.GetDatanode())
		if err != nil {
			b.Err = err
			return b
		}
		b.Objects = append(b.Objects, svc)
		return b
	}

	for _, datanodeSpec := range b.Cluster.GetDatanodeGroups() {
		svc, err := b.generateDatanodeService(datanodeSpec)
		if err != nil {
			b.Err = err
			return b
		}
		b.Objects = append(b.Objects, svc)
	}

	return b
}

func (b *datanodeBuilder) BuildConfigMap() deployer.Builder {
	if b.Err != nil {
		return b
	}

	if b.Cluster.GetDatanode() == nil && len(b.Cluster.GetDatanodeGroups()) == 0 {
		return b
	}

	if b.Cluster.GetDatanode() != nil {
		cm, err := b.GenerateConfigMap(b.Cluster.GetDatanode())
		if err != nil {
			b.Err = err
			return b
		}
		b.Objects = append(b.Objects, cm)
	}

	for _, datanodeSpec := range b.Cluster.GetDatanodeGroups() {
		cm, err := b.GenerateConfigMap(datanodeSpec)
		if err != nil {
			b.Err = err
			return b
		}
		b.Objects = append(b.Objects, cm)
	}

	return b
}

func (b *datanodeBuilder) BuildStatefulSet() deployer.Builder {
	if b.Err != nil {
		return b
	}

	if b.Cluster.GetDatanode() == nil && len(b.Cluster.GetDatanodeGroups()) == 0 {
		return b
	}

	if b.Cluster.GetMonitoring().IsEnabled() && b.Cluster.GetMonitoring().GetVector() != nil {
		cm, err := b.GenerateVectorConfigMap()
		if err != nil {
			b.Err = err
			return b
		}
		b.Objects = append(b.Objects, cm)
	}

	if b.Cluster.GetDatanode() != nil {
		sts, err := b.generateDatanodeStatefulSet(nil, b.Cluster.GetDatanode())
		if err != nil {
			b.Err = err
			return b
		}
		b.Objects = append(b.Objects, sts)
		return b
	}

	for i, datanodeSpec := range b.Cluster.GetDatanodeGroups() {
		groupID := int32(i)
		sts, err := b.generateDatanodeStatefulSet(&groupID, datanodeSpec)
		if err != nil {
			b.Err = err
			return b
		}
		b.Objects = append(b.Objects, sts)
	}

	return b
}

func (b *datanodeBuilder) BuildPodMonitor() deployer.Builder {
	if b.Err != nil {
		return b
	}

	if b.Cluster.GetDatanode() == nil && len(b.Cluster.GetDatanodeGroups()) == 0 {
		return b
	}

	if b.Cluster.Spec.PrometheusMonitor == nil || !b.Cluster.Spec.PrometheusMonitor.Enabled {
		return b
	}

	if b.Cluster.GetDatanode() != nil {
		pm, err := b.GeneratePodMonitor(b.Cluster.Namespace, common.ResourceName(b.Cluster.Name, b.RoleKind))
		if err != nil {
			b.Err = err
			return b
		}
		b.Objects = append(b.Objects, pm)

		return b
	}

	for _, datanodeSpec := range b.Cluster.GetDatanodeGroups() {
		pm, err := b.GeneratePodMonitor(b.Cluster.Namespace, common.ResourceName(b.Cluster.Name, b.RoleKind, datanodeSpec.GetName()))
		if err != nil {
			b.Err = err
			return b
		}
		b.Objects = append(b.Objects, pm)
	}

	return b
}

func (b *datanodeBuilder) generateDatanodeStatefulSet(groupID *int32, spec *v1alpha1.DatanodeSpec) (*appsv1.StatefulSet, error) {
	resourceName := common.ResourceName(b.Cluster.Name, b.RoleKind, spec.GetName())

	sts := &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "StatefulSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      resourceName,
			Namespace: b.Cluster.Namespace,
			Labels: map[string]string{
				constant.GreptimeDBComponentName: resourceName,
			},
		},
		Spec: appsv1.StatefulSetSpec{
			PodManagementPolicy: appsv1.ParallelPodManagement,
			ServiceName:         resourceName,
			Replicas:            spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					constant.GreptimeDBComponentName: resourceName,
				},
			},
			Template: b.generatePodTemplateSpec(spec, groupID),
			UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
				Type:          appsv1.RollingUpdateStatefulSetStrategyType,
				RollingUpdate: spec.RollingUpdate,
			},
		},
	}

	if !spec.GetFileStorage().IsUseEmptyDir() {
		sts.Spec.VolumeClaimTemplates = b.generatePVCs(spec)
	}

	configData, err := dbconfig.FromCluster(b.Cluster, spec)
	if err != nil {
		return nil, err
	}

	sts.Spec.Template.Annotations = util.MergeStringMap(sts.Spec.Template.Annotations,
		map[string]string{deployer.ConfigHash: util.CalculateConfigHash(configData)})

	return sts, nil
}

func (b *datanodeBuilder) generateDatanodeService(spec *v1alpha1.DatanodeSpec) (*corev1.Service, error) {
	resourceName := common.ResourceName(b.Cluster.Name, b.RoleKind, spec.GetName())

	svc := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: b.Cluster.Namespace,
			Name:      resourceName,
			Labels: map[string]string{
				constant.GreptimeDBComponentName: resourceName,
			},
		},
		Spec: corev1.ServiceSpec{
			ClusterIP: corev1.ClusterIPNone,
			Selector: map[string]string{
				constant.GreptimeDBComponentName: resourceName,
			},
			Ports: b.servicePorts(spec),
		},
	}

	return svc, nil
}

func (b *datanodeBuilder) generateMainContainerArgs(spec *v1alpha1.DatanodeSpec) []string {
	return []string{
		"datanode", "start",
		"--metasrv-addrs", fmt.Sprintf("%s.%s:%d", common.ResourceName(b.Cluster.Name, v1alpha1.MetaRoleKind),
			b.Cluster.Namespace, b.Cluster.Spec.Meta.RPCPort),
		"--http-addr", common.GetBindAddress(b.Cluster.Spec.EnableIPv6, spec.HTTPPort),
		"--config-file", path.Join(constant.GreptimeDBConfigDir, constant.GreptimeDBConfigFileName),
	}
}

func (b *datanodeBuilder) generatePodTemplateSpec(spec *v1alpha1.DatanodeSpec, groupID *int32) corev1.PodTemplateSpec {
	podTemplateSpec := b.GeneratePodTemplateSpec(spec.Template)

	if len(spec.Template.MainContainer.Args) == 0 {
		// Setup main container args.
		podTemplateSpec.Spec.Containers[constant.MainContainerIndex].Args = append(b.generateMainContainerArgs(spec), spec.Template.MainContainer.ExtraArgs...)
	}

	podTemplateSpec.Spec.Containers[constant.MainContainerIndex].Ports = b.containerPorts(spec)
	podTemplateSpec.Spec.Containers[constant.MainContainerIndex].Env = append(podTemplateSpec.Spec.Containers[constant.MainContainerIndex].Env, b.env(v1alpha1.DatanodeRoleKind)...)

	b.mountConfigDir(podTemplateSpec)
	b.addVolumeMounts(podTemplateSpec, spec)
	b.addInitConfigDirVolume(podTemplateSpec, common.ResourceName(b.Cluster.Name, b.RoleKind, spec.GetName()))

	if spec.GetFileStorage().IsUseEmptyDir() {
		b.addEmptyDirAsStorage(podTemplateSpec, spec)
	}

	if logging := spec.GetLogging(); logging != nil &&
		!logging.IsOnlyLogToStdout() && !logging.IsPersistentWithData() {
		b.AddLogsVolume(podTemplateSpec, logging.GetLogsDir())
	}

	if b.Cluster.GetMonitoring().IsEnabled() && b.Cluster.GetMonitoring().GetVector() != nil {
		b.AddVectorConfigVolume(podTemplateSpec)
		b.AddAuditLogsVolumeForVector(podTemplateSpec)
		b.AddVectorSidecar(podTemplateSpec, v1alpha1.DatanodeRoleKind)
	}

	podTemplateSpec.Spec.InitContainers = append(podTemplateSpec.Spec.InitContainers, *b.generateInitializer(spec, groupID))
	podTemplateSpec.Labels = util.MergeStringMap(podTemplateSpec.Labels, map[string]string{
		constant.GreptimeDBComponentName: common.ResourceName(b.Cluster.Name, b.RoleKind, spec.GetName()),
	})

	return *podTemplateSpec
}

func (b *datanodeBuilder) generatePVCs(spec *v1alpha1.DatanodeSpec) []corev1.PersistentVolumeClaim {
	var claims []corev1.PersistentVolumeClaim

	// It's always not nil because it's the default value.
	if fs := spec.GetFileStorage(); fs != nil {
		claims = append(claims, *common.FileStorageToPVC(b.Cluster.Name, spec.GetName(), fs, common.FileStorageTypeDatanode, v1alpha1.DatanodeRoleKind))
	}

	// Allocate the standalone WAL storage for the raft-engine.
	if fs := b.Cluster.GetWALProvider().GetRaftEngineWAL().GetFileStorage(); fs != nil {
		claims = append(claims, *common.FileStorageToPVC(b.Cluster.Name, spec.GetName(), fs, common.FileStorageTypeWAL, v1alpha1.DatanodeRoleKind))
	}

	// Allocate the standalone cache file storage for the datanode.
	if fs := b.Cluster.GetObjectStorageProvider().GetCacheFileStorage(); fs != nil {
		claims = append(claims, *common.FileStorageToPVC(b.Cluster.Name, spec.GetName(), fs, common.FileStorageTypeCache, v1alpha1.DatanodeRoleKind))
	}

	return claims
}

func (b *datanodeBuilder) generateInitializer(spec *v1alpha1.DatanodeSpec, groupID *int32) *corev1.Container {
	initializer := &corev1.Container{
		Name:  "initializer",
		Image: b.Cluster.Spec.Initializer.Image,
		Command: []string{
			"greptimedb-initializer",
		},
		Args: []string{
			"--config-path", path.Join(constant.GreptimeDBConfigDir, constant.GreptimeDBConfigFileName),
			"--init-config-path", path.Join(constant.GreptimeDBInitConfigDir, constant.GreptimeDBConfigFileName),
			"--rpc-port", fmt.Sprintf("%d", spec.RPCPort),
			"--component-kind", string(b.RoleKind),
		},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      constant.ConfigVolumeName,
				MountPath: constant.GreptimeDBConfigDir,
			},
			{
				Name:      constant.InitConfigVolumeName,
				MountPath: constant.GreptimeDBInitConfigDir,
			},
		},

		// TODO(zyy17): the datanode don't support to accept hostname.
		Env: b.env(v1alpha1.DatanodeRoleKind),
	}

	if groupID != nil {
		initializer.Args = append(initializer.Args, "--datanode-group-id", fmt.Sprintf("%d", *groupID))
	}

	if spec.GetStartNodeID() != nil {
		initializer.Args = append(initializer.Args, "--start-node-id", fmt.Sprintf("%d", *spec.GetStartNodeID()))
	}

	return initializer
}

func (b *datanodeBuilder) mountConfigDir(template *corev1.PodTemplateSpec) {
	// The empty-dir will be modified by initializer.
	template.Spec.Volumes = append(template.Spec.Volumes, corev1.Volume{
		Name: constant.ConfigVolumeName,
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		},
	})

	template.Spec.Containers[constant.MainContainerIndex].VolumeMounts =
		append(template.Spec.Containers[constant.MainContainerIndex].VolumeMounts,
			corev1.VolumeMount{
				Name:      constant.ConfigVolumeName,
				MountPath: constant.GreptimeDBConfigDir,
			},
		)
}

func (b *datanodeBuilder) addVolumeMounts(template *corev1.PodTemplateSpec, spec *v1alpha1.DatanodeSpec) {
	if fs := spec.GetFileStorage(); fs != nil {
		template.Spec.Containers[constant.MainContainerIndex].VolumeMounts =
			append(template.Spec.Containers[constant.MainContainerIndex].VolumeMounts,
				corev1.VolumeMount{
					Name:      fs.GetName(),
					MountPath: fs.GetMountPath(),
				},
			)
	}

	if fs := b.Cluster.GetWALProvider().GetRaftEngineWAL().GetFileStorage(); fs != nil {
		template.Spec.Containers[constant.MainContainerIndex].VolumeMounts =
			append(template.Spec.Containers[constant.MainContainerIndex].VolumeMounts,
				corev1.VolumeMount{
					Name:      fs.GetName(),
					MountPath: fs.GetMountPath(),
				},
			)
	}

	if fs := b.Cluster.GetObjectStorageProvider().GetCacheFileStorage(); fs != nil {
		template.Spec.Containers[constant.MainContainerIndex].VolumeMounts =
			append(template.Spec.Containers[constant.MainContainerIndex].VolumeMounts,
				corev1.VolumeMount{
					Name:      fs.GetName(),
					MountPath: fs.GetMountPath(),
				},
			)
	}
}

// The init-config volume is used for initializer.
func (b *datanodeBuilder) addInitConfigDirVolume(template *corev1.PodTemplateSpec, configMapName string) {
	template.Spec.Volumes = append(template.Spec.Volumes, corev1.Volume{
		Name: constant.InitConfigVolumeName,
		VolumeSource: corev1.VolumeSource{
			// Mount the configmap as init-config.
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: configMapName,
				},
			},
		},
	})
}

func (b *datanodeBuilder) addEmptyDirAsStorage(template *corev1.PodTemplateSpec, spec *v1alpha1.DatanodeSpec) {
	storageSize := resource.MustParse(spec.GetFileStorage().GetSize())
	template.Spec.Volumes = append(template.Spec.Volumes, corev1.Volume{
		Name: spec.GetFileStorage().GetName(),
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{
				SizeLimit: &storageSize,
			},
		},
	})
}

func (b *datanodeBuilder) servicePorts(spec *v1alpha1.DatanodeSpec) []corev1.ServicePort {
	return []corev1.ServicePort{
		{
			Name:     "rpc",
			Protocol: corev1.ProtocolTCP,
			Port:     spec.RPCPort,
		},
		{
			Name:     "http",
			Protocol: corev1.ProtocolTCP,
			Port:     spec.HTTPPort,
		},
	}
}

func (b *datanodeBuilder) containerPorts(spec *v1alpha1.DatanodeSpec) []corev1.ContainerPort {
	return []corev1.ContainerPort{
		{
			Name:          "rpc",
			Protocol:      corev1.ProtocolTCP,
			ContainerPort: spec.RPCPort,
		},
		{
			Name:          "http",
			Protocol:      corev1.ProtocolTCP,
			ContainerPort: spec.HTTPPort,
		},
	}
}
