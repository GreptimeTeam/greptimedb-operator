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

package metrics

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/metrics"

	greptimev1alpha1 "github.com/GreptimeTeam/greptimedb-operator/apis/v1alpha1"
	"github.com/GreptimeTeam/greptimedb-operator/controllers/common"
	"github.com/GreptimeTeam/greptimedb-operator/controllers/constant"
)

const (
	// MetricPrefix is the prefix of all metrics.
	MetricPrefix = "greptimedb_operator"
)

var (
	podStartupDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: metricName("pod_startup_duration_seconds"),
		Help: "The duration from pod startup to all containers are ready.",

		// Exponential buckets from 1s to 10min.
		Buckets: prometheus.ExponentialBucketsRange(1, 600, 12),
	},
		[]string{"namespace", "resource", "pod", "node", "role"},
	)

	podSchedulingDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: metricName("pod_scheduling_duration_seconds"),
		Help: "The duration from pod startup to scheduled.",

		// Exponential buckets from 1s to 10min.
		Buckets: prometheus.ExponentialBucketsRange(1, 600, 12),
	},
		[]string{"namespace", "resource", "pod", "node", "role"},
	)

	podInitializingDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: metricName("pod_initializing_duration_seconds"),
		Help: "The duration from pod scheduled to initialized.",

		// Exponential buckets from 1s to 10min.
		Buckets: prometheus.ExponentialBucketsRange(1, 600, 12),
	},
		[]string{"namespace", "resource", "pod", "node", "role"},
	)

	podContainerStartupDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: metricName("pod_container_startup_duration_seconds"),
		Help: "The duration from pod initialized to container running.",

		// Exponential buckets from 1s to 10min.
		Buckets: prometheus.ExponentialBucketsRange(1, 600, 12),
	},
		[]string{"namespace", "resource", "pod", "container", "node", "role"},
	)

	podContainerReadyDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: metricName("pod_container_ready_duration_seconds"),
		Help: "The duration from pod started to container ready.",

		// Exponential buckets from 1s to 10min.
		Buckets: prometheus.ExponentialBucketsRange(1, 600, 12),
	},
		[]string{"namespace", "resource", "pod", "container", "node", "role"},
	)

	podImagePullingDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: metricName("pod_image_pulling_duration_milliseconds"),
		Help: "The duration for pod image pulling.",

		// Exponential buckets from 1s to 10min.
		Buckets: prometheus.ExponentialBucketsRange(1, 600, 12),
	},
		[]string{"namespace", "resource", "pod", "node", "role", "image"},
	)
)

var (
	ErrEmptyEvents = errors.New("no events found")
)

func init() {
	metrics.Registry.MustRegister(podStartupDuration)
	metrics.Registry.MustRegister(podSchedulingDuration)
	metrics.Registry.MustRegister(podInitializingDuration)
	metrics.Registry.MustRegister(podContainerStartupDuration)
	metrics.Registry.MustRegister(podImagePullingDuration)
}

// MetricsCollector is used to collect pod metrics.
type MetricsCollector struct {
	client client.Client
}

// NewMetricsCollector creates a new MetricsCollector.
func NewMetricsCollector() (*MetricsCollector, error) {
	// Create a new K8s client for metrics collection.
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, err
	}

	c, err := client.New(cfg, client.Options{})
	if err != nil {
		return nil, err
	}

	return &MetricsCollector{client: c}, nil
}

// CollectClusterPodMetrics collects pod metrics for a cluster.
func (c *MetricsCollector) CollectClusterPodMetrics(ctx context.Context, cluster *greptimev1alpha1.GreptimeDBCluster) error {
	if cluster.GetMeta() != nil {
		if err := c.collectPodMetricsByRole(ctx, cluster, greptimev1alpha1.MetaComponentKind); err != nil {
			return err
		}
	}

	if cluster.GetDatanode() != nil {
		if err := c.collectPodMetricsByRole(ctx, cluster, greptimev1alpha1.DatanodeComponentKind); err != nil {
			return err
		}
	}

	if cluster.GetFrontend() != nil {
		if err := c.collectPodMetricsByRole(ctx, cluster, greptimev1alpha1.FrontendComponentKind); err != nil {
			return err
		}
	}

	if cluster.GetFlownode() != nil {
		if err := c.collectPodMetricsByRole(ctx, cluster, greptimev1alpha1.FlownodeComponentKind); err != nil {
			return err
		}
	}

	if cluster.GetFrontendGroup() != nil {
		if err := c.collectPodMetricsByRole(ctx, cluster, greptimev1alpha1.FrontendComponentKind); err != nil {
			return err
		}
	}

	return nil
}

