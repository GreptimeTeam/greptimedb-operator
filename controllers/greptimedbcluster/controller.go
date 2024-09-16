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
	"errors"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/GreptimeTeam/greptimedb-operator/apis/v1alpha1"
	"github.com/GreptimeTeam/greptimedb-operator/cmd/operator/app/options"
	"github.com/GreptimeTeam/greptimedb-operator/controllers/greptimedbcluster/deployers"
	"github.com/GreptimeTeam/greptimedb-operator/pkg/deployer"
)

const (
	greptimedbClusterFinalizer = "greptimedbcluster.greptime.io/finalizer"
)

var (
	defaultRequeueAfter = 5 * time.Second
)

// Reconciler reconciles a GreptimeDBCluster object
type Reconciler struct {
	client.Client

	Scheme    *runtime.Scheme
	Deployers []deployer.Deployer
	Recorder  record.EventRecorder
}

func Setup(mgr ctrl.Manager, _ *options.Options) error {
	reconciler := &Reconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Recorder: mgr.GetEventRecorderFor("greptimedbcluster-controller"),
	}

	// sync will execute the sync logic of multiple deployers in order.
	reconciler.Deployers = []deployer.Deployer{
		deployers.NewMetaDeployer(mgr),
		deployers.NewDatanodeDeployer(mgr),
		deployers.NewFrontendDeployer(mgr),
		deployers.NewFlownodeDeployer(mgr),
	}

	return reconciler.SetupWithManager(mgr)
}

