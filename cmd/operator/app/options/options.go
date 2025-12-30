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

package options

import (
	"time"

	"github.com/spf13/pflag"
)

const (
	defaultMetricsAddr             = ":8080"
	defaultHealthProbeAddr         = ":9494"
	defaultAdmissionWebhookPort    = 8082
	defaultAdmissionWebhookCertDir = "/etc/webhook-tls"
	defaultProfilingPort           = 8083
)

type Options struct {
	MetricsAddr             string
	HealthProbeAddr         string
	EnableLeaderElection    bool
	LeaseDuration           time.Duration
	RenewDeadline           time.Duration
	RetryPeriod             time.Duration
	EnableAPIServer         bool
	APIServerPort           int32
	EnablePodMetrics        bool
	EnableAdmissionWebhook  bool
	AdmissionWebhookPort    int
	AdmissionWebhookCertDir string
	EnableProfiling         bool
	ProfilingPort           int32
}

func NewDefaultOptions() *Options {
	return &Options{
		MetricsAddr:             defaultMetricsAddr,
		HealthProbeAddr:         defaultHealthProbeAddr,
		EnableLeaderElection:    false,
		EnablePodMetrics:        false,
		EnableAdmissionWebhook:  false,
		AdmissionWebhookPort:    defaultAdmissionWebhookPort,
		AdmissionWebhookCertDir: defaultAdmissionWebhookCertDir,
		EnableProfiling:         false,
		ProfilingPort:           defaultProfilingPort,
	}
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.MetricsAddr, "metrics-bind-address", o.MetricsAddr, "The address the metric endpoint binds to.")
	fs.StringVar(&o.HealthProbeAddr, "health-probe-bind-address", o.HealthProbeAddr, "The address the probe endpoint binds to.")
	fs.BoolVar(&o.EnableLeaderElection, "enable-leader-election", o.EnableLeaderElection, "Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	fs.DurationVar(&o.LeaseDuration, "leader-election-lease-duration", o.LeaseDuration, "LeaseDuration is the duration that non-leader candidates will wait to force acquire leadership. This is measured against time of last observed ack. Default is 15 seconds.")
	fs.DurationVar(&o.RenewDeadline, "leader-election-renew-deadline", o.RenewDeadline, "RenewDeadline is the duration that the acting controlplane will retry refreshing leadership before giving up. Default is 10 seconds.")
	fs.DurationVar(&o.RetryPeriod, "leader-election-retry-period", o.RetryPeriod, "RetryPeriod is the duration the LeaderElector clients should wait between tries of actions. Default is 2 seconds.")
	fs.BoolVar(&o.EnablePodMetrics, "enable-pod-metrics", o.EnablePodMetrics, "Enable fetching PodMetrics from metrics-server.")
	fs.BoolVar(&o.EnableAdmissionWebhook, "enable-admission-webhook", o.EnableAdmissionWebhook, "Enable admission webhook for GreptimeDB operator.")
	fs.IntVar(&o.AdmissionWebhookPort, "admission-webhook-port", o.AdmissionWebhookPort, "The port the admission webhook binds to.")
	fs.StringVar(&o.AdmissionWebhookCertDir, "admission-webhook-cert-dir", o.AdmissionWebhookCertDir, "The directory that contains the server key and certificate.")
	fs.BoolVar(&o.EnableProfiling, "enable-profiling", o.EnableProfiling, "Enable pprof performance profiling (exposes /debug/pprof endpoints).")
	fs.Int32Var(&o.ProfilingPort, "profiling-port", o.ProfilingPort, "The port that pprof profiling HTTP server binds to (e.g., for accessing /debug/pprof).")
}
