package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

const (
	ClusterError    float64 = 0
	ClusterDeleted  float64 = 1
	ClusterStarting float64 = 2
	ClusterRunning  float64 = 3
)

type Metrics struct {
	ClusterStatus *prometheus.GaugeVec
	ClusterUpdate *prometheus.GaugeVec
	ClusterCount  prometheus.Gauge
}

func NewMetrics() *Metrics {
	m := Metrics{
		ClusterStatus: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "greptimedb_cluster_status",
			Help: "When the greptimedb cluster running is 3, starting is 2, Deleted is 1, otherwise 0",
		}, []string{"name", "namespace"}),
		ClusterUpdate: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "greptimedb_cluster_update_count",
			Help: "Number of greptimedb cluster update count",
		}, []string{"name", "namespace"}),
		ClusterCount: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "greptimedb_cluster_total_count",
			Help: "Number of greptimedb cluster total count",
		}),
	}
	metrics.Registry.MustRegister(
		m.ClusterStatus,
		m.ClusterUpdate,
		m.ClusterCount,
	)

	return &m
}
