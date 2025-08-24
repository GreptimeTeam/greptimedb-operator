// Copyright 2022 Greptime Team
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1alpha1

import (
	"fmt"
	"reflect"
	"strings"

	"dario.cat/mergo"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
)

// SetDefaults sets the default values for the GreptimeDBCluster.
func (in *GreptimeDBCluster) SetDefaults() error {
	if in == nil {
		return nil
	}

	// Set the version of the GreptimeDBClusterSpec if it is not set.
	in.Spec.Version = getVersionFromImage(in.GetBaseMainContainer().GetImage())

	// Merge the default settings into the GreptimeDBClusterSpec.
	if err := mergo.Merge(&in.Spec, in.defaultSpec(), mergo.WithTransformers(intOrStringTransformer{})); err != nil {
		return err
	}

	// Merge the default spec into the datanode groups and frontend groups.
	if err := in.mergeDefaultGroups(); err != nil {
		return err
	}

	return nil
}

// We need to execute another merge operation for slice struct because mergo still don't support to merge Slice without override(https://github.com/darccio/mergo/issues/233).
func (in *GreptimeDBCluster) mergeDefaultGroups() error {
	for _, datanodeGroup := range in.GetDatanodeGroups() {
		if err := mergo.Merge(datanodeGroup, in.defaultDatanode()); err != nil {
			return err
		}
	}

	for _, frontend := range in.GetFrontendGroups() {
		if err := mergo.Merge(frontend, in.defaultFrontend()); err != nil {
			return err
		}
	}

	return nil
}

// MergeWithBaseTemplate merges the base template with the component's template.
func (in *GreptimeDBCluster) MergeWithBaseTemplate() error {
	mergeFuncs := []func() error{
		in.mergeFrontendTemplate,
		in.mergeMetaTemplate,
		in.mergeDatanodeTemplate,
		in.mergeFlownodeTemplate,
	}

	for _, mergeFunc := range mergeFuncs {
		if err := mergeFunc(); err != nil {
			return err
		}
	}

	return nil
}

// MergeWithGlobalLogging will merge the base logging settings into the component's logging settings.
// If the component's logging settings is not set, it will be set to the base logging settings.
func (in *GreptimeDBCluster) MergeWithGlobalLogging() error {
	if in.GetLogging() == nil {
		return nil
	}

	for _, spec := range in.getAllLoggingSpecs() {
		if err := mergo.Merge(spec, in.GetLogging().DeepCopy()); err != nil {
			return err
		}

		if in.GetMonitoring().IsEnabled() {
			// Set the default logging format to JSON if monitoring is enabled.
			spec.Format = LogFormatJSON
		}
	}

	return nil
}

// MergeWithGlobalTracing will merge the global tracing settings into the component's tracing settings.
// If the component's tracing settings is not set, it will be set to the global tracing settings.
func (in *GreptimeDBCluster) MergeWithGlobalTracing() error {
	if in.GetTracing() == nil {
		return nil
	}

	for _, spec := range in.getAllTracingSpecs() {
		if err := mergo.Merge(spec, in.GetTracing().DeepCopy()); err != nil {
			return err
		}
	}

	return nil
}

func (in *GreptimeDBCluster) defaultSpec() *GreptimeDBClusterSpec {
	var defaultSpec = &GreptimeDBClusterSpec{
		Base: &PodTemplateSpec{
			MainContainer: &MainContainerSpec{
				StartupProbe:   defaultStartupProbe(),
				LivenessProbe:  defaultLivenessProbe(),
				ReadinessProbe: defaultReadinessProbe(),
			},
		},
		Initializer:    &InitializerSpec{Image: DefaultInitializerImage},
		HTTPPort:       DefaultHTTPPort,
		RPCPort:        DefaultRPCPort,
		MySQLPort:      DefaultMySQLPort,
		PostgreSQLPort: DefaultPostgreSQLPort,
		Version:        DefaultVersion,
		Meta:           in.defaultMeta(),
	}

	if in.GetFrontend() != nil {
		defaultSpec.Frontend = in.defaultFrontend()
	}

	for range in.GetFrontendGroups() {
		defaultSpec.FrontendGroups = append(defaultSpec.FrontendGroups, in.defaultFrontend())
	}

	if in.GetFlownode() != nil {
		defaultSpec.Flownode = in.defaultFlownodeSpec()
	}

	if in.GetDatanode() != nil {
		defaultSpec.Datanode = in.defaultDatanode()
	}

	for range in.GetDatanodeGroups() {
		defaultSpec.DatanodeGroups = append(defaultSpec.DatanodeGroups, in.defaultDatanode())
	}

	defaultSpec.Logging = defaultLogging()

	if in.GetMonitoring().IsEnabled() {
		defaultSpec.Monitoring = &MonitoringSpec{
			Standalone:     in.defaultMonitoringStandaloneSpec(),
			LogsCollection: &LogsCollectionSpec{},
			Vector: &VectorSpec{
				Image: DefaultVectorImage,
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse(DefaultVectorCPURequest),
						corev1.ResourceMemory: resource.MustParse(DefaultVectorMemoryRequest),
					},
					Limits: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse(DefaultVectorCPULimit),
						corev1.ResourceMemory: resource.MustParse(DefaultVectorMemoryLimit),
					},
				},
			},
		}

		// Set the default logging format to JSON if monitoring is enabled.
		defaultSpec.Logging.Format = LogFormatJSON

		// Set the default tracing configuration if monitoring is enabled.
		defaultSpec.Tracing = &TracingSpec{
			Enabled:     ptr.To(true),
			SampleRatio: "1.0",
			Endpoint: fmt.Sprintf("%s.%s:%d/v1/otlp",
				in.Name+"-monitor-"+string(StandaloneRoleKind),
				in.Namespace,
				DefaultHTTPPort,
			),
		}
	}

	return defaultSpec
}

