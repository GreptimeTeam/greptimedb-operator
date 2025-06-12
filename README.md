# GreptimeDB Operator

[![license](https://img.shields.io/github/license/GreptimeTeam/greptimedb-operator)](https://github.com/GreptimeTeam/greptimedb-operator/blob/main/LICENSE)
[![report](https://goreportcard.com/badge/github.com/GreptimeTeam/greptimedb-operator)](https://goreportcard.com/report/github.com/GreptimeTeam/greptimedb-operator)
[![GitHub release](https://img.shields.io/github/tag/GreptimeTeam/greptimedb-operator.svg?label=release)](https://github.com/GreptimeTeam/greptimedb-operator/releases)
[![GoDoc](https://img.shields.io/badge/Godoc-reference-blue.svg)](https://godoc.org/github.com/GreptimeTeam/greptimedb-operator)
[![Docker Pulls](https://img.shields.io/docker/pulls/greptime/greptimedb-operator)](https://hub.docker.com/r/greptime/greptimedb-operator)

## Overview

The GreptimeDB Operator manages the [GreptimeDB](https://github.com/GrepTimeTeam/greptimedb) resources on [Kubernetes](https://kubernetes.io/) using the [Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/). It is like an autopilot that automates the deployment, provisioning, and orchestration of the GreptimeDB cluster and standalone.

## Features

The GreptimeDB Operator includes, but is not limited to, the following features:

- **Automated Provisioning**

  Automates the deployment of the GreptimeDB cluster and standalone on Kubernetes by providing CRD `GreptimeDBCluster` and `GreptimeDBStandalone`.

- **Multi-Cloud Support**

  Users can deploy the GreptimeDB on any Kubernetes cluster, including on-premises and cloud environments(like AWS, GCP, Aliyun etc.).

- **Scaling**

  Scale the GreptimeDB cluster as easily as changing the `replicas` field in the `GreptimeDBCluster` CR.

- **Monitoring Bootstrap**

  Bootstrap the GreptimeDB monitoring stack for the GreptimeDB cluster by providing the `monitoring` field in the `GreptimeDBCluster` CR.

## Prerequisites

The GreptimeDB Operator requires at least Kubernetes `1.18.0`.

## Compatibility Matrix

| GreptimeDB Operator | API group/version      | Supported GreptimeDB version |
|---------------------|------------------------|------------------------------|
| < v0.2.0            | `greptime.io/v1alpha1` | < v0.12.0                    |
| ≥ v0.2.0            | `greptime.io/v1alpha1` | ≥ v0.12.0                    |

## Quick Start

The fastest way to install the GreptimeDB Operator is to use `bundle.yaml`:

```console
kubectl apply -f \
  https://github.com/GreptimeTeam/greptimedb-operator/releases/latest/download/bundle.yaml \
  --server-side 
```

The `greptimedb-operator` will be installed in the `greptimedb-admin` namespace. When the `greptimedb-operator` is running, you can see the following output:

```console
$ kubectl get pods -n greptimedb-admin
NAME                                   READY   STATUS    RESTARTS   AGE
greptimedb-operator-7947d785b5-b668p   1/1     Running   0          2m18s
```

Once the operator is running, you can experience the GreptimeDB by creating a basic standalone instance:

```console
cat <<EOF | kubectl apply -f -
apiVersion: greptime.io/v1alpha1
kind: GreptimeDBStandalone
metadata:
  name: basic
spec:
  base:
    main:
      image: greptime/greptimedb:latest
EOF
```

When the standalone is running, you can see the following output:

```console
$ kubectl get greptimedbstandalones basic
NAME    PHASE     VERSION   AGE
basic   Running   latest    75s
```

You can use `kubectl port-forward` to access the GreptimeDB:

```console
kubectl port-forward svc/basic-standalone 4001:4001 4002:4002 4003:4003 4000:4000
```

Please refer to the [quick-start](https://docs.greptime.com/getting-started/quick-start) to try more examples.

## Examples

The GreptimeDB Operator provides a set of examples to help you understand how to use the GreptimeDB Operator. You can find the examples in the [examples](./examples/README.md) directory.

## Deployment

For production use, we recommend deploying the GreptimeDB Operator with the GreptimeDB official Helm [chart](https://github.com/GreptimeTeam/helm-charts). 

## Documentation

For more information, please refer to the following documentation:

- [User Guide](https://docs.greptime.com/user-guide/deployments-administration/deploy-on-kubernetes/overview)

- [API References](./docs/api-references/docs.md)

## License

greptimedb-operator uses the [Apache 2.0 license](./LICENSE) to strike a balance between
open contributions and allowing you to use the software however you want.
