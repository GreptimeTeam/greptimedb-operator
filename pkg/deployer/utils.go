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

package deployer

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	LastAppliedResourceSpec = "controller.greptime.io/last-applied-resource-spec"
)

// CreateObjectIfNotExist creates Kubernetes object if it does not exist, otherwise returns the existing object.
func CreateObjectIfNotExist(ctx context.Context, c client.Client, source, newObject client.Object) (client.Object, error) {
	err := c.Get(ctx, client.ObjectKey{Namespace: newObject.GetNamespace(), Name: newObject.GetName()}, source)

	// If the object does not exist, create it.
	if errors.IsNotFound(err) {
		if err := c.Create(ctx, newObject); err != nil {
			return nil, err
		}
		return nil, nil
	}

	// Other errors happen.
	if err != nil {
		return nil, err
	}

	// The object already exists, return it.
	return source, nil
}

// IsObjectSpecEqual checks if the spec of the object is equal to other one.
func IsObjectSpecEqual(objectA, objectB client.Object) (bool, error) {
	objectASpecStr, ok := objectA.GetAnnotations()[LastAppliedResourceSpec]
	if !ok {
		return false, fmt.Errorf("the objectA object '%s' does not have annotation '%s'",
			client.ObjectKeyFromObject(objectA), LastAppliedResourceSpec)
	}

	objectBSpecStr, ok := objectB.GetAnnotations()[LastAppliedResourceSpec]
	if !ok {
		return false, fmt.Errorf("the objectB object '%s' does not have annotation '%s'",
			client.ObjectKeyFromObject(objectB), LastAppliedResourceSpec)
	}

	var objectASpec, objectBSpec interface{}

	if err := json.Unmarshal([]byte(objectASpecStr), &objectASpec); err != nil {
		return false, err
	}

	if err := json.Unmarshal([]byte(objectBSpecStr), &objectBSpec); err != nil {
		return false, err
	}

	return reflect.DeepEqual(objectASpec, objectBSpec), nil
}

// IsDeploymentReady checks if the deployment is ready.
// TODO(zyy17): Maybe it's not a accurate way to detect the statefulset is ready.
func IsDeploymentReady(deployment *appsv1.Deployment) bool {
	if deployment == nil {
		return false
	}

	for _, cond := range deployment.Status.Conditions {
		if cond.Type == appsv1.DeploymentProgressing {
			if cond.Reason == "NewReplicaSetAvailable" &&
				deployment.Status.ReadyReplicas == *deployment.Spec.Replicas {
				return true
			}
		}
	}

	return false
}

// IsStatefulSetReady checks if the statefulset is ready.
// TODO(zyy17): Maybe it's not a accurate way to detect the deployment is ready.
func IsStatefulSetReady(sts *appsv1.StatefulSet) bool {
	if sts == nil {
		return false
	}

	return sts.Status.ReadyReplicas == *sts.Spec.Replicas
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

// setLastAppliedResourceSpecAnnotation sets the last applied resource spec as annotation for updating purpose.
func setLastAppliedResourceSpecAnnotation(object client.Object, spec interface{}) error {
	data, err := json.Marshal(spec)
	if err != nil {
		return err
	}

	annotations := mergeAnnotations(object.GetAnnotations(), map[string]string{LastAppliedResourceSpec: string(data)})
	object.SetAnnotations(annotations)

	return nil
}

func mergeAnnotations(input, new map[string]string) map[string]string {
	if len(input) == 0 {
		return new
	}
	for k, v := range new {
		input[k] = v
	}
	return input
}
