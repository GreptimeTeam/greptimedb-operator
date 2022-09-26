# Overview

This chart deploy GreptimeDB operator.

## Getting Started

### Install

```
# Deploy greptimedb-operator in default namespace.
$ helm install gtcloud greptimedb-operator

# Deploy greptimedb-operator in new namespace.
$ helm install gtcloud greptimedb-operator --namespace greptimedb-operator-system --create-namespace
```

### Uninstall

```
$ helm uninstall greptimedb-operator
```
