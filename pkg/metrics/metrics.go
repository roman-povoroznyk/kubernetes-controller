package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// DeploymentMetrics tracks deployment operations
	DeploymentOperationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kubernetes_controller_deployment_operations_total",
			Help: "Total number of deployment operations",
		},
		[]string{"operation", "namespace", "cluster"},
	)

	// ReconciliationDuration tracks reconciliation time
	ReconciliationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "kubernetes_controller_reconciliation_duration_seconds",
			Help:    "Time spent on reconciliation operations",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"controller", "namespace"},
	)

	// BusinessRuleValidations tracks business rule validation results
	BusinessRuleValidations = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kubernetes_controller_business_rule_validations_total",
			Help: "Total number of business rule validations",
		},
		[]string{"rule", "result", "namespace"},
	)

	// ActiveClusters tracks number of active clusters
	ActiveClusters = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "kubernetes_controller_active_clusters",
			Help: "Number of active clusters being managed",
		},
	)

	// InformerCacheSync tracks informer cache sync status
	InformerCacheSync = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "kubernetes_controller_informer_cache_synced",
			Help: "Informer cache sync status (1 = synced, 0 = not synced)",
		},
		[]string{"informer", "cluster"},
	)
)

// RecordDeploymentOperation records a deployment operation
func RecordDeploymentOperation(operation, namespace, cluster string) {
	DeploymentOperationsTotal.WithLabelValues(operation, namespace, cluster).Inc()
}

// RecordBusinessRuleValidation records a business rule validation
func RecordBusinessRuleValidation(rule, result, namespace string) {
	BusinessRuleValidations.WithLabelValues(rule, result, namespace).Inc()
}

// SetActiveClusters sets the number of active clusters
func SetActiveClusters(count int) {
	ActiveClusters.Set(float64(count))
}

// SetInformerCacheSync sets informer cache sync status
func SetInformerCacheSync(informer, cluster string, synced bool) {
	value := 0.0
	if synced {
		value = 1.0
	}
	InformerCacheSync.WithLabelValues(informer, cluster).Set(value)
}
