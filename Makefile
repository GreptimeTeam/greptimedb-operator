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

# Build the container image of greptimedb-operator.
IMAGE_REGISTRY ?= localhost:5001
IMAGE_REPO ?= ${IMAGE_REGISTRY}/greptime
IMAGE_TAG ?= latest
DOCKER_BUILD_OPTIONS ?= --network host
OPERATOR_DOCKERFILE = ./docker/operator/Dockerfile
INITIALIZER_DOCKERFILE = ./docker/initializer/Dockerfile

ARCH := $(shell uname -m)
ifeq ($(ARCH),x86_64)
	ARCH := amd64
endif

MANIFESTS_DIR = ./manifests

# Use the kubernetes version to run the tests.
KUBERNETES_VERSION = 1.32.0

# Arguments for running the e2e.
E2E_CLUSTER_NAME ?= greptimedb-operator-e2e
E2E_TIMEOUT ?= 10m

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

LDFLAGS = $(shell ./hack/version.sh)

# Setting SHELL to bash allows bash commands to be executed by recipes.
# This is a requirement for 'setup-envtest.sh' in the test target.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

.PHONY: all
all: build

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-31s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: manifests
manifests: kustomize controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) rbac:roleName=greptimedb-operator-role crd:maxDescLen=0 webhook paths="./..." output:crd:artifacts:config=config/crd/resources
	$(KUSTOMIZE) build config/crd > ${MANIFESTS_DIR}/crds.yaml
	$(KUSTOMIZE) build config/default > ${MANIFESTS_DIR}/bundle.yaml

.PHONY: generate
generate: kustomize controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: generate-client
generate-client: ## Generate the pkg/client/{clientset,informers,listers} code.
	./hack/client/generate-client.sh

.PHONY: check-code-generation
check-code-generation: ## Check code generation.
	echo "Checking code generation"
	./hack/check-code-generation.sh

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: create-e2e-cluster
create-e2e-cluster: ## Create a kind cluster for e2e tests.
	./tests/e2e/setup/create-cluster.sh ${E2E_CLUSTER_NAME} v${KUBERNETES_VERSION}

.PHONY: e2e
e2e: create-e2e-cluster ## Run e2e tests.
	go test -timeout ${E2E_TIMEOUT} -v ./tests/e2e/... || (./tests/e2e/setup/diagnostic-cluster.sh ; exit 1)

.PHONY: lint
lint: golangci-lint ## Run lint.
	$(GOLANGCI_LINT) run -v ./...

.PHONY: test
test: manifests generate fmt vet envtest ## Run tests.
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(KUBERNETES_VERSION) -p path)" go test \
	./controllers/...     \
	./apis/...            \
	./pkg/...             \
	./cmd/initializer/... \
	-coverprofile cover.out

.PHONY: check-api-docs
check-api-docs: api-docs ## Check docs
	@git diff --quiet || \
    (echo "Need to update documentation, please run 'make api-docs'"; \
	exit 1)

.PHONY: check-client-generation
check-client-generation: generate-client ## Check client generation.
	@git diff --quiet || \
    (echo "Need to update client code, please run 'make generate-client'"; \
	exit 1)

.PHONY: kind-up
kind-up: ## Create the kind cluster for developing.
	./hack/kind/3-nodes-with-local-registry.sh

##@ Build

.PHONY: build
build: generate fmt vet ## Build greptimedb-operator binary.
	GO111MODULE=on CGO_ENABLED=0 go build -ldflags '${LDFLAGS}' -o bin/greptimedb-operator ./cmd/operator/main.go

.PHONY: fast-build
fast-build: ## Build greptimedb-operator binary only.
	GO111MODULE=on CGO_ENABLED=0 go build -ldflags '${LDFLAGS}' -o bin/greptimedb-operator ./cmd/operator/main.go

.PHONY: build-operator-for-linux
build-operator-for-linux: ## Build greptimedb-operator binary for linux.
	GO111MODULE=on CGO_ENABLED=0 GOARCH=$(ARCH) GOOS=linux go build -ldflags '${LDFLAGS}' -o bin/greptimedb-operator ./cmd/operator/main.go

.PHONY: build-initializer-for-linux
build-initializer-for-linux: ## Build greptimedb-initializer binary for linux.
	GO111MODULE=on CGO_ENABLED=0 GOARCH=$(ARCH) GOOS=linux go build -ldflags '${LDFLAGS}' -o bin/greptimedb-initializer ./cmd/initializer/main.go

.PHONY: initializer
initializer: ## Build greptimedb-initializer binary.
	GO111MODULE=on CGO_ENABLED=0 go build -ldflags '${LDFLAGS}' -o bin/greptimedb-initializer ./cmd/initializer/main.go

.PHONY: run
run: manifests generate fmt vet ## Run a controller from your host.
	GO111MODULE=on CGO_ENABLED=0 go run -ldflags '${LDFLAGS}' ./cmd/operator/main.go --enable-leader-election=false

.PHONY: docker-build-operator
docker-build-operator: ## Build docker image with the greptimedb-operator.
	docker build ${DOCKER_BUILD_OPTIONS} -f ${OPERATOR_DOCKERFILE} -t ${IMAGE_REPO}/greptimedb-operator:${IMAGE_TAG} .

