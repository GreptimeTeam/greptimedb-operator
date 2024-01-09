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

package greptimedbstandalone

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/GreptimeTeam/greptimedb-operator/apis/v1alpha1"
	"github.com/GreptimeTeam/greptimedb-operator/cmd/operator/app/options"
)

// Reconciler reconciles a GreptimeDBStandalone object
type Reconciler struct {
	client.Client

	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

func Setup(mgr ctrl.Manager, _ *options.Options) error {
	reconciler := &Reconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Recorder: mgr.GetEventRecorderFor("greptimedbstandalone-controller"),
	}
	return reconciler.SetupWithManager(mgr)
}

//+kubebuilder:rbac:groups=greptime.io,resources=greptimedbstandalones,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=greptime.ioresources=greptimedbstandalones/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=greptime.io,resources=greptimedbstandalones/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	klog.Infof("Reconciling GreptimeDBStandalone: %s", req.NamespacedName)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.GreptimeDBStandalone{}).
		Complete(r)
}
