# GreptimeDB Operator API Reference

## Packages
- [greptime.io/v1alpha1](#greptimeiov1alpha1)


## greptime.io/v1alpha1


### Resource Types
- [GreptimeDBCluster](#greptimedbcluster)
- [GreptimeDBClusterList](#greptimedbclusterlist)
- [GreptimeDBStandalone](#greptimedbstandalone)
- [GreptimeDBStandaloneList](#greptimedbstandalonelist)





#### ComponentSpec



ComponentSpec is the common specification for all components(frontend/meta/datanode).



_Appears in:_
- [DatanodeSpec](#datanodespec)
- [FlownodeSpec](#flownodespec)
- [FrontendSpec](#frontendspec)
- [MetaSpec](#metaspec)

| Field | Description |
| --- | --- |
| `replicas` _integer_ | The number of replicas of the components. |  | Minimum: 0 <br /> |
| `config` _string_ | The content of the configuration file of the component in TOML format. |  |  |
| `template` _[PodTemplateSpec](#podtemplatespec)_ | Template defines the pod template for the component, if not specified, the pod template will use the default value. |  |  |


#### Condition



Condition describes the state of a deployment at a certain point.



_Appears in:_
- [GreptimeDBClusterStatus](#greptimedbclusterstatus)
- [GreptimeDBStandaloneStatus](#greptimedbstandalonestatus)

| Field | Description |
| --- | --- |
| `type` _[ConditionType](#conditiontype)_ | Type of deployment condition. |  |  |
| `lastUpdateTime` _[Time](https://kubernetes.io/docs/reference/generated/kubernetes-api/v/#time-v1-meta)_ | The last time this condition was updated. |  |  |
| `lastTransitionTime` _[Time](https://kubernetes.io/docs/reference/generated/kubernetes-api/v/#time-v1-meta)_ | Last time the condition transitioned from one status to another. |  |  |
| `reason` _string_ | The reason for the condition's last transition. |  |  |
| `message` _string_ | A human-readable message indicating details about the transition. |  |  |


#### ConditionType

_Underlying type:_ _string_





_Appears in:_
- [Condition](#condition)

| Field | Description |
| `Ready` | ConditionTypeReady indicates that the GreptimeDB cluster is ready to serve requests.<br />Every component in the cluster are all ready.<br /> |
| `Progressing` | ConditionTypeProgressing indicates that the GreptimeDB cluster is progressing.<br /> |


#### DatanodeSpec



DatanodeSpec is the specification for datanode component.



_Appears in:_
- [GreptimeDBClusterSpec](#greptimedbclusterspec)

| Field | Description |
| --- | --- |
| `replicas` _integer_ | The number of replicas of the components. |  | Minimum: 0 <br /> |
| `config` _string_ | The content of the configuration file of the component in TOML format. |  |  |
| `template` _[PodTemplateSpec](#podtemplatespec)_ | Template defines the pod template for the component, if not specified, the pod template will use the default value. |  |  |
| `rpcPort` _integer_ | The RPC port of the datanode. |  |  |
| `httpPort` _integer_ | The HTTP port of the datanode. |  |  |
| `storage` _[StorageSpec](#storagespec)_ | Storage is the storage specification for the datanode. |  |  |


#### DatanodeStatus







_Appears in:_
- [GreptimeDBClusterStatus](#greptimedbclusterstatus)

| Field | Description |
| --- | --- |
| `replicas` _integer_ |  |  |  |
| `readyReplicas` _integer_ |  |  |  |


#### FlownodeSpec



FlownodeSpec is the specification for flownode component.



_Appears in:_
- [GreptimeDBClusterSpec](#greptimedbclusterspec)

| Field | Description |
| --- | --- |
| `replicas` _integer_ | The number of replicas of the components. |  | Minimum: 0 <br /> |
| `config` _string_ | The content of the configuration file of the component in TOML format. |  |  |
| `template` _[PodTemplateSpec](#podtemplatespec)_ | Template defines the pod template for the component, if not specified, the pod template will use the default value. |  |  |
| `rpcPort` _integer_ | The gRPC port of the flownode. |  |  |


#### FlownodeStatus







_Appears in:_
- [GreptimeDBClusterStatus](#greptimedbclusterstatus)

| Field | Description |
| --- | --- |
| `replicas` _integer_ |  |  |  |
| `readyReplicas` _integer_ |  |  |  |


#### FrontendSpec



FrontendSpec is the specification for frontend component.



_Appears in:_
- [GreptimeDBClusterSpec](#greptimedbclusterspec)

| Field | Description |
| --- | --- |
| `replicas` _integer_ | The number of replicas of the components. |  | Minimum: 0 <br /> |
| `config` _string_ | The content of the configuration file of the component in TOML format. |  |  |
| `template` _[PodTemplateSpec](#podtemplatespec)_ | Template defines the pod template for the component, if not specified, the pod template will use the default value. |  |  |
| `service` _[ServiceSpec](#servicespec)_ |  |  |  |
| `tls` _[TLSSpec](#tlsspec)_ | The TLS configurations of the frontend. |  |  |


#### FrontendStatus







_Appears in:_
- [GreptimeDBClusterStatus](#greptimedbclusterstatus)

| Field | Description |
| --- | --- |
| `replicas` _integer_ |  |  |  |
| `readyReplicas` _integer_ |  |  |  |


#### GCSStorageProvider







_Appears in:_
- [ObjectStorageProvider](#objectstorageprovider)

| Field | Description |
| --- | --- |
| `bucket` _string_ | The data will be stored in the bucket. |  |  |
| `root` _string_ | The gcs directory path. |  |  |
| `scope` _string_ | The scope for gcs. |  |  |
| `endpoint` _string_ | The endpoint URI of gcs service. |  |  |
| `secretName` _string_ | The secret of storing Credentials for gcs service OAuth2 authentication.<br />The secret must be the same namespace with the GreptimeDBCluster resource. |  |  |


#### GreptimeDBCluster



GreptimeDBCluster is the Schema for the greptimedbclusters API



_Appears in:_
- [GreptimeDBClusterList](#greptimedbclusterlist)

| Field | Description |
| --- | --- |
| `apiVersion` _string_ | `greptime.io/v1alpha1` | | |
| `kind` _string_ | `GreptimeDBCluster` | | |
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `spec` _[GreptimeDBClusterSpec](#greptimedbclusterspec)_ |  |  |  |


#### GreptimeDBClusterList



GreptimeDBClusterList contains a list of GreptimeDBCluster





| Field | Description |
| --- | --- |
| `apiVersion` _string_ | `greptime.io/v1alpha1` | | |
| `kind` _string_ | `GreptimeDBClusterList` | | |
| `metadata` _[ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v/#listmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `items` _[GreptimeDBCluster](#greptimedbcluster) array_ |  |  |  |


#### GreptimeDBClusterSpec



GreptimeDBClusterSpec defines the desired state of GreptimeDBCluster



_Appears in:_
- [GreptimeDBCluster](#greptimedbcluster)

| Field | Description |
| --- | --- |
| `base` _[PodTemplateSpec](#podtemplatespec)_ | Base is the base pod template for all components and can be overridden by template of individual component. |  |  |
| `frontend` _[FrontendSpec](#frontendspec)_ | Frontend is the specification of frontend node. |  |  |
| `meta` _[MetaSpec](#metaspec)_ | Meta is the specification of meta node. |  |  |
| `datanode` _[DatanodeSpec](#datanodespec)_ | Datanode is the specification of datanode node. |  |  |
| `flownode` _[FlownodeSpec](#flownodespec)_ | Flownode is the specification of flownode node. |  |  |
| `httpPort` _integer_ |  |  |  |
| `rpcPort` _integer_ |  |  |  |
| `mysqlPort` _integer_ |  |  |  |
| `postgreSQLPort` _integer_ |  |  |  |
| `enableInfluxDBProtocol` _boolean_ |  |  |  |
| `prometheusMonitor` _[PrometheusMonitorSpec](#prometheusmonitorspec)_ |  |  |  |
| `version` _string_ | The version of greptimedb. |  |  |
| `initializer` _[InitializerSpec](#initializerspec)_ |  |  |  |
| `objectStorage` _[ObjectStorageProvider](#objectstorageprovider)_ |  |  |  |
| `remoteWal` _[RemoteWalProvider](#remotewalprovider)_ |  |  |  |




#### GreptimeDBStandalone



GreptimeDBStandalone is the Schema for the greptimedbstandalones API



_Appears in:_
- [GreptimeDBStandaloneList](#greptimedbstandalonelist)

| Field | Description |
| --- | --- |
| `apiVersion` _string_ | `greptime.io/v1alpha1` | | |
| `kind` _string_ | `GreptimeDBStandalone` | | |
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `spec` _[GreptimeDBStandaloneSpec](#greptimedbstandalonespec)_ |  |  |  |


#### GreptimeDBStandaloneList



GreptimeDBStandaloneList contains a list of GreptimeDBStandalone





| Field | Description |
| --- | --- |
| `apiVersion` _string_ | `greptime.io/v1alpha1` | | |
| `kind` _string_ | `GreptimeDBStandaloneList` | | |
| `metadata` _[ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v/#listmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `items` _[GreptimeDBStandalone](#greptimedbstandalone) array_ |  |  |  |


#### GreptimeDBStandaloneSpec



GreptimeDBStandaloneSpec defines the desired state of GreptimeDBStandalone



_Appears in:_
- [GreptimeDBStandalone](#greptimedbstandalone)

| Field | Description |
| --- | --- |
| `base` _[PodTemplateSpec](#podtemplatespec)_ | Base is the base pod template for all components and can be overridden by template of individual component. |  |  |
| `service` _[ServiceSpec](#servicespec)_ |  |  |  |
| `tls` _[TLSSpec](#tlsspec)_ | The TLS configurations of the greptimedb. |  |  |
| `httpPort` _integer_ |  |  |  |
| `rpcPort` _integer_ |  |  |  |
| `mysqlPort` _integer_ |  |  |  |
| `postgreSQLPort` _integer_ |  |  |  |
| `enableInfluxDBProtocol` _boolean_ |  |  |  |
| `prometheusMonitor` _[PrometheusMonitorSpec](#prometheusmonitorspec)_ |  |  |  |
| `version` _string_ | The version of greptimedb. |  |  |
| `initializer` _[InitializerSpec](#initializerspec)_ |  |  |  |
| `objectStorage` _[ObjectStorageProvider](#objectstorageprovider)_ |  |  |  |
| `localStorage` _[StorageSpec](#storagespec)_ |  |  |  |
| `remoteWal` _[RemoteWalProvider](#remotewalprovider)_ |  |  |  |
| `config` _string_ |  |  |  |




#### InitializerSpec



InitializerSpec is the init container to set up components configurations before running the container.



_Appears in:_
- [GreptimeDBClusterSpec](#greptimedbclusterspec)
- [GreptimeDBStandaloneSpec](#greptimedbstandalonespec)

| Field | Description |
| --- | --- |
| `image` _string_ |  |  |  |


#### KafkaRemoteWal



KafkaRemoteWal is the specification for remote WAL that uses Kafka.



_Appears in:_
- [RemoteWalProvider](#remotewalprovider)

| Field | Description |
| --- | --- |
| `brokerEndpoints` _string array_ |  |  |  |


#### MainContainerSpec



MainContainerSpec describes the specification of the main container of a pod.
Most of the fields of MainContainerSpec are from 'corev1.Container'.



_Appears in:_
- [PodTemplateSpec](#podtemplatespec)

| Field | Description |
| --- | --- |
| `image` _string_ | The main container image name of the component. |  |  |
| `resources` _[ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v/#resourcerequirements-v1-core)_ | The resource requirements of the main container. |  |  |
| `command` _string array_ | Entrypoint array. Not executed within a shell.<br />The container image's ENTRYPOINT is used if this is not provided.<br />Variable references $(VAR_NAME) are expanded using the container's environment. If a variable<br />cannot be resolved, the reference in the input string will be unchanged. Double $$ are reduced<br />to a single $, which allows for escaping the $(VAR_NAME) syntax: i.e. "$$(VAR_NAME)" will<br />produce the string literal "$(VAR_NAME)". Escaped references will never be expanded, regardless<br />of whether the variable exists or not. Cannot be updated.<br />More info: https://kubernetes.io/docs/tasks/inject-data-application/define-command-argument-container/#running-a-command-in-a-shell<br />Command field is from 'corev1.Container.Command'. |  |  |
| `args` _string array_ | Arguments to the entrypoint.<br />The container image's CMD is used if this is not provided.<br />Variable references $(VAR_NAME) are expanded using the container's environment. If a variable<br />cannot be resolved, the reference in the input string will be unchanged. Double $$ are reduced<br />to a single $, which allows for escaping the $(VAR_NAME) syntax: i.e. "$$(VAR_NAME)" will<br />produce the string literal "$(VAR_NAME)". Escaped references will never be expanded, regardless<br />of whether the variable exists or not. Cannot be updated.<br />More info: https://kubernetes.io/docs/tasks/inject-data-application/define-command-argument-container/#running-a-command-in-a-shell<br />Args field is from 'corev1.Container.Args'. |  |  |
| `workingDir` _string_ | Container's working directory.<br />If not specified, the container runtime's default will be used, which<br />might be configured in the container image.<br />Cannot be updated.<br />WorkingDir field is from 'corev1.Container.WorkingDir'. |  |  |
| `env` _[EnvVar](https://kubernetes.io/docs/reference/generated/kubernetes-api/v/#envvar-v1-core) array_ | List of environment variables to set in the container.<br />Cannot be updated.<br />Env field is from 'corev1.Container.Env'. |  |  |
| `livenessProbe` _[Probe](https://kubernetes.io/docs/reference/generated/kubernetes-api/v/#probe-v1-core)_ | Periodic probe of container liveness.<br />Container will be restarted if the probe fails.<br />More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes<br />LivenessProbe field is from 'corev1.Container.LivenessProbe'. |  |  |
| `readinessProbe` _[Probe](https://kubernetes.io/docs/reference/generated/kubernetes-api/v/#probe-v1-core)_ | Periodic probe of container service readiness.<br />Container will be removed from service endpoints if the probe fails.<br />ReadinessProbe field is from 'corev1.Container.LivenessProbe'.<br />More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes |  |  |
| `lifecycle` _[Lifecycle](https://kubernetes.io/docs/reference/generated/kubernetes-api/v/#lifecycle-v1-core)_ | Actions that the management system should take in response to container lifecycle events.<br />Cannot be updated.<br />Lifecycle field is from 'corev1.Container.Lifecycle'. |  |  |
| `imagePullPolicy` _[PullPolicy](https://kubernetes.io/docs/reference/generated/kubernetes-api/v/#pullpolicy-v1-core)_ | Image pull policy.<br />One of Always, Never, IfNotPresent.<br />Defaults to Always if :latest tag is specified, or IfNotPresent otherwise.<br />Cannot be updated.<br />More info: https://kubernetes.io/docs/concepts/containers/images#updating-images<br />ImagePullPolicy field is from 'corev1.Container.ImagePullPolicy'. |  |  |
| `volumeMounts` _[VolumeMount](https://kubernetes.io/docs/reference/generated/kubernetes-api/v/#volumemount-v1-core) array_ | Pod volumes to mount into the container's filesystem.<br />Cannot be updated. |  |  |


#### MetaSpec



MetaSpec is the specification for meta component.



_Appears in:_
- [GreptimeDBClusterSpec](#greptimedbclusterspec)

| Field | Description |
| --- | --- |
| `replicas` _integer_ | The number of replicas of the components. |  | Minimum: 0 <br /> |
| `config` _string_ | The content of the configuration file of the component in TOML format. |  |  |
| `template` _[PodTemplateSpec](#podtemplatespec)_ | Template defines the pod template for the component, if not specified, the pod template will use the default value. |  |  |
| `rpcPort` _integer_ | The RPC port of the meta. |  |  |
| `httpPort` _integer_ | The HTTP port of the meta. |  |  |
| `etcdEndpoints` _string array_ |  |  |  |
| `enableCheckEtcdService` _boolean_ | EnableCheckEtcdService indicates whether to check etcd cluster health when starting meta. |  |  |
| `enableRegionFailover` _boolean_ | EnableRegionFailover indicates whether to enable region failover. |  |  |
| `storeKeyPrefix` _string_ | The meta will store data with this key prefix. |  |  |


#### MetaStatus







_Appears in:_
- [GreptimeDBClusterStatus](#greptimedbclusterstatus)

| Field | Description |
| --- | --- |
| `replicas` _integer_ |  |  |  |
| `readyReplicas` _integer_ |  |  |  |
| `etcdEndpoints` _string array_ |  |  |  |


#### OSSStorageProvider







_Appears in:_
- [ObjectStorageProvider](#objectstorageprovider)

| Field | Description |
| --- | --- |
| `bucket` _string_ | The data will be stored in the bucket. |  |  |
| `region` _string_ | The region of the bucket. |  |  |
| `endpoint` _string_ | The endpoint of the bucket. |  |  |
| `secretName` _string_ | The secret of storing the credentials of access key id and secret access key.<br />The secret must be the same namespace with the GreptimeDBCluster resource. |  |  |
| `root` _string_ | The OSS directory path. |  |  |


#### ObjectStorageProvider



ObjectStorageProvider defines the storage provider for the cluster. The data will be stored in the storage.



_Appears in:_
- [GreptimeDBClusterSpec](#greptimedbclusterspec)
- [GreptimeDBStandaloneSpec](#greptimedbstandalonespec)

| Field | Description |
| --- | --- |
| `s3` _[S3StorageProvider](#s3storageprovider)_ |  |  |  |
| `oss` _[OSSStorageProvider](#ossstorageprovider)_ |  |  |  |
| `gcs` _[GCSStorageProvider](#gcsstorageprovider)_ |  |  |  |
| `cachePath` _string_ |  |  |  |
| `cacheCapacity` _string_ |  |  |  |


#### Phase

_Underlying type:_ _string_

Phase define the phase of the cluster or standalone.



_Appears in:_
- [GreptimeDBClusterStatus](#greptimedbclusterstatus)
- [GreptimeDBStandaloneStatus](#greptimedbstandalonestatus)

| Field | Description |
| `Starting` | PhaseStarting means the controller start to create cluster.<br /> |
| `Running` | PhaseRunning means all the components of cluster is ready.<br /> |
| `Updating` | PhaseUpdating means the cluster is updating.<br /> |
| `Error` | PhaseError means some kind of error happen in reconcile.<br /> |
| `Terminating` | PhaseTerminating means the cluster is terminating.<br /> |


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

| Field | Description |
| --- | --- |
| `annotations` _object (keys:string, values:string)_ | The annotations to be created to the pod. |  |  |
| `labels` _object (keys:string, values:string)_ | The labels to be created to the pod. |  |  |
| `main` _[MainContainerSpec](#maincontainerspec)_ | MainContainer defines the specification of the main container of the pod. |  |  |
| `nodeSelector` _object (keys:string, values:string)_ | NodeSelector is a selector which must be true for the pod to fit on a node.<br />Selector which must match a node's labels for the pod to be scheduled on that node.<br />More info: https://kubernetes.io/docs/concepts/configuration/assign-pod-node/<br />NodeSelector field is from 'corev1.PodSpec.NodeSelector'. |  |  |
| `initContainers` _[Container](https://kubernetes.io/docs/reference/generated/kubernetes-api/v/#container-v1-core) array_ | List of initialization containers belonging to the pod.<br />Init containers are executed in order prior to containers being started. If any<br />init container fails, the pod is considered to have failed and is handled according<br />to its restartPolicy. The name for an init container or normal container must be<br />unique among all containers.<br />Init containers may not have Lifecycle actions, Readiness probes, Liveness probes, or Startup probes.<br />The resourceRequirements of an init container are taken into account during scheduling<br />by finding the highest request/limit for each resource type, and then using the max of<br />that value or the sum of the normal containers. Limits are applied to init containers<br />in a similar fashion.<br />Init containers cannot currently be added or removed.<br />Cannot be updated.<br />More info: https://kubernetes.io/docs/concepts/workloads/pods/init-containers/<br />InitContainers field is from 'corev1.PodSpec.InitContainers'. |  |  |
| `restartPolicy` _[RestartPolicy](https://kubernetes.io/docs/reference/generated/kubernetes-api/v/#restartpolicy-v1-core)_ | Restart policy for all containers within the pod.<br />One of Always, OnFailure, Never.<br />Default to Always.<br />More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#restart-policy<br />RestartPolicy field is from 'corev1.PodSpec.RestartPolicy'. |  |  |
| `terminationGracePeriodSeconds` _integer_ | Optional duration in seconds the pod needs to terminate gracefully. May be decreased in delete request.<br />Value must be non-negative integer. The value zero indicates stop immediately via<br />the kill signal (no opportunity to shut down).<br />If this value is nil, the default grace period will be used instead.<br />The grace period is the duration in seconds after the processes running in the pod are sent<br />a termination signal and the time when the processes are forcibly halted with a kill signal.<br />Set this value longer than the expected cleanup time for your process.<br />Defaults to 30 seconds.<br />TerminationGracePeriodSeconds field is from 'corev1.PodSpec.TerminationGracePeriodSeconds'. |  |  |
| `activeDeadlineSeconds` _integer_ | Optional duration in seconds the pod may be active on the node relative to<br />StartTime before the system will actively try to mark it failed and kill associated containers.<br />Value must be a positive integer.<br />ActiveDeadlineSeconds field is from 'corev1.PodSpec.ActiveDeadlineSeconds'. |  |  |
| `dnsPolicy` _[DNSPolicy](https://kubernetes.io/docs/reference/generated/kubernetes-api/v/#dnspolicy-v1-core)_ | Set DNS policy for the pod.<br />Defaults to "ClusterFirst".<br />Valid values are 'ClusterFirstWithHostNet', 'ClusterFirst', 'Default' or 'None'.<br />DNS parameters given in DNSConfig will be merged with the policy selected with DNSPolicy.<br />To have DNS options set along with hostNetwork, you have to specify DNS policy<br />explicitly to 'ClusterFirstWithHostNet'.<br />DNSPolicy field is from 'corev1.PodSpec.DNSPolicy'. |  |  |
| `serviceAccountName` _string_ | ServiceAccountName is the name of the ServiceAccount to use to run this pod.<br />More info: https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/<br />ServiceAccountName field is from 'corev1.PodSpec.ServiceAccountName'. |  |  |
| `hostNetwork` _boolean_ | Host networking requested for this pod. Use the host's network namespace.<br />If this option is set, the ports that will be used must be specified.<br />Default to false.<br />HostNetwork field is from 'corev1.PodSpec.HostNetwork'. |  |  |
| `imagePullSecrets` _[LocalObjectReference](https://kubernetes.io/docs/reference/generated/kubernetes-api/v/#localobjectreference-v1-core) array_ | ImagePullSecrets is an optional list of references to secrets in the same namespace to use for pulling any of the images used by this PodSpec.<br />If specified, these secrets will be passed to individual puller implementations for them to use.<br />More info: https://kubernetes.io/docs/concepts/containers/images#specifying-imagepullsecrets-on-a-pod<br />ImagePullSecrets field is from 'corev1.PodSpec.ImagePullSecrets'. |  |  |
| `affinity` _[Affinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v/#affinity-v1-core)_ | If specified, the pod's scheduling constraints<br />Affinity field is from 'corev1.PodSpec.Affinity'. |  |  |
| `tolerations` _[Toleration](https://kubernetes.io/docs/reference/generated/kubernetes-api/v/#toleration-v1-core) array_ | If specified, the pod's tolerations. |  |  |
| `schedulerName` _string_ | If specified, the pod will be dispatched by specified scheduler.<br />If not specified, the pod will be dispatched by default scheduler.<br />SchedulerName field is from 'corev1.PodSpec.SchedulerName'. |  |  |
| `additionalContainers` _[Container](https://kubernetes.io/docs/reference/generated/kubernetes-api/v/#container-v1-core) array_ | For most time, there is one main container in a pod(frontend/meta/datanode).<br />If specified, additional containers will be added to the pod as sidecar containers. |  |  |
| `volumes` _[Volume](https://kubernetes.io/docs/reference/generated/kubernetes-api/v/#volume-v1-core) array_ | List of volumes that can be mounted by containers belonging to the pod. |  |  |


#### PrometheusMonitorSpec



PrometheusMonitorSpec defines the PodMonitor configuration.



_Appears in:_
- [GreptimeDBClusterSpec](#greptimedbclusterspec)
- [GreptimeDBStandaloneSpec](#greptimedbstandalonespec)

| Field | Description |
| --- | --- |
| `enabled` _boolean_ | Enable a Prometheus PodMonitor |  |  |
| `labels` _object (keys:string, values:string)_ | Prometheus PodMonitor labels. |  |  |
| `interval` _string_ | Interval at which metrics should be scraped |  |  |


#### RemoteWalProvider



RemoteWalProvider defines the remote wal provider for the cluster.



_Appears in:_
- [GreptimeDBClusterSpec](#greptimedbclusterspec)
- [GreptimeDBStandaloneSpec](#greptimedbstandalonespec)

| Field | Description |
| --- | --- |
| `kafka` _[KafkaRemoteWal](#kafkaremotewal)_ |  |  |  |


#### S3StorageProvider







_Appears in:_
- [ObjectStorageProvider](#objectstorageprovider)

| Field | Description |
| --- | --- |
| `bucket` _string_ | The data will be stored in the bucket. |  |  |
| `region` _string_ | The region of the bucket. |  |  |
| `endpoint` _string_ | The endpoint of the bucket. |  |  |
| `secretName` _string_ | The secret of storing the credentials of access key id and secret access key.<br />The secret must be the same namespace with the GreptimeDBCluster resource. |  |  |
| `root` _string_ | The S3 directory path. |  |  |


#### ServiceSpec







_Appears in:_
- [FrontendSpec](#frontendspec)
- [GreptimeDBStandaloneSpec](#greptimedbstandalonespec)

| Field | Description |
| --- | --- |
| `type` _[ServiceType](https://kubernetes.io/docs/reference/generated/kubernetes-api/v/#servicetype-v1-core)_ | type determines how the Service is exposed. |  |  |
| `annotations` _object (keys:string, values:string)_ | Additional annotations for the service |  |  |
| `labels` _object (keys:string, values:string)_ | Additional labels for the service |  |  |
| `loadBalancerClass` _string_ | loadBalancerClass is the class of the load balancer implementation this Service belongs to. |  |  |


#### SlimPodSpec



SlimPodSpec is a slimmed down version of corev1.PodSpec.
Most of the fields in SlimPodSpec are copied from corev1.PodSpec.



_Appears in:_
- [PodTemplateSpec](#podtemplatespec)

| Field | Description |
| --- | --- |
| `nodeSelector` _object (keys:string, values:string)_ | NodeSelector is a selector which must be true for the pod to fit on a node.<br />Selector which must match a node's labels for the pod to be scheduled on that node.<br />More info: https://kubernetes.io/docs/concepts/configuration/assign-pod-node/<br />NodeSelector field is from 'corev1.PodSpec.NodeSelector'. |  |  |
| `initContainers` _[Container](https://kubernetes.io/docs/reference/generated/kubernetes-api/v/#container-v1-core) array_ | List of initialization containers belonging to the pod.<br />Init containers are executed in order prior to containers being started. If any<br />init container fails, the pod is considered to have failed and is handled according<br />to its restartPolicy. The name for an init container or normal container must be<br />unique among all containers.<br />Init containers may not have Lifecycle actions, Readiness probes, Liveness probes, or Startup probes.<br />The resourceRequirements of an init container are taken into account during scheduling<br />by finding the highest request/limit for each resource type, and then using the max of<br />that value or the sum of the normal containers. Limits are applied to init containers<br />in a similar fashion.<br />Init containers cannot currently be added or removed.<br />Cannot be updated.<br />More info: https://kubernetes.io/docs/concepts/workloads/pods/init-containers/<br />InitContainers field is from 'corev1.PodSpec.InitContainers'. |  |  |
| `restartPolicy` _[RestartPolicy](https://kubernetes.io/docs/reference/generated/kubernetes-api/v/#restartpolicy-v1-core)_ | Restart policy for all containers within the pod.<br />One of Always, OnFailure, Never.<br />Default to Always.<br />More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#restart-policy<br />RestartPolicy field is from 'corev1.PodSpec.RestartPolicy'. |  |  |
| `terminationGracePeriodSeconds` _integer_ | Optional duration in seconds the pod needs to terminate gracefully. May be decreased in delete request.<br />Value must be non-negative integer. The value zero indicates stop immediately via<br />the kill signal (no opportunity to shut down).<br />If this value is nil, the default grace period will be used instead.<br />The grace period is the duration in seconds after the processes running in the pod are sent<br />a termination signal and the time when the processes are forcibly halted with a kill signal.<br />Set this value longer than the expected cleanup time for your process.<br />Defaults to 30 seconds.<br />TerminationGracePeriodSeconds field is from 'corev1.PodSpec.TerminationGracePeriodSeconds'. |  |  |
| `activeDeadlineSeconds` _integer_ | Optional duration in seconds the pod may be active on the node relative to<br />StartTime before the system will actively try to mark it failed and kill associated containers.<br />Value must be a positive integer.<br />ActiveDeadlineSeconds field is from 'corev1.PodSpec.ActiveDeadlineSeconds'. |  |  |
| `dnsPolicy` _[DNSPolicy](https://kubernetes.io/docs/reference/generated/kubernetes-api/v/#dnspolicy-v1-core)_ | Set DNS policy for the pod.<br />Defaults to "ClusterFirst".<br />Valid values are 'ClusterFirstWithHostNet', 'ClusterFirst', 'Default' or 'None'.<br />DNS parameters given in DNSConfig will be merged with the policy selected with DNSPolicy.<br />To have DNS options set along with hostNetwork, you have to specify DNS policy<br />explicitly to 'ClusterFirstWithHostNet'.<br />DNSPolicy field is from 'corev1.PodSpec.DNSPolicy'. |  |  |
| `serviceAccountName` _string_ | ServiceAccountName is the name of the ServiceAccount to use to run this pod.<br />More info: https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/<br />ServiceAccountName field is from 'corev1.PodSpec.ServiceAccountName'. |  |  |
| `hostNetwork` _boolean_ | Host networking requested for this pod. Use the host's network namespace.<br />If this option is set, the ports that will be used must be specified.<br />Default to false.<br />HostNetwork field is from 'corev1.PodSpec.HostNetwork'. |  |  |
| `imagePullSecrets` _[LocalObjectReference](https://kubernetes.io/docs/reference/generated/kubernetes-api/v/#localobjectreference-v1-core) array_ | ImagePullSecrets is an optional list of references to secrets in the same namespace to use for pulling any of the images used by this PodSpec.<br />If specified, these secrets will be passed to individual puller implementations for them to use.<br />More info: https://kubernetes.io/docs/concepts/containers/images#specifying-imagepullsecrets-on-a-pod<br />ImagePullSecrets field is from 'corev1.PodSpec.ImagePullSecrets'. |  |  |
| `affinity` _[Affinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v/#affinity-v1-core)_ | If specified, the pod's scheduling constraints<br />Affinity field is from 'corev1.PodSpec.Affinity'. |  |  |
| `tolerations` _[Toleration](https://kubernetes.io/docs/reference/generated/kubernetes-api/v/#toleration-v1-core) array_ | If specified, the pod's tolerations. |  |  |
| `schedulerName` _string_ | If specified, the pod will be dispatched by specified scheduler.<br />If not specified, the pod will be dispatched by default scheduler.<br />SchedulerName field is from 'corev1.PodSpec.SchedulerName'. |  |  |
| `additionalContainers` _[Container](https://kubernetes.io/docs/reference/generated/kubernetes-api/v/#container-v1-core) array_ | For most time, there is one main container in a pod(frontend/meta/datanode).<br />If specified, additional containers will be added to the pod as sidecar containers. |  |  |
| `volumes` _[Volume](https://kubernetes.io/docs/reference/generated/kubernetes-api/v/#volume-v1-core) array_ | List of volumes that can be mounted by containers belonging to the pod. |  |  |


#### StorageRetainPolicyType

_Underlying type:_ _string_





_Appears in:_
- [StorageSpec](#storagespec)

| Field | Description |
| `Retain` | StorageRetainPolicyTypeRetain is the default options.<br />The storage(PVCs) will be retained when the cluster is deleted.<br /> |
| `Delete` | StorageRetainPolicyTypeDelete specify that the storage will be deleted when the associated StatefulSet delete.<br /> |


#### StorageSpec



StorageSpec will generate PVC.



_Appears in:_
- [DatanodeSpec](#datanodespec)
- [GreptimeDBStandaloneSpec](#greptimedbstandalonespec)

| Field | Description |
| --- | --- |
| `name` _string_ | The name of the storage. |  |  |
| `storageClassName` _string_ | The name of the storage class to use for the volume. |  |  |
| `storageSize` _string_ | The size of the storage. |  | Pattern: `(^([+-]?[0-9.]+)([eEinumkKMGTP]*[-+]?[0-9]*)$)` <br /> |
| `mountPath` _string_ | The mount path of the storage in datanode container. |  |  |
| `storageRetainPolicy` _[StorageRetainPolicyType](#storageretainpolicytype)_ | The PVCs will retain or delete when the cluster is deleted, default to Retain. |  | Enum: [Retain Delete] <br /> |
| `walDir` _string_ | The wal directory of the storage. |  |  |
| `dataHome` _string_ | The datahome directory. |  |  |


#### TLSSpec







_Appears in:_
- [FrontendSpec](#frontendspec)
- [GreptimeDBStandaloneSpec](#greptimedbstandalonespec)

| Field | Description |
| --- | --- |
| `secretName` _string_ | The secret name of the TLS certificate, and it must be in the same namespace of the cluster.<br />The secret must contain keys named ca.crt, tls.crt and tls.key. |  |  |


