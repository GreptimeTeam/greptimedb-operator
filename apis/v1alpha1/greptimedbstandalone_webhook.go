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
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var greptimedbstandalonelog = logf.Log.WithName("greptimedbstandalone-resource")

func (r *GreptimeDBStandalone) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(liyang): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-greptime-io-v1alpha1-greptimedbstandalone,mutating=false,failurePolicy=fail,sideEffects=None,groups=greptime.io,resources=greptimedbstandalones,verbs=create;update,versions=v1alpha1,name=vgreptimedbstandalone.kb.io,admissionReviewVersions=v1

var _ admission.CustomValidator = &GreptimeDBStandalone{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *GreptimeDBStandalone) ValidateCreate(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	greptimedbstandalonelog.Info("validate create", "name", r.Name)

	_, ok := obj.(*GreptimeDBStandalone)
	if !ok {
		return nil, fmt.Errorf("unexpected type: %T", obj)
	}

	if err := r.Validate(); err != nil {
		return nil, err
	}

	return nil, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *GreptimeDBStandalone) ValidateUpdate(_ context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	greptimedbstandalonelog.Info("validate update", "name", r.Name)

	_, ok := newObj.(*GreptimeDBStandalone)
	if !ok {
		return nil, fmt.Errorf("unexpected type: %T", newObj)
	}

	if err := r.Validate(); err != nil {
		return nil, err
	}

	return nil, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *GreptimeDBStandalone) ValidateDelete(_ context.Context, _ runtime.Object) (admission.Warnings, error) {
	// FIXME(liyang): Unnecessary validation when object deletion.
	return nil, nil
}
