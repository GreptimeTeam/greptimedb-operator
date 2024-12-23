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

package app

import (
	"flag"
	"os"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/spf13/cobra"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/klog/v2"
	"k8s.io/klog/v2/textlogger"
	podmetricsv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	"github.com/GreptimeTeam/greptimedb-operator/apis/v1alpha1"
	"github.com/GreptimeTeam/greptimedb-operator/cmd/operator/app/options"
	"github.com/GreptimeTeam/greptimedb-operator/cmd/operator/app/version"
	"github.com/GreptimeTeam/greptimedb-operator/controllers/greptimedbcluster"
	"github.com/GreptimeTeam/greptimedb-operator/controllers/greptimedbstandalone"
	"github.com/GreptimeTeam/greptimedb-operator/pkg/apiserver"
)

const (
	leaderElectionID = "greptimedb-operator"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	// Add Kubernetes client-go scheme.
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	// Add Kubernetes API extensions.
	utilruntime.Must(apiextensionsv1.AddToScheme(scheme))

	// Add GreptimeDB CRD.
	utilruntime.Must(v1alpha1.AddToScheme(scheme))

	// Add prometheus-operator's CRDs for monitoring(PodMonitor and ServiceMonitor).
	utilruntime.Must(monitoringv1.AddToScheme(scheme))

	// Add [PodMetrics](https://github.com/kubernetes/metrics/blob/master/pkg/apis/metrics/v1beta1/types.go) for fetching PodMetrics from metrics-server.
	utilruntime.Must(podmetricsv1beta1.AddToScheme(scheme))

	// Add admission webhook scheme.
	utilruntime.Must(admissionv1beta1.AddToScheme(scheme))

	// +kubebuilder:scaffold:scheme
}

func NewOperatorCommand() *cobra.Command {
	o := options.NewDefaultOptions()

	command := &cobra.Command{
		Use:   "greptimedb-operator",
		Short: "greptimedb-operator manages GreptimeDB clusters atop Kubernetes.",
		Run: func(cmd *cobra.Command, args []string) {
			ctrl.SetLogger(textlogger.NewLogger(textlogger.NewConfig()))
			setupLog := ctrl.Log.WithName("setup")
			cfg := ctrl.GetConfigOrDie()

			webhookServer := webhook.NewServer(webhook.Options{})
			if o.EnableAdmissionWebhook {
				webhookServerOptions := webhook.Options{
					Port:    o.AdmissionWebhookPort,
					CertDir: o.AdmissionWebhookCertDir,
				}
				webhookServer = webhook.NewServer(webhookServerOptions)
			}
			mgr, err := ctrl.NewManager(cfg, ctrl.Options{
				Scheme:                 scheme,
				HealthProbeBindAddress: o.HealthProbeAddr,
				LeaderElection:         o.EnableLeaderElection,
				LeaderElectionID:       leaderElectionID,
				Metrics: metricsserver.Options{
					BindAddress: o.MetricsAddr,
				},
				WebhookServer: webhookServer,
			})
			if err != nil {
				setupLog.Error(err, "unable to start manager")
				os.Exit(1)
			}

			if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
				setupLog.Error(err, "unable to setup healthz check")
				os.Exit(1)
			}

			if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
				setupLog.Error(err, "unable to setup readyz check")
				os.Exit(1)
			}

			if err := greptimedbcluster.Setup(mgr, o); err != nil {
				setupLog.Error(err, "unable to setup controller", "controller", "greptimedbcluster")
				os.Exit(1)
			}

			if err := greptimedbstandalone.Setup(mgr, o); err != nil {
				setupLog.Error(err, "unable to setup controller", "controller", "greptimedbstandalone")
				os.Exit(1)
			}

			if o.EnableAdmissionWebhook {
				if err := (&v1alpha1.GreptimeDBCluster{}).SetupWebhookWithManager(mgr); err != nil {
					setupLog.Error(err, "unable to setup admission webhook", "controller", "greptimedbcluster")
					os.Exit(1)
				}
				if err := (&v1alpha1.GreptimeDBStandalone{}).SetupWebhookWithManager(mgr); err != nil {
					setupLog.Error(err, "unable to setup admission webhook", "controller", "greptimedbstandalone")
					os.Exit(1)
				}
			}

			if o.EnableAPIServer {
				server, err := apiserver.NewServer(mgr, &apiserver.Options{
					Port:             o.APIServerPort,
					EnablePodMetrics: o.EnablePodMetrics,
				})
				if err != nil {
					setupLog.Error(err, "unable to create API server")
					os.Exit(1)
				}

				go func() {
					if err := server.Run(); err != nil {
						setupLog.Error(err, "unable to run API service")
						os.Exit(1)
					}
				}()
			}

			// +kubebuilder:scaffold:builder

			setupLog.Info("starting manager")
			if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
				setupLog.Error(err, "unable to start manager")
				os.Exit(1)
			}
		},
	}

	o.AddFlags(command.Flags())
	klog.InitFlags(nil)
	command.Flags().AddGoFlagSet(flag.CommandLine)

	command.AddCommand(version.NewVersionCommand())

	return command
}
