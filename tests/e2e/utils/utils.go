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
	"math/rand"
	"net"
	"os"
	"os/exec"
	"reflect"
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

// TODO(zyy17): We should use the sqlness in the future.

// RunSQLTest runs the sql test on the frontend ingress IP.
func RunSQLTest(ctx context.Context, frontendAddr string, isDistributed bool) error {
	cfg := mysql.Config{
		Net:                  "tcp",
		Addr:                 frontendAddr,
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
		createDistributedTableSQL = `CREATE TABLE test_table(
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
						  )
						  engine=mito;`

		createStandaloneTableSQL = `CREATE TABLE test_table(
							ts TIMESTAMP DEFAULT current_timestamp(),
							n INT,
							row_id INT,
							PRIMARY KEY(n),
							TIME INDEX (ts)
						  )
						  engine=mito;`

		insertDataSQL = `INSERT INTO test_table(n, row_id, ts) VALUES (?, ?, ?);`
		selectDataSQL = `SELECT * FROM test_table`

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
		_, err = conn.ExecContext(ctx, insertDataSQL, i, i, timestampInMillisecond())
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

// RunFlowTest runs the greptimedb flow tests on the frontend ingress IP.
func RunFlowTest(ctx context.Context, frontendAddr string) error {
	cfg := mysql.Config{
		Net:                  "tcp",
		Addr:                 frontendAddr,
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

	createInputTable := `CREATE TABLE numbers_input (
	    number INT,
	    ts TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	    PRIMARY KEY(number),
	    TIME INDEX(ts)
	);`
	_, err = conn.ExecContext(ctx, createInputTable)
	if err != nil {
		return err
	}

	createOutputTable := `CREATE TABLE sum_number_output (
	    sum_number BIGINT,
	    start_window TIMESTAMP TIME INDEX,
	    end_window TIMESTAMP,
	    update_at TIMESTAMP,
	);`
	_, err = conn.ExecContext(ctx, createOutputTable)
	if err != nil {
		return err
	}

	createFlow := fmt.Sprintf(`CREATE FLOW test_flow SINK TO sum_number_output AS
		SELECT sum(number) FROM numbers_input GROUP BY tumble(ts, '5 second', %d);`, timestampInMillisecond())
	_, err = conn.ExecContext(ctx, createFlow)
	if err != nil {
		return err
	}

	insertDataSQL := `INSERT INTO numbers_input(number, ts) VALUES (?, ?);`
	selectDataSQL := `SELECT sum_number FROM sum_number_output;`

	for i := 0; i < 20; i++ {
		_, err = conn.ExecContext(ctx, insertDataSQL, i, timestampInMillisecond())
		if err != nil {
			return err
		}
	}

	// Wait for the flow to process the data.
	time.Sleep(5 * time.Second)

	var sumNumber int
	if err := conn.QueryRowContext(ctx, selectDataSQL).Scan(&sumNumber); err != nil {
		return err
	}

	var numbers []int
	results, err := conn.QueryContext(ctx, selectDataSQL)
	if err != nil {
		return err
	}

	for results.Next() {
		var number int
		if err := results.Scan(&number); err != nil {
			return err
		}
		numbers = append(numbers, number)
	}

	if !reflect.DeepEqual(numbers, []int{190}) {
		return fmt.Errorf("unexpected number of rows returned: %d", numbers)
	}

	// Keep inserting the data.
	for i := 21; i < 42; i++ {
		_, err = conn.ExecContext(ctx, insertDataSQL, i, timestampInMillisecond())
		if err != nil {
			return err
		}
	}

	results, err = conn.QueryContext(ctx, selectDataSQL)
	if err != nil {
		return err
	}

	for results.Next() {
		var number int
		if err := results.Scan(&number); err != nil {
			return err
		}
		numbers = append(numbers, number)
	}

	// There is two windows in the flow.
	if reflect.DeepEqual(numbers, []int{190, 841}) {
		return fmt.Errorf("unexpected number of rows returned: %d", numbers)
	}

	return nil
}

func timestampInMillisecond() int64 {
	now := time.Now()
	return now.Unix()*1000 + int64(now.Nanosecond())/1e6
}

// PortForward uses kubectl to port forward the service to a local random available port.
func PortForward(ctx context.Context, namespace, serviceName string, sourcePort int) (string, error) {
	localDestinationPort, err := getRandomAvailablePort(10000, 30000)
	if err != nil {
		return "", err
	}

	// Start the port forwarding in a goroutine.
	go func() {
		fmt.Printf("Use kubectl to port forwarding %s/%s:%d to %d\n", namespace, serviceName, sourcePort, localDestinationPort)
		cmd := exec.CommandContext(ctx, "kubectl", "port-forward", "-n", namespace, fmt.Sprintf("svc/%s", serviceName), fmt.Sprintf("%d:%d", localDestinationPort, sourcePort))
		if err := cmd.Run(); err != nil {
			return
		}
	}()

	return fmt.Sprintf("localhost:%d", localDestinationPort), nil
}

// KillPortForwardProcess kills the port forwarding process.
func KillPortForwardProcess() {
	cmd := exec.Command("pkill", "-f", "kubectl port-forward")
	if err := cmd.Run(); err != nil {
		panic(err)
	}
}

func getRandomAvailablePort(minPort, maxPort int) (int, error) {
	for i := 0; i < 100; i++ {
		port := rand.Intn(maxPort-minPort+1) + minPort
		listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			continue
		}
		listener.Close()
		return port, nil
	}

	return 0, fmt.Errorf("failed to get a random available port")
}
