#!/usr/bin/env bash

set -euo pipefail

# The Go module path.
GO_MODULE=github.com/GreptimeTeam/greptimedb-operator

# The output generated client directory(relative to the Go module path).
OUTPUT_DIR=pkg/client

# The boilerplate file path.
BOILERPLATE_FILE=hack/boilerplate.go.txt

# The Go bin directory which to download the *-gen binaries.
GOBIN=$(pwd)/bin

# The code generator version.
CODE_GENERATOR_VERSION=v0.30.0

# The local path of the kube_codegen.sh.
KUBE_CODEGEN_SCRIPT="${GOBIN}/kube_codegen.sh"

# Create the Go bin directory if not exists.
if [ ! -d "${GOBIN}" ]; then
  mkdir -p "${GOBIN}"
fi

# Download kube_codegen.sh if not present
if [ ! -f "${KUBE_CODEGEN_SCRIPT}" ]; then
  echo "Downloading kube_codegen.sh..."
  curl -sSL "https://raw.githubusercontent.com/kubernetes/code-generator/refs/tags/${CODE_GENERATOR_VERSION}/kube_codegen.sh" -o "${KUBE_CODEGEN_SCRIPT}"
  chmod +x "${KUBE_CODEGEN_SCRIPT}"
fi

# Source the kube_codegen.sh for using the `kube::codegen::*` functions.
source "${KUBE_CODEGEN_SCRIPT}"

# Create backup of existing client directory
BACKUP_DIR=$(mktemp -d)
echo "Backing up original client directory to ${BACKUP_DIR}"

if [ -d "$(pwd)/${OUTPUT_DIR}" ]; then
  mv "$(pwd)/${OUTPUT_DIR}" "${BACKUP_DIR}"
fi

echo "Generating client code..."
GOBIN=${GOBIN} kube::codegen::gen_client \
  "$(pwd)" \
  --output-pkg "${GO_MODULE}/${OUTPUT_DIR}" \
  --output-dir "$(pwd)/${OUTPUT_DIR}" \
  --clientset-name clientset \
  --versioned-name versioned \
  --listers-name listers \
  --applyconfig-name applyconfiguration \
  --boilerplate "${BOILERPLATE_FILE}" \
  --with-watch

# Recover the original client directory if the generation failed.
if [ $? -ne 0 ]; then
  echo "Client generation failed, recovering the original client directory..."
  mv "${BACKUP_DIR}/client" "$(pwd)/${OUTPUT_DIR}"
  exit 1
fi

echo "Client generation completed successfully"
