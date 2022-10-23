package greptimedbcluster

import (
	"context"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
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
	greptimedbClusterFinalizer = "greptimedbcluster.greptime.io/finalizer"
)

var (
	defaultRequeueAfter = 5 * time.Second
)

// Reconciler reconciles a GreptimeDBCluster object
type Reconciler struct {
	client.Client

	Scheme   *runtime.Scheme
	Recorder record.EventRecorder

	deployers map[v1alpha1.ComponentKind]deployer.Deployer
}

func Setup(mgr ctrl.Manager, _ *options.Options) error {
	reconciler := &Reconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Recorder: mgr.GetEventRecorderFor("greptimedbcluster-controller"),
	}

	reconciler.deployers = map[v1alpha1.ComponentKind]deployer.Deployer{
		v1alpha1.FrontendComponentKind: deployers.NewFrontendDeployer(mgr),
		v1alpha1.DatanodeComponentKind: deployers.NewDatanodeDeployer(mgr),
		v1alpha1.MetaComponentKind:     deployers.NewMetaDeployer(mgr),
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
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;create;
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;delete;
// +kubebuilder:rbac:groups=core,resources=persistentvolumeclaims,verbs=get;list;watch;create;update;delete;
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=monitoring.coreos.com,resources=podmonitors,verbs=get;list;watch;create;update;patch;delete

// Reconcile is reconciliation loop for GreptimeDBCluster.
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	klog.V(2).Infof("Reconciling GreptimeDBCluster: %s", req.NamespacedName)

	cluster := new(v1alpha1.GreptimeDBCluster)
	if err := r.Get(ctx, req.NamespacedName, cluster); err != nil {
		if errors.IsNotFound(err) {
			klog.V(2).Infof("Resource not found: %s", req.NamespacedName)
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// The object is being deleted.
	if !cluster.ObjectMeta.DeletionTimestamp.IsZero() {
		return r.delete(ctx, cluster)
	}

	if err := r.addFinalizer(ctx, cluster); err != nil {
		return ctrl.Result{}, err
	}

	// TODO(zyy17): Add validation for cluster.

	if err := cluster.SetDefaults(); err != nil {
		return ctrl.Result{}, err
	}

	return r.sync(ctx, cluster)
}

// sync executes the sync logic of multiple components in order.
func (r *Reconciler) sync(ctx context.Context, cluster *v1alpha1.GreptimeDBCluster) (ctrl.Result, error) {
	if cluster.Spec.Meta != nil {
		componentDeployer := r.deployers[v1alpha1.MetaComponentKind]

		err := componentDeployer.Sync(ctx, cluster, componentDeployer)
		if err == deployer.ErrSyncNotReady {
			return ctrl.Result{}, nil
		}
		if err != nil {
			return ctrl.Result{RequeueAfter: defaultRequeueAfter}, err
		}
	}

	if cluster.Spec.EnablePrometheusMonitor {
		if err := r.syncPodMonitor(ctx, cluster); err != nil {
			klog.Infof("Sync pod monitor error: %v", err)
			return ctrl.Result{}, err
		}
	}

	if cluster.Spec.Datanode != nil {
		componentDeployer := r.deployers[v1alpha1.DatanodeComponentKind]

		err := componentDeployer.Sync(ctx, cluster, componentDeployer)
		if err == deployer.ErrSyncNotReady {
			return ctrl.Result{}, nil
		}
		if err != nil {
			return ctrl.Result{RequeueAfter: defaultRequeueAfter}, err
		}
	}

	if cluster.Spec.Frontend != nil {
		componentDeployer := r.deployers[v1alpha1.FrontendComponentKind]

		err := componentDeployer.Sync(ctx, cluster, componentDeployer)
		if err == deployer.ErrSyncNotReady {
			return ctrl.Result{}, nil
		}
		if err != nil {
			return ctrl.Result{RequeueAfter: defaultRequeueAfter}, err
		}
	}

	if cluster.Status.GetCondition(v1alpha1.GreptimeDBClusterReady) == nil {
		klog.Infof("The GreptimeDB cluster: '%s/%s' is ready", cluster.Namespace, cluster.Name)
		cluster.Status.SetCondtion(*v1alpha1.NewCondition(v1alpha1.GreptimeDBClusterReady, corev1.ConditionTrue, "", "the cluster is ready"))
		if err := r.updateStatus(ctx, cluster); err != nil {
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

	for _, d := range r.deployers {
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

func (r *Reconciler) updateStatus(ctx context.Context, cluster *v1alpha1.GreptimeDBCluster, opts ...client.UpdateOption) error {
	status := cluster.DeepCopy().Status
	return retry.RetryOnConflict(retry.DefaultBackoff, func() (err error) {
		objectKey := client.ObjectKey{Namespace: cluster.Namespace, Name: cluster.Name}
		if err = r.Client.Get(ctx, objectKey, cluster); err != nil {
			return
		}
		cluster.Status = status
		return r.Status().Update(ctx, cluster, opts...)
	})
}
