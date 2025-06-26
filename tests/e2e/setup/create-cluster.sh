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
KUBERNETES_VERSION=${2:-"v1.32.3"}
TEST_NAMESPACE=default

# The argument for deploying local registry.
REGISTRY_NAME=kind-registry
REGISTRY_PORT=5001

# The argument for deploying etcd cluster.
ETCD_NAMESPACE=etcd-cluster
ETCD_CHART_VERSION=9.0.0

# The argument for deploying ingress nginx controller.
INGRESS_NGINX_CONTROLLER_NAMESPACE=ingress-nginx
INGRESS_NGINX_CONTROLLER_CHART_VERSION=4.12.0

# The argument for deploying Kafka cluster.
KAFKA_NAMESPACE=kafka
KAFKA_CLUSTER_NAME=kafka-wal
KAFKA_OPERATOR_VERSION=0.46.0

# The argument for deploying postgresql.
POSTGRESQL_NAMESPACE=postgresql
POSTGRESQL_CHART_VERSION=16.7.4
# The password for the default admin user `postgres`.
POSTGRESQL_ADMIN_PASSWORD=gt-operator-e2e
POSTGRESQL_DATABASE=metasrv

# The argument for deploying mysql.
MYSQL_NAMESPACE=mysql
MYSQL_CHART_VERSION=13.0.0
# The password for the default root user `root`.
MYSQL_ROOT_PASSWORD=gt-operator-e2e
MYSQL_DATABASE=metasrv

# The default timeout for waiting for resources to be ready.
DEFAULT_TIMEOUT=300s

# We always use the latest released greptimedb image for testing.
GREPTIMEDB_IMAGE=greptime/greptimedb:latest

# We always use the latest released vector image for testing.
VECTOR_IMAGE=timberio/vector:nightly-alpine

# Define the color for the output.
RED='\033[1;31m'
GREEN='\033[1;32m'
BLUE='\033[1;34m'
RESET='\033[0m'

function check_prerequisites() {
  echo -e "${GREEN}=> Check prerequisites...${RESET}"

  if ! hash docker 2>/dev/null; then
    echo -e "${RED}docker command is not found! You can download docker here: https://docs.docker.com/get-docker/${RESET}"
    exit
  fi

  if ! hash kind 2>/dev/null; then
    echo "${RED}kind command is not found! You can download kind here: https://kind.sigs.k8s.io/docs/user/quick-start/#installing-from-release-binaries${RESET}"
    exit
  fi

  if ! hash kubectl 2>/dev/null; then
    echo "${RED}kubectl command is not found! You can download kubectl here: https://kubernetes.io/docs/tasks/tools/${RESET}"
    exit
  fi

  if ! hash helm 2>/dev/null; then
    echo "${RED}helm command is not found! You can download helm here: https://helm.sh/docs/intro/install/${RESET}"
    exit
  fi

  echo -e "${GREEN}<= All prerequisites are met.${RESET}"
}

function start_local_registry() {
  echo -e "${GREEN}=> Start local registry...${RESET}"
  # create registry container unless it already exists
  if [ "$(docker inspect -f '{{.State.Running}}' "${REGISTRY_NAME}" 2>/dev/null || true)" != 'true' ]; then
    docker run -d --rm \
      -p "127.0.0.1:${REGISTRY_PORT}:5000" \
      --name "${REGISTRY_NAME}" \
      registry:2
  fi
  echo -e "${GREEN}<= Local registry is running.${RESET}"
}

function build_operator_image() {
  echo -e "${GREEN}=> Build and push operator image...${RESET}"
  make docker-build-operator
  make docker-push-operator
  echo -e "${GREEN}<= Operator image is built and pushed.${RESET}"
}

function build_initializer_image() {
  echo -e "${GREEN}=> Build and push initializer image...${RESET}"
  make docker-build-initializer
  make docker-push-initializer
  echo -e "${GREEN}<= Initializer image is built and pushed.${RESET}"
}