func (in *GreptimeDBCluster) defaultFrontend() *FrontendSpec {
	defaultSpec := &FrontendSpec{
		ComponentSpec: ComponentSpec{
			Template: &PodTemplateSpec{},
			Logging:  &LoggingSpec{},
			Tracing:  &TracingSpec{},
			Internal: &Internal{},
		},
		RPCPort:        DefaultRPCPort,
		HTTPPort:       DefaultHTTPPort,
		MySQLPort:      DefaultMySQLPort,
		PostgreSQLPort: DefaultPostgreSQLPort,
		Service: &ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
		},
		RollingUpdate: defaultRollingUpdateForDeployment(),
		SlowQuery:     defaultSlowQuery(),
	}

	if in.GetFrontend().GetReplicas() == nil {
		defaultSpec.Replicas = ptr.To(int32(DefaultReplicas))
	}

	if internal := in.GetFrontend().GetInternal(); internal != nil && internal.IsEnabled() {
		defaultSpec.Internal.Enabled = true
		defaultSpec.Internal.RPCPort = DefaultInternalRPCPort
	}

	return defaultSpec
}

func (in *GreptimeDBCluster) defaultMeta() *MetaSpec {
	defaultSpec := &MetaSpec{
		ComponentSpec: ComponentSpec{
			Template: &PodTemplateSpec{},
			Logging:  &LoggingSpec{},
			Tracing:  &TracingSpec{},
		},
		RPCPort:              DefaultMetaRPCPort,
		HTTPPort:             DefaultHTTPPort,
		EnableRegionFailover: ptr.To(false),
		RollingUpdate:        defaultRollingUpdateForDeployment(),
	}

	if in.GetMeta().GetReplicas() == nil {
		defaultSpec.Replicas = ptr.To(int32(DefaultReplicas))
	}

	return defaultSpec
}

func (in *GreptimeDBCluster) defaultDatanode() *DatanodeSpec {
	defaultSpec := &DatanodeSpec{
		ComponentSpec: ComponentSpec{
			Template: &PodTemplateSpec{},
			Logging:  &LoggingSpec{},
			Tracing:  &TracingSpec{},
		},
		RPCPort:       DefaultRPCPort,
		HTTPPort:      DefaultHTTPPort,
		Storage:       defaultDatanodeStorage(),
		RollingUpdate: defaultRollingUpdateForStatefulSet(),
	}

	if in.GetDatanode().GetReplicas() == nil {
		defaultSpec.Replicas = ptr.To(int32(DefaultReplicas))
	}

	return defaultSpec
}

func (in *GreptimeDBCluster) defaultFlownodeSpec() *FlownodeSpec {
	defaultSpec := &FlownodeSpec{
		ComponentSpec: ComponentSpec{
			Template: &PodTemplateSpec{},
			Logging:  &LoggingSpec{},
			Tracing:  &TracingSpec{},
		},
		RPCPort:  DefaultRPCPort,
		HTTPPort: DefaultHTTPPort,
	}

	if in.GetFlownode().GetReplicas() == nil {
		defaultSpec.Replicas = ptr.To(int32(DefaultReplicas))
	}

	return defaultSpec
}

