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
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"

	k8sutils "github.com/GreptimeTeam/greptimedb-operator/pkg/util/k8s"
)

var (
	ErrSyncNotReady = fmt.Errorf("the sync process is not ready")
)

// Deployer is the interface that abstracts the deployment of a component.
type Deployer interface {
	Syncer
	ComponentOperator
}

// Hook is a function that will be executed before or after the Sync().
type Hook func(ctx context.Context, crdObject client.Object) error

// Syncer use the deployer to sync the CRD object.
type Syncer interface {
	Sync(ctx context.Context, crdObject client.Object, operator ComponentOperator) error
}

// ComponentOperator is the interface that define the behaviors of a deployer.
type ComponentOperator interface {
	// Generate generates the multiple Kubernetes objects base on the CRD object.
	Generate(crdObject client.Object) ([]client.Object, error)

	// Apply creates or update the Kubernetes objects that generated by Render().
	// If the object is not existed, it will be created.
	// If the object is existed, it will be updated if the object is different.
	Apply(ctx context.Context, objects []client.Object) error

	// CleanUp cleans up the resources that created by the deployer.
	CleanUp(ctx context.Context, crdObject client.Object) error

	// CheckAndUpdateStatus checks if the status of Kubernetes objects are ready and update the status.
	CheckAndUpdateStatus(ctx context.Context, crdObject client.Object) (bool, error)

	// PreSyncHooks returns the hooks that will be executed before the core login of Sync().
	PreSyncHooks() []Hook

	// PostSyncHooks returns the hooks that will be executed after the core login of Sync().
	PostSyncHooks() []Hook
}

var _ Deployer = &DefaultDeployer{}

// DefaultDeployer implement some common behaviors of the ComponentOperator interface.
type DefaultDeployer struct {
	client.Client
}

func (d *DefaultDeployer) Sync(ctx context.Context, crdObject client.Object, operator ComponentOperator) error {
	objects, err := operator.Generate(crdObject)
	if err != nil {
		return err
	}

	if len(objects) == 0 {
		return nil
	}

	if err := d.onPrevSync(ctx, crdObject, operator); err != nil {
		return err
	}

	if err := operator.Apply(ctx, objects); err != nil {
		return err
	}

	ready, err := operator.CheckAndUpdateStatus(ctx, crdObject)
	if err != nil {
		return err
	}

	if !ready {
		return ErrSyncNotReady
	}

	if err := d.onPostSync(ctx, crdObject, operator); err != nil {
		return err
	}

	return nil
}

func (d *DefaultDeployer) Generate(_ client.Object) ([]client.Object, error) {
	return nil, nil
}

func (d *DefaultDeployer) Apply(ctx context.Context, objects []client.Object) error {
	for _, newObject := range objects {
		oldObject, err := k8sutils.CreateObjectIfNotExist(ctx, d.Client, d.sourceObject(newObject), newObject)
		if err != nil {
			return err
		}

		if oldObject != nil {
			equal, err := k8sutils.IsObjectSpecEqual(oldObject, newObject, LastAppliedResourceSpec)
			if err != nil {
				return err
			}

			// If the spec is not equal, update the object.
			if !equal {
				if err := d.Client.Patch(ctx, newObject, client.MergeFrom(oldObject)); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (d *DefaultDeployer) CleanUp(_ context.Context, _ client.Object) error {
	return nil
}

func (d *DefaultDeployer) CheckAndUpdateStatus(_ context.Context, _ client.Object) (bool, error) {
	return false, nil
}

func (d *DefaultDeployer) PreSyncHooks() []Hook {
	return nil
}

func (d *DefaultDeployer) PostSyncHooks() []Hook {
	return nil
}

// onPrevSync executes the PreSyncHooks.
func (d *DefaultDeployer) onPrevSync(ctx context.Context, crdObject client.Object, operator ComponentOperator) error {
	for _, hook := range operator.PreSyncHooks() {
		if err := hook(ctx, crdObject); err != nil {
			return err
		}
	}
	return nil
}

// onPostSync executes the PostSyncHooks.
func (d *DefaultDeployer) onPostSync(ctx context.Context, crdObject client.Object, operator ComponentOperator) error {
	for _, hook := range operator.PostSyncHooks() {
		if err := hook(ctx, crdObject); err != nil {
			return err
		}
	}
	return nil
}

// sourceObject create the unstructured object from the given object.
func (d *DefaultDeployer) sourceObject(input client.Object) client.Object {
	u := &unstructured.Unstructured{}

	// MUST set the APIVersion and Kind.
	u.SetGroupVersionKind(input.GetObjectKind().GroupVersionKind())

	return u
}
