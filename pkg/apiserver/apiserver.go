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

package apiserver

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	podmetricsv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	greptimev1alpha1 "github.com/GreptimeTeam/greptimedb-operator/apis/v1alpha1"
	"github.com/GreptimeTeam/greptimedb-operator/controllers/common"
	"github.com/GreptimeTeam/greptimedb-operator/controllers/constant"
)

const (
	// APIGroup is the base path for the HTTP API.
	APIGroup = "/api/v1alpha1"
)

// Server provides an HTTP API to interact with Greptime CRD resources.
type Server struct {
	client.Client

	port             int32
	enablePodMetrics bool
}

// Options represents the options for the Server.
type Options struct {
	// Port is the port that the API service will listen on.
	Port int32

	// EnablePodMetrics indicates whether to enable fetching PodMetrics from metrics-server.
	EnablePodMetrics bool
}

// GreptimeDBCluster represents a GreptimeDBCluster resource that is returned by the API.
// This struct is used to serialize the GreptimeDBCluster resource into JSON.
type GreptimeDBCluster struct {
	// Name is the name of the GreptimeDBCluster.
	Name string `json:"name"`

	// Namespace is the namespace of the GreptimeDBCluster.
	Namespace string `json:"namespace"`

	// Spec is the spec of the GreptimeDBCluster.
	Spec *greptimev1alpha1.GreptimeDBClusterSpec `json:"spec,omitempty"`

	// Status is the status of the GreptimeDBCluster.
	Status *greptimev1alpha1.GreptimeDBClusterStatus `json:"status,omitempty"`

	// Topology is the deployment topology of the GreptimeDBCluster.
	Topology *GreptimeDBClusterTopology `json:"topology,omitempty"`
}

// GreptimeDBClusterTopology represents the deployment topology of a GreptimeDBCluster.
type GreptimeDBClusterTopology struct {
	// Meta represents the meta component of the GreptimeDBCluster.
	Meta []*Pod `json:"meta,omitempty"`

	// Datanode represents the datanode component of the GreptimeDBCluster.
	Datanode []*Pod `json:"datanode,omitempty"`

	// Frontend represents the frontend component of the GreptimeDBCluster.
	Frontend []*Pod `json:"frontend,omitempty"`

	// Flownode represents the flownode component of the GreptimeDBCluster.
	Flownode []*Pod `json:"flownode,omitempty"`
}

// Pod is a simplified representation of a Kubernetes Pod.
type Pod struct {
	// Name is the name of the Pod.
	Name string `json:"name"`

	// Namespace is the namespace of the Pod.
	Namespace string `json:"namespace"`

	// IP is the IP address of the Pod.
	IP string `json:"ip"`

	// Node is the name of the node where the Pod is running.
	Node string `json:"node"`

	// Resource is the resources of all containers in the Pod.
	Resources []*Resource `json:"resources,omitempty"`

	// Status is the status of the Pod.
	Status string `json:"status"`

	// StartTime is the time when the Pod started.
	StartTime *metav1.Time `json:"startTime,omitempty"`
}

// Resource represents the resource of a container.
type Resource struct {
	// Name is the name of the container.
	Name string `json:"name"`

	// Request is the resource request of the container.
	Request corev1.ResourceList `json:"request,omitempty"`

	// Limit is the resource limit of the container.
	Limit corev1.ResourceList `json:"limit,omitempty"`

	// Usage is the resource usage of the container.
	Usage corev1.ResourceList `json:"usage,omitempty"`
}

// Response represents a response returned by the API.
type Response struct {
	// Success indicates whether the request was successful.
	Success bool `json:"success"`

	// Code is the status code that defined by the Server.
	Code int `json:"code,omitempty"`

	// Message is additional message returned by the API.
	Message string `json:"message,omitempty"`

	// Data is the data returned by the API.
	Data interface{} `json:"data,omitempty"`
}

