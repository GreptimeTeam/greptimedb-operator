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
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

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

	port int
}

// Options represents the options for the Server.
type Options struct {
	// Port is the port that the HTTP service will listen on.
	Port int
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

	// Resource is the resource requirements of the Pod.
	Resource corev1.ResourceRequirements `json:"resource,omitempty"`

	// Status is the status of the Pod.
	Status string `json:"status"`

	// StartTime is the time when the Pod started.
	StartTime *metav1.Time `json:"startTime,omitempty"`
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
func NewServer(client client.Client, opts *Options) *Server {
	return &Server{
		Client: client,
		port:   opts.Port,
	}
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
		topology, err := s.getTopology(c, cluster.Namespace, cluster.Name)
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

	if errors.IsNotFound(err) {
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

	topology, err := s.getTopology(c, namespace, name)
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

func (s *Server) getTopology(ctx context.Context, namespace, name string) (*GreptimeDBClusterTopology, error) {
	topology := new(GreptimeDBClusterTopology)

	for _, kind := range []greptimev1alpha1.ComponentKind{
		greptimev1alpha1.MetaComponentKind,
		greptimev1alpha1.DatanodeComponentKind,
		greptimev1alpha1.FrontendComponentKind,
		greptimev1alpha1.FlownodeComponentKind,
	} {
		if err := s.getPods(ctx, namespace, name, kind, topology); err != nil {
			return nil, err
		}
	}

	return topology, nil
}

func (s *Server) getPods(ctx context.Context, namespace, name string, kind greptimev1alpha1.ComponentKind, topology *GreptimeDBClusterTopology) error {
	var pods corev1.PodList
	if err := s.List(ctx, &pods, client.InNamespace(namespace),
		client.MatchingLabels{constant.GreptimeDBComponentName: common.ResourceName(name, kind)}); err != nil {
		return err
	}

	switch kind {
	case greptimev1alpha1.MetaComponentKind:
		for _, pod := range pods.Items {
			topology.Meta = append(topology.Meta, s.covertToPod(&pod))
		}
	case greptimev1alpha1.DatanodeComponentKind:
		for _, pod := range pods.Items {
			topology.Datanode = append(topology.Datanode, s.covertToPod(&pod))
		}
	case greptimev1alpha1.FrontendComponentKind:
		for _, pod := range pods.Items {
			topology.Frontend = append(topology.Frontend, s.covertToPod(&pod))
		}
	case greptimev1alpha1.FlownodeComponentKind:
		for _, pod := range pods.Items {
			topology.Flownode = append(topology.Flownode, s.covertToPod(&pod))
		}
	}

	return nil
}

func (s *Server) covertToPod(pod *corev1.Pod) *Pod {
	return &Pod{
		Name:      pod.Name,
		Namespace: pod.Namespace,
		IP:        pod.Status.PodIP,
		Status:    string(pod.Status.Phase),
		Node:      pod.Spec.NodeName,
		Resource:  pod.Spec.Containers[constant.MainContainerIndex].Resources,
		StartTime: pod.Status.StartTime,
	}
}