func (c *MetricsCollector) collectPodMetricsByRole(ctx context.Context, cluster *greptimev1alpha1.GreptimeDBCluster, role greptimev1alpha1.ComponentKind) error {
	var pods []corev1.Pod
	if cluster.GetFrontendGroup() == nil {
		var err error
		pods, err = c.getPods(ctx, cluster, role, "")
		if err != nil {
			return err
		}
	} else {
		for _, frontend := range cluster.GetFrontendGroup() {
			frontendPods, err := c.getPods(ctx, cluster, role, frontend.Name)
			if err != nil {
				return err
			}
			pods = append(pods, frontendPods...)
		}
	}

	for _, pod := range pods {
		if err := c.collectPodMetrics(ctx, cluster.Name, &pod, role); err != nil {
			return err
		}
	}

	return nil
}

func (c *MetricsCollector) collectPodMetrics(ctx context.Context, clusterName string, pod *corev1.Pod, role greptimev1alpha1.ComponentKind) error {
	if pod.Status.Conditions == nil {
		return nil
	}

	startTime := pod.GetCreationTimestamp().Time
	scheduledTime, err := c.getPodConditionTime(&pod.Status, corev1.PodScheduled)
	if err != nil {
		return fmt.Errorf("get pod scheduled time: %v", err)
	}

	initializedTime, err := c.getPodConditionTime(&pod.Status, corev1.PodInitialized)
	if err != nil {
		return fmt.Errorf("get pod initialized time: %v", err)
	}

	readyTime, err := c.getPodConditionTime(&pod.Status, corev1.PodReady)
	if err != nil {
		return fmt.Errorf("get pod ready time: %v", err)
	}

	// Collect pod scheduling duration.
	// The calculation is: pod.Status.Conditions[corev1.PodScheduled].LastTransitionTime.Time - pod.GetCreationTimestamp().Time.
	duration := scheduledTime.Sub(startTime)
	podSchedulingDuration.WithLabelValues(
		pod.Namespace, clusterName, pod.Name, pod.Spec.NodeName, string(role),
	).Observe(duration.Seconds())
	klog.Infof("pod '%s/%s' scheduling duration: '%v'", pod.Namespace, pod.Name, duration)

	// Collect pod initializing duration.
	// The calculation is: pod.Status.Conditions[corev1.PodInitialized].LastTransitionTime.Time - pod.Status.Conditions[corev1.PodScheduled].LastTransitionTime.Time.
	duration = initializedTime.Sub(*scheduledTime)
	podInitializingDuration.WithLabelValues(
		pod.Namespace, clusterName, pod.Name, pod.Spec.NodeName, string(role),
	).Observe(duration.Seconds())
	klog.Infof("pod '%s/%s' from scheduled to initialized duration: '%v'", pod.Namespace, pod.Name, duration)

	// Collect container startup and ready duration.
	for _, container := range pod.Status.ContainerStatuses {
		// The calculation is: pod.Status.ContainerStatuses[*].State.Running.StartedAt - pod.Status.Conditions[corev1.PodInitialized].LastTransitionTime.Time.
		startupDuration := container.State.Running.StartedAt.Time.Sub(*initializedTime)
		podContainerStartupDuration.WithLabelValues(
			pod.Namespace, clusterName, pod.Name, container.Name, pod.Spec.NodeName, string(role),
		).Observe(startupDuration.Seconds())

		// The calculation is: pod.Status.Conditions[corev1.PodReady].LastTransitionTime.Time - pod.Status.ContainerStatuses[*].State.Running.StartedAt.
		readyDuration := readyTime.Sub(container.State.Running.StartedAt.Time)
		podContainerReadyDuration.WithLabelValues(
			pod.Namespace, clusterName, pod.Name, container.Name, pod.Spec.NodeName, string(role),
		).Observe(readyDuration.Seconds())

		klog.Infof("pod '%s/%s' container '%s' from initialized to running duration: '%v', from running to ready duration: '%v'", pod.Namespace, pod.Name, container.Name, startupDuration, readyDuration)
	}

	// Collect pod startup duration.
	// The calculation is: pod.Status.Conditions[corev1.PodReady].LastTransitionTime.Time - pod.GetCreationTimestamp().Time.
	duration = readyTime.Sub(startTime)
	podStartupDuration.WithLabelValues(
		pod.Namespace, clusterName, pod.Name, pod.Spec.NodeName, string(role),
	).Observe(duration.Seconds())
	klog.Infof("pod '%s/%s' from created to ready duration: '%v'", pod.Namespace, pod.Name, duration)

	if err := c.collectPodImagePullingDuration(ctx, clusterName, pod, role); err != nil {
		return err
	}

	return nil
}

