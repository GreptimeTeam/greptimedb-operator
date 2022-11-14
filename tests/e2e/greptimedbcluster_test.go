package e2e

import (
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"os/exec"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/go-sql-driver/mysql"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/GreptimeTeam/greptimedb-operator/apis/v1alpha1"
)

const (
	createTableSQL = `CREATE TABLE dist_table (
                        ts TIMESTAMP DEFAULT current_timestamp(),
                        n INT,
    					row_id INT,
                        TIME INDEX (ts)
                     )
                     PARTITION BY RANGE COLUMNS (n) (
    				 	PARTITION r0 VALUES LESS THAN (5),
    					PARTITION r1 VALUES LESS THAN (9),
    					PARTITION r2 VALUES LESS THAN (MAXVALUE),
					)
					engine=mito`

	insertDataSQLStr = "INSERT INTO dist_table(n, row_id) VALUES (%d, %d)"

	selectDataSQL = `SELECT * FROM dist_table`

	testRowIDNum = 12
)

var (
	defaultQueryTimeout = 5 * time.Second
)

// TestData is the schema of test data in SQL table.
type TestData struct {
	timestamp string
	n         int32
	rowID     int32
}

var _ = Describe("Basic test greptimedbcluster controller", func() {
	It("Bootstrap cluster", func() {
		testCluster, err := readClusterConfig()
		Expect(err).NotTo(HaveOccurred(), "failed to read testCluster config")

		err = k8sClient.Create(ctx, testCluster)
		Expect(err).NotTo(HaveOccurred(), "failed to create greptimedbcluster")

		By("Check the status of testCluster")
		Eventually(func() bool {
			var cluster v1alpha1.GreptimeDBCluster
			if err := k8sClient.Get(ctx, client.ObjectKey{Namespace: testCluster.Namespace,
				Name: testCluster.Name}, &cluster); err != nil {
				return false
			}

			for _, condition := range cluster.Status.Conditions {
				if condition.Type == v1alpha1.GreptimeDBClusterReady &&
					condition.Status == corev1.ConditionTrue {
					return true
				}
			}
			return false
		}, 120*time.Second, time.Second).Should(BeTrue())

		go func() {
			forwardRequest(testCluster.Name)
		}()

		By("Connecting GreptimeDB")
		var db *sql.DB
		Eventually(func() error {
			cfg := mysql.Config{
				Net:                  "tcp",
				Addr:                 "127.0.0.1:3306",
				User:                 "",
				Passwd:               "",
				DBName:               "",
				AllowNativePasswords: true,
			}

			db, err = sql.Open("mysql", cfg.FormatDSN())
			if err != nil {
				return err
			}

			_, err = db.Conn(context.TODO())
			if err != nil {
				return err
			}

			return nil
		}, 60*time.Second, time.Second).ShouldNot(HaveOccurred())

		By("Execute SQL queries after connecting")

		ctx, cancel := context.WithTimeout(context.Background(), defaultQueryTimeout)
		defer cancel()

		_, err = db.ExecContext(ctx, createTableSQL)
		Expect(err).NotTo(HaveOccurred(), "failed to create SQL table")

		ctx, cancel = context.WithTimeout(context.Background(), defaultQueryTimeout)
		defer cancel()
		for rowID := 1; rowID <= testRowIDNum; rowID++ {
			insertDataSQL := fmt.Sprintf(insertDataSQLStr, rowID, rowID)
			_, err = db.ExecContext(ctx, insertDataSQL)
			Expect(err).NotTo(HaveOccurred(), "failed to insert data")
		}

		ctx, cancel = context.WithTimeout(context.Background(), defaultQueryTimeout)
		defer cancel()
		results, err := db.QueryContext(ctx, selectDataSQL)
		Expect(err).NotTo(HaveOccurred(), "failed to get data")

		var data []TestData
		for results.Next() {
			var d TestData
			err = results.Scan(&d.timestamp, &d.n, &d.rowID)
			Expect(err).NotTo(HaveOccurred(), "failed to scan data that query from db")
			data = append(data, d)
		}
		Expect(len(data) == testRowIDNum).Should(BeTrue(), "get the wrong data from db")

		By("Delete cluster")
		err = k8sClient.Delete(ctx, testCluster)
		Expect(err).NotTo(HaveOccurred(), "failed to delete cluster")
		Eventually(func() error {
			// The cluster will be deleted eventually.
			return k8sClient.Get(ctx, client.ObjectKey{Name: testCluster.Namespace, Namespace: testCluster.Namespace}, testCluster)
		}, 30*time.Second, time.Second).Should(HaveOccurred())
	})
})

func readClusterConfig() (*v1alpha1.GreptimeDBCluster, error) {
	data, err := ioutil.ReadFile("./testdata/cluster.yaml")
	if err != nil {
		return nil, err
	}

	var cluster v1alpha1.GreptimeDBCluster
	if err := yaml.Unmarshal(data, &cluster); err != nil {
		return nil, err
	}

	return &cluster, nil
}

// FIXME(zyy17): When the e2e exit, the kubectl process still active background.
func forwardRequest(clusterName string) {
	for {
		cmd := exec.Command("kubectl", "port-forward", fmt.Sprintf("svc/%s-frontend", clusterName), "3306:3306")
		if err := cmd.Run(); err != nil {
			klog.Errorf("Failed to port forward:%v", err)
			return
		}
	}
}
