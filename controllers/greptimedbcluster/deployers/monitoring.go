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

package deployers

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/fs"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/avast/retry-go"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/GreptimeTeam/greptimedb-operator/apis/v1alpha1"
	"github.com/GreptimeTeam/greptimedb-operator/controllers/common"
	"github.com/GreptimeTeam/greptimedb-operator/controllers/greptimedbcluster/deployers/config"
	"github.com/GreptimeTeam/greptimedb-operator/pkg/deployer"
)

type MonitoringDeployer struct {
	*CommonDeployer
}

var _ deployer.Deployer = &MonitoringDeployer{}

func NewMonitoringDeployer(mgr ctrl.Manager) *MonitoringDeployer {
	return &MonitoringDeployer{
		CommonDeployer: NewFromManager(mgr),
	}
}

func (d *MonitoringDeployer) NewBuilder(crdObject client.Object) deployer.Builder {
	return &monitoringBuilder{CommonBuilder: d.NewCommonBuilder(crdObject, v1alpha1.StandaloneKind)}
}

func (d *MonitoringDeployer) Generate(crdObject client.Object) ([]client.Object, error) {
	objects, err := d.NewBuilder(crdObject).
		BuildGreptimeDBStandalone().
		BuildConfigMap().
		SetControllerAndAnnotation().
		Generate()

	if err != nil {
		return nil, err
	}

	return objects, nil
}

func (d *MonitoringDeployer) CheckAndUpdateStatus(ctx context.Context, crdObject client.Object) (bool, error) {
	cluster, err := d.GetCluster(crdObject)
	if err != nil {
		return false, err
	}

	if !cluster.GetMonitoring().IsEnabled() || cluster.GetMonitoring().GetStandalone() == nil {
		return true, nil
	}

	var (
		standalone = new(v1alpha1.GreptimeDBStandalone)

		objectKey = client.ObjectKey{
			Namespace: cluster.Namespace,
			Name:      common.MonitoringServiceName(cluster.Name),
		}
	)

	err = d.Get(ctx, objectKey, standalone)
	if errors.IsNotFound(err) {
		return false, nil
	}

	if cluster.GetMonitoring().IsEnabled() && standalone.Status.StandalonePhase == v1alpha1.PhaseRunning {
		if err := d.createPipeline(cluster); err != nil {
			klog.Errorf("failed to create pipeline for standalone, err: '%v'", err)
			return false, err
		}

		cluster.Status.Monitoring.InternalDNSName = fmt.Sprintf("%s.%s.svc.cluster.local", common.ResourceName(common.MonitoringServiceName(cluster.Name), v1alpha1.StandaloneKind), cluster.Namespace)
		if err := UpdateStatus(ctx, cluster, d.Client); err != nil {
			klog.Errorf("Failed to update status: %s", err)
		}
		return true, nil
	}

	return false, nil
}

func (d *MonitoringDeployer) createPipeline(cluster *v1alpha1.GreptimeDBCluster) error {
	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	fw, err := w.CreateFormFile("file", "pipeline.yaml")
	if err != nil {
		return err
	}

	pipeline, err := d.defaultPipeline()
	if err != nil {
		return err
	}

	// If the pipeline is specified in the CR, use that instead.
	if p := cluster.GetMonitoring().GetLogsCollection().GetPipeline().GetData(); p != "" {
		pipeline = p
	}

	_, err = io.Copy(fw, strings.NewReader(pipeline))
	if err != nil {
		return err
	}
	w.Close()

	standaloneName := common.ResourceName(common.MonitoringServiceName(cluster.Name), v1alpha1.StandaloneKind)

	// FIXME(zyy17): Make the port configurable.
	svc := fmt.Sprintf("%s.%s.svc.cluster.local:%d", standaloneName, cluster.Namespace, v1alpha1.DefaultHTTPPort)
	hc := &http.Client{
		Timeout: 5 * time.Second,
	}

	operation := func() error {
		req, err := http.NewRequest("POST", fmt.Sprintf("http://%s/v1/events/pipelines/%s", svc, common.LogsPipelineName(cluster.Namespace, cluster.Name)), &b)
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", w.FormDataContentType())

		resp, err := hc.Do(req)
		if err != nil {
			klog.Warningf("failed to create pipeline: %v", err)
			return err
		}
		defer resp.Body.Close()

		responseBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("failed to create pipeline: '%s'", string(responseBody))
		}

		return nil
	}

	// The server may not be ready to accept the request, so we retry a few times.
	if err = retry.Do(
		operation,
		retry.Attempts(10),
		retry.Delay(500*time.Millisecond),
		retry.DelayType(retry.FixedDelay),
	); err != nil {
		return err
	}

	return nil
}

// defaultPipeline returns the default pipeline that will be used by the standalone greptimedb instance to collect greptimedb logs.
func (d *MonitoringDeployer) defaultPipeline() (string, error) {
	data, err := fs.ReadFile(config.DefaultPipeline, "pipeline.yaml")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

var _ deployer.Builder = &monitoringBuilder{}

type monitoringBuilder struct {
	*CommonBuilder
}

func (b *monitoringBuilder) BuildGreptimeDBStandalone() deployer.Builder {
	if !b.Cluster.GetMonitoring().IsEnabled() || b.Cluster.GetMonitoring().GetStandalone() == nil {
		return b
	}

	if b.Err != nil {
		return b
	}

	standalone := &v1alpha1.GreptimeDBStandalone{
		TypeMeta: metav1.TypeMeta{
			Kind:       "GreptimeDBStandalone",
			APIVersion: "greptime.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      common.MonitoringServiceName(b.Cluster.Name),
			Namespace: b.Cluster.Namespace,
		},
		Spec: *b.Cluster.GetMonitoring().GetStandalone().DeepCopy(),
	}

	b.Objects = append(b.Objects, standalone)

	return b
}

func (b *monitoringBuilder) BuildConfigMap() deployer.Builder {
	if !b.Cluster.GetMonitoring().IsEnabled() || b.Cluster.GetMonitoring().GetVector() == nil {
		return b
	}

	if b.Err != nil {
		return b
	}

	cm, err := b.GenerateVectorConfigMap()
	if err != nil {
		b.Err = err
		return b
	}

	b.Objects = append(b.Objects, cm)

	return b
}

func (b *monitoringBuilder) Generate() ([]client.Object, error) {
	return b.Objects, b.Err
}