function pull_greptimedb_image() {
  echo -e "${GREEN}=> Pull and push greptimedb image...${RESET}"
  docker pull "$GREPTIMEDB_IMAGE"

  # Note: For testing convenience, we tag the greptimedb image to localhost:5001/greptimedb:latest.
  # After pushing the image to the local registry, we can use the image in the kind cluster as localhost:5001/greptimedb:latest.
  docker tag "$GREPTIMEDB_IMAGE" localhost:${REGISTRY_PORT}/greptime/greptimedb:latest
  docker push localhost:${REGISTRY_PORT}/greptime/greptimedb:latest
  echo -e "${GREEN}<= Greptimedb image is pulled and pushed.${RESET}"
}

function pull_vector_image() {
  echo -e "${GREEN}=> Pull and push vector image...${RESET}"
  docker pull "$VECTOR_IMAGE"
  docker tag "$VECTOR_IMAGE" localhost:${REGISTRY_PORT}/timberio/vector:nightly-alpine
  docker push localhost:${REGISTRY_PORT}/timberio/vector:nightly-alpine
  echo -e "${GREEN}<= Vector image is pulled and pushed.${RESET}"
}

function create_kind_cluster() {
  echo -e "${GREEN}=> Create kind cluster...${RESET}"
  # check cluster
  for cluster in $(kind get clusters); do
    if [ "$cluster" = "${CLUSTER_NAME}" ]; then
      echo "Use the existed cluster $cluster"
      kubectl config use-context kind-"$cluster"
      return
    fi
  done

  # create a cluster with the local registry enabled in containerd
  cat <<EOF | kind create cluster --name "${CLUSTER_NAME}" --image=kindest/node:"${KUBERNETES_VERSION}" --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
containerdConfigPatches:
- |-
  [plugins."io.containerd.grpc.v1.cri".registry.mirrors."localhost:${REGISTRY_PORT}"]
    endpoint = ["http://${REGISTRY_NAME}:5000"]
nodes:
- role: control-plane
- role: worker
- role: worker
EOF

  # connect the registry to the cluster network if not already connected
  if [ "$(docker inspect -f='{{json .NetworkSettings.Networks.kind}}' "${REGISTRY_NAME}")" = 'null' ]; then
    docker network connect "kind" "${REGISTRY_NAME}"
  fi

  # Document the local registry
  # https://github.com/kubernetes/enhancements/tree/master/keps/sig-cluster-lifecycle/generic/1755-communicating-a-local-registry
  cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: local-registry-hosting
  namespace: kube-public
data:
  localRegistryHosting.v1: |
    host: "localhost:${REGISTRY_PORT}"
    help: "https://kind.sigs.k8s.io/docs/user/local-registry/"
EOF

  echo -e "${GREEN}<= Kind cluster is created.${RESET}"
}

function deploy_greptimedb_operator() {
  echo -e "${GREEN}=> Deploy GreptimeDB operator...${RESET}"
  make deploy
  echo -e "${GREEN}<= GreptimeDB operator is deployed.${RESET}"
}

# Deploy etcd cluster that used for metasrv.
function deploy_etcd_cluster() {
  echo -e "${GREEN}=> Deploy etcd cluster...${RESET}"
  helm upgrade --install etcd oci://registry-1.docker.io/bitnamicharts/etcd \
    --set replicaCount=1 \
    --set auth.rbac.create=false \
    --set auth.rbac.token.enabled=false \
    --namespace "$ETCD_NAMESPACE" \
    --create-namespace \
    --version "$ETCD_CHART_VERSION"
  echo -e "${GREEN}<= etcd cluster is deployed.${RESET}"
}

