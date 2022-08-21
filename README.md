# GreptimeDB Operator

## Overview

The GreptimeDB Operator manages GreptimeDB clusters on [Kubernetes](https://kubernetes.io/) by using [Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/).

## Getting Started

### Prerequisites

- Kubernetes 1.18 or higher version is required

  You can use [kind](https://kind.sigs.k8s.io/) to create your own Kubernetes cluster:

  ```
  $ kind create cluster
  ```

  If you want to deploy Kubernetes with local registry, you can use the following commands (the scripts is modified from [here](https://kind.sigs.k8s.io/docs/user/local-registry/)):

  ```
  $ ./greptimedb-operator/hack/kind/3-nodes-with-local-registry.sh
  ```

- [kubectl](https://kubernetes.io/docs/tasks/tools/)

- Go 1.18 or higher version is required

### Usages

- Install the CRDs

  ```
  $ make install
  ```
  
- Run the operator locally

  ```
  $ make run
  ```
  
- Manager the basic greptimedb cluster

  ```
  # Create the cluster
  $ kubectl apply -f ./config/samples/basic/cluster.yaml
  
  # Delete the cluster
  $ kubectl delete -f ./config/samples/basic/cluster.yaml
  ```

- Deploy the basic cluster of one Datanode

  ```
  # Deploy the cluster that only has one Datanode.
  $ kubectl apply -f ./config/samples/basic-datanode/cluster.yaml

  # Port forward the service to your host.
  $ kubectl port-forward svc/basic-datanode 3306:3306

  # Use mysql client to connect service.
  $ mysql -h 127.0.0.1 -P 3306
  ```

  After connecting to the cluster, you can [run your own SQL](https://github.com/GreptimeTeam/greptimedb).

  Make sure your [greptimedb](https://github.com/GreptimeTeam/greptimedb) image is already in your local registry of kind, you can push your greptimedb image:

  ```
  # Build image in greptimedb repo.
  $ docker build --network host -f docker/Dockerfile -t localhost:5001/greptimedb .

  # Push the image to local registry
  $ docker push localhost:5001/greptimedb
  ```
