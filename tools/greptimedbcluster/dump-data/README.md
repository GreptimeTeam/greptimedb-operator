# Dump data Script

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
  Enter GreptimeDBCluster name: ${The name of the GreptimeDBCluster}
  Enter GreptimeDBCluster namespace: ${The namespace of the GreptimeDBCluster}
  Enter the name of the ETCD StatefulSet: ${The name of the ETCD StatefulSet}
  Enter the namespace of the ETCD StatefulSet: ${The namespace of the ETCD StatefulSet}
  ```

3. The script will then begin collecting data and display progress as it processes:
  ```bash
  Enter GreptimeDBCluster name: basic
  Enter GreptimeDBCluster namespace: default
  Enter the name of the ETCD StatefulSet: etcd
  Enter the namespace of the ETCD StatefulSet: etcd-cluster
  Starting GreptimeDB environment information collection...
  [1/6] Collecting global resources...
  [2/6] Collecting cluster (default/basic) information...
  [3/6] Collecting ETCD status...
  [4/6] Collecting pod details and logs...
  Processing Pod: basic-datanode-0
  Processing Pod: basic-datanode-1
  Processing Pod: basic-datanode-2
  Processing Pod: basic-frontend-759768ff9d-pwrvl
  Processing Pod: basic-meta-5f76ccd8bb-bzhww
  [5/6] Collecting information_schema table...
  [6/6] Packaging collected information...
  tar: Removing leading '/' from member names
  Collection complete! All data saved to: /tmp/greptime_dump_20250326_180308.tar.gz
  Contents:
  -rw-r--r--  1 greptime  wheel   176B  3 26 18:03 /tmp/greptime_dump_20250326_180308/1_global_resources.txt
  -rw-r--r--  1 greptime  wheel    28K  3 26 18:03 /tmp/greptime_dump_20250326_180308/2_cluster_default_basic.txt
  -rw-r--r--  1 greptime  wheel   3.4K  3 26 18:03 /tmp/greptime_dump_20250326_180308/3_etcd_status.txt
  -rw-r--r--  1 greptime  wheel    12K  3 26 18:03 /tmp/greptime_dump_20250326_180308/4_pod_basic-datanode-0_info.txt
  -rw-r--r--  1 greptime  wheel    12K  3 26 18:03 /tmp/greptime_dump_20250326_180308/4_pod_basic-datanode-1_info.txt
  -rw-r--r--  1 greptime  wheel    12K  3 26 18:03 /tmp/greptime_dump_20250326_180308/4_pod_basic-datanode-2_info.txt
  -rw-r--r--  1 greptime  wheel   8.6K  3 26 18:03 /tmp/greptime_dump_20250326_180308/4_pod_basic-frontend-759768ff9d-pwrvl_info.txt
  -rw-r--r--  1 greptime  wheel   8.3K  3 26 18:03 /tmp/greptime_dump_20250326_180308/4_pod_basic-meta-5f76ccd8bb-bzhww_info.txt
  -rw-r--r--  1 greptime  wheel   1.8K  3 26 18:03 /tmp/greptime_dump_20250326_180308/5_information_schema_table.txt

  /tmp/greptime_dump_20250326_180308/pod_logs:
  total 424
  -rw-r--r--  1 greptime  wheel    21K  3 26 18:03 basic-datanode-0.log
  -rw-r--r--  1 greptime  wheel   112B  3 26 18:03 basic-datanode-0_previous.log
  -rw-r--r--  1 greptime  wheel    20K  3 26 18:03 basic-datanode-1.log
  -rw-r--r--  1 greptime  wheel   112B  3 26 18:03 basic-datanode-1_previous.log
  -rw-r--r--  1 greptime  wheel    20K  3 26 18:03 basic-datanode-2.log
  -rw-r--r--  1 greptime  wheel   112B  3 26 18:03 basic-datanode-2_previous.log
  -rw-r--r--  1 greptime  wheel    15K  3 26 18:03 basic-frontend-759768ff9d-pwrvl.log
  -rw-r--r--  1 greptime  wheel   124B  3 26 18:03 basic-frontend-759768ff9d-pwrvl_previous.log
  -rw-r--r--  1 greptime  wheel    11K  3 26 18:03 basic-meta-5f76ccd8bb-bzhww.log
  -rw-r--r--  1 greptime  wheel   116B  3 26 18:03 basic-meta-5f76ccd8bb-bzhww_previous.log
  -rw-r--r--  1 greptime  wheel    31K  3 26 18:03 etcd-0.log
  -rw-r--r--  1 greptime  wheel    28K  3 26 18:03 etcd-1.log
  -rw-r--r--  1 greptime  wheel    35K  3 26 18:03 etcd-2.log
  ```

## Output Structure
```bash
/tmp/greptime_dump_<timestamp>/
├── 1_global_resources.txt           # Global resources information
├── 2_cluster_<namespace>_<name>.txt # Cluster-specific details
├── 3_etcd_status.txt                # ETCD pod status and health information
├── 4_pod_<pod-name>_info.txt        # Pod descriptions and configurations
├── 5_information_schema_table.txt   # Results from SQL queries
├── pod_logs/
│   ├── pod_<pod-name>_info.txt       # Current logs of the pod 
│   └── pod_<pod-name>_previous.log   # Previous logs of the pod
```
