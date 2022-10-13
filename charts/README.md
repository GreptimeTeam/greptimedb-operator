# Charts

This directory contains a collection of Helm charts to deploy GreptimeDB cluster.

- [greptimedb-operator](./greptimedb-operator/README.md): A Helm chart for GreptimeDB Operator.
- [greptimedb](./greptimedb/README.md): A Helm chart for GreptimeDB cluster.
- [etcd](./etcd/README.md): A Helm chart for etcd cluster(**NOT** production ready).

## Prerequisites

- [Helm](https://helm.sh/docs/intro/install/)

## Getting Started

If you want to deploy GreptimeDB cluster, you can use the following command:

1. Deploy GreptimeDB operator

   ```
   # Deploy greptimedb-operator in new namespace.
   $ helm install gtcloud greptimedb-operator --namespace greptimedb-operator-system --create-namespace
   ```

2. Deploy GreptimeDB cluster

   ```
   $ helm install mydb greptimedb
   ```
   
3. Use kubectl port-forward to access GreptimeDB cluster

   ```
   $ kubectl port-forward svc/mydb-frontend 3306:3306
   ```

You also can list the current releases by `helm` command:

```
$ helm list --all-namespaces
```

If you want to terminate the GreptimeDB cluster, you can use the following command:

```
$ helm uninstall mydb
$ helm uninstall gtcloud -n greptimedb-operator-system
```

The CRDs of GreptimeDB are not deleted [by default](https://helm.sh/docs/topics/charts/#limitations-on-crds). You can delete them by the following command:

```
$ kubectl delete crds greptimedbclusters.greptime.io
```
