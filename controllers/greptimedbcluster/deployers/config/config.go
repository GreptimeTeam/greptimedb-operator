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

package config

import (
	"embed"
)

//go:embed vector-config-template.yaml
var VectorConfigTemplate embed.FS

//go:embed logs-pipeline.yaml
var DefaultLogsPipeline embed.FS

//go:embed slow-queries-pipeline.yaml
var DefaultSlowQueriesPipeline embed.FS
