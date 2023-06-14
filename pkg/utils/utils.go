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

package utils

import (
	"context"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetK8sResource returns a native K8s resource by namespace and name.
func GetK8sResource(namespace, name string, obj client.Object) error {
	c, err := client.New(ctrl.GetConfigOrDie(), client.Options{})
	if err != nil {
		return err
	}

	if err := c.Get(context.Background(), client.ObjectKey{Namespace: namespace, Name: name}, obj); err != nil {
		return err
	}

	return nil
}

// MergeStringMap merges two string maps.
func MergeStringMap(origin, new map[string]string) map[string]string {
	if origin == nil {
		origin = make(map[string]string)
	}

	for k, v := range new {
		origin[k] = v
	}

	return origin
}
