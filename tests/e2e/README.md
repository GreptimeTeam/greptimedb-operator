# GreptimeDB Operator E2E

## Overview

`tests/e2e` contains end-to-end tests for the GreptimeDB Operator. It tests the function of the greptimedb-operator from the **user's perspective**.

## How it works

The E2E needs to be run in a Kubernetes cluster. It will use the [kind](https://kind.sigs.k8s.io/) to create a K8s and deploy the greptimedb-operator and other dependencies.

The following scripts are used to run the E2E:

- `tests/e2e/setup/create-cluster.sh`
   - Create a Kubernetes cluster using kind;
   - Pull the latest release GreptimeDB image;
   - Build the `greptimedb-operator` and `greptimedb-initializer` image and deploy it to the cluster;
   - Deploy the dependencies (e.g., etcd, Kafka, etc);
   
- `tests/e2e/setup/delete-cluster.sh`: Destroy the E2E cluster;
- `tests/e2e/setup/diagnostic-cluster.sh`: Output the cluster information when the E2E fails;

When the cluster is ready, Ginkgo will run the E2E tests. For each scenario(`tests/e2e/testdata`), the E2E will deploy the resources, run the related tests, check the results, and clean up the resources.

## How to run

In the root directory of the project, run the following commands:

```console
make e2e
```
