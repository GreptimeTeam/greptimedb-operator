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

	"dario.cat/mergo"
	"github.com/sergi/go-diff/diffmatchpatch"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/yaml"
)

func TestClusterSetDefaults(t *testing.T) {
	const (
		testDir        = "testdata/defaulting/greptimedbcluster/setdefaults"
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

func TestClusterMerge(t *testing.T) {
	const (
		testDir        = "testdata/defaulting/greptimedbcluster/merge"
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

			if err := input.MergeWithBaseTemplate(); err != nil {
				t.Fatalf("failed to merge template: %v", err)
			}

			if err := input.MergeWithGlobalLogging(); err != nil {
				t.Fatalf("failed to merge logging: %v", err)
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

func TestStandaloneSetDefaults(t *testing.T) {
	const (
		testDir        = "testdata/defaulting/greptimedbstandalone"
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
				input  GreptimeDBStandalone
				expect GreptimeDBStandalone
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

func TestIntOrStringTransformer(t *testing.T) {
	type foo struct {
		Val *intstr.IntOrString
	}
	type testStruct struct {
		Src    foo
		Dst    foo
		Expect foo
	}

	tests := []testStruct{
		{
			Src:    foo{Val: &intstr.IntOrString{Type: intstr.String, StrVal: "1"}},
			Dst:    foo{Val: &intstr.IntOrString{Type: intstr.Int, IntVal: 10}},
			Expect: foo{Val: &intstr.IntOrString{Type: intstr.Int, IntVal: 10}},
		},
		{
			Src:    foo{Val: &intstr.IntOrString{Type: intstr.String, StrVal: "25%"}},
			Dst:    foo{Val: &intstr.IntOrString{Type: intstr.String, StrVal: "75%"}},
			Expect: foo{Val: &intstr.IntOrString{Type: intstr.String, StrVal: "75%"}},
		},
		{
			Src:    foo{Val: &intstr.IntOrString{Type: intstr.String, StrVal: "25%"}},
			Dst:    foo{},
			Expect: foo{Val: &intstr.IntOrString{Type: intstr.String, StrVal: "25%"}},
		},
		{
			Src:    foo{Val: &intstr.IntOrString{Type: intstr.Int, IntVal: 10}},
			Dst:    foo{},
			Expect: foo{Val: &intstr.IntOrString{Type: intstr.Int, IntVal: 10}},
		},
	}

	for i, tt := range tests {
		if err := mergo.Merge(&tt.Dst, &tt.Src, mergo.WithTransformers(intOrStringTransformer{})); err != nil {
			t.Errorf("test [%d] failed: %v", i, err)
		}

		if !reflect.DeepEqual(tt.Dst, tt.Expect) {
			t.Errorf("test [%d] failed: expected '%v', got '%v'", i, tt.Expect, tt.Dst)
		}
	}
}
