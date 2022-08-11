package app

import (
	"flag"
	"os"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/klog/v2"
	"k8s.io/klog/v2/klogr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"

	"github.com/greptime/greptimedb-operator/apis/v1alpha1"
	"github.com/greptime/greptimedb-operator/cmd/operator/app/options"
	"github.com/greptime/greptimedb-operator/cmd/operator/app/version"
	"github.com/greptime/greptimedb-operator/controllers/greptimedbcluster"
)

const (
	leaderElectionID = "greptimedb-operator"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(v1alpha1.AddToScheme(scheme))

	// +kubebuilder:scaffold:scheme
}

func NewOperatorCommand() *cobra.Command {
	o := options.NewDefaultOptions()

	command := &cobra.Command{
		Use:   "greptimedb-operator",
		Short: "greptimedb-operator manages GreptimeDB clusters atop Kubernetes.",
		Run: func(cmd *cobra.Command, args []string) {
			ctrl.SetLogger(klogr.New())
			setupLog := ctrl.Log.WithName("setup")
			cfg := ctrl.GetConfigOrDie()

			mgr, err := ctrl.NewManager(cfg, ctrl.Options{
				Scheme:                 scheme,
				MetricsBindAddress:     o.MetricsAddr,
				HealthProbeBindAddress: o.HealthProbeAddr,
				LeaderElection:         o.EnableLeaderElection,
				LeaderElectionID:       leaderElectionID,
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
