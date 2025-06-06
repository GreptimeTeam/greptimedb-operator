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

package constant

const (
	GreptimeDBComponentName = "app.greptime.io/component"
	GreptimeDBConfigDir     = "/etc/greptimedb"
	GreptimeDBTLSDir        = "/etc/greptimedb/tls"

	// GreptimeDBInitConfigDir used for greptimedb-initializer.
	GreptimeDBInitConfigDir = "/etc/greptimedb-init"

	GreptimeDBConfigFileName = "config.toml"
	MainContainerIndex       = 0
	ConfigVolumeName         = "config"
	InitConfigVolumeName     = "init-config"
	TLSVolumeName            = "tls"
	DefaultTLSMode           = "prefer"

	// LogsTableName is the table name of storing greptimedb logs.
	LogsTableName = "_gt_logs"

	// DefaultVectorConfigName is the default name of vector config.
	DefaultVectorConfigName = "vector-config"

	// DefaultLogsVolumeName is the default name of logs volume.
	DefaultLogsVolumeName = "logs"
)
