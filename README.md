# GreptimeDB Operator

## Overview

The GreptimeDB Operator manages [GreptimeDB](https://github.com/GrepTimeTeam/greptimedb) clusters on [Kubernetes](https://kubernetes.io/) by using [Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/).

The GreptimeDB operator abstract the model of maintaining the high available GreptimeDB cluster, you can create you own cluster as easy as possible:

```console
cat <<EOF | kubectl apply -f -
apiVersion: greptime.io/v1alpha1
kind: GreptimeDBCluster
metadata:
  name: basic
spec:
  base:
    main:
      image: greptime/greptimedb
  frontend:
    replicas: 1
  meta:
    replicas: 1
    etcdEndpoints:
      - "etcd.default:2379"
  datanode:
    replicas: 3
EOF
```

## Getting Started

### Prerequisites

- **Kubernetes 1.18 or higher version is required**

  You can use [kind](https://kind.sigs.k8s.io/) to create your own Kubernetes cluster:

  ```console
  kind create cluster
  ```

  If you want to deploy Kubernetes with local registry, you can use the following commands:

  ```console
  make kind-up
  ```

  It will create the cluster with 3 nodes and local registry.

- **kubectl**

  You can download the `kubectl` tool from the [page](https://kubernetes.io/docs/tasks/tools/).
  
- **Helm**

  You can follow the [guide](https://helm.sh/docs/intro/install/) to  install Helm.

### Quick start

You can use Helm chart of greptimedb-operator in [helm-charts](https://github.com/GreptimeTeam/helm-charts/blob/main/charts/greptimedb-operator/README.md) to start your operator quickly.

## Development

### About `make` targets

We can use `make` to handle most development, you can use the targets that list by the following command:

```console
make help
```

### Run operator on host

1. Install the CRDs:

   ```console
   make install
   ```

2. Run the operator on your host(make sure your Kubernetes is ready):

   ```console
   make run
   ```

### Deploy operator on self-managed Kubernetes

1. Build the image of operator

   ```console
   make docker-build-operator
   ```

   the default image URL is:

   ```console
   localhost:5001/greptime/greptimedb-operator:latest
   ```

   You can prefer your registry and tag:

   ```console
   make docker-build-operator IMAGE_REPO=<your-image-repo> IMAGE_TAG=<your-image-tag>
   ```

   **Note**: If you use the `IMAGE_REPO` or `IMAGE_TAG` in `make docker-build-operator`, you also have to use them again in the following command.

2. Push the image

   ```console
   make docker-push-operator
   ```

3. Deploy the operator in your self-managed Kubernetes

   ```console
   make deploy
   ```

   The operator will deploy in `greptimedb-admin` namespace:
   
   ```console
   kubectl get pod -n greptimedb-admin
   NAME                                   READY   STATUS    RESTARTS   AGE
   greptimedb-operator-7b4496c84d-bpwbm   1/1     Running   0          76s
   ```

   If you want to delete the deployment, you can:

   ```console
   make undeploy
   ```
   

If you want to build `greptimedb-initializer` which it's the init-container to process the initialization(for example: allocating node id for datanode), you can use the following commands that similar to above:

```console
# Building initializer.
make initializer

# Building initializer image, also can use IMAGE_REPO and IMAGE_NAME.
make docker-build-initializer

# Pushing initializer image, also can use IMAGE_REPO and IMAGE_NAME.
make docker-push-initializer
```

### Testing

1. Run unit test

   ```console
   make test
   ```

2. Run e2e test

   ```console
   make e2e
   ```

## License

greptimedb-operator uses the [Apache 2.0 license](./LICENSE) to strike a balance between
open contributions and allowing you to use the software however you want.
