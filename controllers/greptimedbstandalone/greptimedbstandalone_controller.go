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

package greptimedbstandalone

import (
	"context"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/GreptimeTeam/greptimedb-operator/apis/v1alpha1"
	"github.com/GreptimeTeam/greptimedb-operator/cmd/operator/app/options"
	"github.com/GreptimeTeam/greptimedb-operator/pkg/deployer"
)

const (
	greptimedbStandaloneFinalizer = "greptimedbstandalone.greptime.io/finalizer"
)

var (
	defaultRequeueAfter = 5 * time.Second
)

// Reconciler reconciles a GreptimeDBStandalone object
type Reconciler struct {
	client.Client

	Scheme   *runtime.Scheme
	Deployer deployer.Deployer
	Recorder record.EventRecorder
}

func Setup(mgr ctrl.Manager, _ *options.Options) error {
	reconciler := &Reconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Recorder: mgr.GetEventRecorderFor("greptimedbstandalone-controller"),
		Deployer: NewStandaloneDeployer(mgr),
	}
	return reconciler.SetupWithManager(mgr)
}

// +kubebuilder:rbac:groups=greptime.io,resources=greptimedbstandalones,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=greptime.ioresources=greptimedbstandalones/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=greptime.io,resources=greptimedbstandalones/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;create;
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;delete;
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;create;
// +kubebuilder:rbac:groups=core,resources=events,verbs=get;list;watch;create;patch;
// +kubebuilder:rbac:groups=core,resources=persistentvolumeclaims,verbs=get;list;watch;create;update;delete;
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=monitoring.coreos.com,resources=podmonitors,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apiextensions.k8s.io,resources=customresourcedefinitions,verbs=get;list;watch

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, err error) {
	klog.Infof("Reconciling GreptimeDBStandalone: %s", req.NamespacedName)

	standalone := new(v1alpha1.GreptimeDBStandalone)
	if err := r.Get(ctx, req.NamespacedName, standalone); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// The object is being deleted.
	if !standalone.ObjectMeta.DeletionTimestamp.IsZero() {
		return r.delete(ctx, standalone)
	}

	if err = r.addFinalizer(ctx, standalone); err != nil {
		r.Recorder.Eventf(standalone, corev1.EventTypeWarning, "AddFinalizerFailed", fmt.Sprintf("Add finalizer failed: %v", err))
		return
	}

	if err = standalone.SetDefaults(); err != nil {
		r.Recorder.Eventf(standalone, corev1.EventTypeWarning, "SetDefaultValuesFailed", fmt.Sprintf("Set default values failed: %v", err))
		return
	}

	return r.sync(ctx, standalone)
}

// SetupWithManager sets up the controller with the Manager.
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.GreptimeDBStandalone{}).
		Owns(&corev1.Service{}).
		Owns(&appsv1.StatefulSet{}).
		Complete(r)
}

func (r *Reconciler) sync(ctx context.Context, standalone *v1alpha1.GreptimeDBStandalone) (ctrl.Result, error) {
	err := r.Deployer.Sync(ctx, standalone, r.Deployer)
	if err != nil {
		return ctrl.Result{RequeueAfter: defaultRequeueAfter}, err
	}
	return ctrl.Result{}, nil
}

func (r *Reconciler) addFinalizer(ctx context.Context, standalone *v1alpha1.GreptimeDBStandalone) error {
	if !controllerutil.ContainsFinalizer(standalone, greptimedbStandaloneFinalizer) {
		controllerutil.AddFinalizer(standalone, greptimedbStandaloneFinalizer)
		if err := r.Update(ctx, standalone); err != nil {
			return err
		}
	}

	return nil
}

func (r *Reconciler) delete(ctx context.Context, standalone *v1alpha1.GreptimeDBStandalone) (ctrl.Result, error) {
	if !controllerutil.ContainsFinalizer(standalone, greptimedbStandaloneFinalizer) {
		klog.V(2).Info("Skipping as it does not have a finalizer")
		return ctrl.Result{}, nil
	}

	if err := r.Deployer.CleanUp(ctx, standalone); err != nil {
		return ctrl.Result{}, err
	}

	// remove our finalizer from the list.
	controllerutil.RemoveFinalizer(standalone, greptimedbStandaloneFinalizer)

	if err := r.Update(ctx, standalone); err != nil {
		return ctrl.Result{}, err
	}

	klog.Infof("Delete GreptimeDB cluster '%s/%s'", standalone.Namespace, standalone.Name)

	return ctrl.Result{}, nil
}
