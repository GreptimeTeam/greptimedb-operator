package options

import (
	"github.com/spf13/pflag"
)

const (
	defaultMetricsAddr     = ":8080"
	defaultHealthProbeAddr = ":9494"
)

type Options struct {
	MetricsAddr          string
	HealthProbeAddr      string
	EnableLeaderElection bool
}

func NewDefaultOptions() *Options {
	return &Options{
		MetricsAddr:     defaultMetricsAddr,
		HealthProbeAddr: defaultHealthProbeAddr,
	}
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.MetricsAddr, "metrics-bind-address", o.MetricsAddr, "The address the metric endpoint binds to.")
	fs.StringVar(&o.HealthProbeAddr, "health-probe-bind-address", o.HealthProbeAddr, "The address the probe endpoint binds to.")
	fs.BoolVar(&o.EnableLeaderElection, "enable-leader-election", o.EnableLeaderElection, "Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
}
