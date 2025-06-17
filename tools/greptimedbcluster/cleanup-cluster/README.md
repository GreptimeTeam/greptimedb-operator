# GreptimeDBCluster Cleanup Script

This script is designed to clean up a GreptimeDBCluster and its associated etcd data. It performs the following steps:

1. Delete a specified GreptimeDBCluster.
2. Delete the persistent volume claims (PVCs) associated with that GreptimeDB Datanodes.
3. Cleans up all data from etcd.

Note: Please proceed with caution when executing this script, as it will permanently delete the GreptimeDBCluster and associated data.

## Prerequisites

- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) or [docker](https://docs.docker.com/get-docker/) is installed.
- `kubeconfig` should be configured to access your Kubernetes cluster.
- Ensure `etcdctl` is installed and can be executed inside the etcd pods.
- jq tool installed for JSON processing.

## Script Usage

1. Find the script file: `./tools/greptimedbcluster/clean-cluster/cleanup-cluster.sh`.
2. Run the script:
   ```bash
   ./tools/greptimedbcluster/cleanup-cluster/cleanup-cluster.sh \
   --cluster basic \
   --namespace default \
   --etcd etcd \
   --etcd-namespace etcd-cluster
   ```

During execution, you will need to enter:
- `--cluster`: The name of the GreptimeDBCluster
- `--namespace`: The namespace of the GreptimeDBCluster
- `--etcd`: The name of the ETCD StatefulSet
- `--etcd-namespace`: The namespace of the ETCD StatefulSet

3. Output:

```bash
=> Check prerequisites...
<= All prerequisites are met.
Would you like to cleanup the GreptimeDBCluster 'basic' in namespace 'default', delete the datanode PVCs, and remove the metadata in etcd StatefulSet 'etcd' in namespace 'etcd-cluster'? type 'yes' or 'no':
yes
Please confirm again, would you like to cleanup the GreptimeDBCluster 'basic' in namespace 'default', delete the datanode PVCs, and remove the metadata in etcd StatefulSet 'etcd' in namespace 'etcd-cluster'? type 'yes' or 'no':
yes
Deleting GreptimeDBCluster 'basic' in namespace 'default'...
Successfully deleted GreptimeDBCluster 'basic' in namespace 'default'.
Deleting PVC: datanode-basic-datanode-0
Successfully deleted PVC 'datanode-basic-datanode-0'.
Deleting monitor PVC: datanode-basic-monitor-standalone-0
Successfully deleted monitor PVC 'datanode-basic-monitor-standalone-0'.
Starting etcd data cleanup...
Checking pod etcd-0 for leader status...
Checking pod etcd-1 for leader status...
Leader found: etcd-1 (member_id=15463123987720296471)
Deleting all data from etcd...
7
Successfully deleted all data from etcd.
Cleanup completed successfully.
Please remember to check and manually delete any associated Object Storage files, RDS and Kafka data if exist.
```

## Error Handling

The script checks for the following conditions and provides relevant error messages:

1. The specified GreptimeDBCluster does not exist.
2. No PVCs were found.
3. No running etcd pods were found or the leader pod could not be detected.
4. Failure to delete etcd data.

## Next

After executing the script, you will also need to manually delete object storage files, kafka and RDS backend data, etc.
