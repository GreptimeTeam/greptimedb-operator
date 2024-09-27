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
