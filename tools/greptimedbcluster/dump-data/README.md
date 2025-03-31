# Dump data Script

This script collects various information about a GreptimeDB cluster and its associated resources in Kubernetes. It gathers global resource information, specific cluster details, ETCD status, pod information, and SQL queries execution results, packaging all the data into a compressed file.

## Requirements

- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) or [docker](https://docs.docker.com/get-docker/) is installed.
- Kubernetes cluster with access to the GreptimeDB environment.
- jq tool installed for JSON processing.
- Ensure `etcdctl` is installed and can be executed inside the etcd pods.

## Features
- Collects global and cluster-specific information.
- Gathers ETCD pod statuses and health check results.
- Collects detailed pod descriptions and logs, both current and previous.
- Executes SQL queries against the GreptimeDB cluster and fetches relevant tables.
- Packages all collected information into a compressed .tar.gz archive.

## Usage

1. Run the script:
  ```bash
  ./tools/greptimedbcluster/dump-data/dump-data.sh --cluster basic --namespace default --etcd etcd --etcd-namespace etcd-cluster --output-dir
 /tmp
  ```

2. During execution, you will need to enter:     

`--cluster`: The name of the GreptimeDBCluster.    
`--namespace`: The namespace of the GreptimeDBCluster.    
`--etcd`: The name of the ETCD StatefulSet.    
`--etcd-namespace`: The namespace of the ETCD StatefulSet.    
`--output-dir`: The output directory.   


3. The script will then begin collecting data and display progress as it processes:
  ```bash
  Output directory: /tmp/greptime_dump_20250401_005417
=> Check prerequisites...
<= All prerequisites are met.
[1/6] Collecting global resources...
[2/6] Collecting cluster (default/basic) information...
[3/6] Collecting ETCD info...
[4/6] Collecting pod details and logs...
[5/6] Collecting information_schema table...
[6/6] Packaging collected information...
Collection complete! All data saved to: /tmp/greptime_dump_20250401_005417.tar.gz
Contents:
-rw-r--r--  1 greptime  wheel   176B  4  1 00:54 /tmp/greptime_dump_20250401_005417/1_global_resources.txt
-rw-r--r--  1 greptime  wheel    28K  4  1 00:54 /tmp/greptime_dump_20250401_005417/2_cluster_default_basic.txt
-rw-r--r--  1 greptime  wheel   9.7K  4  1 00:54 /tmp/greptime_dump_20250401_005417/3_etcd_status.txt
-rw-r--r--  1 greptime  wheel    13K  4  1 00:54 /tmp/greptime_dump_20250401_005417/4_pod_basic-datanode-0_info.txt
-rw-r--r--  1 greptime  wheel   8.7K  4  1 00:54 /tmp/greptime_dump_20250401_005417/4_pod_basic-frontend-759768ff9d-qq48s_info.txt
-rw-r--r--  1 greptime  wheel   8.4K  4  1 00:54 /tmp/greptime_dump_20250401_005417/4_pod_basic-meta-5f76ccd8bb-5c2bs_info.txt
-rw-r--r--  1 greptime  wheel   1.6K  4  1 00:54 /tmp/greptime_dump_20250401_005417/5_information_schema_table.txt

/tmp/greptime_dump_20250401_005417/pod_logs:
total 792
-rw-r--r--  1 greptime  wheel    20K  4  1 00:54 basic-datanode-0.log
-rw-r--r--  1 greptime  wheel   112B  4  1 00:54 basic-datanode-0_previous.log
-rw-r--r--  1 greptime  wheel    15K  4  1 00:54 basic-frontend-759768ff9d-qq48s.log
-rw-r--r--  1 greptime  wheel   124B  4  1 00:54 basic-frontend-759768ff9d-qq48s_previous.log
-rw-r--r--  1 greptime  wheel    10K  4  1 00:54 basic-meta-5f76ccd8bb-5c2bs.log
-rw-r--r--  1 greptime  wheel   116B  4  1 00:54 basic-meta-5f76ccd8bb-5c2bs_previous.log
-rw-r--r--  1 greptime  wheel   123K  4  1 00:54 etcd-0.log
-rw-r--r--  1 greptime  wheel   125K  4  1 00:54 etcd-1.log
-rw-r--r--  1 greptime  wheel    81K  4  1 00:54 etcd-2.log
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