func (in *GreptimeDBCluster) defaultMonitoringStandaloneSpec() *GreptimeDBStandaloneSpec {
	standalone := new(GreptimeDBStandalone)
	standalone.Spec = *standalone.defaultSpec()

	if image := in.GetBaseMainContainer().GetImage(); image != "" {
		standalone.Spec.Base.MainContainer.Image = image
	} else {
		standalone.Spec.Base.MainContainer.Image = DefaultGreptimeDBImage
	}

	standalone.Spec.Version = getVersionFromImage(standalone.Spec.Base.MainContainer.Image)

	// For better performance and easy management, the monitoring standalone use file storage by default.
	standalone.Spec.DatanodeStorage = &DatanodeStorageSpec{
		DataHome: DefaultDataHome,
		FileStorage: &FileStorage{
			Name:                DefaultDatanodeFileStorageName,
			StorageSize:         DefaultDataSizeForMonitoring,
			MountPath:           DefaultDataHome,
			StorageRetainPolicy: DefaultStorageRetainPolicyType,
		},
	}

	return &standalone.Spec
}

func (in *GreptimeDBCluster) mergeFrontendTemplate() error {
	for _, frontend := range in.Spec.FrontendGroups {
		if err := in.doMergeWithBaseTemplate(frontend.Template, frontend.HTTPPort); err != nil {
			return err
		}
	}

	if frontend := in.GetFrontend(); frontend != nil {
		if err := in.doMergeWithBaseTemplate(frontend.Template, frontend.HTTPPort); err != nil {
			return err
		}
	}

	return nil
}

func (in *GreptimeDBCluster) mergeMetaTemplate() error {
	if meta := in.GetMeta(); meta != nil {
		if err := in.doMergeWithBaseTemplate(meta.Template, meta.HTTPPort); err != nil {
			return err
		}
	}

	return nil
}

func (in *GreptimeDBCluster) mergeDatanodeTemplate() error {
	for _, datanode := range in.GetDatanodeGroups() {
		if err := in.doMergeWithBaseTemplate(datanode.Template, datanode.HTTPPort); err != nil {
			return err
		}
	}

	if datanode := in.GetDatanode(); datanode != nil {
		if err := in.doMergeWithBaseTemplate(datanode.Template, datanode.HTTPPort); err != nil {
			return err
		}
	}

	return nil
}

func (in *GreptimeDBCluster) mergeFlownodeTemplate() error {
	if flownode := in.GetFlownode(); flownode != nil {
		if err := in.doMergeWithBaseTemplate(flownode.Template, flownode.HTTPPort); err != nil {
			return err
		}
	}

	return nil
}

// doMergeWithBaseTemplate merges the base template with the component's template.
// If the component's template is not set, it will be set to the base template.
func (in *GreptimeDBCluster) doMergeWithBaseTemplate(dst *PodTemplateSpec, port int32) error {
	if dst == nil {
		dst = &PodTemplateSpec{}
	}

	// Use DeepCopy to avoid the same pointer.
	if err := mergo.Merge(dst, in.DeepCopy().Spec.Base); err != nil {
		return err
	}

	// Reconfigure the probe settings based on the HTTP port.
	in.configureProbePort(dst, port)

	return nil
}

func (in *GreptimeDBCluster) configureProbePort(spec *PodTemplateSpec, port int32) {
	spec.MainContainer.StartupProbe.HTTPGet.Port = intstr.FromInt32(port)
	spec.MainContainer.LivenessProbe.HTTPGet.Port = intstr.FromInt32(port)
	spec.MainContainer.ReadinessProbe.HTTPGet.Port = intstr.FromInt32(port)
}

func (in *GreptimeDBStandalone) SetDefaults() error {
	if in == nil {
		return nil
	}

	in.Spec.Version = getVersionFromImage(in.GetBaseMainContainer().GetImage())

	if err := mergo.Merge(&in.Spec, in.defaultSpec(), mergo.WithTransformers(intOrStringTransformer{})); err != nil {
		return err
	}

	return nil
}

func (in *GreptimeDBStandalone) defaultSpec() *GreptimeDBStandaloneSpec {
	var defaultSpec = &GreptimeDBStandaloneSpec{
		Base: &PodTemplateSpec{
			MainContainer: &MainContainerSpec{
				StartupProbe:   defaultStartupProbe(),
				LivenessProbe:  defaultLivenessProbe(),
				ReadinessProbe: defaultReadinessProbe(),
			},
		},
		HTTPPort:       DefaultHTTPPort,
		RPCPort:        DefaultRPCPort,
		MySQLPort:      DefaultMySQLPort,
		PostgreSQLPort: DefaultPostgreSQLPort,
		Version:        DefaultVersion,
		Service: &ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
		},
		Logging:         defaultLogging(),
		DatanodeStorage: defaultDatanodeStorage(),
		RollingUpdate:   defaultRollingUpdateForStatefulSet(),
		SlowQuery:       defaultSlowQuery(),
	}

	return defaultSpec
}

