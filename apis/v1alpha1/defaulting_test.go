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

package v1alpha1

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/sergi/go-diff/diffmatchpatch"
	"sigs.k8s.io/yaml"
)

func TestClusterSetDefaults(t *testing.T) {
	const (
		testDir        = "testdata/greptimedbcluster"
		inputFileName  = "input.yaml"
		expectFileName = "expect.yaml"
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

			expectFile := filepath.Join(testDir, entry.Name(), expectFileName)
			expectData, err := os.ReadFile(expectFile)
			if err != nil {
				t.Fatalf("failed to read %s: %v", expectFile, err)
			}

			var (
				input  GreptimeDBCluster
				expect GreptimeDBCluster
			)
			if err := yaml.Unmarshal(inputData, &input); err != nil {
				t.Fatalf("failed to unmarshal %s: %v", inputFile, err)
			}
			if err := yaml.Unmarshal(expectData, &expect); err != nil {
				t.Fatalf("failed to unmarshal %s: %v", expectFile, err)
			}

			if err := input.SetDefaults(); err != nil {
				t.Fatalf("failed to set defaults: %v", err)
			}

			if !reflect.DeepEqual(input, expect) {
				rawInputData, err := yaml.Marshal(input)
				if err != nil {
					t.Fatalf("failed to marshal: %v", err)
				}

				rawExpectData, err := yaml.Marshal(expect)
				if err != nil {
					t.Fatalf("failed to marshal: %v", err)
				}

				// Use diffmatchpatch to get a human-readable diff.
				dmp := diffmatchpatch.New()
				t.Errorf("unexpected result for %s:\n%s", entry.Name(), dmp.DiffPrettyText(dmp.DiffMain(string(rawExpectData), string(rawInputData), false)))
			}
		}
	}
}
