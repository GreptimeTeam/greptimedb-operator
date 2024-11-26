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

package k8s

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
func IsObjectSpecEqual(oldObject, newObject client.Object, annotationKey string) (bool, error) {
	oldObjectSpecStr, ok := oldObject.GetAnnotations()[annotationKey]
	if !ok {
		return false, fmt.Errorf("the objectA object '%s' does not have annotation '%s'",
			client.ObjectKeyFromObject(oldObject), annotationKey)
	}

	newObjectSpecStr, ok := newObject.GetAnnotations()[annotationKey]
	if !ok {
		return false, fmt.Errorf("the objectB object '%s' does not have annotation '%s'",
			client.ObjectKeyFromObject(newObject), annotationKey)
	}

	var oldObjectSpec, newObjectSpec interface{}

	if err := json.Unmarshal([]byte(oldObjectSpecStr), &oldObjectSpec); err != nil {
		return false, err
	}

	if err := json.Unmarshal([]byte(newObjectSpecStr), &newObjectSpec); err != nil {
		return false, err
	}

	return reflect.DeepEqual(oldObjectSpec, newObjectSpec), nil
}

// IsObjectLabelsEqual checks if the labels of the object is equal to other one.
func IsObjectLabelsEqual(oldObjectLabels, newObjectLabels map[string]string) bool {
	return reflect.DeepEqual(oldObjectLabels, newObjectLabels)
}

// IsDeploymentReady checks if the deployment is ready.
func IsDeploymentReady(deployment *appsv1.Deployment) bool {
	if deployment == nil {
		return false
	}

	if deployment.Status.ObservedGeneration != deployment.Generation {
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
func IsStatefulSetReady(sts *appsv1.StatefulSet) bool {
	if sts == nil {
		return false
	}

	if sts.Status.ObservedGeneration != sts.Generation {
		return false
	}

	return sts.Status.ReadyReplicas == *sts.Spec.Replicas && sts.Status.CurrentReplicas == *sts.Spec.Replicas
}

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

// SourceObject create the unstructured object from the given object.
func SourceObject(input client.Object) client.Object {
	u := &unstructured.Unstructured{}

	// MUST set the APIVersion and Kind.
	u.SetGroupVersionKind(input.GetObjectKind().GroupVersionKind())

	return u
}

// GetSecretsData returns data of according keys from the secret.
func GetSecretsData(namespace, name string, keys []string) ([][]byte, error) {
	var secret corev1.Secret
	if err := GetK8sResource(namespace, name, &secret); err != nil {
		return nil, err
	}

	if secret.Data == nil {
		return nil, fmt.Errorf("secret '%s/%s' is empty", namespace, name)
	}

	var values [][]byte
	for _, key := range keys {
		value := secret.Data[key]
		if value == nil {
			return nil, fmt.Errorf("secret '%s/%s' does not have key '%s'", namespace, name, key)
		}
		values = append(values, value)
	}

	return values, nil
}
