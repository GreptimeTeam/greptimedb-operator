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

function dump_basic_info() {
  kubectl get pods -A -o wide
  kubectl get nodes
}

function dump_events() {
  kubectl get events --sort-by=.metadata.creationTimestamp -A
}

function dump_e2e_pods_details() {
  kubectl get pods -A | grep -E 'e2e|greptimedb-operator' | awk '{print $2 " " $1}' | while read -r line; do
    namespace=$(echo "$line" | awk '{print $2}')
    pod=$(echo "$line" | awk '{print $1}')
    echo "===> Describing pod $pod in namespace $namespace"
    kubectl describe pod "$pod" -n "$namespace"
    echo "===> Start dumping logs for pod $pod in namespace $namespace"
    kubectl logs "$pod" -n "$namespace"
    echo "<=== Finish dumping logs for pod $pod in namespace $namespace"
  done
}

function main() {
  dump_basic_info
  dump_events
  dump_e2e_pods_details
}

main
