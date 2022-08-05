# GreptimeDB Operator

## Overview

The GreptimeDB Operator manages GreptimeDB clusters on [Kubernetes](https://kubernetes.io/) by using [Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/).

## Getting Started

### Prerequisites

- Kubernetes 1.18 or higher version is required

  You can use [kind](https://kind.sigs.k8s.io/) to create your own Kubernetes cluster:

  ```bash
  $ kind create cluster
  ```

- [kubectl](https://kubernetes.io/docs/tasks/tools/)

- Go 1.18 or higher version is required

### Usages

- Install the CRDs

  ```bash
  $ make install
  ```
  
- Run the operator locally

  ```bash
  $ make run
  ```
  
- Manager the basic greptimedb cluster

  ```bash
  # Create the cluster
  $ kubectl apply -f ./config/samples/basic/cluster.yaml
  
  # Delete the cluster
  $ kubectl delete -f ./config/samples/basic/cluster.yaml
  ```