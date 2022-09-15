#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

CLUSTER=greptimedb-operator-e2e
REGISTRY_NAME=kind-registry
REGISTRY_PORT=5001

TEST_ETCD_IMAGE=ghcr.io/greptimeteam/etcd:latest
TEST_META_IMAGE=ghcr.io/greptimeteam/meta-mock:latest
TEST_FRONTEND_IMAGE=ghcr.io/greptimeteam/frontend-mock:latest
TEST_GREPTIMEDB_IMAGE=ghcr.io/greptimeteam/db-test:latest

function check_prerequisites() {
    if ! hash docker 2>/dev/null; then
        echo "docker command is not found! You can download docker here: https://docs.docker.com/get-docker/"
        exit
    fi

    if ! hash kind 2>/dev/null; then
        echo "kind command is not found! You can download kind here: https://kind.sigs.k8s.io/docs/user/quick-start/#installing-from-release-binaries"
        exit
    fi

    if ! hash kubectl 2>/dev/null; then
        echo "kubectl command is not found! You can download kubectl here: https://kubernetes.io/docs/tasks/tools/"
        exit
    fi
}

function start_local_registry() {
    # create registry container unless it already exists
    if [ "$(docker inspect -f '{{.State.Running}}' "${REGISTRY_NAME}" 2>/dev/null || true)" != 'true' ]; then
        docker run \
        -d --restart=always -p "127.0.0.1:${REGISTRY_PORT}:5000" --name "${REGISTRY_NAME}" \
        registry:2
    fi
}

function create_kind_cluster() {
    # check cluster
    for cluster in $(kind get clusters); do
      if [ "$cluster" = "${CLUSTER}" ]; then
          echo "Use the existed cluster $cluster"
          kubectl config use-context kind-"$cluster"
          return
      fi
    done

    # create a cluster with the local registry enabled in containerd
    cat <<EOF | kind create cluster --name "${CLUSTER}" --config=-
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
}

function deploy_etcd() {
    kubectl apply -f ./hack/etcd/etcd-basic.yaml
    if ! kubectl rollout status --watch --timeout=180s statefulset/etcd; then
       echo "Deploy etcd failed"
       exit
    fi
    kubectl port-forward svc/etcd 2379:2379 &
}

function pull_images_and_push_to_local_registry() {
    docker pull "${TEST_ETCD_IMAGE}"
    docker tag "${TEST_ETCD_IMAGE}" localhost:5001/greptime/etcd:latest
    docker push localhost:5001/greptime/etcd:latest

    docker pull "${TEST_META_IMAGE}"
    docker tag "${TEST_META_IMAGE}" localhost:5001/greptime/meta:latest
    docker push localhost:5001/greptime/meta:latest

    docker pull "${TEST_FRONTEND_IMAGE}"
    docker tag "${TEST_FRONTEND_IMAGE}" localhost:5001/greptime/frontend:latest
    docker push localhost:5001/greptime/frontend:latest

    docker pull "${TEST_GREPTIMEDB_IMAGE}"
    docker tag "${TEST_GREPTIMEDB_IMAGE}" localhost:5001/greptime/greptimedb:latest
    docker push localhost:5001/greptime/greptimedb:latest
}

check_prerequisites
start_local_registry
create_kind_cluster
deploy_etcd
pull_images_and_push_to_local_registry
