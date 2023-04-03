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
	"fmt"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
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

	Scheme   *runtime.Scheme
	Recorder record.EventRecorder

	Deployers []deployer.Deployer
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
// +kubebuilder:rbac:groups=apiextensions.k8s.io,resources=customresourcedefinitions,verbs=get;list;watch

// Reconcile is reconciliation loop for GreptimeDBCluster.
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, err error) {
	klog.V(2).Infof("Reconciling GreptimeDBCluster: %s", req.NamespacedName)

	cluster := new(v1alpha1.GreptimeDBCluster)
	if err := r.Get(ctx, req.NamespacedName, cluster); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	defer func() {
		if err != nil {
			if err := r.setClusterPhase(ctx, cluster, v1alpha1.ClusterError); err != nil {
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
		return
	}

	if err = r.validate(ctx, cluster); err != nil {
		return
	}

	if err = cluster.SetDefaults(); err != nil {
		return
	}

	if len(cluster.Status.ClusterPhase) == 0 {
		klog.Infof("Start to create the cluster '%s/%s'", cluster.Namespace, cluster.Name)
		if err = r.setClusterPhase(ctx, cluster, v1alpha1.ClusterStarting); err != nil {
			return
		}
	}

	return r.sync(ctx, cluster)
}

func (r *Reconciler) sync(ctx context.Context, cluster *v1alpha1.GreptimeDBCluster) (ctrl.Result, error) {
	for _, d := range r.Deployers {
		err := d.Sync(ctx, cluster, d)
		if err == deployer.ErrSyncNotReady {
			if cluster.Status.ClusterPhase != v1alpha1.ClusterStarting {
				if err := r.setClusterPhase(ctx, cluster, v1alpha1.ClusterStarting); err != nil {
					return ctrl.Result{}, err
				}
			}
			return ctrl.Result{}, nil
		}
		if err != nil {
			return ctrl.Result{RequeueAfter: defaultRequeueAfter}, err
		}
	}

	if cluster.Status.ClusterPhase != v1alpha1.ClusterRunning {
		// FIXME(zyy17): The logging maybe duplicated because the status update will trigger another reconcile.
		klog.Infof("The GreptimeDB cluster '%s/%s' is ready", cluster.Namespace, cluster.Name)
		cluster.Status.SetCondition(*v1alpha1.NewCondition(v1alpha1.GreptimeDBClusterReady, corev1.ConditionTrue, "ClusterReady", "the cluster is ready"))
		cluster.Status.ClusterPhase = v1alpha1.ClusterRunning
		if err := deployers.UpdateStatus(ctx, cluster, r.Client); err != nil {
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

func (r *Reconciler) setClusterPhase(ctx context.Context, cluster *v1alpha1.GreptimeDBCluster, phase v1alpha1.ClusterPhase) error {
	cluster.Status.ClusterPhase = phase
	return deployers.UpdateStatus(ctx, cluster, r.Client)
}

func (r *Reconciler) validate(ctx context.Context, cluster *v1alpha1.GreptimeDBCluster) error {
	if cluster.Spec.Meta == nil && cluster.Spec.Datanode == nil && cluster.Spec.Frontend == nil {
		return fmt.Errorf("no components spec in cluster")
	}

	if cluster.Spec.Frontend != nil && cluster.Spec.Frontend.TLS != nil {
		if len(cluster.Spec.Frontend.TLS.SecretName) > 0 {
			tlsSecret := &corev1.Secret{}
			err := r.Get(ctx, types.NamespacedName{Namespace: cluster.Namespace, Name: cluster.Spec.Frontend.TLS.SecretName}, tlsSecret)
			if err != nil {
				return fmt.Errorf("get tls secret '%s' failed, error: '%v'", cluster.Spec.Frontend.TLS.SecretName, err)
			}

			if _, ok := tlsSecret.Data[deployers.CASecretKey]; !ok {
				return fmt.Errorf("tls secret '%s' does not contain key '%s'", cluster.Spec.Frontend.TLS.SecretName, deployers.CASecretKey)
			}

			if _, ok := tlsSecret.Data[deployers.TLSCrtSecretKey]; !ok {
				return fmt.Errorf("tls secret '%s' does not contain key '%s'", cluster.Spec.Frontend.TLS.SecretName, deployers.TLSCrtSecretKey)
			}

			if _, ok := tlsSecret.Data[deployers.TLSKeySecretKey]; !ok {
				return fmt.Errorf("tls secret '%s' does not contain key '%s'", cluster.Spec.Frontend.TLS.SecretName, deployers.TLSKeySecretKey)
			}
		}
	}

	// To detect if the CRD of podmonitor is installed.
	if cluster.Spec.PrometheusMonitor != nil {
		if cluster.Spec.PrometheusMonitor.Enabled {
			// CheckPodMonitorCRDInstall is used to check if the CRD of podmonitor is installed, it is not used to create the podmonitor.
			err := r.checkPodMonitorCRDInstall(ctx, metav1.GroupKind{
				Group: "monitoring.coreos.com",
				Kind:  "PodMonitor",
			})
			if err != nil {
				if errors.IsNotFound(err) {
					return fmt.Errorf("the crd podmonitors.monitoring.coreos.com is not installed")
				} else {
					return fmt.Errorf("check crd of podmonitors.monitoring.coreos.com is installed error: %v", err)
				}
			}
		}
	}

	if cluster.Spec.Meta != nil {
		if err := r.validateTomlConfig(cluster.Spec.Meta.Config); err != nil {
			return fmt.Errorf("invalid meta toml config: %v", err)
		}
	}

	if cluster.Spec.Datanode != nil {
		if err := r.validateTomlConfig(cluster.Spec.Datanode.Config); err != nil {
			return fmt.Errorf("invalid datanode toml config: %v", err)
		}
	}

	if cluster.Spec.Frontend != nil {
		if err := r.validateTomlConfig(cluster.Spec.Frontend.Config); err != nil {
			return fmt.Errorf("invalid frontend toml config: %v", err)
		}
	}

	if cluster.Spec.StorageProvider != nil {
		if cluster.Spec.StorageProvider.Local != nil && cluster.Spec.StorageProvider.S3 != nil {
			return fmt.Errorf("only one storage provider can be specified")
		}

		if cluster.Spec.StorageProvider.S3 != nil {
			storageCredentialsSecret := &corev1.Secret{}
			err := r.Get(ctx, types.NamespacedName{Namespace: cluster.Namespace, Name: cluster.Spec.StorageProvider.S3.SecretName}, storageCredentialsSecret)
			if err != nil {
				return fmt.Errorf("get storage credentials secret '%s' failed, error: '%v'", cluster.Spec.StorageProvider.S3.SecretName, err)
			}

			if _, ok := storageCredentialsSecret.Data[deployers.AccessKeyIDSecretKey]; !ok {
				return fmt.Errorf("credentials secret '%s' does not contain access key id '%s'", cluster.Spec.StorageProvider.S3.SecretName, deployers.AccessKeyIDSecretKey)
			}

			if _, ok := storageCredentialsSecret.Data[deployers.SecretAccessKeySecretKey]; !ok {
				return fmt.Errorf("credentials secret '%s' does not contain secret access key '%s'", cluster.Spec.StorageProvider.S3.SecretName, deployers.SecretAccessKeySecretKey)
			}
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