# Deploy postgresql that used for meta backend testing.
function deploy_postgresql() {
  echo -e "${GREEN}=> Deploy postgresql...${RESET}"
  helm upgrade --install pg oci://registry-1.docker.io/bitnamicharts/postgresql \
    --set auth.postgresPassword="$POSTGRESQL_ADMIN_PASSWORD" \
    --set auth.database="$POSTGRESQL_DATABASE" \
    --namespace "$POSTGRESQL_NAMESPACE" \
    --create-namespace \
    --version "$POSTGRESQL_CHART_VERSION"

  # Create postgres-credentials secret.
  kubectl create secret generic postgresql-credentials \
    --namespace "$TEST_NAMESPACE" \
    --from-literal=username=postgres \
    --from-literal=password="$POSTGRESQL_ADMIN_PASSWORD"
}

# Deploy mysql that used for meta backend testing.
function deploy_mysql() {
  echo -e "${GREEN}=> Deploy mysql...${RESET}"
  helm upgrade --install mysql oci://registry-1.docker.io/bitnamicharts/mysql \
    --set auth.rootPassword="$MYSQL_ROOT_PASSWORD" \
    --set auth.database="$MYSQL_DATABASE" \
    --namespace "$MYSQL_NAMESPACE" \
    --create-namespace \
    --version "$MYSQL_CHART_VERSION"

  # Create mysql-credentials secret.
  kubectl create secret generic mysql-credentials \
    --namespace "$TEST_NAMESPACE" \
    --from-literal=username=root \
    --from-literal=password="$MYSQL_ROOT_PASSWORD"
}

# Deploy ingress nginx controller that used for frontend ingress testing.
function deploy_ingress_nginx_controller() {
  echo -e "${GREEN}=> Deploy ingress nginx controller...${RESET}"
  helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
  helm upgrade --install ingress-nginx ingress-nginx/ingress-nginx \
    --namespace "$INGRESS_NGINX_CONTROLLER_NAMESPACE" \
    --set controller.ingressClassResource.enabled=true \
    --set controller.ingressClassResource.name=nginx \
    --set controller.ingressClassResource.controllerValue=k8s.io/ingress-nginx \
    --set controller.ingressClass=nginx \
    --create-namespace \
    --version "$INGRESS_NGINX_CONTROLLER_CHART_VERSION"
  echo -e "${GREEN}<= Ingress nginx controller is deployed.${RESET}"
}

# Deploy Kafka cluster using Strimzi for remote WAL testing.
function deploy_kafka_cluster() {
  echo -e "${GREEN}=> Deploy Kafka cluster...${RESET}"
  kubectl create namespace "$KAFKA_NAMESPACE"
  kubectl create -f "https://github.com/strimzi/strimzi-kafka-operator/releases/download/${KAFKA_OPERATOR_VERSION}/strimzi-cluster-operator-${KAFKA_OPERATOR_VERSION}.yaml" -n "$KAFKA_NAMESPACE"

  # Wait for the CRDs to be created.
  kubectl wait \
    --for=condition=Established \
    crd/kafkabridges.kafka.strimzi.io \
    crd/kafkaconnects.kafka.strimzi.io \
    crd/kafkaconnectors.kafka.strimzi.io \
    crd/kafkanodepools.kafka.strimzi.io \
    crd/kafkarebalances.kafka.strimzi.io \
    crd/kafkas.kafka.strimzi.io \
    crd/kafkatopics.kafka.strimzi.io \
    crd/kafkausers.kafka.strimzi.io \
    crd/strimzipodsets.core.strimzi.io \
    --timeout="$DEFAULT_TIMEOUT"

  kubectl apply -f ./tests/e2e/setup/kafka-wal.yaml -n "$KAFKA_NAMESPACE"
  echo -e "${GREEN}<= Kafka cluster is deployed.${RESET}"
}

