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

package dbconfig

import (
	"fmt"

	"github.com/GreptimeTeam/greptimedb-operator/apis/v1alpha1"
	"github.com/GreptimeTeam/greptimedb-operator/pkg/dbconfig/datanode"
	"github.com/GreptimeTeam/greptimedb-operator/pkg/dbconfig/frontend"
	"github.com/GreptimeTeam/greptimedb-operator/pkg/dbconfig/metasrv"
)

// FromClusterCRD is the factory method to create dbconfig from GreptimeDBCluster CRD.
func FromClusterCRD(cluster *v1alpha1.GreptimeDBCluster, componentKind v1alpha1.ComponentKind) ([]byte, error) {
	if componentKind == v1alpha1.MetaComponentKind {
		metasrvConfig, err := metasrv.FromClusterCRD(cluster)
		if err != nil {
			return nil, err
		}
		return metasrvConfig.Marshal()
	}

	if componentKind == v1alpha1.FrontendComponentKind {
		frontendConfig, err := frontend.FromClusterCRD(cluster)
		if err != nil {
			return nil, err
		}
		return frontendConfig.Marshal()
	}

	if componentKind == v1alpha1.DatanodeComponentKind {
		datanodeConfig, err := datanode.FromClusterCRD(cluster)
		if err != nil {
			return nil, err
		}
		return datanodeConfig.Marshal()
	}

	return nil, fmt.Errorf("unknown component kind: %v", componentKind)
}
