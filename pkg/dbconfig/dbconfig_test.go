// Copyright 2023 Greptime Team
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

package dbconfig

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/GreptimeTeam/greptimedb-operator/apis/v1alpha1"
)

func TestFromClusterCRD(t *testing.T) {
	testCluster := &v1alpha1.GreptimeDBCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-cluster",
			Namespace: "default",
		},
		Spec: v1alpha1.GreptimeDBClusterSpec{
			StorageProvider: &v1alpha1.StorageProvider{
				S3: &v1alpha1.S3StorageProvider{
					Prefix: "testcluster",
					Bucket: "testbucket",
				},
			},
		},
	}

	_, err := FromClusterCRD(testCluster, v1alpha1.MetaComponentKind)
	if err != nil {
		t.Fatalf("failed to generate metasrv config from cluster CRD: %v", err)
	}

	_, err = FromClusterCRD(testCluster, v1alpha1.DatanodeComponentKind)
	if err != nil {
		t.Fatalf("failed to generate datanode config from cluster CRD: %v", err)
	}

	_, err = FromClusterCRD(testCluster, v1alpha1.FrontendComponentKind)
	if err != nil {
		t.Fatalf("failed to generate frontend config from cluster CRD: %v", err)
	}
}
