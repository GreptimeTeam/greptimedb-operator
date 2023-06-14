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

package deployer

import (
	"encoding/json"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type Builder interface {
	BuildService() Builder
	BuildDeployment() Builder
	BuildStatefulSet() Builder
	BuildConfigMap() Builder
	BuildPodMonitor() Builder
	Generate() ([]client.Object, error)
}

var _ Builder = &DefaultBuilder{}

type DefaultBuilder struct {
	Scheme  *runtime.Scheme
	Objects []client.Object

	// record error for builder pattern.
	Err error
}

func (d *DefaultBuilder) BuildService() Builder {
	return d
}

func (d *DefaultBuilder) BuildDeployment() Builder {
	return d
}

func (d *DefaultBuilder) BuildStatefulSet() Builder {
	return d
}

func (d *DefaultBuilder) BuildConfigMap() Builder {
	return d
}

func (d *DefaultBuilder) BuildPodMonitor() Builder {
	return d
}

func (d *DefaultBuilder) Generate() ([]client.Object, error) {
	return d.Objects, d.Err
}

// SetControllerAndAnnotation sets the controller reference and annotation for the object.
func SetControllerAndAnnotation(owner, controlled client.Object, scheme *runtime.Scheme, spec interface{}) error {
	if err := controllerutil.SetControllerReference(owner, controlled, scheme); err != nil {
		return err
	}

	if err := setLastAppliedResourceSpecAnnotation(controlled, spec); err != nil {
		return err
	}

	return nil
}

func MergeStringMap(origin, new map[string]string) map[string]string {
	if origin == nil {
		origin = make(map[string]string)
	}

	for k, v := range new {
		origin[k] = v
	}

	return origin
}

// setLastAppliedResourceSpecAnnotation sets the last applied resource spec as annotation for updating purpose.
func setLastAppliedResourceSpecAnnotation(object client.Object, spec interface{}) error {
	data, err := json.Marshal(spec)
	if err != nil {
		return err
	}

	annotations := MergeStringMap(object.GetAnnotations(), map[string]string{LastAppliedResourceSpec: string(data)})
	object.SetAnnotations(annotations)

	return nil
}
