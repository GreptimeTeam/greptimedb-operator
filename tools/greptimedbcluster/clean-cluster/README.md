# GreptimeDBCluster Cleanup Script

This script is designed to clean up a GreptimeDBCluster and its associated etcd data. It performs the following steps:

1. Deletes a specified GreptimeDBCluster.
2. Deletes the persistent volume claims (PVCs) associated with that GreptimeDB Datanodes.
3. Cleans up all data from etcd.

Note: Please proceed with caution when executing this script, as it will permanently delete the GreptimeDBCluster and associated data.

## Prerequisites

- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) must be installed.
- `kubectl` should be configured to access your Kubernetes cluster.
- Ensure `etcdctl` is installed and can be executed inside the etcd pods.

## Script Usage

1. Find the script file: `./tools/greptimedbcluster/clean-cluster/clean-cluster.sh`.
2. Run the script:
   ```bash
   ./tools/greptimedbcluster/clean-cluster/clean-cluster.sh
   ```

During execution, the script will prompt you for:
- The name of the GreptimeDBCluster
- The namespace of the GreptimeDBCluster
- The name of the ETCD StatefulSet
- The namespace of the ETCD StatefulSet

## Example

```bash
# ./tools/greptimedbcluster/clean-cluster/clean-cluster.sh
Enter the name of the GreptimeDBCluster: basic
Enter the namespace of the GreptimeDBCluster: default
Enter the name of the ETCD StatefulSet: etcd
Enter the namespace of the ETCD StatefulSet: etcd-cluster
greptimedbcluster.greptime.io "basic" deleted
Successfully deleted GreptimeDBCluster 'basic' in namespace 'default'.
Deleting PVC: datanode-basic-datanode-0
persistentvolumeclaim "datanode-basic-datanode-0" deleted
Successfully deleted PVC 'datanode-basic-datanode-0'.
Deleting PVC: datanode-basic-datanode-1
persistentvolumeclaim "datanode-basic-datanode-1" deleted
Successfully deleted PVC 'datanode-basic-datanode-1'.
Deleting PVC: datanode-basic-datanode-2
persistentvolumeclaim "datanode-basic-datanode-2" deleted
Successfully deleted PVC 'datanode-basic-datanode-2'.
Starting etcd data cleanup...
Checking pod etcd-0 for leader status...
Checking pod etcd-1 for leader status...
Leader found: etcd-1 (member_id=15463123987720296471)
Deleting all data from etcd...
Successfully deleted all data from etcd.
```

## Error Handling

The script checks for the following conditions and provides relevant error messages:

1. The specified GreptimeDBCluster does not exist.
2. No PVCs were found.
3. No running etcd pods were found or the leader pod could not be detected.
4. Failure to delete etcd data.

## Next

After executing the script, you will also need to manually delete object storage files and PG backend data, etc.
