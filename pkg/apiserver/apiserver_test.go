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
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"

	greptimev1alpha1 "github.com/GreptimeTeam/greptimedb-operator/apis/v1alpha1"
)

const (
	TestPort      = 18081
	TestNamespace = "default"
	TestCluster   = "testcluster"
)

func TestHTTPGetService(t *testing.T) {
	// Start the HTTP service.
	svc := NewServer(&FakeClient{}, &Options{Port: TestPort})
	go func() {
		if err := svc.Run(); err != nil {
			t.Errorf("failed to start HTTP service: %v", err)
		}
	}()

	// Wait for the HTTP service to start.
	time.Sleep(1 * time.Second)

	var (
		testAPI    = fmt.Sprintf("http://localhost:%d%s", TestPort, APIGroup)
		httpClient = &http.Client{}
	)

	tests := []struct {
		url      string
		httpCode int
		success  bool
	}{
		{
			url:      fmt.Sprintf("%s/namespaces/all/clusters", testAPI),
			httpCode: http.StatusOK,
			success:  true,
		},
		{
			url:      fmt.Sprintf("%s/namespaces/foo/clusters/bar", testAPI),
			httpCode: http.StatusNotFound,
			success:  true,
		},
		{
			url:      fmt.Sprintf("%s/namespaces/%s/clusters/%s", testAPI, TestNamespace, TestCluster),
			httpCode: http.StatusOK,
			success:  true,
		},
		{
			url:      fmt.Sprintf("%s/namespaces/default/clusters", testAPI),
			httpCode: http.StatusOK,
			success:  true,
		},
		{
			url:      fmt.Sprintf("%s/namespaces/foo/clusters/", testAPI),
			httpCode: http.StatusNotFound,
			success:  true,
		},
	}

	for i, tt := range tests {
		req, err := http.NewRequest(http.MethodGet, tt.url, nil)
		if err != nil {
			t.Errorf("[%d] failed to create a GET request for url '%s': '%v'", i, tt.url, err)
		}

		resp, err := httpClient.Do(req)
		if err != nil {
			t.Errorf("[%d] failed to send a GET request to url '%s': '%v'", i, tt.url, err)
		}
		resp.Body.Close()

		if resp.StatusCode != tt.httpCode {
			t.Errorf("[%d] expected status code %d, got %d, url: '%s'", i, tt.httpCode, resp.StatusCode, tt.url)
		}
	}
}

var _ client.Client = &FakeClient{}

// FakeClient is a fake implementation of the client.Client interface.
type FakeClient struct{}

func (f *FakeClient) Get(_ context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	if key.String() == fmt.Sprintf("%s/%s", TestNamespace, TestCluster) {
		cluster := obj.(*greptimev1alpha1.GreptimeDBCluster)
		cluster.Name = TestCluster
		cluster.Namespace = TestNamespace
		_ = cluster.SetDefaults()
		return nil
	} else {
		return errors.NewNotFound(schema.GroupResource{
			Group:    "greptime.io",
			Resource: "GreptimeDBCluster",
		}, key.Name)
	}
}

func (f *FakeClient) SubResource(_ string) client.SubResourceClient {
	return nil
}

func (f *FakeClient) GroupVersionKindFor(_ runtime.Object) (schema.GroupVersionKind, error) {
	return schema.GroupVersionKind{}, nil
}

func (f *FakeClient) IsObjectNamespaced(_ runtime.Object) (bool, error) {
	return true, nil
}

func (f *FakeClient) List(_ context.Context, list client.ObjectList, opts ...client.ListOption) error {
	listOpts := &client.ListOptions{}
	for _, opt := range opts {
		opt.ApplyToList(listOpts)
	}

	if listOpts.Namespace != "" && listOpts.Namespace != TestNamespace {
		return nil
	}

	switch v := list.(type) {
	case *greptimev1alpha1.GreptimeDBClusterList:
		cluster := new(greptimev1alpha1.GreptimeDBCluster)
		cluster.Name = TestCluster
		cluster.Namespace = TestNamespace
		_ = cluster.SetDefaults()
		v.Items = append(v.Items, *cluster)
	case *corev1.PodList:
		pod := new(corev1.Pod)
		pod.Name = "testpod"
		pod.Namespace = TestNamespace
		pod.Spec.Containers = []corev1.Container{
			{
				Name: "testcontainer",
				Resources: corev1.ResourceRequirements{
					Limits: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("1"),
						corev1.ResourceMemory: resource.MustParse("1Gi"),
					},
					Requests: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("1"),
						corev1.ResourceMemory: resource.MustParse("1Gi"),
					},
				},
			},
		}
		v.Items = append(v.Items, *pod)
	default:
		return errors.NewBadRequest("unknown list type")
	}

	return nil
}

func (f *FakeClient) Create(_ context.Context, _ client.Object, _ ...client.CreateOption) error {
	return nil
}

func (f *FakeClient) Delete(_ context.Context, _ client.Object, _ ...client.DeleteOption) error {
	return nil
}

func (f *FakeClient) Update(_ context.Context, _ client.Object, _ ...client.UpdateOption) error {
	return nil
}

func (f *FakeClient) Patch(_ context.Context, _ client.Object, _ client.Patch, _ ...client.PatchOption) error {
	return nil
}

func (f *FakeClient) DeleteAllOf(_ context.Context, _ client.Object, _ ...client.DeleteAllOfOption) error {
	return nil
}

func (f *FakeClient) Status() client.StatusWriter {
	return nil
}

func (f *FakeClient) Scheme() *runtime.Scheme {
	return nil
}

func (f *FakeClient) RESTMapper() meta.RESTMapper {
	return nil
}
