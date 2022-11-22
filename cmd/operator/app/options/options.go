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
