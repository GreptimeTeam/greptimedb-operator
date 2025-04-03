# GreptimeDB Operator API References

## Packages
- [greptime.io/v1alpha1](#greptimeiov1alpha1)


## greptime.io/v1alpha1


### Resource Types
- [GreptimeDBCluster](#greptimedbcluster)
- [GreptimeDBClusterList](#greptimedbclusterlist)
- [GreptimeDBStandalone](#greptimedbstandalone)
- [GreptimeDBStandaloneList](#greptimedbstandalonelist)



#### AZBlobStorage



AZBlobStorage defines the Azure Blob storage specification.



_Appears in:_
- [ObjectStorageProviderSpec](#objectstorageproviderspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `container` _string_ | The data will be stored in the container. |  |  |
| `secretName` _string_ | The secret of storing the credentials of account name and account key.<br />The secret should contain keys named `account-name` and `account-key`.<br />The secret must be the same namespace with the GreptimeDBCluster resource. |  |  |
| `root` _string_ | The Blob directory path. |  |  |
| `endpoint` _string_ | The Blob Storage endpoint. |  |  |


#### CacheStorage



CacheStorage defines the cache storage specification.



_Appears in:_
- [ObjectStorageProviderSpec](#objectstorageproviderspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `fs` _[FileStorage](#filestorage)_ | Storage is the storage specification for the cache.<br />If the storage is not specified, the cache will use DatanodeStorageSpec. |  |  |
| `cacheCapacity` _string_ | CacheCapacity is the capacity of the cache. |  |  |




#### ComponentSpec



ComponentSpec is the common specification for all components(`frontend`/`meta`/`datanode`/`flownode`).



_Appears in:_
- [DatanodeSpec](#datanodespec)
- [FlownodeSpec](#flownodespec)
- [FrontendSpec](#frontendspec)
- [MetaSpec](#metaspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `replicas` _integer_ | The number of replicas of the components. |  | Minimum: 0 <br /> |
| `config` _string_ | The content of the configuration file of the component in TOML format. |  |  |
| `template` _[PodTemplateSpec](#podtemplatespec)_ | Template defines the pod template for the component, if not specified, the pod template will use the default value. |  |  |
| `logging` _[LoggingSpec](#loggingspec)_ | Logging defines the logging configuration for the component. |  |  |


#### Condition



Condition describes the state of a deployment at a certain point.



_Appears in:_
- [GreptimeDBClusterStatus](#greptimedbclusterstatus)
- [GreptimeDBStandaloneStatus](#greptimedbstandalonestatus)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `type` _[ConditionType](#conditiontype)_ | Type of deployment condition. |  |  |
| `lastUpdateTime` _[Time](https://kubernetes.io/docs/reference/generated/kubernetes-api/v/#time-v1-meta)_ | The last time this condition was updated. |  |  |
| `lastTransitionTime` _[Time](https://kubernetes.io/docs/reference/generated/kubernetes-api/v/#time-v1-meta)_ | Last time the condition transitioned from one status to another. |  |  |
| `reason` _string_ | The reason for the condition's last transition. |  |  |
| `message` _string_ | A human-readable message indicating details about the transition. |  |  |


#### ConditionType

_Underlying type:_ _string_

ConditionType is the type of the condition.



_Appears in:_
- [Condition](#condition)

| Field | Description |
| --- | --- |
| `Ready` | ConditionTypeReady indicates that the GreptimeDB cluster is ready to serve requests.<br />Every component in the cluster are all ready.<br /> |
| `Progressing` | ConditionTypeProgressing indicates that the GreptimeDB cluster is progressing.<br /> |


#### DatanodeSpec



DatanodeSpec is the specification for datanode component.



_Appears in:_
- [GreptimeDBClusterSpec](#greptimedbclusterspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `replicas` _integer_ | The number of replicas of the components. |  | Minimum: 0 <br /> |
| `config` _string_ | The content of the configuration file of the component in TOML format. |  |  |
| `template` _[PodTemplateSpec](#podtemplatespec)_ | Template defines the pod template for the component, if not specified, the pod template will use the default value. |  |  |
| `logging` _[LoggingSpec](#loggingspec)_ | Logging defines the logging configuration for the component. |  |  |
| `rpcPort` _integer_ | RPCPort is the gRPC port of the datanode. |  | Maximum: 65535 <br />Minimum: 0 <br /> |
| `httpPort` _integer_ | HTTPPort is the HTTP port of the datanode. |  | Maximum: 65535 <br />Minimum: 0 <br /> |
| `storage` _[DatanodeStorageSpec](#datanodestoragespec)_ | Storage is the default file storage of the datanode. For example, WAL, cache, index etc. |  |  |
| `rollingUpdate` _[RollingUpdateStatefulSetStrategy](https://kubernetes.io/docs/reference/generated/kubernetes-api/v/#rollingupdatestatefulsetstrategy-v1-apps)_ | RollingUpdate is the rolling update configuration. We always use `RollingUpdate` strategy. |  |  |


#### DatanodeStatus



DatanodeStatus is the status of datanode node.



_Appears in:_
- [GreptimeDBClusterStatus](#greptimedbclusterstatus)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `replicas` _integer_ | Replicas is the number of replicas of the datanode. |  |  |
| `readyReplicas` _integer_ | ReadyReplicas is the number of ready replicas of the datanode. |  |  |


#### DatanodeStorageSpec



DatanodeStorageSpec defines the storage specification for the datanode.



_Appears in:_
- [DatanodeSpec](#datanodespec)
- [GreptimeDBStandaloneSpec](#greptimedbstandalonespec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `dataHome` _string_ | DataHome is the home directory of the data. |  |  |
| `fs` _[FileStorage](#filestorage)_ | FileStorage is the file storage configuration. |  |  |


#### FileStorage



FileStorage defines the file storage specification. It is used to generate the PVC that will be mounted to the container.



_Appears in:_
- [CacheStorage](#cachestorage)
- [DatanodeStorageSpec](#datanodestoragespec)
- [RaftEngineWAL](#raftenginewal)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `name` _string_ | Name is the name of the PVC that will be created. |  |  |
| `storageClassName` _string_ | StorageClassName is the name of the StorageClass to use for the PVC. |  |  |
| `storageSize` _string_ | StorageSize is the size of the storage. |  | Pattern: `(^([+-]?[0-9.]+)([eEinumkKMGTP]*[-+]?[0-9]*)$)` <br /> |
| `mountPath` _string_ | MountPath is the path where the storage will be mounted in the container. |  |  |
| `storageRetainPolicy` _[StorageRetainPolicyType](#storageretainpolicytype)_ | StorageRetainPolicy is the policy of the storage. It can be `Retain` or `Delete`. |  | Enum: [Retain Delete] <br /> |
| `labels` _object (keys:string, values:string)_ | Labels is the labels for the PVC. |  |  |
| `annotations` _object (keys:string, values:string)_ | Annotations is the annotations for the PVC. |  |  |




#### FlownodeSpec



FlownodeSpec is the specification for flownode component.



_Appears in:_
- [GreptimeDBClusterSpec](#greptimedbclusterspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `replicas` _integer_ | The number of replicas of the components. |  | Minimum: 0 <br /> |
| `config` _string_ | The content of the configuration file of the component in TOML format. |  |  |
| `template` _[PodTemplateSpec](#podtemplatespec)_ | Template defines the pod template for the component, if not specified, the pod template will use the default value. |  |  |
| `logging` _[LoggingSpec](#loggingspec)_ | Logging defines the logging configuration for the component. |  |  |
| `rpcPort` _integer_ | The gRPC port of the flownode. |  | Maximum: 65535 <br />Minimum: 0 <br /> |
| `httpPort` _integer_ | The HTTP port of the flownode. |  | Maximum: 65535 <br />Minimum: 0 <br /> |
| `rollingUpdate` _[RollingUpdateStatefulSetStrategy](https://kubernetes.io/docs/reference/generated/kubernetes-api/v/#rollingupdatestatefulsetstrategy-v1-apps)_ | RollingUpdate is the rolling update configuration. We always use `RollingUpdate` strategy. |  |  |


#### FlownodeStatus



FlownodeStatus is the status of flownode node.



_Appears in:_
- [GreptimeDBClusterStatus](#greptimedbclusterstatus)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `replicas` _integer_ | Replicas is the number of replicas of the flownode. |  |  |
| `readyReplicas` _integer_ | ReadyReplicas is the number of ready replicas of the flownode. |  |  |


#### FrontendSpec



FrontendSpec is the specification for frontend component.



_Appears in:_
- [GreptimeDBClusterSpec](#greptimedbclusterspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `replicas` _integer_ | The number of replicas of the components. |  | Minimum: 0 <br /> |
| `config` _string_ | The content of the configuration file of the component in TOML format. |  |  |
| `template` _[PodTemplateSpec](#podtemplatespec)_ | Template defines the pod template for the component, if not specified, the pod template will use the default value. |  |  |
| `logging` _[LoggingSpec](#loggingspec)_ | Logging defines the logging configuration for the component. |  |  |
| `name` _string_ | Name is the name of the frontend. |  |  |
| `rpcPort` _integer_ | RPCPort is the gRPC port of the frontend. |  | Maximum: 65535 <br />Minimum: 0 <br /> |
| `httpPort` _integer_ | HTTPPort is the HTTP port of the frontend. |  | Maximum: 65535 <br />Minimum: 0 <br /> |
| `mysqlPort` _integer_ | MySQLPort is the MySQL port of the frontend. |  | Maximum: 65535 <br />Minimum: 0 <br /> |
| `postgreSQLPort` _integer_ | PostgreSQLPort is the PostgreSQL port of the frontend. |  | Maximum: 65535 <br />Minimum: 0 <br /> |
| `service` _[ServiceSpec](#servicespec)_ | Service is the service configuration of the frontend. |  |  |
| `tls` _[TLSSpec](#tlsspec)_ | TLS is the TLS configuration of the frontend. |  |  |
| `rollingUpdate` _[RollingUpdateDeployment](https://kubernetes.io/docs/reference/generated/kubernetes-api/v/#rollingupdatedeployment-v1-apps)_ | RollingUpdate is the rolling update configuration. We always use `RollingUpdate` strategyt. |  |  |


#### FrontendStatus



FrontendStatus is the status of frontend node.



_Appears in:_
- [GreptimeDBClusterStatus](#greptimedbclusterstatus)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `replicas` _integer_ | Replicas is the number of replicas of the frontend. |  |  |
| `readyReplicas` _integer_ | ReadyReplicas is the number of ready replicas of the frontend. |  |  |


#### GCSStorage



GCSStorage defines the Google GCS storage specification.



_Appears in:_
- [ObjectStorageProviderSpec](#objectstorageproviderspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `bucket` _string_ | The data will be stored in the bucket. |  |  |
| `root` _string_ | The gcs directory path. |  |  |
| `secretName` _string_ | The secret of storing Credentials for gcs service OAuth2 authentication.<br />The secret should contain keys named `service-account-key`.<br />The secret must be the same namespace with the GreptimeDBCluster resource. |  |  |
| `scope` _string_ | The scope for gcs. |  |  |
| `endpoint` _string_ | The endpoint URI of gcs service. |  |  |


#### GreptimeDBCluster



GreptimeDBCluster is the Schema for the greptimedbclusters API



_Appears in:_
- [GreptimeDBClusterList](#greptimedbclusterlist)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `greptime.io/v1alpha1` | | |
| `kind` _string_ | `GreptimeDBCluster` | | |
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `spec` _[GreptimeDBClusterSpec](#greptimedbclusterspec)_ | Spec is the specification of the desired state of the GreptimeDBCluster. |  |  |


#### GreptimeDBClusterList



GreptimeDBClusterList contains a list of GreptimeDBCluster





| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `greptime.io/v1alpha1` | | |
| `kind` _string_ | `GreptimeDBClusterList` | | |
| `metadata` _[ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v/#listmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `items` _[GreptimeDBCluster](#greptimedbcluster) array_ |  |  |  |


#### GreptimeDBClusterSpec



GreptimeDBClusterSpec defines the desired state of GreptimeDBCluster



_Appears in:_
- [GreptimeDBCluster](#greptimedbcluster)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `base` _[PodTemplateSpec](#podtemplatespec)_ | Base is the base pod template for all components and can be overridden by template of individual component. |  |  |
| `frontend` _[FrontendSpec](#frontendspec)_ | Frontend is the specification of frontend node. |  |  |
| `meta` _[MetaSpec](#metaspec)_ | Meta is the specification of meta node. |  |  |
| `datanode` _[DatanodeSpec](#datanodespec)_ | Datanode is the specification of datanode node. |  |  |
| `flownode` _[FlownodeSpec](#flownodespec)_ | Flownode is the specification of flownode node. |  |  |
| `frontends` _[FrontendSpec](#frontendspec) array_ | Frontends is a group of frontend nodes. |  |  |
| `httpPort` _integer_ | HTTPPort is the HTTP port of the greptimedb cluster. |  | Maximum: 65535 <br />Minimum: 0 <br /> |
| `rpcPort` _integer_ | RPCPort is the RPC port of the greptimedb cluster. |  | Maximum: 65535 <br />Minimum: 0 <br /> |
| `mysqlPort` _integer_ | MySQLPort is the MySQL port of the greptimedb cluster. |  | Maximum: 65535 <br />Minimum: 0 <br /> |
| `postgreSQLPort` _integer_ | PostgreSQLPort is the PostgreSQL port of the greptimedb cluster. |  | Maximum: 65535 <br />Minimum: 0 <br /> |
| `prometheusMonitor` _[PrometheusMonitorSpec](#prometheusmonitorspec)_ | PrometheusMonitor is the specification for creating PodMonitor or ServiceMonitor. |  |  |
| `version` _string_ | Version is the version of greptimedb. |  |  |
| `initializer` _[InitializerSpec](#initializerspec)_ | Initializer is the init container to set up components configurations before running the container. |  |  |
| `objectStorage` _[ObjectStorageProviderSpec](#objectstorageproviderspec)_ | ObjectStorageProvider is the storage provider for the greptimedb cluster. |  |  |
| `wal` _[WALProviderSpec](#walproviderspec)_ | WALProvider is the WAL provider for the greptimedb cluster. |  |  |
| `logging` _[LoggingSpec](#loggingspec)_ | The global logging configuration for all components. It can be overridden by the logging configuration of individual component. |  |  |
| `monitoring` _[MonitoringSpec](#monitoringspec)_ | Monitoring is the specification for monitor bootstrapping. It will create a standalone greptimedb instance to monitor the cluster. |  |  |
| `ingress` _[IngressSpec](#ingressspec)_ | Ingress is the Ingress configuration of the frontend. |  |  |




#### GreptimeDBStandalone



GreptimeDBStandalone is the Schema for the greptimedbstandalones API



_Appears in:_
- [GreptimeDBStandaloneList](#greptimedbstandalonelist)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `greptime.io/v1alpha1` | | |
| `kind` _string_ | `GreptimeDBStandalone` | | |
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `spec` _[GreptimeDBStandaloneSpec](#greptimedbstandalonespec)_ | Spec is the specification of the desired state of the GreptimeDBStandalone. |  |  |


#### GreptimeDBStandaloneList



GreptimeDBStandaloneList contains a list of GreptimeDBStandalone





| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `greptime.io/v1alpha1` | | |
| `kind` _string_ | `GreptimeDBStandaloneList` | | |
| `metadata` _[ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v/#listmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `items` _[GreptimeDBStandalone](#greptimedbstandalone) array_ |  |  |  |


#### GreptimeDBStandaloneSpec



GreptimeDBStandaloneSpec defines the desired state of GreptimeDBStandalone



_Appears in:_
- [GreptimeDBStandalone](#greptimedbstandalone)
- [MonitoringSpec](#monitoringspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `base` _[PodTemplateSpec](#podtemplatespec)_ | Base is the base pod template for all components and can be overridden by template of individual component. |  |  |
| `service` _[ServiceSpec](#servicespec)_ | Service is the service configuration of greptimedb. |  |  |
| `tls` _[TLSSpec](#tlsspec)_ | The TLS configurations of the greptimedb. |  |  |
| `httpPort` _integer_ | HTTPPort is the port of the greptimedb http service. |  | Maximum: 65535 <br />Minimum: 0 <br /> |
| `rpcPort` _integer_ | RPCPort is the port of the greptimedb rpc service. |  | Maximum: 65535 <br />Minimum: 0 <br /> |
| `mysqlPort` _integer_ | MySQLPort is the port of the greptimedb mysql service. |  | Maximum: 65535 <br />Minimum: 0 <br /> |
| `postgreSQLPort` _integer_ | PostgreSQLPort is the port of the greptimedb postgresql service. |  | Maximum: 65535 <br />Minimum: 0 <br /> |
| `prometheusMonitor` _[PrometheusMonitorSpec](#prometheusmonitorspec)_ | PrometheusMonitor is the specification for creating PodMonitor or ServiceMonitor. |  |  |
| `version` _string_ | Version is the version of the greptimedb. |  |  |
| `initializer` _[InitializerSpec](#initializerspec)_ | Initializer is the init container to set up components configurations before running the container. |  |  |
| `objectStorage` _[ObjectStorageProviderSpec](#objectstorageproviderspec)_ | ObjectStorageProvider is the storage provider for the greptimedb cluster. |  |  |
| `datanodeStorage` _[DatanodeStorageSpec](#datanodestoragespec)_ | DatanodeStorage is the default file storage of the datanode. For example, WAL, cache, index etc. |  |  |
| `wal` _[WALProviderSpec](#walproviderspec)_ | WALProvider is the WAL provider for the greptimedb cluster. |  |  |
| `config` _string_ | The content of the configuration file of the component in TOML format. |  |  |
| `logging` _[LoggingSpec](#loggingspec)_ | Logging defines the logging configuration for the component. |  |  |
| `rollingUpdate` _[RollingUpdateStatefulSetStrategy](https://kubernetes.io/docs/reference/generated/kubernetes-api/v/#rollingupdatestatefulsetstrategy-v1-apps)_ | RollingUpdate is the rolling update configuration. We always use `RollingUpdate` strategy. |  |  |




#### IngressBackend



IngressBackend defines the Ingress backend configuration.



_Appears in:_
- [IngressRule](#ingressrule)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `name` _string_ | Name is the referenced frontend name. |  |  |
| `path` _string_ | Path is matched against the path of an incoming request. |  |  |
| `pathType` _[PathType](https://kubernetes.io/docs/reference/generated/kubernetes-api/v/#pathtype-v1-networking)_ | PathType determines the interpretation of the path matching. |  |  |


#### IngressRule



IngressRule defines the Ingress rule configuration.



_Appears in:_
- [IngressSpec](#ingressspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `host` _string_ | Host is the fully qualified domain name of a network host. |  |  |
| `backends` _[IngressBackend](#ingressbackend) array_ | IngressBackend is the Ingress backend configuration. |  |  |


#### IngressSpec



IngressSpec defines the Ingress configuration.



_Appears in:_
- [GreptimeDBClusterSpec](#greptimedbclusterspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `annotations` _object (keys:string, values:string)_ | Annotations is the annotations for the ingress. |  |  |
| `labels` _object (keys:string, values:string)_ | Labels is the labels for the ingress. |  |  |
| `ingressClassName` _string_ | IngressClassName is the name of an IngressClass. |  |  |
| `rules` _[IngressRule](#ingressrule) array_ | IngressRule is a list of host rules used to configure the Ingress. |  |  |
| `tls` _[IngressTLS](https://kubernetes.io/docs/reference/generated/kubernetes-api/v/#ingresstls-v1-networking) array_ | TLS is the Ingress TLS configuration. |  |  |


#### InitializerSpec



InitializerSpec is the init container to set up components configurations before running the container.



_Appears in:_
- [GreptimeDBClusterSpec](#greptimedbclusterspec)
- [GreptimeDBStandaloneSpec](#greptimedbstandalonespec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `image` _string_ | The image of the initializer. |  |  |


#### KafkaWAL



KafkaWAL is the specification for Kafka remote WAL.



_Appears in:_
- [WALProviderSpec](#walproviderspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `brokerEndpoints` _string array_ | BrokerEndpoints is the list of Kafka broker endpoints. |  |  |


#### LogFormat

_Underlying type:_ _string_





_Appears in:_
- [LoggingSpec](#loggingspec)

| Field | Description |
| --- | --- |
| `json` | LogFormatJSON is the `json` format of the logging.<br /> |
| `text` | LogFormatText is the `text` format of the logging.<br /> |


#### LogPipeline



LogPipeline is the specification for log pipeline.



_Appears in:_
- [LogsCollectionSpec](#logscollectionspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `data` _string_ | The content of the pipeline configuration file in YAML format. |  |  |


#### LoggingLevel

_Underlying type:_ _string_

LoggingLevel is the level of the logging.



_Appears in:_
- [LoggingSpec](#loggingspec)

| Field | Description |
| --- | --- |
| `info` | LoggingLevelInfo is the `info` level of the logging.<br /> |
| `error` | LoggingLevelError is the `error` level of the logging.<br /> |
| `warn` | LoggingLevelWarn is the `warn` level of the logging.<br /> |
| `debug` | LoggingLevelDebug is the `debug` level of the logging.<br /> |


#### LoggingSpec



LoggingSpec defines the logging configuration for the component.



_Appears in:_
- [ComponentSpec](#componentspec)
- [DatanodeSpec](#datanodespec)
- [FlownodeSpec](#flownodespec)
- [FrontendSpec](#frontendspec)
- [GreptimeDBClusterSpec](#greptimedbclusterspec)
- [GreptimeDBStandaloneSpec](#greptimedbstandalonespec)
- [MetaSpec](#metaspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `level` _[LoggingLevel](#logginglevel)_ | Level is the level of the logging. |  | Enum: [info error warn debug] <br /> |
| `filters` _string array_ | Filters is the filters of the logging.<br />User can use [EnvFilter](https://docs.rs/tracing-subscriber/latest/tracing_subscriber/filter/struct.EnvFilter.html) to filter the logging.<br />We can use `target[span\{field=value\}]=level` to filter the logging by target and span field.<br />For example, "mito2=debug" will filter the logging of target `mito2` to `debug` level. |  |  |
| `logsDir` _string_ | LogsDir is the directory path of the logs. |  |  |
| `persistentWithData` _boolean_ | PersistentWithData indicates whether to persist the log with the datanode data storage. It **ONLY** works for the datanode component.<br />If false, the log will be stored in ephemeral storage. |  |  |
| `onlyLogToStdout` _boolean_ | OnlyLogToStdout indicates whether to only log to stdout. If true, the log will not be stored in the storage even if the storage is configured. |  |  |
| `format` _[LogFormat](#logformat)_ | Format is the format of the logging. |  | Enum: [json text] <br /> |
| `slowQuery` _[SlowQuery](#slowquery)_ | SlowQuery is the slow query configuration. |  |  |


#### LogsCollectionSpec



LogsCollectionSpec is the specification for cluster logs collection.



_Appears in:_
- [MonitoringSpec](#monitoringspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `pipeline` _[LogPipeline](#logpipeline)_ | The specification of the log pipeline. |  |  |


#### MainContainerSpec



MainContainerSpec describes the specification of the main container of a pod.
Most of the fields of MainContainerSpec are from 'corev1.Container'.



_Appears in:_
- [PodTemplateSpec](#podtemplatespec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `image` _string_ | The main container image name of the component. |  |  |
| `resources` _[ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v/#resourcerequirements-v1-core)_ | The resource requirements of the main container. |  |  |
| `command` _string array_ | Entrypoint array. Not executed within a shell.<br />The container image's ENTRYPOINT is used if this is not provided.<br />Variable references `$(VAR_NAME)` are expanded using the container's environment. If a variable<br />cannot be resolved, the reference in the input string will be unchanged. Double `$$` are reduced<br />to a single `$`, which allows for escaping the `$(VAR_NAME)` syntax: i.e. `$$(VAR_NAME)` will<br />produce the string literal `$(VAR_NAME)`. Escaped references will never be expanded, regardless<br />of whether the variable exists or not. Cannot be updated.<br />More info: `https://kubernetes.io/docs/tasks/inject-data-application/define-command-argument-container/#running-a-command-in-a-shell`<br />Command field is from `corev1.Container.Command`. |  |  |
| `args` _string array_ | Arguments to the entrypoint.<br />The container image's CMD is used if this is not provided.<br />Variable references `$(VAR_NAME)` are expanded using the container's environment. If a variable<br />cannot be resolved, the reference in the input string will be unchanged. Double `$$` are reduced<br />to a single `$`, which allows for escaping the `$(VAR_NAME)` syntax: i.e. `$$(VAR_NAME)` will<br />produce the string literal `$(VAR_NAME)`. Escaped references will never be expanded, regardless<br />of whether the variable exists or not. Cannot be updated.<br />More info: `https://kubernetes.io/docs/tasks/inject-data-application/define-command-argument-container/#running-a-command-in-a-shell`<br />Args field is from `corev1.Container.Args`. |  |  |
| `workingDir` _string_ | Container's working directory.<br />If not specified, the container runtime's default will be used, which<br />might be configured in the container image.<br />Cannot be updated.<br />WorkingDir field is from `corev1.Container.WorkingDir`. |  |  |
| `env` _[EnvVar](https://kubernetes.io/docs/reference/generated/kubernetes-api/v/#envvar-v1-core) array_ | List of environment variables to set in the container.<br />Cannot be updated.<br />Env field is from `corev1.Container.Env`. |  |  |
| `livenessProbe` _[Probe](https://kubernetes.io/docs/reference/generated/kubernetes-api/v/#probe-v1-core)_ | Periodic probe of container liveness.<br />Container will be restarted if the probe fails.<br />More info: `https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes`<br />LivenessProbe field is from `corev1.Container.LivenessProbe`. |  |  |
| `readinessProbe` _[Probe](https://kubernetes.io/docs/reference/generated/kubernetes-api/v/#probe-v1-core)_ | Periodic probe of container service readiness.<br />Container will be removed from service endpoints if the probe fails.<br />ReadinessProbe field is from `corev1.Container.LivenessProbe`.<br />More info: `https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes` |  |  |
| `startupProbe` _[Probe](https://kubernetes.io/docs/reference/generated/kubernetes-api/v/#probe-v1-core)_ | StartupProbe indicates that the Pod has successfully initialized.<br />If specified, no other probes are executed until this completes successfully.<br />If this probe fails, the Pod will be restarted, just as if the livenessProbe failed.<br />This can be used to provide different probe parameters at the beginning of a Pod's lifecycle,<br />when it might take a long time to load data or warm a cache, than during steady-state operation.<br />This cannot be updated.<br />More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes |  |  |
| `lifecycle` _[Lifecycle](https://kubernetes.io/docs/reference/generated/kubernetes-api/v/#lifecycle-v1-core)_ | Actions that the management system should take in response to container lifecycle events.<br />Cannot be updated.<br />Lifecycle field is from `corev1.Container.Lifecycle`. |  |  |
| `imagePullPolicy` _[PullPolicy](https://kubernetes.io/docs/reference/generated/kubernetes-api/v/#pullpolicy-v1-core)_ | Image pull policy.<br />One of `Always`, `Never`, `IfNotPresent`.<br />Defaults to `Always` if `:latest` tag is specified, or IfNotPresent otherwise.<br />Cannot be updated.<br />More info: `https://kubernetes.io/docs/concepts/containers/images#updating-images`<br />ImagePullPolicy field is from `corev1.Container.ImagePullPolicy`. |  |  |
| `volumeMounts` _[VolumeMount](https://kubernetes.io/docs/reference/generated/kubernetes-api/v/#volumemount-v1-core) array_ | Pod volumes to mount into the container's filesystem.<br />Cannot be updated. |  |  |
| `securityContext` _[SecurityContext](https://kubernetes.io/docs/reference/generated/kubernetes-api/v/#securitycontext-v1-core)_ | SecurityContext holds container-level security attributes and common settings. |  |  |


#### MetaSpec



MetaSpec is the specification for meta component.



_Appears in:_
- [GreptimeDBClusterSpec](#greptimedbclusterspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `replicas` _integer_ | The number of replicas of the components. |  | Minimum: 0 <br /> |
| `config` _string_ | The content of the configuration file of the component in TOML format. |  |  |
| `template` _[PodTemplateSpec](#podtemplatespec)_ | Template defines the pod template for the component, if not specified, the pod template will use the default value. |  |  |
| `logging` _[LoggingSpec](#loggingspec)_ | Logging defines the logging configuration for the component. |  |  |
| `rpcPort` _integer_ | RPCPort is the gRPC port of the meta. |  | Maximum: 65535 <br />Minimum: 0 <br /> |
| `httpPort` _integer_ | HTTPPort is the HTTP port of the meta. |  | Maximum: 65535 <br />Minimum: 0 <br /> |
| `etcdEndpoints` _string array_ | EtcdEndpoints is the endpoints of the etcd cluster. |  |  |
| `enableCheckEtcdService` _boolean_ | EnableCheckEtcdService indicates whether to check etcd cluster health when starting meta. |  |  |
| `enableRegionFailover` _boolean_ | EnableRegionFailover indicates whether to enable region failover. |  |  |
| `storeKeyPrefix` _string_ | StoreKeyPrefix is the prefix of the key in the etcd. We can use it to isolate the data of different clusters. |  |  |
| `rollingUpdate` _[RollingUpdateDeployment](https://kubernetes.io/docs/reference/generated/kubernetes-api/v/#rollingupdatedeployment-v1-apps)_ | RollingUpdate is the rolling update configuration. We always use `RollingUpdate` strategyt. |  |  |


#### MetaStatus



MetaStatus is the status of meta node.



_Appears in:_
- [GreptimeDBClusterStatus](#greptimedbclusterstatus)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `replicas` _integer_ | Replicas is the number of replicas of the meta. |  |  |
| `readyReplicas` _integer_ | ReadyReplicas is the number of ready replicas of the meta. |  |  |
| `etcdEndpoints` _string array_ | EtcdEndpoints is the endpoints of the etcd cluster. |  |  |


#### MonitoringSpec



MonitoringSpec is the specification for monitor bootstrapping. It will create a standalone greptimedb instance to monitor the cluster.



_Appears in:_
- [GreptimeDBClusterSpec](#greptimedbclusterspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `enabled` _boolean_ | Enabled indicates whether to enable the monitoring service. |  |  |
| `standalone` _[GreptimeDBStandaloneSpec](#greptimedbstandalonespec)_ | The specification of the standalone greptimedb instance. |  |  |
| `logsCollection` _[LogsCollectionSpec](#logscollectionspec)_ | The specification of cluster logs collection. |  |  |
| `vector` _[VectorSpec](#vectorspec)_ | The specification of the vector instance. |  |  |


#### MonitoringStatus



MonitoringStatus is the status of the monitoring service.



_Appears in:_
- [GreptimeDBClusterStatus](#greptimedbclusterstatus)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `internalDNSName` _string_ | InternalDNSName is the internal DNS name of the monitoring service. For example, 'mycluster-standalone-monitor.default.svc.cluster.local'. |  |  |


#### OSSStorage



OSSStorage defines the Aliyun OSS storage specification.



_Appears in:_
- [ObjectStorageProviderSpec](#objectstorageproviderspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `bucket` _string_ | The data will be stored in the bucket. |  |  |
| `region` _string_ | The region of the bucket. |  |  |
| `secretName` _string_ | The secret of storing the credentials of access key id and secret access key.<br />The secret should contain keys named `access-key-id` and `secret-access-key`.<br />The secret must be the same namespace with the GreptimeDBCluster resource. |  |  |
| `root` _string_ | The OSS directory path. |  |  |
| `endpoint` _string_ | The endpoint of the bucket. |  |  |




#### ObjectStorageProviderSpec



ObjectStorageProviderSpec defines the object storage provider for the cluster. The data will be stored in the storage.



_Appears in:_
- [GreptimeDBClusterSpec](#greptimedbclusterspec)
- [GreptimeDBStandaloneSpec](#greptimedbstandalonespec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `s3` _[S3Storage](#s3storage)_ | S3 is the AWS S3 storage configuration. |  |  |
| `oss` _[OSSStorage](#ossstorage)_ | OSS is the Aliyun OSS storage configuration. |  |  |
| `gcs` _[GCSStorage](#gcsstorage)_ | GCS is the Google cloud storage configuration. |  |  |
| `azblob` _[AZBlobStorage](#azblobstorage)_ | AZBlob is the Azure Blob storage configuration. |  |  |
| `cache` _[CacheStorage](#cachestorage)_ | Cache is the cache storage configuration for object storage. |  |  |


#### Phase

_Underlying type:_ _string_

Phase define the phase of the cluster or standalone.



_Appears in:_
- [GreptimeDBClusterStatus](#greptimedbclusterstatus)
- [GreptimeDBStandaloneStatus](#greptimedbstandalonestatus)

| Field | Description |
| --- | --- |
| `Starting` | PhaseStarting means the controller start to create cluster or standalone.<br /> |
| `Running` | PhaseRunning means all the components of cluster or standalone is ready.<br /> |
| `Updating` | PhaseUpdating means the cluster or standalone is updating.<br /> |
| `Error` | PhaseError means some kind of error happen in reconcile.<br /> |
| `Terminating` | PhaseTerminating means the cluster or standalone is terminating.<br /> |


#### PodTemplateSpec



PodTemplateSpec defines the template for a pod of cluster.



_Appears in:_
- [ComponentSpec](#componentspec)
- [DatanodeSpec](#datanodespec)
- [FlownodeSpec](#flownodespec)
- [FrontendSpec](#frontendspec)
- [GreptimeDBClusterSpec](#greptimedbclusterspec)
- [GreptimeDBStandaloneSpec](#greptimedbstandalonespec)
- [MetaSpec](#metaspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `annotations` _object (keys:string, values:string)_ | The annotations to be created to the pod. |  |  |
| `labels` _object (keys:string, values:string)_ | The labels to be created to the pod. |  |  |
| `main` _[MainContainerSpec](#maincontainerspec)_ | MainContainer defines the specification of the main container of the pod. |  |  |
| `nodeSelector` _object (keys:string, values:string)_ | NodeSelector is a selector which must be true for the pod to fit on a node.<br />Selector which must match a node's labels for the pod to be scheduled on that node.<br />More info: `https://kubernetes.io/docs/concepts/configuration/assign-pod-node/`<br />NodeSelector field is from `corev1.PodSpec.NodeSelector`. |  |  |
| `initContainers` _[Container](https://kubernetes.io/docs/reference/generated/kubernetes-api/v/#container-v1-core) array_ | List of initialization containers belonging to the pod.<br />Init containers are executed in order prior to containers being started. If any<br />init container fails, the pod is considered to have failed and is handled according<br />to its restartPolicy. The name for an init container or normal container must be<br />unique among all containers.<br />Init containers may not have Lifecycle actions, Readiness probes, Liveness probes, or Startup probes.<br />The resourceRequirements of an init container are taken into account during scheduling<br />by finding the highest request/limit for each resource type, and then using the max of<br />that value or the sum of the normal containers. Limits are applied to init containers<br />in a similar fashion.<br />Init containers cannot currently be added or removed.<br />Cannot be updated.<br />More info: `https://kubernetes.io/docs/concepts/workloads/pods/init-containers/`<br />InitContainers field is from `corev1.PodSpec.InitContainers`. |  |  |
| `restartPolicy` _[RestartPolicy](https://kubernetes.io/docs/reference/generated/kubernetes-api/v/#restartpolicy-v1-core)_ | Restart policy for all containers within the pod.<br />One of `Always`, `OnFailure`, `Never`.<br />Default to `Always`.<br />More info: `https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#restart-policy`<br />RestartPolicy field is from `corev1.PodSpec.RestartPolicy`. |  |  |
| `terminationGracePeriodSeconds` _integer_ | Optional duration in seconds the pod needs to terminate gracefully. May be decreased in delete request.<br />Value must be non-negative integer. The value zero indicates stop immediately via<br />the kill signal (no opportunity to shut down).<br />If this value is nil, the default grace period will be used instead.<br />The grace period is the duration in seconds after the processes running in the pod are sent<br />a termination signal and the time when the processes are forcibly halted with a kill signal.<br />Set this value longer than the expected cleanup time for your process.<br />Defaults to 30 seconds.<br />TerminationGracePeriodSeconds field is from `corev1.PodSpec.TerminationGracePeriodSeconds`. |  |  |
| `activeDeadlineSeconds` _integer_ | Optional duration in seconds the pod may be active on the node relative to<br />StartTime before the system will actively try to mark it failed and kill associated containers.<br />Value must be a positive integer.<br />ActiveDeadlineSeconds field is from `corev1.PodSpec.ActiveDeadlineSeconds`. |  |  |
| `dnsPolicy` _[DNSPolicy](https://kubernetes.io/docs/reference/generated/kubernetes-api/v/#dnspolicy-v1-core)_ | Set DNS policy for the pod.<br />Defaults to `ClusterFirst`.<br />Valid values are `ClusterFirstWithHostNet`, `ClusterFirst`, `Default` or `None`.<br />DNS parameters given in DNSConfig will be merged with the policy selected with DNSPolicy.<br />To have DNS options set along with hostNetwork, you have to specify DNS policy<br />explicitly to `ClusterFirstWithHostNet`.<br />DNSPolicy field is from `corev1.PodSpec.DNSPolicy`. |  |  |
| `serviceAccountName` _string_ | ServiceAccountName is the name of the ServiceAccount to use to run this pod.<br />More info: `https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/`<br />ServiceAccountName field is from `corev1.PodSpec.ServiceAccountName`. |  |  |
| `hostNetwork` _boolean_ | Host networking requested for this pod. Use the host's network namespace.<br />If this option is set, the ports that will be used must be specified.<br />Default to `false`.<br />HostNetwork field is from `corev1.PodSpec.HostNetwork`. |  |  |
| `imagePullSecrets` _[LocalObjectReference](https://kubernetes.io/docs/reference/generated/kubernetes-api/v/#localobjectreference-v1-core) array_ | ImagePullSecrets is an optional list of references to secrets in the same namespace to use for pulling any of the images used by this PodSpec.<br />If specified, these secrets will be passed to individual puller implementations for them to use.<br />More info: `https://kubernetes.io/docs/concepts/containers/images#specifying-imagepullsecrets-on-a-pod`<br />ImagePullSecrets field is from `corev1.PodSpec.ImagePullSecrets`. |  |  |
| `affinity` _[Affinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v/#affinity-v1-core)_ | If specified, the pod's scheduling constraints<br />Affinity field is from `corev1.PodSpec.Affinity`. |  |  |
| `tolerations` _[Toleration](https://kubernetes.io/docs/reference/generated/kubernetes-api/v/#toleration-v1-core) array_ | If specified, the pod's tolerations. |  |  |
| `schedulerName` _string_ | If specified, the pod will be dispatched by specified scheduler.<br />If not specified, the pod will be dispatched by default scheduler.<br />SchedulerName field is from `corev1.PodSpec.SchedulerName`. |  |  |
| `additionalContainers` _[Container](https://kubernetes.io/docs/reference/generated/kubernetes-api/v/#container-v1-core) array_ | For most time, there is one main container in a pod(`frontend`/`meta`/`datanode`/`flownode`).<br />If specified, additional containers will be added to the pod as sidecar containers. |  |  |
| `volumes` _[Volume](https://kubernetes.io/docs/reference/generated/kubernetes-api/v/#volume-v1-core) array_ | List of volumes that can be mounted by containers belonging to the pod. |  |  |
| `securityContext` _[PodSecurityContext](https://kubernetes.io/docs/reference/generated/kubernetes-api/v/#podsecuritycontext-v1-core)_ | SecurityContext holds pod-level security attributes and common container settings. |  |  |


#### PrometheusMonitorSpec



PrometheusMonitorSpec defines the PodMonitor configuration.



_Appears in:_
- [GreptimeDBClusterSpec](#greptimedbclusterspec)
- [GreptimeDBStandaloneSpec](#greptimedbstandalonespec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `enabled` _boolean_ | Enabled indicates whether the PodMonitor is enabled. |  |  |
| `labels` _object (keys:string, values:string)_ | Labels is the labels for the PodMonitor. |  |  |
| `interval` _string_ | Interval is the scape interval for the PodMonitor. |  |  |


#### RaftEngineWAL



RaftEngineWAL is the specification for local WAL that uses raft-engine.



_Appears in:_
- [WALProviderSpec](#walproviderspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `fs` _[FileStorage](#filestorage)_ | FileStorage is the file storage configuration for the raft-engine WAL.<br />If the file storage is not specified, WAL will use DatanodeStorageSpec. |  |  |


#### S3Storage



S3Storage defines the S3 storage specification.



_Appears in:_
- [ObjectStorageProviderSpec](#objectstorageproviderspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `bucket` _string_ | The data will be stored in the bucket. |  |  |
| `region` _string_ | The region of the bucket. |  |  |
| `secretName` _string_ | The secret of storing the credentials of access key id and secret access key.<br />The secret should contain keys named `access-key-id` and `secret-access-key`.<br />The secret must be the same namespace with the GreptimeDBCluster resource. |  |  |
| `root` _string_ | The S3 directory path. |  |  |
| `endpoint` _string_ | The endpoint of the bucket. |  |  |


#### ServiceSpec



ServiceSpec defines the service configuration for the component.



_Appears in:_
- [FrontendSpec](#frontendspec)
- [GreptimeDBStandaloneSpec](#greptimedbstandalonespec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `type` _[ServiceType](https://kubernetes.io/docs/reference/generated/kubernetes-api/v/#servicetype-v1-core)_ | Type is the type of the service. |  |  |
| `annotations` _object (keys:string, values:string)_ | Annotations is the annotations for the service. |  |  |
| `labels` _object (keys:string, values:string)_ | Labels is the labels for the service. |  |  |
| `loadBalancerClass` _string_ | LoadBalancerClass is the class of the load balancer. |  |  |


#### SlimPodSpec



SlimPodSpec is a slimmed down version of corev1.PodSpec.
Most of the fields in SlimPodSpec are copied from `corev1.PodSpec`.



_Appears in:_
- [PodTemplateSpec](#podtemplatespec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `nodeSelector` _object (keys:string, values:string)_ | NodeSelector is a selector which must be true for the pod to fit on a node.<br />Selector which must match a node's labels for the pod to be scheduled on that node.<br />More info: `https://kubernetes.io/docs/concepts/configuration/assign-pod-node/`<br />NodeSelector field is from `corev1.PodSpec.NodeSelector`. |  |  |
| `initContainers` _[Container](https://kubernetes.io/docs/reference/generated/kubernetes-api/v/#container-v1-core) array_ | List of initialization containers belonging to the pod.<br />Init containers are executed in order prior to containers being started. If any<br />init container fails, the pod is considered to have failed and is handled according<br />to its restartPolicy. The name for an init container or normal container must be<br />unique among all containers.<br />Init containers may not have Lifecycle actions, Readiness probes, Liveness probes, or Startup probes.<br />The resourceRequirements of an init container are taken into account during scheduling<br />by finding the highest request/limit for each resource type, and then using the max of<br />that value or the sum of the normal containers. Limits are applied to init containers<br />in a similar fashion.<br />Init containers cannot currently be added or removed.<br />Cannot be updated.<br />More info: `https://kubernetes.io/docs/concepts/workloads/pods/init-containers/`<br />InitContainers field is from `corev1.PodSpec.InitContainers`. |  |  |
| `restartPolicy` _[RestartPolicy](https://kubernetes.io/docs/reference/generated/kubernetes-api/v/#restartpolicy-v1-core)_ | Restart policy for all containers within the pod.<br />One of `Always`, `OnFailure`, `Never`.<br />Default to `Always`.<br />More info: `https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#restart-policy`<br />RestartPolicy field is from `corev1.PodSpec.RestartPolicy`. |  |  |
| `terminationGracePeriodSeconds` _integer_ | Optional duration in seconds the pod needs to terminate gracefully. May be decreased in delete request.<br />Value must be non-negative integer. The value zero indicates stop immediately via<br />the kill signal (no opportunity to shut down).<br />If this value is nil, the default grace period will be used instead.<br />The grace period is the duration in seconds after the processes running in the pod are sent<br />a termination signal and the time when the processes are forcibly halted with a kill signal.<br />Set this value longer than the expected cleanup time for your process.<br />Defaults to 30 seconds.<br />TerminationGracePeriodSeconds field is from `corev1.PodSpec.TerminationGracePeriodSeconds`. |  |  |
| `activeDeadlineSeconds` _integer_ | Optional duration in seconds the pod may be active on the node relative to<br />StartTime before the system will actively try to mark it failed and kill associated containers.<br />Value must be a positive integer.<br />ActiveDeadlineSeconds field is from `corev1.PodSpec.ActiveDeadlineSeconds`. |  |  |
| `dnsPolicy` _[DNSPolicy](https://kubernetes.io/docs/reference/generated/kubernetes-api/v/#dnspolicy-v1-core)_ | Set DNS policy for the pod.<br />Defaults to `ClusterFirst`.<br />Valid values are `ClusterFirstWithHostNet`, `ClusterFirst`, `Default` or `None`.<br />DNS parameters given in DNSConfig will be merged with the policy selected with DNSPolicy.<br />To have DNS options set along with hostNetwork, you have to specify DNS policy<br />explicitly to `ClusterFirstWithHostNet`.<br />DNSPolicy field is from `corev1.PodSpec.DNSPolicy`. |  |  |
| `serviceAccountName` _string_ | ServiceAccountName is the name of the ServiceAccount to use to run this pod.<br />More info: `https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/`<br />ServiceAccountName field is from `corev1.PodSpec.ServiceAccountName`. |  |  |
| `hostNetwork` _boolean_ | Host networking requested for this pod. Use the host's network namespace.<br />If this option is set, the ports that will be used must be specified.<br />Default to `false`.<br />HostNetwork field is from `corev1.PodSpec.HostNetwork`. |  |  |
| `imagePullSecrets` _[LocalObjectReference](https://kubernetes.io/docs/reference/generated/kubernetes-api/v/#localobjectreference-v1-core) array_ | ImagePullSecrets is an optional list of references to secrets in the same namespace to use for pulling any of the images used by this PodSpec.<br />If specified, these secrets will be passed to individual puller implementations for them to use.<br />More info: `https://kubernetes.io/docs/concepts/containers/images#specifying-imagepullsecrets-on-a-pod`<br />ImagePullSecrets field is from `corev1.PodSpec.ImagePullSecrets`. |  |  |
| `affinity` _[Affinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v/#affinity-v1-core)_ | If specified, the pod's scheduling constraints<br />Affinity field is from `corev1.PodSpec.Affinity`. |  |  |
| `tolerations` _[Toleration](https://kubernetes.io/docs/reference/generated/kubernetes-api/v/#toleration-v1-core) array_ | If specified, the pod's tolerations. |  |  |
| `schedulerName` _string_ | If specified, the pod will be dispatched by specified scheduler.<br />If not specified, the pod will be dispatched by default scheduler.<br />SchedulerName field is from `corev1.PodSpec.SchedulerName`. |  |  |
| `additionalContainers` _[Container](https://kubernetes.io/docs/reference/generated/kubernetes-api/v/#container-v1-core) array_ | For most time, there is one main container in a pod(`frontend`/`meta`/`datanode`/`flownode`).<br />If specified, additional containers will be added to the pod as sidecar containers. |  |  |
| `volumes` _[Volume](https://kubernetes.io/docs/reference/generated/kubernetes-api/v/#volume-v1-core) array_ | List of volumes that can be mounted by containers belonging to the pod. |  |  |
| `securityContext` _[PodSecurityContext](https://kubernetes.io/docs/reference/generated/kubernetes-api/v/#podsecuritycontext-v1-core)_ | SecurityContext holds pod-level security attributes and common container settings. |  |  |


#### SlowQuery



SlowQuery defines the slow query configuration. It only works for the datanode component.



_Appears in:_
- [LoggingSpec](#loggingspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `enabled` _boolean_ | Enabled indicates whether the slow query is enabled. |  |  |
| `threshold` _string_ | Threshold is the threshold of the slow query. Default to `10s`. |  | Pattern: `^([0-9]+(\.[0-9]+)?(ns\|us\|µs\|ms\|s\|m\|h))+$` <br /> |
| `sampleRatio` _string_ | SampleRatio is the sampling ratio of slow query log. The value should be in the range of (0, 1]. Default to `1.0`. |  | Pattern: `^(0?\.\d+\|1(\.0+)?)$` <br />Type: string <br /> |


#### StorageRetainPolicyType

_Underlying type:_ _string_

StorageRetainPolicyType is the type of the storage retain policy.



_Appears in:_
- [FileStorage](#filestorage)

| Field | Description |
| --- | --- |
| `Retain` | StorageRetainPolicyTypeRetain is the default options.<br />The storage(PVCs) will be retained when the cluster is deleted.<br /> |
| `Delete` | StorageRetainPolicyTypeDelete specify that the storage will be deleted when the associated StatefulSet delete.<br /> |


#### TLSSpec



TLSSpec defines the TLS configurations for the component.



_Appears in:_
- [FrontendSpec](#frontendspec)
- [GreptimeDBStandaloneSpec](#greptimedbstandalonespec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `secretName` _string_ | SecretName is the name of the secret that contains the TLS certificates.<br />The secret must be in the same namespace with the greptime resource.<br />The secret must contain keys named `tls.crt` and `tls.key`. |  |  |


#### VectorSpec



VectorSpec is the specification for vector instance.



_Appears in:_
- [MonitoringSpec](#monitoringspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `image` _string_ | The image of the vector instance. |  |  |
| `resources` _[ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v/#resourcerequirements-v1-core)_ | The resources of the vector instance. |  |  |


#### WALProviderSpec



WALProviderSpec defines the WAL provider for the cluster.



_Appears in:_
- [GreptimeDBClusterSpec](#greptimedbclusterspec)
- [GreptimeDBStandaloneSpec](#greptimedbstandalonespec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `raftEngine` _[RaftEngineWAL](#raftenginewal)_ | RaftEngineWAL is the specification for local WAL that uses raft-engine. |  |  |
| `kafka` _[KafkaWAL](#kafkawal)_ | KafkaWAL is the specification for remote WAL that uses Kafka. |  |  |


