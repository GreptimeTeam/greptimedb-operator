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

	return nil
}

// MergeTemplate merges the base template with the component's template.
func (in *GreptimeDBCluster) MergeTemplate() error {
	mergeFuncs := []func() error{
		in.mergeMetaTemplate,
		in.mergeDatanodeTemplate,
		in.mergeFlownodeTemplate,
	}

	if in.GetFrontendGroup() != nil {
		mergeFuncs = append(mergeFuncs, in.mergeFrontendGroupTemplate)
	} else {
		mergeFuncs = append(mergeFuncs, in.mergeFrontendTemplate)
	}

	for _, mergeFunc := range mergeFuncs {
		if err := mergeFunc(); err != nil {
			return err
		}
	}

	return nil
}

// MergeLogging merges the logging settings into the component's logging settings.
func (in *GreptimeDBCluster) MergeLogging() error {
	loggingSpecs := []*LoggingSpec{
		in.GetMeta().GetLogging(),
		in.GetDatanode().GetLogging(),
		in.GetFlownode().GetLogging(),
	}

	if in.GetFrontendGroup() != nil {
		for _, frontend := range in.GetFrontendGroup() {
			loggingSpecs = append(loggingSpecs, frontend.GetLogging())
		}
	} else {
		loggingSpecs = append(loggingSpecs, in.GetFrontend().GetLogging())
	}

	for _, logging := range loggingSpecs {
		if logging == nil {
			continue
		}
		if err := in.doMergeLogging(logging, in.GetLogging(), in.GetMonitoring().IsEnabled()); err != nil {
			return err
		}
	}

	return nil
}

func (in *GreptimeDBCluster) doMergeLogging(input, global *LoggingSpec, isEnableMonitoring bool) error {
	if input == nil || global == nil {
		return nil
	}

	if err := mergo.Merge(input, global.DeepCopy()); err != nil {
		return err
	}

	if isEnableMonitoring {
		// Set the default logging format to JSON if monitoring is enabled.
		input.Format = LogFormatJSON
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
		Datanode:       in.defaultDatanode(),
	}

	if in.GetFrontendGroup() == nil {
		defaultSpec.Frontend = in.defaultFrontend()
	} else {
		defaultSpec.FrontendGroup = in.defaultFrontendGroup()
	}

	if in.GetFlownode() != nil {
		defaultSpec.Flownode = in.defaultFlownodeSpec()
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
						corev1.ResourceCPU:    resource.MustParse(DefaultVectorCPURequest),
						corev1.ResourceMemory: resource.MustParse(DefaultVectorMemoryRequest),
					},
				},
			},
		}

		// Set the default logging format to JSON if monitoring is enabled.
		defaultSpec.Logging.Format = LogFormatJSON

		// Turn on the slow query log by default if monitoring is enabled.
		defaultSpec.Logging.SlowQuery = &SlowQuery{
			Enabled:     true,
			Threshold:   "10s",
			SampleRatio: "1.0",
		}
	}

	return defaultSpec
}

func (in *GreptimeDBCluster) defaultFrontend() *FrontendSpec {
	return &FrontendSpec{
		ComponentSpec: ComponentSpec{
			Template: &PodTemplateSpec{},
			Replicas: ptr.To(int32(DefaultReplicas)),
			Logging:  &LoggingSpec{},
		},
		RPCPort:        DefaultRPCPort,
		HTTPPort:       DefaultHTTPPort,
		MySQLPort:      DefaultMySQLPort,
		PostgreSQLPort: DefaultPostgreSQLPort,
		Service: &ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
		},
		RollingUpdate: defaultRollingUpdateForDeployment(),
	}
}

func (in *GreptimeDBCluster) defaultFrontendGroup() []*FrontendSpec {
	var frontendGroup []*FrontendSpec
	var (
		replicas       *int32
		rpcPort        = DefaultRPCPort
		httpPort       = DefaultHTTPPort
		mysqlPort      = DefaultMySQLPort
		postgresqlPort = DefaultPostgreSQLPort
		rollingUpdate  = defaultRollingUpdateForDeployment()
	)

	for _, frontend := range in.GetFrontendGroup() {
		if frontend.Replicas != nil {
			replicas = frontend.Replicas
		}
		if frontend.RPCPort != 0 {
			rpcPort = frontend.RPCPort
		}
		if frontend.HTTPPort != 0 {
			httpPort = frontend.HTTPPort
		}
		if frontend.MySQLPort != 0 {
			mysqlPort = frontend.MySQLPort
		}
		if frontend.PostgreSQLPort != 0 {
			postgresqlPort = frontend.PostgreSQLPort
		}
		if frontend.RollingUpdate != nil {
			rollingUpdate = frontend.RollingUpdate
		}
		frontendSpec := &FrontendSpec{
			Name: frontend.Name,
			ComponentSpec: ComponentSpec{
				Template: &PodTemplateSpec{},
				Replicas: replicas,
				Logging:  &LoggingSpec{},
			},
			RPCPort:        rpcPort,
			HTTPPort:       httpPort,
			MySQLPort:      mysqlPort,
			PostgreSQLPort: postgresqlPort,
			Service: &ServiceSpec{
				Type: corev1.ServiceTypeClusterIP,
			},
			RollingUpdate: rollingUpdate,
		}
		frontendGroup = append(frontendGroup, frontendSpec)
	}

	if err := mergo.Merge(&in.Spec.FrontendGroup, frontendGroup, mergo.WithSliceDeepCopy); err != nil {
		return nil
	}

	return frontendGroup
}