func defaultDatanodeStorage() *DatanodeStorageSpec {
	return &DatanodeStorageSpec{
		DataHome: DefaultDataHome,
		FileStorage: &FileStorage{
			Name:                DefaultDatanodeFileStorageName,
			StorageSize:         DefaultDataSize,
			MountPath:           DefaultDataHome,
			StorageRetainPolicy: DefaultStorageRetainPolicyType,
		},
	}
}

func defaultLogging() *LoggingSpec {
	return &LoggingSpec{
		Level:              DefaultLoggingLevel,
		LogsDir:            DefaultLogsDir,
		Format:             LogFormatText,
		PersistentWithData: ptr.To(false),
		OnlyLogToStdout:    ptr.To(false),
	}
}

func defaultSlowQuery() *SlowQuery {
	return &SlowQuery{
		Enabled:     true,
		Threshold:   "30s",
		SampleRatio: "1.0",
		TTL:         "90d",
		RecordType:  SlowQueryRecordTypeSystemTable,
	}
}

func getVersionFromImage(imageURL string) string {
	tokens := strings.Split(imageURL, "/")
	if len(tokens) > 0 {
		imageTag := tokens[len(tokens)-1]
		tokens = strings.Split(imageTag, ":")
		if len(tokens) == 2 {
			return tokens[1]
		}
	}
	return DefaultVersion
}

func defaultStartupProbe() *corev1.Probe {
	// When StartupProbe is successful, the liveness probe and readiness probe will be enabled.
	return &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			HTTPGet: &corev1.HTTPGetAction{
				Path: DefaultHealthEndpoint,
				Port: intstr.FromInt32(DefaultHTTPPort),
			},
		},
		PeriodSeconds: 5,

		// The StartupProbe can try up to 60 * 5 = 300 seconds to start the container.
		// For some scenarios, the datanode may take a long time to start, so we set the failure threshold to 60.
		FailureThreshold: 60,
	}
}

func defaultLivenessProbe() *corev1.Probe {
	// If the liveness probe fails, the container will be restarted.
	return &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			HTTPGet: &corev1.HTTPGetAction{
				Path: DefaultHealthEndpoint,
				Port: intstr.FromInt32(DefaultHTTPPort),
			},
		},
		PeriodSeconds:    5,
		FailureThreshold: 10,
	}
}

func defaultReadinessProbe() *corev1.Probe {
	// If the readiness probe fails, the container will be removed from the service endpoints.
	return &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			HTTPGet: &corev1.HTTPGetAction{
				Path: DefaultHealthEndpoint,
				Port: intstr.FromInt32(DefaultHTTPPort),
			},
		},
		PeriodSeconds:    5,
		FailureThreshold: 10,
	}
}

// Same as the default rolling update strategy of Deployment.
func defaultRollingUpdateForDeployment() *appsv1.RollingUpdateDeployment {
	return &appsv1.RollingUpdateDeployment{
		MaxUnavailable: &intstr.IntOrString{Type: intstr.String, StrVal: "25%"},
		MaxSurge:       &intstr.IntOrString{Type: intstr.String, StrVal: "25%"},
	}
}

// Same as the default rolling update strategy of StatefulSet.
func defaultRollingUpdateForStatefulSet() *appsv1.RollingUpdateStatefulSetStrategy {
	return &appsv1.RollingUpdateStatefulSetStrategy{
		Partition: ptr.To(int32(0)),
		MaxUnavailable: &intstr.IntOrString{
			Type:   intstr.Int,
			IntVal: 1,
		},
	}
}

// This transformer handles merging of intstr.IntOrString values.
// The `Type` field in IntOrString is an int starting from 0, which means it would be considered "empty" during merging and get overwritten.
// We want to preserve the original Type of the destination value while only merging the actual int/string content.
type intOrStringTransformer struct{}

func (t intOrStringTransformer) Transformer(typ reflect.Type) func(dst, src reflect.Value) error {
	if typ != reflect.TypeOf(&intstr.IntOrString{}) {
		return nil
	}

	return func(dst, src reflect.Value) error {
		if dst.IsNil() || src.IsNil() {
			return nil
		}

		dstVal, srcVal := dst.Interface().(*intstr.IntOrString), src.Interface().(*intstr.IntOrString)

		// Don't override the type of dst.
		if dstVal.Type == intstr.Int {
			if dstVal.IntVal == 0 {
				dstVal.IntVal = srcVal.IntVal
			}
			dstVal.StrVal = ""
		}

		if dstVal.Type == intstr.String {
			if dstVal.StrVal == "" {
				dstVal.StrVal = srcVal.StrVal
			}
			dstVal.IntVal = 0
		}

		return nil
	}
}
