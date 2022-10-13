# Overview

This chart bootstraps an etcd cluster on a [Kubernetes](http://kubernetes.io) cluster using the [Helm](https://helm.sh) package manager.

## Getting Started

- Install and uninstall etcd cluster

  ```
  # Install etcd.
  $ helm install mydb ./etcd

  # Uninstall etcd.
  $ helm uninstall mydb
  ```
  
- Install and uninstall etcd cluster in the given namespace

  ```
  # Install etcd on the specific namespace(if you want to create the namespace).
  $ helm install mydb ./etcd --namespace etcd --create-namespace
  
  # Uninstall etcd.
  $ helm install mydb -n etcd
  ```
  
- Test your etcd cluster

  ```
  # Create the diagnostic pod.
  $ kubectl run etcd-client --image bitnami/etcd:latest --command sleep infinity
  
  # List the members of etcd cluster.
  $ kubectl exec etcd-client -- etcdctl --endpoints=mydb-etcd-svc.default:2379 member list
  
  # Get etcd endpoints status.
  $ kubectl exec etcd-client -- etcdctl --endpoints=mydb-etcd-0.mydb-etcd-svc:2379,mydb-etcd-1.mydb-etcd-svc:2379,mydb-etcd-2.mydb-etcd-svc:2379 endpoint status -w table
  ```
