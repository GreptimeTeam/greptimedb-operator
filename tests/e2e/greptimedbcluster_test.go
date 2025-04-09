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

package e2e

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/GreptimeTeam/greptimedb-operator/tests/e2e/greptimedbcluster"
)

var _ = Describe("Test GreptimeDBCluster", func() {
	var (
		ctx = context.Background()
	)

	AfterEach(func() {
		err := h.CleanEtcdData(ctx)
		Expect(err).NotTo(HaveOccurred(), "failed to clean etcd data")
	})

	It("Test a basic cluster", func() {
		greptimedbcluster.TestBasicCluster(ctx, h)
	})

	It("Test a cluster that enables remote wal", func() {
		greptimedbcluster.TestClusterEnableRemoteWal(ctx, h)
	})

	It("Test a cluster that enables flow", func() {
		greptimedbcluster.TestClusterEnableFlow(ctx, h)
	})

	It("Test scaling a cluster up and down", func() {
		greptimedbcluster.TestScaleCluster(ctx, h)
	})

	It("Test a cluster with dedicated WAL", func() {
		greptimedbcluster.TestClusterDedicatedWAL(ctx, h)
	})

	It("Test a cluster that enables monitoring", func() {
		greptimedbcluster.TestClusterEnableMonitoring(ctx, h)
	})

	It("Test a cluster with frontend group", func() {
		greptimedbcluster.TestClusterFrontendGroup(ctx, h)
	})
})
