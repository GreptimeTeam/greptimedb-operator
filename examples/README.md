# Examples of Using the GreptimeDB Operator

## Cluster

The following examples suppose that you have installed the etcd cluster in the `etcd` namespace with the service endpoint `etcd.etcd-cluster.svc.cluster.local:2379`.

- [Basic](./cluster/basic/cluster.yaml): Create a basic GreptimeDB cluster.
- [S3](./cluster/s3/cluster.yaml): Create a GreptimeDB cluster with S3 storage.
- [GCS](./cluster/gcs/cluster.yaml): Create a GreptimeDB cluster with Google GCS storage.
- [OSS](./cluster/oss/cluster.yaml): Create a GreptimeDB cluster with Aliyun OSS storage.
- [Flownode](./cluster/flownode/cluster.yaml): Create a GreptimeDB cluster with `flownode` enabled. By adding the `flownode` configuration, you can use [continuous aggregation](https://docs.greptime.com/user-guide/continuous-aggregation/overview) in the GreptimeDB cluster.
- [TLS Service](./cluster/tls-service/cluster.yaml): Create a GreptimeDB cluster with TLS service.
- [Prometheus Monitoring](./cluster/prometheus-monitor/cluster.yaml): Create a GreptimeDB cluster with Prometheus monitoring. Please ensure you have already installed prometheus-operator and created a Prometheus instance with the label `release=prometheus`.
- [Kafka Remote WAL](./cluster/kafka-remote-wal/cluster.yaml): Create a GreptimeDB cluster with Kafka remote WAL. Please ensure you have installed the Kafka cluster in the `kafka` namespace with the service endpoint `kafka-bootstrap.kafka.svc.cluster.local:9092`.
- [Add Custom Config](./cluster/add-custom-config/cluster.yaml): Create a GreptimeDB cluster with custom configuration by using the `config` field.
- [AWS NLB](./cluster/aws-nlb/cluster.yaml): Create a GreptimeDB cluster with the AWS NLB service. Please ensure you have already configured it.
- [Standalone WAL](./cluster/standalone-wal/cluster.yaml): Create a GreptimeDB cluster with standalone storage for WAL.
- [Configure Logging](./cluster/configure-logging/cluster.yaml): Create a GreptimeDB cluster with custom logging configuration.
- [Enable Monitoring Bootstrap](./cluster/enable-monitoring/cluster.yaml): Create a GreptimeDB cluster with monitoring enabled.

## Standalone

- [Basic](./standalone/basic/standalone.yaml): Create a basic GreptimeDB standalone.
- [S3](./standalone/s3/standalone.yaml): Create a GreptimeDB standalone with S3 storage.
- [GCS](./standalone/gcs/standalone.yaml): Create a GreptimeDB standalone with Google GCS storage.
- [OSS](./standalone/oss/standalone.yaml): Create a GreptimeDB standalone with Aliyun OSS storage.
