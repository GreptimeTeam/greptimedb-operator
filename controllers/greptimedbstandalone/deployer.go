// Copyright 2024 Greptime Team
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

package greptimedbstandalone

import (
	"context"
	"fmt"
	"path"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/GreptimeTeam/greptimedb-operator/apis/v1alpha1"
	"github.com/GreptimeTeam/greptimedb-operator/controllers/common"
	"github.com/GreptimeTeam/greptimedb-operator/controllers/constant"
	"github.com/GreptimeTeam/greptimedb-operator/pkg/dbconfig"
	"github.com/GreptimeTeam/greptimedb-operator/pkg/deployer"
	"github.com/GreptimeTeam/greptimedb-operator/pkg/util"
	k8sutil "github.com/GreptimeTeam/greptimedb-operator/pkg/util/k8s"
)

type StandaloneDeployer struct {
	Scheme *runtime.Scheme

	client.Client
	deployer.DefaultDeployer
}

var _ deployer.Deployer = &StandaloneDeployer{}

func NewStandaloneDeployer(mgr ctrl.Manager) *StandaloneDeployer {
	return &StandaloneDeployer{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),

		DefaultDeployer: deployer.DefaultDeployer{
			Client: mgr.GetClient(),
		},
	}
}

func (s *StandaloneDeployer) NewBuilder(crdObject client.Object) deployer.Builder {
	sb := &standaloneBuilder{
		DefaultBuilder: &deployer.DefaultBuilder{
			Scheme: s.Scheme,
			Owner:  crdObject,
		},
	}

	standalone, err := s.getStandalone(crdObject)
	if err != nil {
		sb.Err = err
	}
	sb.standalone = standalone

	return sb
}

func (s *StandaloneDeployer) Generate(crdObject client.Object) ([]client.Object, error) {
	objects, err := s.NewBuilder(crdObject).
		BuildService().
		BuildConfigMap().
		BuildStatefulSet().
		SetControllerAndAnnotation().
		Generate()

	if err != nil {
		return nil, err
	}

	return objects, nil
}

func (s *StandaloneDeployer) CleanUp(ctx context.Context, crdObject client.Object) error {
	return nil
}

func (s *StandaloneDeployer) CheckAndUpdateStatus(ctx context.Context, crdObject client.Object) (bool, error) {
	standalone, err := s.getStandalone(crdObject)
	if err != nil {
		return false, err
	}

	var (
		sts = new(appsv1.StatefulSet)

		objectKey = client.ObjectKey{
			Namespace: standalone.Namespace,
			Name:      standalone.Name,
		}
	)

	err = s.Get(ctx, objectKey, sts)
	if errors.IsNotFound(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return k8sutil.IsStatefulSetReady(sts), nil
}

func (s *StandaloneDeployer) getStandalone(crdObject client.Object) (*v1alpha1.GreptimeDBStandalone, error) {
	standalone, ok := crdObject.(*v1alpha1.GreptimeDBStandalone)
	if !ok {
		return nil, fmt.Errorf("the object is not a GreptimeDBStandalone")
	}
	return standalone, nil
}

var _ deployer.Builder = &standaloneBuilder{}

type standaloneBuilder struct {
	standalone *v1alpha1.GreptimeDBStandalone
	*deployer.DefaultBuilder
}

func (s *standaloneBuilder) BuildService() deployer.Builder {
	if s.Err != nil {
		return s
	}

	svc := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   s.standalone.Namespace,
			Name:        common.ResourceName(s.standalone.Name, v1alpha1.StandaloneKind),
			Annotations: s.standalone.Spec.Service.Annotations,
			Labels:      s.standalone.Spec.Service.Labels,
		},
		Spec: corev1.ServiceSpec{
			Type: s.standalone.Spec.Service.Type,
			Selector: map[string]string{
				constant.GreptimeDBComponentName: s.standalone.Name,
			},
			Ports:             s.servicePorts(),
			LoadBalancerClass: s.standalone.Spec.Service.LoadBalancerClass,
		},
	}

	s.Objects = append(s.Objects, svc)

	return s
}

func (s *standaloneBuilder) BuildConfigMap() deployer.Builder {
	if s.Err != nil {
		return s
	}

	configData, err := dbconfig.FromStandalone(s.standalone)
	if err != nil {
		s.Err = err
		return s
	}

	cm, err := common.GenerateConfigMap(s.standalone.Namespace, s.standalone.Name, v1alpha1.StandaloneKind, configData)
	if err != nil {
		s.Err = err
		return s
	}

	s.Objects = append(s.Objects, cm)

	return s
}

