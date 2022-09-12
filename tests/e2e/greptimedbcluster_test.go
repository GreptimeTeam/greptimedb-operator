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
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/GreptimeTeam/greptimedb-operator/apis/v1alpha1"
)

const (
	createTableSQL = `CREATE TABLE monitor (
                        host STRING,
                        ts BIGINT,
                        cpu DOUBLE DEFAULT 0,
                        memory DOUBLE,
                        TIME INDEX (ts),
                      PRIMARY KEY(ts,host)) ENGINE=mito WITH(regions=1)`
	insertDataSQL = `INSERT INTO monitor(host, cpu, memory, ts) VALUES ('host1', 66.6, 1024, 1660897955)`
	selectDataSQL = `SELECT * FROM monitor`
)

// Metric is the schema of test data in SQL table.
type Metric struct {
	hostname  string
	timestamp int32
	cpu       float64
	memory    float64
}

var _ = Describe("Basic test greptimedbcluster controller", func() {
	It("Bootstrap cluster", func() {
		testCluster, err := readClusterConfig()
		Expect(err).NotTo(HaveOccurred(), "failed to read testCluster config")

		err = k8sClient.Create(ctx, testCluster)
		Expect(err).NotTo(HaveOccurred(), "failed to create greptimedbcluster")

		By("Check the status of testCluster")
		Eventually(func() bool {
			var cluser v1alpha1.GreptimeDBCluster
			if err := k8sClient.Get(ctx, client.ObjectKey{Namespace: testCluster.Namespace,
				Name: testCluster.Name}, &cluser); err != nil {
				return false
			}

			for _, condition := range cluser.Status.Conditions {
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
		_, err = db.Query(createTableSQL)
		Expect(err).NotTo(HaveOccurred(), "failed to create SQL table")

		_, err = db.Query(insertDataSQL)
		Expect(err).NotTo(HaveOccurred(), "failed to insert data")

		results, err := db.Query(selectDataSQL)
		Expect(err).NotTo(HaveOccurred(), "failed to get data")

		var metrics []Metric
		for results.Next() {
			var m Metric
			err = results.Scan(&m.hostname, &m.cpu, &m.memory, &m.timestamp)
			Expect(err).NotTo(HaveOccurred(), "failed to scan data that query from db")
			metrics = append(metrics, m)
		}
		Expect(len(metrics) == 1).Should(BeTrue(), "get the wrong data from db")

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
		cmd.Run()
	}
}
