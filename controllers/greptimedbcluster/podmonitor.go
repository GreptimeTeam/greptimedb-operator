package greptimedbcluster

import (
	"context"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/GreptimeTeam/greptimedb-operator/apis/v1alpha1"
)

const (
	MetricPortName = "metrics"
	MetricPath     = "/metrics"
	Interval       = "30s"
)

type SyncPodMonitorFunc func(ctx context.Context, cluster *v1alpha1.GreptimeDBCluster) error

func (r *Reconciler) syncPodMonitor(ctx context.Context, cluster *v1alpha1.GreptimeDBCluster) error {
	var actions []SyncPodMonitorFunc
	if cluster.Spec.Meta != nil {
		actions = append(actions, r.syncMetaPodMonitor)
	}
	if cluster.Spec.Datanode != nil {
		actions = append(actions, r.syncDatanodePodMonitor)
	}
	if cluster.Spec.Frontend != nil {
		actions = append(actions, r.syncFrontendPodMonitor)
	}

	for index, action := range actions {
		err := action(ctx, cluster)
		if err != nil {
			return err
		}
		if index == len(actions)-1 {
			klog.Infof("Creating pod monitor successful")
			return nil
		}
	}

	return nil
}

func (r *Reconciler) syncMetaPodMonitor(ctx context.Context, cluster *v1alpha1.GreptimeDBCluster) error {
	namespaceName := types.NamespacedName{
		Namespace: cluster.Namespace,
		Name:      cluster.Name + "-meta",
	}

	pm := &monitoringv1.PodMonitor{}
	err := r.Get(ctx, namespaceName, pm)
	if meta.IsNoMatchError(err) {
		klog.Info("Pod monitor is not match")
		return nil
	}

	if errors.IsNotFound(err) {
		podMonitor := &monitoringv1.PodMonitor{
			ObjectMeta: metav1.ObjectMeta{
				Name:      namespaceName.Name,
				Namespace: namespaceName.Namespace,
			},
		}
		if err := r.buildMetaPodMonitor(cluster, podMonitor); err != nil {
			return err
		}

		klog.Infof("Create meta pod monitor: %s in %s namespace", podMonitor.Name, podMonitor.Namespace)
		return r.Create(ctx, podMonitor)
	}
	if err != nil {
		return err
	}

	return nil
}

func (r *Reconciler) buildMetaPodMonitor(cluster *v1alpha1.GreptimeDBCluster, podMonitor *monitoringv1.PodMonitor) error {
	var metaLabels = map[string]string{
		greptimeDBApplication: cluster.Name + "-meta",
	}

	podMonitor.Spec.PodMetricsEndpoints = []monitoringv1.PodMetricsEndpoint{
		{
			Path:        MetricPath,
			Port:        MetricPortName,
			Interval:    Interval,
			HonorLabels: true,
		},
	}
	podMonitor.Spec.NamespaceSelector = monitoringv1.NamespaceSelector{
		MatchNames: []string{
			cluster.Namespace,
		},
	}
	podMonitor.Spec.Selector.MatchLabels = metaLabels

	if err := controllerutil.SetControllerReference(cluster, podMonitor, r.Scheme); err != nil {
		return err
	}

	return nil
}

func (r *Reconciler) syncDatanodePodMonitor(ctx context.Context, cluster *v1alpha1.GreptimeDBCluster) error {
	namespaceName := types.NamespacedName{
		Namespace: cluster.Namespace,
		Name:      cluster.Name + "-datanode",
	}

	pm := &monitoringv1.PodMonitor{}
	err := r.Get(ctx, namespaceName, pm)
	if meta.IsNoMatchError(err) {
		klog.Info("Pod monitor is not match")
		return nil
	}

	if errors.IsNotFound(err) {
		podMonitor := &monitoringv1.PodMonitor{
			ObjectMeta: metav1.ObjectMeta{
				Name:      namespaceName.Name,
				Namespace: namespaceName.Namespace,
			},
		}
		if err := r.buildDatanodePodMonitor(cluster, podMonitor); err != nil {
			return err
		}

		klog.Infof("Create datanode pod monitor: %s in %s namespace", podMonitor.Name, podMonitor.Namespace)
		return r.Create(ctx, podMonitor)
	}
	if err != nil {
		return err
	}

	return nil
}

func (r *Reconciler) buildDatanodePodMonitor(cluster *v1alpha1.GreptimeDBCluster, podMonitor *monitoringv1.PodMonitor) error {

	var datanodeLabels = map[string]string{
		greptimeDBApplication: cluster.Name + "-datanode",
	}
	podMonitor.Spec.PodMetricsEndpoints = []monitoringv1.PodMetricsEndpoint{
		{
			Path:        MetricPath,
			Port:        MetricPortName,
			Interval:    Interval,
			HonorLabels: true,
		},
	}
	podMonitor.Spec.NamespaceSelector = monitoringv1.NamespaceSelector{
		MatchNames: []string{
			cluster.Namespace,
		},
	}
	podMonitor.Spec.Selector.MatchLabels = datanodeLabels

	if err := controllerutil.SetControllerReference(cluster, podMonitor, r.Scheme); err != nil {
		return err
	}

	return nil
}

func (r *Reconciler) syncFrontendPodMonitor(ctx context.Context, cluster *v1alpha1.GreptimeDBCluster) error {
	namespaceName := types.NamespacedName{
		Namespace: cluster.Namespace,
		Name:      cluster.Name + "-frontend",
	}

	pm := &monitoringv1.PodMonitor{}
	err := r.Get(ctx, namespaceName, pm)
	if meta.IsNoMatchError(err) {
		klog.Info("Pod monitor is not match")
		return nil
	}

	if errors.IsNotFound(err) {
		podMonitor := &monitoringv1.PodMonitor{
			ObjectMeta: metav1.ObjectMeta{
				Name:      namespaceName.Name,
				Namespace: namespaceName.Namespace,
			},
		}
		if err := r.buildFrontendPodMonitor(cluster, podMonitor); err != nil {
			return err
		}

		klog.Infof("Create frontend pod monitor: %s in %s namespace", podMonitor.Name, podMonitor.Namespace)
		return r.Create(ctx, podMonitor)
	}
	if err != nil {
		return err
	}

	return nil
}

func (r *Reconciler) buildFrontendPodMonitor(cluster *v1alpha1.GreptimeDBCluster, podMonitor *monitoringv1.PodMonitor) error {
	var frontendLabels = map[string]string{
		greptimeDBApplication: cluster.Name + "-frontend",
	}
	podMonitor.Spec.PodMetricsEndpoints = []monitoringv1.PodMetricsEndpoint{
		{
			Path:        MetricPath,
			Port:        MetricPortName,
			Interval:    Interval,
			HonorLabels: true,
		},
	}
	podMonitor.Spec.NamespaceSelector = monitoringv1.NamespaceSelector{
		MatchNames: []string{
			cluster.Namespace,
		},
	}
	podMonitor.Spec.Selector.MatchLabels = frontendLabels

	if err := controllerutil.SetControllerReference(cluster, podMonitor, r.Scheme); err != nil {
		return err
	}

	return nil
}