.PHONY: docker-build-initializer
docker-build-initializer: ## Build docker image with the greptimedb-initializer.
	docker build ${DOCKER_BUILD_OPTIONS} -f ${INITIALIZER_DOCKERFILE} -t ${IMAGE_REPO}/greptimedb-initializer:${IMAGE_TAG} .

.PHONY: docker-push-operator
docker-push-operator: ## Push docker image with the greptimedb-operator.
	docker push ${IMAGE_REPO}/greptimedb-operator:${IMAGE_TAG}

.PHONY: docker-push-initializer
docker-push-initializer: ## Push docker image with the greptimedb-initializer.
	docker push ${IMAGE_REPO}/greptimedb-initializer:${IMAGE_TAG}

.PHONY: build-operator-local-test-image
build-operator-local-test-image: build-operator-for-linux ## Build the docker image for the greptimedb-operator locally for testing.
	@cp bin/greptimedb-operator greptimedb-operator # Copy the binary to the root directory because the bin directory is ignored in the .dockerignore file.
	docker build -t ${IMAGE_REPO}/greptimedb-operator:${IMAGE_TAG} . -f docker/operator/Dockerfile.local
	@rm greptimedb-operator

.PHONY: build-initializer-local-test-image
build-initializer-local-test-image: build-initializer-for-linux ## Build the docker image for the greptimedb-initializer locally for testing.
	@cp bin/greptimedb-initializer greptimedb-initializer # Copy the binary to the root directory because the bin directory is ignored in the .dockerignore file.
	docker build -t ${IMAGE_REPO}/greptimedb-initializer:${IMAGE_TAG} . -f docker/initializer/Dockerfile.local
	@rm greptimedb-initializer

.PHONY: push-operator-local-test-image
push-operator-local-test-image: ## Push the docker image for the greptimedb-operator locally for testing.
	docker push ${IMAGE_REPO}/greptimedb-operator:${IMAGE_TAG}

.PHONY: push-initializer-local-test-image
push-initializer-local-test-image: ## Push the docker image for the greptimedb-initializer locally for testing.
	docker push ${IMAGE_REPO}/greptimedb-initializer:${IMAGE_TAG}

##@ Deployment

ifndef ignore-not-found
  ignore-not-found = false
endif

.PHONY: install
install: manifests kustomize ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | kubectl apply -f - --server-side

.PHONY: uninstall
uninstall: manifests kustomize ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/crd | kubectl delete --ignore-not-found=$(ignore-not-found) -f -

.PHONY: deploy
deploy: manifests kustomize ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMAGE_REPO}/greptimedb-operator:${IMAGE_TAG}
	$(KUSTOMIZE) build config/default | kubectl apply -f - --server-side

.PHONY: undeploy
undeploy: ## Undeploy controller from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/default | kubectl delete --ignore-not-found=$(ignore-not-found) -f -

##@ Documentation

.PHONY: api-docs
api-docs: crd-ref-docs ## Generate api references docs.
	$(CRD_REF_DOCS) \
		--source-path=./apis \
		--renderer=markdown \
		--output-path=./docs/api-references/docs.md \
		--templates-dir=./docs/api-references/template/ \
		--config=./docs/api-references/config.yaml

##@ Build Dependencies

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
KUSTOMIZE ?= $(LOCALBIN)/kustomize
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
ENVTEST ?= $(LOCALBIN)/setup-envtest
GOLANGCI_LINT ?= $(LOCALBIN)/golangci-lint
CRD_REF_DOCS ?= $(LOCALBIN)/crd-ref-docs

## Tool Versions
KUSTOMIZE_VERSION ?= v5.5.0
CONTROLLER_TOOLS_VERSION ?= v0.17.3
CRD_REF_DOCS_VERSION ?= v0.1.0
GOLANGCI_LINT_VERSION ?= v2.1.6

.PHONY: kustomize
kustomize: $(KUSTOMIZE) ## Download kustomize locally if necessary.
$(KUSTOMIZE): $(LOCALBIN)
	GOBIN=$(LOCALBIN) GO111MODULE=on go install sigs.k8s.io/kustomize/kustomize/v5@${KUSTOMIZE_VERSION}

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary.
$(CONTROLLER_GEN): $(LOCALBIN)
	GOBIN=$(LOCALBIN) GO111MODULE=on go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)

.PHONY: envtest
envtest: $(ENVTEST) ## Download envtest-setup locally if necessary.
$(ENVTEST): $(LOCALBIN)
	GOBIN=$(LOCALBIN) GO111MODULE=on go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest

.PHONY: golangci-lint
golangci-lint: ## Install golangci-lint.
	GOBIN=$(LOCALBIN) GO111MODULE=on go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@${GOLANGCI_LINT_VERSION}

.PHONY: crd-ref-docs
crd-ref-docs: ## Install crd-ref-docs.
	GOBIN=$(LOCALBIN) GO111MODULE=on go install github.com/elastic/crd-ref-docs@$(CRD_REF_DOCS_VERSION)