func (c *MetricsCollector) getPodConditionTime(podStatus *corev1.PodStatus, conditionType corev1.PodConditionType) (*time.Time, error) {
	for _, condition := range podStatus.Conditions {
		if condition.Type == conditionType {
			return &condition.LastTransitionTime.Time, nil
		}
	}

	return nil, fmt.Errorf("condition %s not found", conditionType)
}

func (c *MetricsCollector) getPods(ctx context.Context, cluster *greptimev1alpha1.GreptimeDBCluster, componentKind greptimev1alpha1.ComponentKind, frontendName string) ([]corev1.Pod, error) {
	resourceName := common.ResourceName(cluster.Name, componentKind)
	if len(frontendName) != 0 && componentKind == greptimev1alpha1.FrontendComponentKind {
		resourceName = common.FrontendGroupResourceName(cluster.Name, componentKind, frontendName)
	}
	selector := metav1.LabelSelector{
		MatchLabels: map[string]string{
			constant.GreptimeDBComponentName: resourceName,
		},
	}

	pods := &corev1.PodList{}
	if err := c.client.List(ctx, pods,
		client.InNamespace(cluster.Namespace),
		client.MatchingLabels(selector.MatchLabels),
		client.MatchingFields{"status.phase": string(corev1.PodRunning)},
	); err != nil {
		return nil, err
	}

	return pods.Items, nil
}

func (c *MetricsCollector) collectPodImagePullingDuration(ctx context.Context, clusterName string, pod *corev1.Pod, role greptimev1alpha1.ComponentKind) error {
	imageName, duration, err := c.getImagePullingDurationFromEvents(ctx, pod)
	if errors.Is(err, ErrEmptyEvents) { // Maybe the events are deleted by the garbage collector.
		return nil
	}

	if imageName == "" || duration == 0 {
		return nil
	}

	if err != nil {
		return fmt.Errorf("get image pulling duration: %v", err)
	}

	klog.Infof("pod '%s/%s' image pulling '%s' duration: '%v'", pod.Namespace, pod.Name, imageName, duration)
	podImagePullingDuration.WithLabelValues(
		pod.Namespace, clusterName, pod.Name, pod.Spec.NodeName, string(role), imageName,
	).Observe(float64(duration.Milliseconds()))

	return nil
}

func (c *MetricsCollector) getImagePullingDurationFromEvents(ctx context.Context, pod *corev1.Pod) (string, time.Duration, error) {
	const (
		prompt = "Successfully pulled image"
	)

	events := &corev1.EventList{}
	if err := c.client.List(ctx, events,
		client.InNamespace(pod.Namespace), client.MatchingFields{"involvedObject.uid": string(pod.UID)},
	); err != nil && !k8serrors.IsNotFound(err) {
		return "", 0, err
	}

	if len(events.Items) == 0 {
		return "", 0, ErrEmptyEvents
	}

	for _, event := range events.Items {
		if strings.Contains(event.Message, prompt) {
			imageName, err := parseImageName(event.Message)
			if err != nil {
				return "", 0, err
			}

			duration, err := parseImagePullingDuration(event.Message)
			if err != nil {
				return "", 0, err
			}

			return imageName, duration, nil
		}
	}

	return "", 0, nil
}

func parseImagePullingDuration(message string) (time.Duration, error) {
	// Parse the message to get the image pulling duration.
	// The message format is like: 'Successfully pulled image "greptime/greptimedb:latest" in 314.950966ms (6.159733714s including waiting)'.
	// The regex is to capture the duration that generated by `time.Since()`.
	re := regexp.MustCompile(`in (\d+(?:\.\d+)?(?:h|ms|µs|ns|m|s))`)
	matches := re.FindStringSubmatch(message)
	if len(matches) == 0 {
		return 0, fmt.Errorf("parse image pulling duration from message: %s", message)
	}

	duration, err := time.ParseDuration(matches[1])
	if err != nil {
		return 0, fmt.Errorf("parse duration %q: %v", matches[1], err)
	}

	return duration, nil
}

func parseImageName(message string) (string, error) {
	re := regexp.MustCompile(`"([^"]+)"`)
	matches := re.FindStringSubmatch(message)
	if len(matches) == 0 {
		return "", fmt.Errorf("failed to parse image name from message: %s", message)
	}

	return matches[1], nil
}

func metricName(name string) string {
	return fmt.Sprintf("%s_%s", MetricPrefix, name)
}
