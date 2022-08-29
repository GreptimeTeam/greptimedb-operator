#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

CLUSTER=kind-greptimedb-playground

function check_prerequisites() {
    if ! hash docker 2>/dev/null; then
        echo "❌ docker command is not found! You can download docker here: https://docs.docker.com/get-docker/"
        exit
    fi

    if ! hash kind 2>/dev/null; then
        echo "❌ kind command is not found! You can download kind here: https://kind.sigs.k8s.io/docs/user/quick-start/#installing-from-release-binaries"
        exit
    fi

    if ! hash kubectl 2>/dev/null; then
        echo "❌ kubectl command is not found! You can download kubectl here: https://kubernetes.io/docs/tasks/tools/"
        exit
    fi
}

function create_local_kubernetes() {
    if kubectl cluster-info > /dev/null 2>&1; then
        current_cluster=$(kubectl config current-context)
        if [ "$current_cluster" = "${CLUSTER}" ]; then
            echo "🚀 Use the existed cluster ${current_cluster}"
            return
        fi
    fi

    echo "🚀 Creating local Kubernetes by kind..."
    ./hack/kind/3-nodes-with-local-registry.sh
}

function process_meta_image() {
    docker pull grygt/meta:latest
    docker tag grygt/meta:latest localhost:5001/greptime/meta:latest
    docker push localhost:5001/greptime/meta:latest
}

function process_etcd_image() {
    docker pull grygt/etcd:latest
    docker tag grygt/etcd:latest localhost:5001/greptime/etcd:latest
    docker push localhost:5001/greptime/etcd:latest
}

function process_frontend_image() {
    docker pull grygt/frontend:latest
    docker tag grygt/frontend:latest localhost:5001/greptime/frontend:latest
    docker push localhost:5001/greptime/frontend:latest
}

function process_db_image() {
    docker pull grygt/db:latest
    docker tag grygt/db:latest localhost:5001/greptime/greptimedb:latest
    docker push localhost:5001/greptime/greptimedb:latest
}

function process_operator_image() {
    docker pull grygt/operator:latest
    docker tag grygt/operator:latest localhost:5001/greptime/greptimedb-operator:latest
    docker push localhost:5001/greptime/greptimedb-operator:latest
}

function process_images() {
    process_meta_image
    process_etcd_image
    process_frontend_image
    process_db_image
    process_operator_image
}

function deploy_greptime_operator() {
    kubectl apply -f ./manifests/greptimedb-operator-deployment.yaml
    if ! kubectl wait deployment -n greptimedb-operator-system greptimedb-operator --for condition=Available=True --timeout=90s; then
       echo "❌ Deploy greptimedb-operator failed, please check your system."
       exit
    fi
}

function deploy_greptimedb_cluster() {
    kubectl apply -f ./config/samples/mock/cluster.yaml

    if ! kubectl wait greptimedbclusters.greptime.cloud -n default mock --for condition=Ready=True --timeout=90s; then
       echo "❌ Deploy greptimedb cluster failed, please check your system."
       exit
    fi
}

function forward_mysql_request() {
    kubectl port-forward svc/mock-frontend 3306:3306
}

# The entrypoint.
echo "💼 Checking prerequisites..."
check_prerequisites
echo "✅ You already have everything needed to bootstrap the cluster"

create_local_kubernetes

echo "🚀 Downloading images..."
process_images 1> /dev/null
echo "✅ Finish to download images"

echo "🚀 Deploying greptimedb-operator..."
deploy_greptime_operator 1> /dev/null
echo "✅ Finish to deploy greptimedb-operator"

echo "🚀 Deploying greptimedb cluster..."
deploy_greptimedb_cluster 1> /dev/null
echo "✅ Finish to deploy greptimedb cluster"

echo "🎉 Your first GreptimeDB cluster is UP!"

echo "🚀 Connect GreptimeDB with MySQL client: mysql -h 127.0.0.1 -P 3306"
forward_mysql_request