// NewServer creates a new Server with the given client and options.
func NewServer(mgr manager.Manager, opts *Options) (*Server, error) {
	// Create a NoCache client to avoid watching.
	client, err := client.New(mgr.GetConfig(), client.Options{Scheme: mgr.GetScheme(), Mapper: mgr.GetRESTMapper()})
	if err != nil {
		return nil, err
	}

	return &Server{
		Client:           client,
		port:             opts.Port,
		enablePodMetrics: opts.EnablePodMetrics,
	}, nil
}

// Run starts the HTTP service and listens on the specified port.
func (s *Server) Run() error {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	api := router.Group(APIGroup)

	// Get all clusters in a specific namespace. If the namespace is 'all', it will return all clusters in all namespaces.
	api.GET("/namespaces/:namespace/clusters", s.getClusters)

	// Get a specific cluster in a specific namespace.
	api.GET("/namespaces/:namespace/clusters/:name", s.getCluster)

	klog.Infof("HTTP service is running on port %d", s.port)

	return router.Run(fmt.Sprintf(":%d", s.port))
}

func (s *Server) getClusters(c *gin.Context) {
	namespace := c.Param("namespace")
	if namespace == "" {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "namespace is required",
		})
		return
	}

	var listOptions []client.ListOption
	if namespace != "all" {
		listOptions = append(listOptions, client.InNamespace(namespace))
	}

	var clusters greptimev1alpha1.GreptimeDBClusterList
	if err := s.List(c, &clusters, listOptions...); err != nil {
		klog.Errorf("failed to list GreptimeDBCluster: %v", err)
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: "failed to get all clusters",
		})
		return
	}

	if len(clusters.Items) == 0 {
		c.JSON(http.StatusNotFound, Response{
			Success: true,
			Message: "no cluster found",
		})
		return
	}

	var data []GreptimeDBCluster
	for _, cluster := range clusters.Items {
		topology, err := s.getTopology(c, cluster)
		if err != nil {
			klog.Errorf("failed to get topology for cluster %s/%s: %v", cluster.Namespace, cluster.Name, err)
			c.JSON(http.StatusInternalServerError, Response{
				Success: false,
				Message: fmt.Sprintf("failed to get topology for cluster '%s' from namespace '%s'", cluster.Name, cluster.Namespace),
			})
			return
		}

		data = append(data, GreptimeDBCluster{
			Name:      cluster.Name,
			Namespace: cluster.Namespace,
			Spec:      &cluster.Spec,
			Status:    &cluster.Status,
			Topology:  topology,
		})
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    data,
	})
}

func (s *Server) getCluster(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	if namespace == "" {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "namespace is required",
		})
		return
	}

	if name == "" {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "name is required",
		})
		return
	}

	var cluster greptimev1alpha1.GreptimeDBCluster
	err := s.Get(c, client.ObjectKey{Namespace: namespace, Name: name}, &cluster)

	if apierrors.IsNotFound(err) {
		c.JSON(http.StatusNotFound, Response{
			Success: true,
			Message: fmt.Sprintf("cluster '%s' is not found in namespace '%s'", name, namespace),
		})
		return
	}

	if err != nil {
		klog.Errorf("failed to get GreptimeDBCluster %s/%s: %v", namespace, name, err)
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: fmt.Sprintf("failed to get cluster '%s' from namespace '%s'", name, namespace),
		})
		return
	}

	topology, err := s.getTopology(c, cluster)
	if err != nil {
		klog.Errorf("failed to get topology for cluster %s/%s: %v", namespace, name, err)
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: fmt.Sprintf("failed to get topology for cluster '%s' from namespace '%s'", name, namespace),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data: GreptimeDBCluster{
			Name:      cluster.Name,
			Namespace: cluster.Namespace,
			Spec:      &cluster.Spec,
			Status:    &cluster.Status,
			Topology:  topology,
		},
	})
}

