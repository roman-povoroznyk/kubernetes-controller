// pkg/metrics/metrics.go
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds all Prometheus metrics
type Metrics struct {
	// Deployment metrics
	DeploymentEvents *prometheus.CounterVec
	DeploymentStatus *prometheus.GaugeVec
	
	// Reconciliation metrics
	ReconciliationDuration *prometheus.HistogramVec
	ReconciliationErrors   *prometheus.CounterVec
	
	// Controller metrics
	ControllerRestarts *prometheus.CounterVec
	ControllerUptime   *prometheus.GaugeVec
	
	// Cluster metrics
	ClusterConnections *prometheus.GaugeVec
	ClusterLatency     *prometheus.HistogramVec
	
	// Cache metrics
	CacheHits   *prometheus.CounterVec
	CacheMisses *prometheus.CounterVec
	CacheSize   *prometheus.GaugeVec
}

// New creates a new Metrics instance
func New() *Metrics {
	return &Metrics{
		DeploymentEvents: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "k6s_deployment_events_total",
				Help: "Total number of deployment events processed",
			},
			[]string{"cluster", "namespace", "event_type"},
		),
		
		DeploymentStatus: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "k6s_deployment_status",
				Help: "Current status of deployments",
			},
			[]string{"cluster", "namespace", "deployment", "status"},
		),
		
		ReconciliationDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name: "k6s_reconciliation_duration_seconds",
				Help: "Duration of reconciliation operations",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"cluster", "controller"},
		),
		
		ReconciliationErrors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "k6s_reconciliation_errors_total",
				Help: "Total number of reconciliation errors",
			},
			[]string{"cluster", "controller", "error_type"},
		),
		
		ControllerRestarts: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "k6s_controller_restarts_total",
				Help: "Total number of controller restarts",
			},
			[]string{"controller", "reason"},
		),
		
		ControllerUptime: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "k6s_controller_uptime_seconds",
				Help: "Controller uptime in seconds",
			},
			[]string{"controller"},
		),
		
		ClusterConnections: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "k6s_cluster_connections",
				Help: "Current number of cluster connections",
			},
			[]string{"cluster", "status"},
		),
		
		ClusterLatency: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name: "k6s_cluster_latency_seconds",
				Help: "Cluster API latency",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"cluster", "operation"},
		),
		
		CacheHits: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "k6s_cache_hits_total",
				Help: "Total number of cache hits",
			},
			[]string{"cache_type"},
		),
		
		CacheMisses: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "k6s_cache_misses_total",
				Help: "Total number of cache misses",
			},
			[]string{"cache_type"},
		),
		
		CacheSize: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "k6s_cache_size",
				Help: "Current cache size",
			},
			[]string{"cache_type"},
		),
	}
}

// RecordDeploymentEvent records a deployment event
func (m *Metrics) RecordDeploymentEvent(cluster, namespace, eventType string) {
	m.DeploymentEvents.WithLabelValues(cluster, namespace, eventType).Inc()
}

// SetDeploymentStatus sets the deployment status
func (m *Metrics) SetDeploymentStatus(cluster, namespace, deployment, status string, value float64) {
	m.DeploymentStatus.WithLabelValues(cluster, namespace, deployment, status).Set(value)
}

// RecordReconciliationDuration records reconciliation duration
func (m *Metrics) RecordReconciliationDuration(cluster, controller string, duration float64) {
	m.ReconciliationDuration.WithLabelValues(cluster, controller).Observe(duration)
}

// RecordReconciliationError records a reconciliation error
func (m *Metrics) RecordReconciliationError(cluster, controller, errorType string) {
	m.ReconciliationErrors.WithLabelValues(cluster, controller, errorType).Inc()
}

// RecordControllerRestart records a controller restart
func (m *Metrics) RecordControllerRestart(controller, reason string) {
	m.ControllerRestarts.WithLabelValues(controller, reason).Inc()
}

// SetControllerUptime sets the controller uptime
func (m *Metrics) SetControllerUptime(controller string, uptime float64) {
	m.ControllerUptime.WithLabelValues(controller).Set(uptime)
}

// SetClusterConnections sets the number of cluster connections
func (m *Metrics) SetClusterConnections(cluster, status string, count float64) {
	m.ClusterConnections.WithLabelValues(cluster, status).Set(count)
}

// RecordClusterLatency records cluster API latency
func (m *Metrics) RecordClusterLatency(cluster, operation string, latency float64) {
	m.ClusterLatency.WithLabelValues(cluster, operation).Observe(latency)
}

// RecordCacheHit records a cache hit
func (m *Metrics) RecordCacheHit(cacheType string) {
	m.CacheHits.WithLabelValues(cacheType).Inc()
}

// RecordCacheMiss records a cache miss
func (m *Metrics) RecordCacheMiss(cacheType string) {
	m.CacheMisses.WithLabelValues(cacheType).Inc()
}

// SetCacheSize sets the cache size
func (m *Metrics) SetCacheSize(cacheType string, size float64) {
	m.CacheSize.WithLabelValues(cacheType).Set(size)
}
