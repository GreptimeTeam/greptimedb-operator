#!/usr/bin/env bash
# Copyright 2022 Greptime Team
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

TMP_DIR=/tmp/greptimedb-operator-tmp
GENERATED_CODE_DIRS=(config manifests apis)

function copy_original_files() {
    rm -rf "${TMP_DIR}"
    mkdir -p "${TMP_DIR}"
    for dir in "${GENERATED_CODE_DIRS[@]}"; do
        cp -r "${dir}" "${TMP_DIR}"/"${dir}"
    done
}

function run_code_generation() {
    make manifests
    make generate
}

function check_diff() {
    for dir in "${GENERATED_CODE_DIRS[@]}"; do
        diff -Naupr "${dir}" "${TMP_DIR}"/"${dir}" || (echo \'"${dir}"/\' is out of date, please run \'make manifests\' and \'make generate\' && exit 1)
    done
}

copy_original_files
run_code_generation
check_diff
