#!/usr/bin/env bash

# The Go module path.
GO_MODULE=github.com/GreptimeTeam/greptimedb-operator

# The output generated client directory(relative to the Go module path).
OUTPUT_DIR=pkg/client

# The boilerplate file path.
BOILERPLATE_FILE=hack/boilerplate.go.txt

# The Go bin directory which to download the *-gen binaries.
GOBIN=$(pwd)/bin

CODE_GENERATOR_VERSION=v0.30.0

# If the kube_codegen.sh is already in ${GOBIN}, skip the download.
if [ ! -f ${GOBIN}/kube_codegen.sh ]; then
  # Download kube_codegen.sh to ${GOBIN}
  curl -sSL https://raw.githubusercontent.com/kubernetes/code-generator/refs/tags/${CODE_GENERATOR_VERSION}/kube_codegen.sh -o ${GOBIN}/kube_codegen.sh
fi

source ${GOBIN}/kube_codegen.sh

BACKUP_DIR=$(mktemp -d)

echo "Backup the original client directory to ${BACKUP_DIR}"

# Backup the original client directory.
mv $(pwd)/${OUTPUT_DIR} ${BACKUP_DIR}

GOBIN=${GOBIN} kube::codegen::gen_client \
  $(pwd) \
  --output-pkg ${GO_MODULE}/${OUTPUT_DIR} \
  --output-dir $(pwd)/${OUTPUT_DIR} \
  --clientset-name clientset \
  --versioned-name versioned \
  --listers-name listers \
  --applyconfig-name applyconfiguration \
  --boilerplate ${BOILERPLATE_FILE} \
  --with-watch

# Recover the original client directory if the generation failed.
if [ $? -ne 0 ]; then
  echo "Client generation failed, recovering the original client directory."
  mv ${BACKUP_DIR}/client $(pwd)/${OUTPUT_DIR}
fi