func (in *GreptimeDBCluster) defaultMeta() *MetaSpec {
	return &MetaSpec{
		ComponentSpec: ComponentSpec{
			Template: &PodTemplateSpec{},
			Replicas: ptr.To(int32(DefaultReplicas)),
			Logging:  &LoggingSpec{},
		},
		RPCPort:              DefaultMetaRPCPort,
		HTTPPort:             DefaultHTTPPort,
		EnableRegionFailover: ptr.To(false),
		RollingUpdate:        defaultRollingUpdateForDeployment(),
	}
}

func (in *GreptimeDBCluster) defaultDatanode() *DatanodeSpec {
	return &DatanodeSpec{
		ComponentSpec: ComponentSpec{
			Template: &PodTemplateSpec{},
			Replicas: ptr.To(int32(DefaultReplicas)),
			Logging:  &LoggingSpec{},
		},
		RPCPort:       DefaultRPCPort,
		HTTPPort:      DefaultHTTPPort,
		Storage:       defaultDatanodeStorage(),
		RollingUpdate: defaultRollingUpdateForStatefulSet(),
	}
}

func (in *GreptimeDBCluster) defaultFlownodeSpec() *FlownodeSpec {
	return &FlownodeSpec{
		ComponentSpec: ComponentSpec{
			Template: &PodTemplateSpec{},
			Replicas: ptr.To(int32(DefaultReplicas)),
			Logging:  &LoggingSpec{},
		},
		RPCPort: DefaultRPCPort,
	}
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
	if in.Spec.Frontend != nil {
		// Use DeepCopy to avoid the same pointer.
		if err := mergo.Merge(in.Spec.Frontend.Template, in.DeepCopy().Spec.Base); err != nil {
			return err
		}

		// Reconfigure the probe settings based on the HTTP port.
		in.Spec.Frontend.Template.MainContainer.StartupProbe.HTTPGet.Port = intstr.FromInt32(in.Spec.Frontend.HTTPPort)
		in.Spec.Frontend.Template.MainContainer.LivenessProbe.HTTPGet.Port = intstr.FromInt32(in.Spec.Frontend.HTTPPort)
		in.Spec.Frontend.Template.MainContainer.ReadinessProbe.HTTPGet.Port = intstr.FromInt32(in.Spec.Frontend.HTTPPort)
	}

	return nil
}

func (in *GreptimeDBCluster) mergeFrontendGroupTemplate() error {
	for _, frontend := range in.Spec.FrontendGroup {
		if frontend.Template == nil {
			frontend.Template = &PodTemplateSpec{}
		}
		if err := mergo.Merge(frontend.Template, in.DeepCopy().Spec.Base); err != nil {
			return err
		}

		frontend.Template.MainContainer.StartupProbe.HTTPGet.Port = intstr.FromInt32(frontend.HTTPPort)
		frontend.Template.MainContainer.LivenessProbe.HTTPGet.Port = intstr.FromInt32(frontend.HTTPPort)
		frontend.Template.MainContainer.ReadinessProbe.HTTPGet.Port = intstr.FromInt32(frontend.HTTPPort)
	}

	return nil
}

func (in *GreptimeDBCluster) mergeMetaTemplate() error {
	if in.Spec.Meta != nil {
		// Use DeepCopy to avoid the same pointer.
		if err := mergo.Merge(in.Spec.Meta.Template, in.DeepCopy().Spec.Base); err != nil {
			return err
		}

		// Reconfigure the probe settings based on the HTTP port.
		in.Spec.Meta.Template.MainContainer.StartupProbe.HTTPGet.Port = intstr.FromInt32(in.Spec.Meta.HTTPPort)
		in.Spec.Meta.Template.MainContainer.LivenessProbe.HTTPGet.Port = intstr.FromInt32(in.Spec.Meta.HTTPPort)
		in.Spec.Meta.Template.MainContainer.ReadinessProbe.HTTPGet.Port = intstr.FromInt32(in.Spec.Meta.HTTPPort)
	}

	return nil
}

func (in *GreptimeDBCluster) mergeDatanodeTemplate() error {
	if in.Spec.Datanode != nil {
		// Use DeepCopy to avoid the same pointer.
		if err := mergo.Merge(in.Spec.Datanode.Template, in.DeepCopy().Spec.Base); err != nil {
			return err
		}

		// Reconfigure the probe settings based on the HTTP port.
		in.Spec.Datanode.Template.MainContainer.StartupProbe.HTTPGet.Port = intstr.FromInt32(in.Spec.Datanode.HTTPPort)
		in.Spec.Datanode.Template.MainContainer.LivenessProbe.HTTPGet.Port = intstr.FromInt32(in.Spec.Datanode.HTTPPort)
		in.Spec.Datanode.Template.MainContainer.ReadinessProbe.HTTPGet.Port = intstr.FromInt32(in.Spec.Datanode.HTTPPort)
	}

	return nil
}

func (in *GreptimeDBCluster) mergeFlownodeTemplate() error {
	if in.Spec.Flownode != nil {
		// Use DeepCopy to avoid the same pointer.
		if err := mergo.Merge(in.Spec.Flownode.Template, in.DeepCopy().Spec.Base); err != nil {
			return err
		}

		// TODO(zyy17): The flownode does not need liveness probe and will be added in the future.
		in.Spec.Flownode.Template.MainContainer.StartupProbe = nil
		in.Spec.Flownode.Template.MainContainer.LivenessProbe = nil
		in.Spec.Flownode.Template.MainContainer.ReadinessProbe = nil
	}

	return nil
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
		Logging: &LoggingSpec{
			Level:              DefaultLoggingLevel,
			LogsDir:            DefaultLogsDir,
			Format:             LogFormatText,
			PersistentWithData: ptr.To(false),
			OnlyLogToStdout:    ptr.To(false),
		},
		DatanodeStorage: defaultDatanodeStorage(),
		RollingUpdate:   defaultRollingUpdateForStatefulSet(),
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