func (s *standaloneBuilder) BuildStatefulSet() deployer.Builder {
	if s.Err != nil {
		return s
	}

	// Always set replicas to 1 for standalone mode.
	replicas := int32(1)

	sts := &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "StatefulSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      s.standalone.Name,
			Namespace: s.standalone.Namespace,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					constant.GreptimeDBComponentName: s.standalone.Name,
				},
			},
			Template:             s.generatePodTemplateSpec(),
			VolumeClaimTemplates: s.generatePVC(),
		},
	}

	if s.standalone.Spec.ReloadWhenConfigChange {
		sts.SetAnnotations(util.MergeStringMap(sts.GetAnnotations(),
			map[string]string{deployer.ConfigmapReloader: s.standalone.Name}))
	}

	s.Objects = append(s.Objects, sts)

	return s
}

func (s *standaloneBuilder) generatePodTemplateSpec() corev1.PodTemplateSpec {
	template := common.GeneratePodTemplateSpec(v1alpha1.StandaloneKind, s.standalone.Spec.Base)

	if len(s.standalone.Spec.Base.MainContainer.Args) == 0 {
		// Setup main container args.
		template.Spec.Containers[constant.MainContainerIndex].Args = s.generateMainContainerArgs()
	}

	// Add Storage Dir.
	template.Spec.Containers[constant.MainContainerIndex].VolumeMounts =
		append(template.Spec.Containers[constant.MainContainerIndex].VolumeMounts,
			corev1.VolumeMount{
				Name:      s.standalone.Spec.LocalStorage.Name,
				MountPath: s.standalone.Spec.LocalStorage.MountPath,
			},
		)

	template.Spec.Containers[constant.MainContainerIndex].Ports = s.containerPorts()
	template.ObjectMeta.Labels = util.MergeStringMap(template.ObjectMeta.Labels, map[string]string{
		constant.GreptimeDBComponentName: s.standalone.Name,
	})

	common.MountConfigDir(s.standalone.Name, v1alpha1.StandaloneKind, template)

	return *template
}

func (s *standaloneBuilder) generatePVC() []corev1.PersistentVolumeClaim {
	return []corev1.PersistentVolumeClaim{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: s.standalone.Spec.LocalStorage.Name,
			},
			Spec: corev1.PersistentVolumeClaimSpec{
				StorageClassName: s.standalone.Spec.LocalStorage.StorageClassName,
				AccessModes: []corev1.PersistentVolumeAccessMode{
					corev1.ReadWriteOnce,
				},
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceStorage: resource.MustParse(s.standalone.Spec.LocalStorage.StorageSize),
					},
				},
			},
		},
	}
}

func (s *standaloneBuilder) servicePorts() []corev1.ServicePort {
	return []corev1.ServicePort{
		{
			Name:     "grpc",
			Protocol: corev1.ProtocolTCP,
			Port:     s.standalone.Spec.GRPCServicePort,
		},
		{
			Name:     "http",
			Protocol: corev1.ProtocolTCP,
			Port:     s.standalone.Spec.HTTPServicePort,
		},
		{
			Name:     "mysql",
			Protocol: corev1.ProtocolTCP,
			Port:     s.standalone.Spec.MySQLServicePort,
		},
		{
			Name:     "postgres",
			Protocol: corev1.ProtocolTCP,
			Port:     s.standalone.Spec.PostgresServicePort,
		},
		{
			Name:     "opentsdb",
			Protocol: corev1.ProtocolTCP,
			Port:     s.standalone.Spec.OpenTSDBServicePort,
		},
	}
}

func (s *standaloneBuilder) containerPorts() []corev1.ContainerPort {
	return []corev1.ContainerPort{
		{
			Name:          "grpc",
			Protocol:      corev1.ProtocolTCP,
			ContainerPort: s.standalone.Spec.GRPCServicePort,
		},
		{
			Name:          "http",
			Protocol:      corev1.ProtocolTCP,
			ContainerPort: s.standalone.Spec.HTTPServicePort,
		},
		{
			Name:          "mysql",
			Protocol:      corev1.ProtocolTCP,
			ContainerPort: s.standalone.Spec.MySQLServicePort,
		},
		{
			Name:          "postgres",
			Protocol:      corev1.ProtocolTCP,
			ContainerPort: s.standalone.Spec.PostgresServicePort,
		},
		{
			Name:          "opentsdb",
			Protocol:      corev1.ProtocolTCP,
			ContainerPort: s.standalone.Spec.OpenTSDBServicePort,
		},
	}
}

func (s *standaloneBuilder) generateMainContainerArgs() []string {
	return []string{
		"standalone", "start",
		"--data-home", "/data",
		"--rpc-addr", fmt.Sprintf("0.0.0.0:%d", s.standalone.Spec.GRPCServicePort),
		"--mysql-addr", fmt.Sprintf("0.0.0.0:%d", s.standalone.Spec.MySQLServicePort),
		"--http-addr", fmt.Sprintf("0.0.0.0:%d", s.standalone.Spec.HTTPServicePort),
		"--postgres-addr", fmt.Sprintf("0.0.0.0:%d", s.standalone.Spec.PostgresServicePort),
		"--config-file", path.Join(constant.GreptimeDBConfigDir, constant.GreptimeDBConfigFileName),
	}
}
