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

package v1alpha1

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"k8s.io/utils/ptr"
	"sigs.k8s.io/yaml"
)

func TestClusterValidation(t *testing.T) {
	const (
		testDir       = "testdata/validation/greptimedbcluster"
		inputFileName = "input.yaml"
	)

	entries, err := os.ReadDir(testDir)
	if err != nil {
		t.Fatal(err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			inputFile := filepath.Join(testDir, entry.Name(), inputFileName)
			inputData, err := os.ReadFile(inputFile)
			if err != nil {
				t.Errorf("failed to read %s: %v", inputFile, err)
			}

			var input GreptimeDBCluster

			if err := yaml.Unmarshal(inputData, &input); err != nil {
				t.Fatalf("failed to unmarshal %s: %v", inputFile, err)
			}

			if err := input.Validate(); err != nil && !expectError(input.GetName()) {
				t.Errorf("expected no error, but got %v", err)
			}
		}
	}
}

func TestStandaloneValidation(t *testing.T) {
	const (
		testDir       = "testdata/validation/greptimedbstandalone"
		inputFileName = "input.yaml"
	)

	entries, err := os.ReadDir(testDir)
	if err != nil {
		t.Fatal(err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			inputFile := filepath.Join(testDir, entry.Name(), inputFileName)
			inputData, err := os.ReadFile(inputFile)
			if err != nil {
				t.Errorf("failed to read %s: %v", inputFile, err)
			}

			var input GreptimeDBStandalone

			if err := yaml.Unmarshal(inputData, &input); err != nil {
				t.Fatalf("failed to unmarshal %s: %v", inputFile, err)
			}

			if err := input.Validate(); err != nil && !expectError(input.GetName()) {
				t.Errorf("expected no error, but got %v", err)
			}
		}
	}
}

// If the resource name contains the word "error", we expect an error in the validation.
func expectError(name string) bool {
	return strings.Contains(name, "error")
}

func TestDatanodeDrainRequiresObjectStorage(t *testing.T) {
	newCluster := func(datanode *DatanodeSpec, objectStorage *ObjectStorageProviderSpec) *GreptimeDBCluster {
		return &GreptimeDBCluster{
			Spec: GreptimeDBClusterSpec{
				Datanode:              datanode,
				ObjectStorageProvider: objectStorage,
			},
		}
	}
	objectStorage := &ObjectStorageProviderSpec{
		S3: &S3Storage{Bucket: "bucket", Region: "region", Root: "/"},
	}

	t.Run("single datanode without object storage", func(t *testing.T) {
		cluster := newCluster(&DatanodeSpec{
			ComponentSpec: ComponentSpec{Replicas: ptr.To(int32(3))},
			Drain:         &DatanodeDrainSpec{Enabled: true},
		}, nil)
		if err := cluster.Validate(); err == nil {
			t.Fatal("expected validation error, got nil")
		}
	})

	t.Run("single datanode with object storage", func(t *testing.T) {
		cluster := newCluster(&DatanodeSpec{
			ComponentSpec: ComponentSpec{Replicas: ptr.To(int32(3))},
			Drain:         &DatanodeDrainSpec{Enabled: true},
		}, objectStorage)
		if err := cluster.Validate(); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("datanode group without object storage", func(t *testing.T) {
		cluster := newCluster(nil, nil)
		cluster.Spec.DatanodeGroups = []*DatanodeSpec{
			{
				Name:          "read",
				ComponentSpec: ComponentSpec{Replicas: ptr.To(int32(3))},
				Drain:         &DatanodeDrainSpec{Enabled: true},
			},
		}
		if err := cluster.Validate(); err == nil {
			t.Fatal("expected validation error, got nil")
		}
	})

	t.Run("drain disabled without object storage is allowed", func(t *testing.T) {
		cluster := newCluster(&DatanodeSpec{
			ComponentSpec: ComponentSpec{Replicas: ptr.To(int32(3))},
		}, nil)
		if err := cluster.Validate(); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})
}

func TestDatanodeDrainPolicyRotate(t *testing.T) {
	objectStorage := &ObjectStorageProviderSpec{
		S3: &S3Storage{Bucket: "bucket", Region: "region", Root: "/"},
	}

	base := func() *GreptimeDBCluster {
		return &GreptimeDBCluster{
			Spec: GreptimeDBClusterSpec{
				ObjectStorageProvider: objectStorage,
				Datanode: &DatanodeSpec{
					ComponentSpec: ComponentSpec{Replicas: ptr.To(int32(3))},
					Drain: &DatanodeDrainSpec{
						Enabled:   true,
						Policy:    DrainPolicyRotate,
						TempGroup: &TempDatanodeGroupSpec{Replicas: ptr.To(int32(1))},
					},
				},
			},
		}
	}

	t.Run("rotate with temp group and object storage", func(t *testing.T) {
		if err := base().Validate(); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("rotate without temp group", func(t *testing.T) {
		cluster := base()
		cluster.Spec.Datanode.Drain.TempGroup = nil
		if err := cluster.Validate(); err == nil {
			t.Fatal("expected validation error, got nil")
		}
	})

	t.Run("rotate without enableDrain", func(t *testing.T) {
		cluster := base()
		cluster.Spec.Datanode.Drain.Enabled = false
		if err := cluster.Validate(); err == nil {
			t.Fatal("expected validation error, got nil")
		}
	})
}
