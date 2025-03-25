# Dump GreptimeDB data Script

This script collects various information about a GreptimeDB cluster and its associated resources in Kubernetes. It gathers global resource information, specific cluster details, ETCD status, pod information, and SQL queries execution results, packaging all the data into a compressed file.

## Features
- Collects global and cluster-specific information.
- Gathers ETCD pod statuses and health check results.
- Collects detailed pod descriptions and logs, both current and previous.
- Executes SQL queries against the GreptimeDB cluster and fetches relevant tables.
- Packages all collected information into a compressed .tar.gz archive.

## Requirements

- Kubernetes cluster with access to the GreptimeDB environment.
- kubectl CLI tool installed and configured.
- jq tool installed for JSON processing.

## Usage

1. Run the script:
  ```bash
  ./tools/greptimedbcluster/dump-data/dump-data.sh
  ```

2. You will be prompted to enter the names and namespaces for the GreptimeDBCluster and ETCD StatefulSet:
  ```bash
  Enter GreptimeDBCluster name: basic
  Enter GreptimeDBCluster namespace: default
  Enter the name of the ETCD StatefulSet: etcd
  Enter the namespace of the ETCD StatefulSet: etcd-cluster
  ```

3. The script will then begin collecting data and display progress as it processes:
  ```bash
  Collection complete! All data saved to: 
  /tmp/greptime_dump_20250326_020505.tar.gz
  ```

## Output Structure
```bash
/tmp/greptime_dump_<timestamp>/
├── 1_global_resources.txt           # Global resources information
├── 2_cluster_<namespace>_<name>.txt # Cluster-specific details
├── 3_etcd_status.txt                # ETCD pod status and health information
├── 5_information_schema_table.txt   # Results from SQL queries
├── pod_logs/
│   ├── pod_<pod-name>_info.txt       # Pod descriptions and configurations
│   ├── pod_<pod-name>.log            # Current logs of the pod
│   └── pod_<pod-name>_previous.log   # Previous logs of the pod
```
