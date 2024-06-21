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

package utils

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/go-sql-driver/mysql"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/controller-runtime/pkg/client"

	greptimev1alpha1 "github.com/GreptimeTeam/greptimedb-operator/apis/v1alpha1"
	"github.com/GreptimeTeam/greptimedb-operator/controllers/common"
	"github.com/GreptimeTeam/greptimedb-operator/controllers/constant"
)

const (
	DefaultTimeout = 5 * time.Minute
)

// LoadGreptimeCRDFromFile loads GreptimeCRD from the input yaml file.
func LoadGreptimeCRDFromFile(inputFile string, obj client.Object) error {
	data, err := os.ReadFile(inputFile)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(data, obj); err != nil {
		return err
	}

	return nil
}

// GetFrontendServiceIngressIP returns the ingress IP of the frontend service.
// It is assumed that the service is of type LoadBalancer.
func GetFrontendServiceIngressIP(ctx context.Context, k8sClient client.Client, namespace, name string) (string, error) {
	var svc corev1.Service
	if err := k8sClient.Get(ctx, client.ObjectKey{Namespace: namespace, Name: name}, &svc); err != nil {
		return "", err
	}

	if len(svc.Status.LoadBalancer.Ingress) == 0 {
		return "", fmt.Errorf("no ingress found for service %s", name)
	}

	return svc.Status.LoadBalancer.Ingress[0].IP, nil
}

// CleanEtcdData use kubectl to clean up all data in etcd.
func CleanEtcdData(ctx context.Context, namespace, name string) error {
	cmd := exec.CommandContext(ctx, "kubectl", "exec", "-n", namespace, name, "--", "etcdctl", "del", "--prefix", "")
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

// GetPhase returns the phase of the GreptimeDBCluster or GreptimeDBStandalone.
func GetPhase(ctx context.Context, k8sClient client.Client, namespace, name string, obj client.Object) (greptimev1alpha1.Phase, error) {
	if err := k8sClient.Get(ctx, client.ObjectKey{Namespace: namespace, Name: name}, obj); err != nil {
		return "", err
	}

	cluster, ok := obj.(*greptimev1alpha1.GreptimeDBCluster)
	if !ok {
		standalone, ok := obj.(*greptimev1alpha1.GreptimeDBStandalone)
		if !ok {
			return "", fmt.Errorf("unknown object type")
		}

		return standalone.Status.StandalonePhase, nil
	}

	return cluster.Status.ClusterPhase, nil
}

// GetPVCs returns the PVCs of the datanodes.
func GetPVCs(ctx context.Context, k8sClient client.Client, namespace, name string, kind greptimev1alpha1.ComponentKind) ([]corev1.PersistentVolumeClaim, error) {
	selector, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{
		MatchLabels: map[string]string{
			constant.GreptimeDBComponentName: common.ResourceName(name, kind),
		},
	})
	if err != nil {
		return nil, err
	}

	pvcList := new(corev1.PersistentVolumeClaimList)

	if err = k8sClient.List(ctx, pvcList, client.InNamespace(namespace),
		client.MatchingLabelsSelector{Selector: selector}); err != nil {
		return nil, err
	}

	return pvcList.Items, nil
}

// RunSQLTest runs the sql test on the frontend ingress IP.
func RunSQLTest(ctx context.Context, frontendIngressIP string, isDistributed bool) error {
	cfg := mysql.Config{
		Net:                  "tcp",
		Addr:                 fmt.Sprintf("%s:%d", frontendIngressIP, 4002),
		AllowNativePasswords: true,
	}

	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		return err
	}

	conn, err := db.Conn(ctx)
	if err != nil {
		return err
	}

	const (
		createDistributedTableSQL = `CREATE TABLE dist_table(
							ts TIMESTAMP DEFAULT current_timestamp(),
							n INT,
							row_id INT,
							PRIMARY KEY(n),
							TIME INDEX (ts)
						  )
                          PARTITION ON COLUMNS (n) (
						      n < 5,
							  n >= 5 AND n < 9,
							  n >= 9
						  )`

		createStandaloneTableSQL = `CREATE TABLE dist_table(
							ts TIMESTAMP DEFAULT current_timestamp(),
							n INT,
							row_id INT,
							PRIMARY KEY(n),
							TIME INDEX (ts)
						  )
						  engine=mito;`

		insertDataSQL = `INSERT INTO dist_table(n, row_id) VALUES (?, ?);`
		selectDataSQL = `SELECT * FROM dist_table`

		rowsNum = 42
	)

	createTableSQL := createStandaloneTableSQL
	if isDistributed {
		createTableSQL = createDistributedTableSQL
	}

	_, err = conn.ExecContext(ctx, createTableSQL)
	if err != nil {
		return err
	}

	for i := 0; i < rowsNum; i++ {
		_, err = conn.ExecContext(ctx, insertDataSQL, i, i)
		if err != nil {
			return err
		}
	}

	var data []interface{}
	results, err := conn.QueryContext(ctx, selectDataSQL)
	if err != nil {
		return err
	}

	for results.Next() {
		var row struct {
			ts    string
			n     int
			rowID int
		}
		if err := results.Scan(&row.ts, &row.n, &row.rowID); err != nil {
			return err
		}
		data = append(data, row)
	}

	if len(data) != rowsNum {
		return fmt.Errorf("unexpected number of rows returned: %d", len(data))
	}

	return nil
}
