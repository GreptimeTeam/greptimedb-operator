#!/usr/bin/env bash

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
