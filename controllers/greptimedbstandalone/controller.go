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
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/pelletier/go-toml"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
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
// +kubebuilder:rbac:groups=greptime.io,resources=greptimedbstandalones/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=greptime.io,resources=greptimedbstandalones/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;create;
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;delete;
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;
// +kubebuilder:rbac:groups=core,resources=events,verbs=get;list;watch;create;patch;
// +kubebuilder:rbac:groups=core,resources=persistentvolumeclaims,verbs=get;list;watch;create;update;delete;
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=monitoring.coreos.com,resources=podmonitors,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apiextensions.k8s.io,resources=customresourcedefinitions,verbs=get;list;watch

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, err error) {
	klog.V(2).Infof("Reconciling GreptimeDBStandalone: %s", req.NamespacedName)

	standalone := new(v1alpha1.GreptimeDBStandalone)
	if err := r.Get(ctx, req.NamespacedName, standalone); err != nil {
		if k8serrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	defer func() {
		if err != nil {
			r.Recorder.Eventf(standalone, corev1.EventTypeWarning, "ReconcileError", fmt.Sprintf("Reconcile error: %v", err))
			if err := r.setStandaloneStatus(ctx, standalone, v1alpha1.PhaseError); err != nil && !k8serrors.IsNotFound(err) {
				klog.Errorf("Failed to update status: %v", err)
				return
			}
		}
	}()

	// The object is being deleted.
	if !standalone.ObjectMeta.DeletionTimestamp.IsZero() {
		return r.delete(ctx, standalone)
	}

	if err = r.addFinalizer(ctx, standalone); err != nil {
		r.Recorder.Event(standalone, corev1.EventTypeWarning, "AddFinalizerFailed", fmt.Sprintf("Add finalizer failed: %v", err))
		return
	}

	if err = r.validate(ctx, standalone); err != nil {
		r.Recorder.Event(standalone, corev1.EventTypeWarning, "InvalidStandalone", fmt.Sprintf("Invalid standalone: %v", err))
		return
	}

	if err = standalone.SetDefaults(); err != nil {
		r.Recorder.Event(standalone, corev1.EventTypeWarning, "SetDefaultValuesFailed", fmt.Sprintf("Set default values failed: %v", err))
		return
	}

	// Means the standalone is just created.
	if len(standalone.Status.StandalonePhase) == 0 {
		klog.Infof("Start to create the standalone '%s/%s'", standalone.Namespace, standalone.Name)
		if err = r.setStandaloneStatus(ctx, standalone, v1alpha1.PhaseStarting); err != nil {
			return
		}
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
	d := r.Deployer
	err := d.Sync(ctx, standalone, d)

	if errors.Is(err, deployer.ErrSyncNotReady) {
		var (
			currentPhase = standalone.Status.StandalonePhase
			nextPhase    = currentPhase
		)

		// If the standalone is already running, we will set it to updating phase.
		if currentPhase == v1alpha1.PhaseRunning {
			nextPhase = v1alpha1.PhaseUpdating
		}

		// If the standalone is in error phase, we will set it to starting phase.
		if currentPhase == v1alpha1.PhaseError {
			nextPhase = v1alpha1.PhaseStarting
		}

		if err := r.setStandaloneStatus(ctx, standalone, nextPhase); err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	if err != nil {
		return ctrl.Result{RequeueAfter: defaultRequeueAfter}, err
	}

	if standalone.Status.StandalonePhase == v1alpha1.PhaseStarting ||
		standalone.Status.StandalonePhase == v1alpha1.PhaseUpdating {
		standalone.Status.SetCondition(*v1alpha1.NewCondition(v1alpha1.ConditionTypeReady, corev1.ConditionTrue, "StandaloneReady", "the standalone is ready"))
		if err := r.setStandaloneStatus(ctx, standalone, v1alpha1.PhaseRunning); err != nil {
			return ctrl.Result{}, err
		}
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

	klog.Infof("Delete GreptimeDB standalone '%s/%s'", standalone.Namespace, standalone.Name)

	return ctrl.Result{}, nil
}

func (r *Reconciler) setStandaloneStatus(ctx context.Context, standalone *v1alpha1.GreptimeDBStandalone, phase v1alpha1.Phase) error {
	// If the standalone is already in the phase, we will not update it.
	if standalone.Status.StandalonePhase == phase {
		return nil
	}

	standalone.Status.StandalonePhase = phase
	standalone.Status.Version = standalone.Spec.Version

	r.setObservedGeneration(standalone)
	r.recordNormalEventByPhase(standalone)

	return UpdateStatus(ctx, standalone, r.Client)
}

func (r *Reconciler) validate(ctx context.Context, standalone *v1alpha1.GreptimeDBStandalone) error {
	if standalone.Spec.Base == nil {
		return fmt.Errorf("no components spec in standalone")
	}

	if standalone.Spec.TLS != nil {
		if len(standalone.Spec.TLS.SecretName) > 0 {
			tlsSecret := &corev1.Secret{}
			err := r.Get(ctx, types.NamespacedName{Namespace: standalone.Namespace, Name: standalone.Spec.TLS.SecretName}, tlsSecret)
			if err != nil {
				return fmt.Errorf("get tls secret '%s' failed, error: '%v'", standalone.Spec.TLS.SecretName, err)
			}

			if _, ok := tlsSecret.Data[deployers.CASecretKey]; !ok {
				return fmt.Errorf("tls secret '%s' does not contain key '%s'", standalone.Spec.TLS.SecretName, deployers.CASecretKey)
			}

			if _, ok := tlsSecret.Data[deployers.TLSCrtSecretKey]; !ok {
				return fmt.Errorf("tls secret '%s' does not contain key '%s'", standalone.Spec.TLS.SecretName, deployers.TLSCrtSecretKey)
			}

			if _, ok := tlsSecret.Data[deployers.TLSKeySecretKey]; !ok {
				return fmt.Errorf("tls secret '%s' does not contain key '%s'", standalone.Spec.TLS.SecretName, deployers.TLSKeySecretKey)
			}
		}
	}

	// To detect if the CRD of podmonitor is installed.
	if standalone.Spec.PrometheusMonitor != nil {
		if standalone.Spec.PrometheusMonitor.Enabled {
			// CheckPodMonitorCRDInstall is used to check if the CRD of podmonitor is installed, it is not used to create the podmonitor.
			err := r.checkPodMonitorCRDInstall(ctx, metav1.GroupKind{
				Group: "monitoring.coreos.com",
				Kind:  "PodMonitor",
			})
			if err != nil {
				if k8serrors.IsNotFound(err) {
					return fmt.Errorf("the crd podmonitors.monitoring.coreos.com is not installed")
				} else {
					return fmt.Errorf("check crd of podmonitors.monitoring.coreos.com is installed error: %v", err)
				}
			}
		}
	}

	if len(standalone.Spec.Config) > 0 {
		if err := r.validateTomlConfig(standalone.Spec.Config); err != nil {
			return fmt.Errorf("invalid meta toml config: %v", err)
		}
	}

	if standalone.Spec.ObjectStorageProvider != nil {
		if standalone.Spec.ObjectStorageProvider.S3 != nil && standalone.Spec.ObjectStorageProvider.OSS != nil {
			return fmt.Errorf("only one object storage provider can be specified")
		}
	}

	return nil
}

func (r *Reconciler) validateTomlConfig(input string) error {
	if len(input) > 0 {
		data := make(map[string]interface{})
		err := toml.Unmarshal([]byte(input), &data)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Reconciler) checkPodMonitorCRDInstall(ctx context.Context, groupKind metav1.GroupKind) error {
	var crd apiextensionsv1.CustomResourceDefinition
	nameNamespace := types.NamespacedName{Name: fmt.Sprintf("%ss.%s", strings.ToLower(groupKind.Kind), groupKind.Group)}
	err := r.Get(ctx, nameNamespace, &crd)
	if err != nil {
		return err
	}
	return nil
}

func (r *Reconciler) recordNormalEventByPhase(standalone *v1alpha1.GreptimeDBStandalone) {
	switch standalone.Status.StandalonePhase {
	case v1alpha1.PhaseStarting:
		r.Recorder.Event(standalone, corev1.EventTypeNormal, "StartingStandalone", "Standalone is starting")
	case v1alpha1.PhaseRunning:
		r.Recorder.Event(standalone, corev1.EventTypeNormal, "StandaloneIsReady", "Standalone is ready")
	case v1alpha1.PhaseUpdating:
		r.Recorder.Event(standalone, corev1.EventTypeNormal, "UpdatingStandalone", "Standalone is updating")
	case v1alpha1.PhaseTerminating:
		r.Recorder.Event(standalone, corev1.EventTypeNormal, "Terminatingstandalone", "Standalone is terminating")
	}
}

func (r *Reconciler) setObservedGeneration(standalone *v1alpha1.GreptimeDBStandalone) {
	generation := standalone.Generation
	if standalone.Status.ObservedGeneration == nil {
		standalone.Status.ObservedGeneration = &generation
	} else {
		if *standalone.Status.ObservedGeneration != generation {
			standalone.Status.ObservedGeneration = &generation
		}
	}
}

func UpdateStatus(ctx context.Context, input *v1alpha1.GreptimeDBStandalone, kc client.Client, opts ...client.UpdateOption) error {
	standalone := input.DeepCopy()
	status := standalone.Status
	return retry.RetryOnConflict(retry.DefaultBackoff, func() (err error) {
		objectKey := client.ObjectKey{Namespace: standalone.Namespace, Name: standalone.Name}
		if err = kc.Get(ctx, objectKey, standalone); err != nil {
			return
		}
		standalone.Status = status
		return kc.Status().Update(ctx, standalone, opts...)
	})
}