# Deploy cloud-provider-kind for using LoadBalancer service type.
function deploy_cloud_provider_kind() {
  if ! hash cloud-provider-kind 2>/dev/null; then
    go install sigs.k8s.io/cloud-provider-kind@latest
  fi

  echo -e "${GREEN}=> Deploy cloud-provider-kind...${RESET}"
  kubectl label node "${CLUSTER_NAME}-control-plane" node.kubernetes.io/exclude-from-external-load-balancers-

  # Check if the os is darwin or linux.
  # If the os is darwin, we need to use sudo to run cloud-provider-kind.
  nohup cloud-provider-kind > /tmp/cloud-provider-kind-logs 2>&1 &
  echo -e "${GREEN}<= cloud-provider-kind is deployed.${RESET}"
}

function check_kafka_cluster_status() {
  wait_for \
    "kubectl -n $KAFKA_NAMESPACE get kafka $KAFKA_CLUSTER_NAME -o json | jq '.status.conditions[] | select(.type==\"Ready\" and .status==\"True\")' | grep Ready" \
    "${DEFAULT_TIMEOUT%s}"
}

function wait_for() {
  local command="$1"
  local timeout="$2"
  local interval=1
  local elapsed_time=0

  while true; do
    # Turn off the exit-on-error flag to avoid exiting the script.
    set +e
    eval "$command" &> /dev/null
    local status=$?
    set -e

    # Check if command was successful.
    if [[ $status -eq 0 ]]; then
      return 0
    fi

    # Update elapsed time.
    ((elapsed_time += interval))

    # Check if the timeout has been reached.
    if [[ $elapsed_time -ge $timeout ]]; then
      echo "Timeout reached. Command failed."
      return 1
    fi

    # Wait for a specified interval before retrying.
    sleep $interval
  done
}

function wait_all_service_ready() {
  echo -e "${GREEN}=> Wait for all services to be ready...${RESET}"

  # Wait for etcd to be ready.
  kubectl wait \
    --for=condition=Ready \
    pod -l app.kubernetes.io/instance=etcd \
    -n "$ETCD_NAMESPACE" \
    --timeout="$DEFAULT_TIMEOUT"

  # Wait for postgresql to be ready.
  kubectl wait \
    --for=condition=Ready \
    pod -l app.kubernetes.io/instance=pg \
    -n "$POSTGRESQL_NAMESPACE" \
    --timeout="$DEFAULT_TIMEOUT"

  # Wait for mysql to be ready.
  kubectl wait \
    --for=condition=Ready \
    pod -l app.kubernetes.io/instance=mysql \
    -n "$MYSQL_NAMESPACE" \
    --timeout="$DEFAULT_TIMEOUT"

  # Wait for ingress nginx controller to be ready.
  kubectl wait \
    --for=condition=Ready \
    pod -l app.kubernetes.io/instance=ingress-nginx \
    -n "$INGRESS_NGINX_CONTROLLER_NAMESPACE" \
    --timeout="$DEFAULT_TIMEOUT"

  # Wait for kafka to be ready.
  check_kafka_cluster_status

  # Wait for greptimedb-operator to be ready.
  kubectl rollout \
    status deployment/greptimedb-operator \
    -n greptimedb-admin \
    --timeout="$DEFAULT_TIMEOUT"

  echo -e "${GREEN}<= All services are ready.${RESET}"
}

function main() {
  echo -e "${GREEN}=> Start setting up the e2e environment${RESET}"
  echo -e "${BLUE}Cluster: ${CLUSTER_NAME}${RESET}"
  echo -e "${BLUE}Kubernetes: ${KUBERNETES_VERSION}${RESET}"
  echo -e "${BLUE}GreptimeDB: ${GREPTIMEDB_IMAGE}${RESET}"

  check_prerequisites
  start_local_registry
  build_operator_image
  build_initializer_image
  pull_greptimedb_image
  pull_vector_image
  create_kind_cluster
  deploy_cloud_provider_kind
  deploy_greptimedb_operator
  deploy_etcd_cluster
  deploy_postgresql
  deploy_mysql
  deploy_ingress_nginx_controller
  deploy_kafka_cluster
  wait_all_service_ready
}

main "$@"
