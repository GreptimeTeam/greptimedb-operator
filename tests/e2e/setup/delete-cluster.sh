#!/usr/bin/env bash
# Copyright 2024 Greptime Team
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -o errexit
set -o nounset
set -o pipefail

CLUSTER_NAME=${1:-"greptimedb-operator-e2e"}

function delete_kind_cluster() {
  kind delete cluster --name "$CLUSTER_NAME"
}

function destroy_cloud_kind_provider() {
  pkill -9 cloud-provider-kind
}

function stop_registry() {
  docker stop kind-registry
  docker rm kind-registry
}

function main() {
  delete_kind_cluster
  destroy_cloud_kind_provider
  stop_registry
}

main "$@"