func (s *Server) getTopology(ctx context.Context, cluster greptimev1alpha1.GreptimeDBCluster) (*GreptimeDBClusterTopology, error) {
	topology := new(GreptimeDBClusterTopology)

	for _, kind := range []greptimev1alpha1.ComponentKind{
		greptimev1alpha1.MetaComponentKind,
		greptimev1alpha1.DatanodeComponentKind,
		greptimev1alpha1.FrontendComponentKind,
		greptimev1alpha1.FlownodeComponentKind,
	} {
		if err := s.getPods(ctx, cluster, kind, topology); err != nil {
			return nil, err
		}
	}

	return topology, nil
}

func (s *Server) getPods(ctx context.Context, cluster greptimev1alpha1.GreptimeDBCluster, kind greptimev1alpha1.ComponentKind, topology *GreptimeDBClusterTopology) error {
	var internalPods corev1.PodList
	if err := s.List(ctx, &internalPods, client.InNamespace(cluster.Namespace),
		client.MatchingLabels{constant.GreptimeDBComponentName: common.ResourceName(cluster.Name, kind)}); err != nil {
		return err
	}

	// List the frontend groups Pod
	var frontendGroupsPod corev1.PodList
	for _, frontend := range cluster.GetFrontendGroups() {
		if err := s.List(ctx, &frontendGroupsPod, client.InNamespace(cluster.Namespace),
			client.MatchingLabels{constant.GreptimeDBComponentName: common.AdditionalResourceName(cluster.Name, frontend.Name, kind)}); err != nil {
			return err
		}
		internalPods.Items = append(internalPods.Items, frontendGroupsPod.Items...)
	}

	var metrics []*podmetricsv1beta1.PodMetrics
	if s.enablePodMetrics {
		for _, pod := range internalPods.Items {
			podMetrics, err := s.getPodMetrics(ctx, cluster.Namespace, pod.Name)
			if err != nil {
				// Logs the error and continue to get the next pod metrics. We don't want the error of getting podmetrics to block the whole API.
				klog.Errorf("failed to get pod metrics for pod %s/%s: %v", cluster.Namespace, pod.Name, err)
			}
			metrics = append(metrics, podMetrics)
		}
	}

	var pods []*Pod
	// zip internalPods and metrics.
	for i, pod := range internalPods.Items {
		if s.enablePodMetrics {
			pods = append(pods, s.buildPod(&pod, metrics[i]))
		} else {
			pods = append(pods, s.buildPod(&pod, nil))
		}
	}

	switch kind {
	case greptimev1alpha1.MetaComponentKind:
		topology.Meta = pods
	case greptimev1alpha1.DatanodeComponentKind:
		topology.Datanode = pods
	case greptimev1alpha1.FrontendComponentKind:
		topology.Frontend = pods
	case greptimev1alpha1.FlownodeComponentKind:
		topology.Flownode = pods
	}

	return nil
}

func (s *Server) buildPod(internalPod *corev1.Pod, podMetrics *podmetricsv1beta1.PodMetrics) *Pod {
	pod := &Pod{
		Name:      internalPod.Name,
		Namespace: internalPod.Namespace,
		IP:        internalPod.Status.PodIP,
		Status:    string(internalPod.Status.Phase),
		Node:      internalPod.Spec.NodeName,
		StartTime: internalPod.Status.StartTime,
	}

	var resources []*Resource

	for _, container := range internalPod.Spec.Containers {
		resources = append(resources, &Resource{
			Name:    container.Name,
			Request: container.Resources.Requests,
			Limit:   container.Resources.Limits,
		})
	}

	if s.enablePodMetrics && podMetrics != nil {
		for _, container := range podMetrics.Containers {
			for _, resource := range resources {
				if resource.Name == container.Name {
					resource.Usage = container.Usage
				}
			}
		}
	}

	pod.Resources = resources

	return pod
}

func (s *Server) getPodMetrics(ctx context.Context, namespace, name string) (*podmetricsv1beta1.PodMetrics, error) {
	var podMetrics podmetricsv1beta1.PodMetrics
	err := s.Get(ctx, client.ObjectKey{Namespace: namespace, Name: name}, &podMetrics)

	// It's possible when the pod just started, the metrics-server hasn't collected the metrics yet.
	if apierrors.IsNotFound(err) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}
	return &podMetrics, nil
}
