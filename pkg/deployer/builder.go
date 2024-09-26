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
	"fmt"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	greptimev1alpha1 "github.com/GreptimeTeam/greptimedb-operator/apis/v1alpha1"
	"github.com/GreptimeTeam/greptimedb-operator/pkg/util"
)

const (
	LastAppliedResourceSpec = "controller.greptime.io/last-applied-resource-spec"
	ConfigHash              = "controller.greptime.io/config-hash"
)

// Builder is the interface for building K8s resources.
type Builder interface {
	// BuildService builds a K8s service.
	BuildService() Builder

	// BuildDeployment builds a K8s deployment.
	BuildDeployment() Builder

	// BuildStatefulSet builds a K8s statefulset.
	BuildStatefulSet() Builder

	// BuildConfigMap builds a K8s configmap.
	BuildConfigMap() Builder

	// BuildPodMonitor builds a Prometheus podmonitor.
	BuildPodMonitor() Builder

	// BuildGreptimeDBStandalone builds a GreptimeDBStandalone.
	BuildGreptimeDBStandalone() Builder

	// SetControllerAndAnnotation sets the controller reference and annotation for the object.
	SetControllerAndAnnotation() Builder

	// Generate returns the generated K8s resources.
	Generate() ([]client.Object, error)
}

var _ Builder = &DefaultBuilder{}

// DefaultBuilder is the default implementation of Builder.
type DefaultBuilder struct {
	Scheme  *runtime.Scheme
	Objects []client.Object
	Owner   client.Object

	// record error for builder pattern.
	Err error
}

func (b *DefaultBuilder) BuildService() Builder {
	return b
}

func (b *DefaultBuilder) BuildDeployment() Builder {
	return b
}

func (b *DefaultBuilder) BuildStatefulSet() Builder {
	return b
}

func (b *DefaultBuilder) BuildConfigMap() Builder {
	return b
}

func (b *DefaultBuilder) BuildPodMonitor() Builder {
	return b
}

func (b *DefaultBuilder) BuildGreptimeDBStandalone() Builder {
	return b
}

func (b *DefaultBuilder) SetControllerAndAnnotation() Builder {
	var (
		spec       interface{}
		controlled client.Object
	)

	for _, obj := range b.Objects {
		switch v := obj.(type) {
		case *corev1.Service:
			spec = v.Spec
			controlled = v
		case *corev1.ConfigMap:
			spec = v.Data
			controlled = v
		case *appsv1.StatefulSet:
			spec = v.Spec
			controlled = v
		case *appsv1.Deployment:
			spec = v.Spec
			controlled = v
		case *monitoringv1.PodMonitor:
			spec = v.Spec
			controlled = v
		case *greptimev1alpha1.GreptimeDBStandalone:
			spec = v.Spec
			controlled = v
		default:
			b.Err = fmt.Errorf("unsupported object type: %T", obj)
		}

		if err := b.doSetControllerAndAnnotation(b.Owner, controlled, b.Scheme, spec); err != nil {
			b.Err = err
			return b
		}
	}

	return b
}

func (b *DefaultBuilder) Generate() ([]client.Object, error) {
	return b.Objects, b.Err
}

// doSetControllerAndAnnotation sets the controller reference and annotation for the object.
func (b *DefaultBuilder) doSetControllerAndAnnotation(owner, controlled client.Object, scheme *runtime.Scheme, spec interface{}) error {
	if err := controllerutil.SetControllerReference(owner, controlled, scheme); err != nil {
		return err
	}

	if err := b.setLastAppliedResourceSpecAnnotation(controlled, spec); err != nil {
		return err
	}

	return nil
}

// setLastAppliedResourceSpecAnnotation sets the last applied resource spec as annotation for updating purpose.
func (b *DefaultBuilder) setLastAppliedResourceSpecAnnotation(object client.Object, spec interface{}) error {
	data, err := json.Marshal(spec)
	if err != nil {
		return err
	}

	annotations := util.MergeStringMap(object.GetAnnotations(), map[string]string{LastAppliedResourceSpec: string(data)})
	object.SetAnnotations(annotations)

	return nil
}
