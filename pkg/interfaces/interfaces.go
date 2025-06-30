// pkg/interfaces/interfaces.go
package interfaces

import (
	"context"

	"github.com/roman-povoroznyk/kubernetes-controller/k6s/pkg/config"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// ConfigLoader defines interface for loading configuration
type ConfigLoader interface {
	LoadConfig(path string) (*config.Config, error)
	ValidateConfig(*config.Config) error
}

// ClusterManager defines interface for managing Kubernetes clusters
type ClusterManager interface {
	GetClient(clusterName string) (kubernetes.Interface, error)
	GetConfig(clusterName string) (*rest.Config, error)
	TestConnectivity(ctx context.Context, clusterName string) error
	ListClusters() []string
}

// DeploymentWatcher defines interface for watching deployments
type DeploymentWatcher interface {
	Start(ctx context.Context) error
	Stop() error
	AddEventHandler(handler DeploymentEventHandler)
}

// DeploymentEventHandler defines interface for handling deployment events
type DeploymentEventHandler interface {
	OnAdd(deployment *appsv1.Deployment)
	OnUpdate(oldDeployment, newDeployment *appsv1.Deployment)
	OnDelete(deployment *appsv1.Deployment)
}

// MetricsCollector defines interface for metrics collection
type MetricsCollector interface {
	IncrementDeploymentEvents(eventType string, cluster string)
	RecordDeploymentStatus(cluster string, namespace string, status string)
	RecordReconciliationDuration(cluster string, duration float64)
}

// HealthChecker defines interface for health checking
type HealthChecker interface {
	CheckHealth(ctx context.Context) error
	IsReady() bool
}

// ControllerManager defines interface for managing controllers
type ControllerManager interface {
	Start(ctx context.Context) error
	Stop() error
	GetManager() manager.Manager
	AddController(name string, controller interface{}) error
}