// SetupWithManager sets up the controller with the Manager.
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.GreptimeDBCluster{}).
		Owns(&corev1.Service{}).
		Owns(&appsv1.StatefulSet{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}

// +kubebuilder:rbac:groups=greptime.io,resources=greptimedbclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=greptime.io,resources=greptimedbclusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=greptime.io,resources=greptimedbclusters/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;patch;create;
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;patch;watch;create;update;delete;
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;patch;watch;
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;patch;create;update;
// +kubebuilder:rbac:groups=core,resources=events,verbs=get;list;patch;watch;create;
// +kubebuilder:rbac:groups=core,resources=persistentvolumeclaims,verbs=get;list;patch;watch;create;update;delete;
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=monitoring.coreos.com,resources=podmonitors,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apiextensions.k8s.io,resources=customresourcedefinitions,verbs=get;list;watch

// Reconcile is reconciliation loop for GreptimeDBCluster.
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	klog.V(2).Infof("Reconciling GreptimeDBCluster: %s", req.NamespacedName)

	var err error
	cluster := new(v1alpha1.GreptimeDBCluster)
	if err := r.Get(ctx, req.NamespacedName, cluster); err != nil {
		if k8serrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	defer func() {
		if err != nil {
			r.Recorder.Event(cluster, corev1.EventTypeWarning, "ReconcileError", fmt.Sprintf("Reconcile error: %v", err))
			if err := r.updateClusterStatus(ctx, cluster, v1alpha1.PhaseError); err != nil && !k8serrors.IsNotFound(err) {
				klog.Errorf("Failed to update status: %v", err)
				return
			}
		}
	}()

	// The object is being deleted.
	if !cluster.ObjectMeta.DeletionTimestamp.IsZero() {
		return r.delete(ctx, cluster)
	}

	if err = r.addFinalizer(ctx, cluster); err != nil {
		r.Recorder.Event(cluster, corev1.EventTypeWarning, "AddFinalizerFailed", fmt.Sprintf("Add finalizer failed: %v", err))
		return ctrl.Result{}, err
	}

	if err = cluster.Validate(); err != nil {
		r.Recorder.Event(cluster, corev1.EventTypeWarning, "InvalidCluster", fmt.Sprintf("Invalid cluster: %v", err))
		return ctrl.Result{}, err
	}

	if err = cluster.Check(ctx, r.Client); err != nil {
		r.Recorder.Event(cluster, corev1.EventTypeWarning, "InvalidCluster", fmt.Sprintf("Invalid cluster: %v", err))
		return ctrl.Result{}, err
	}

	// Means the cluster is just created.
	if len(cluster.Status.ClusterPhase) == 0 {
		klog.Infof("Start to create the cluster '%s/%s'", cluster.Namespace, cluster.Name)

		if err = cluster.SetDefaults(); err != nil {
			r.Recorder.Event(cluster, corev1.EventTypeWarning, "SetDefaultValuesFailed", fmt.Sprintf("Set default values failed: %v", err))
			return ctrl.Result{}, err
		}

		// Update the default values to the cluster spec.
		if err = r.Update(ctx, cluster); err != nil {
			r.Recorder.Event(cluster, corev1.EventTypeWarning, "UpdateClusterFailed", fmt.Sprintf("Update cluster failed: %v", err))
			return ctrl.Result{}, err
		}

		if err = r.updateClusterStatus(ctx, cluster, v1alpha1.PhaseStarting); err != nil {
			return ctrl.Result{}, err
		}
	}

	return r.sync(ctx, cluster)
}

func (r *Reconciler) sync(ctx context.Context, cluster *v1alpha1.GreptimeDBCluster) (ctrl.Result, error) {
	for _, d := range r.Deployers {
		err := d.Sync(ctx, cluster, d)
		if errors.Is(err, deployer.ErrSyncNotReady) {
			var (
				currentPhase = cluster.Status.ClusterPhase
				nextPhase    = currentPhase
			)

			// If the cluster is already running, we will set it to updating phase.
			if currentPhase == v1alpha1.PhaseRunning {
				nextPhase = v1alpha1.PhaseUpdating
			}

			// If the cluster is in error phase, we will set it to starting phase.
			if currentPhase == v1alpha1.PhaseError {
				nextPhase = v1alpha1.PhaseStarting
			}

			if err := r.updateClusterStatus(ctx, cluster, nextPhase); err != nil {
				return ctrl.Result{}, err
			}

			return ctrl.Result{}, nil
		}

		if err != nil {
			return ctrl.Result{RequeueAfter: defaultRequeueAfter}, err
		}
	}

	if cluster.Status.ClusterPhase == v1alpha1.PhaseStarting ||
		cluster.Status.ClusterPhase == v1alpha1.PhaseUpdating {
		cluster.Status.SetCondition(*v1alpha1.NewCondition(v1alpha1.ConditionTypeReady, corev1.ConditionTrue, "ClusterReady", "the cluster is ready"))
		if err := r.updateClusterStatus(ctx, cluster, v1alpha1.PhaseRunning); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *Reconciler) addFinalizer(ctx context.Context, cluster *v1alpha1.GreptimeDBCluster) error {
	if cluster.Spec.Meta != nil {
		if !controllerutil.ContainsFinalizer(cluster, greptimedbClusterFinalizer) {
			controllerutil.AddFinalizer(cluster, greptimedbClusterFinalizer)
			if err := r.Update(ctx, cluster); err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *Reconciler) delete(ctx context.Context, cluster *v1alpha1.GreptimeDBCluster) (ctrl.Result, error) {
	if !controllerutil.ContainsFinalizer(cluster, greptimedbClusterFinalizer) {
		klog.V(2).Info("Skipping as it does not have a finalizer")
		return ctrl.Result{}, nil
	}

	if cluster.Status.ClusterPhase != v1alpha1.PhaseTerminating {
		if err := r.updateClusterStatus(ctx, cluster, v1alpha1.PhaseTerminating); err != nil {
			return ctrl.Result{}, err
		}

		// Trigger the next reconcile.
		return ctrl.Result{Requeue: true}, nil
	}

	for _, d := range r.Deployers {
		if err := d.CleanUp(ctx, cluster); err != nil {
			return ctrl.Result{}, err
		}
	}

	// remove our finalizer from the list.
	controllerutil.RemoveFinalizer(cluster, greptimedbClusterFinalizer)

	if err := r.Update(ctx, cluster); err != nil {
		return ctrl.Result{}, err
	}

	klog.Infof("Delete GreptimeDB cluster '%s/%s'", cluster.Namespace, cluster.Name)

	return ctrl.Result{}, nil
}

func (r *Reconciler) updateClusterStatus(ctx context.Context, cluster *v1alpha1.GreptimeDBCluster, phase v1alpha1.Phase) error {
	// If the cluster is already in the phase, we will not update it.
	if cluster.Status.ClusterPhase == phase {
		return nil
	}

	cluster.Status.ClusterPhase = phase
	cluster.Status.Version = cluster.Spec.Version

	r.setObservedGeneration(cluster)
	r.recordNormalEventByPhase(cluster)

	return deployers.UpdateStatus(ctx, cluster, r.Client)
}

func (r *Reconciler) recordNormalEventByPhase(cluster *v1alpha1.GreptimeDBCluster) {
	switch cluster.Status.ClusterPhase {
	case v1alpha1.PhaseStarting:
		r.Recorder.Event(cluster, corev1.EventTypeNormal, "StartingCluster", "Cluster is starting")
	case v1alpha1.PhaseRunning:
		r.Recorder.Event(cluster, corev1.EventTypeNormal, "ClusterIsReady", "Cluster is ready")
	case v1alpha1.PhaseUpdating:
		r.Recorder.Event(cluster, corev1.EventTypeNormal, "UpdatingCluster", "Cluster is updating")
	case v1alpha1.PhaseTerminating:
		r.Recorder.Event(cluster, corev1.EventTypeNormal, "TerminatingCluster", "Cluster is terminating")
	}
}

func (r *Reconciler) setObservedGeneration(cluster *v1alpha1.GreptimeDBCluster) {
	generation := cluster.Generation
	if cluster.Status.ObservedGeneration == nil {
		cluster.Status.ObservedGeneration = &generation
	} else if *cluster.Status.ObservedGeneration != generation {
		cluster.Status.ObservedGeneration = &generation
	}
}
