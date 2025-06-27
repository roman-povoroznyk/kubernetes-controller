package server

import (
	"fmt"
	"strings"
	"time"

	"github.com/roman-povoroznyk/kubernetes-controller/internal/informer"
	"github.com/roman-povoroznyk/kubernetes-controller/internal/utils"
	"github.com/valyala/fasthttp"
	corev1 "k8s.io/api/core/v1"
)

// HandleRequests is the main HTTP request handler
func HandleRequests(ctx *fasthttp.RequestCtx) {
	path := string(ctx.Path())
	method := string(ctx.Method())

	// Check for individual resource endpoints first
	if strings.HasPrefix(path, "/deployments/") && path != "/deployments/names" {
		if method != fasthttp.MethodGet {
			utils.HTTP.MethodNotAllowed(ctx)
			return
		}
		deploymentName := strings.TrimPrefix(path, "/deployments/")
		handleDeploymentByName(ctx, deploymentName)
		return
	}

	if strings.HasPrefix(path, "/pods/") && path != "/pods/names" {
		if method != fasthttp.MethodGet {
			utils.HTTP.MethodNotAllowed(ctx)
			return
		}
		podName := strings.TrimPrefix(path, "/pods/")
		handlePodByName(ctx, podName)
		return
	}

	switch path {
	case "/health":
		if method != fasthttp.MethodGet {
			utils.HTTP.MethodNotAllowed(ctx)
			return
		}
		handleHealth(ctx)
	case "/deployments/names":
		if method != fasthttp.MethodGet {
			utils.HTTP.MethodNotAllowed(ctx)
			return
		}
		handleDeploymentNames(ctx)
	case "/deployments":
		if method != fasthttp.MethodGet {
			utils.HTTP.MethodNotAllowed(ctx)
			return
		}
		handleDeployments(ctx)
	case "/pods/names":
		if method != fasthttp.MethodGet {
			utils.HTTP.MethodNotAllowed(ctx)
			return
		}
		handlePodNames(ctx)
	case "/pods":
		if method != fasthttp.MethodGet {
			utils.HTTP.MethodNotAllowed(ctx)
			return
		}
		handlePods(ctx)
	case "/":
		handleDefault(ctx)
	default:
		handleNotFound(ctx)
	}
}

// handleHealth handles health check requests
func handleHealth(ctx *fasthttp.RequestCtx) {
	utils.HTTP.PlainResponse(ctx, fasthttp.StatusOK, "OK")
}

// handleDefault handles the root route
func handleDefault(ctx *fasthttp.RequestCtx) {
	utils.HTTP.PlainResponse(ctx, fasthttp.StatusOK, "Hello from k8s-ctrl FastHTTP server!")
}

// handleNotFound handles requests to non-existent routes
func handleNotFound(ctx *fasthttp.RequestCtx) {
	utils.HTTP.NotFound(ctx, fmt.Sprintf("404 Not Found: %s", ctx.Path()))
}

// Deployment response structures
type DeploymentSummary struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Replicas  int32             `json:"replicas"`
	Ready     int32             `json:"ready"`
	Updated   int32             `json:"updated"`
	Available int32             `json:"available"`
	Age       string            `json:"age"`
	Image     string            `json:"image"`
	Labels    map[string]string `json:"labels,omitempty"`
}

type PodSummary struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Phase     string            `json:"phase"`
	Ready     string            `json:"ready"`
	Restarts  int               `json:"restarts"`
	Age       string            `json:"age"`
	Image     string            `json:"image"`
	Node      string            `json:"node,omitempty"`
	Labels    map[string]string `json:"labels,omitempty"`
}

// handleDeployments returns full deployment information from informer cache
func handleDeployments(ctx *fasthttp.RequestCtx) {
	deployments := informer.GetDeployments()

	var summaries []DeploymentSummary
	for _, dep := range deployments {
		summary := DeploymentSummary{
			Name:      dep.Name,
			Namespace: dep.Namespace,
			Replicas:  utils.K8sHelpers.GetReplicaCount(dep.Spec.Replicas),
			Ready:     dep.Status.ReadyReplicas,
			Updated:   dep.Status.UpdatedReplicas,
			Available: dep.Status.AvailableReplicas,
			Age:       formatAge(dep.CreationTimestamp.Time),
			Image:     utils.K8sHelpers.GetMainContainerImage(dep),
			Labels:    dep.Labels,
		}
		summaries = append(summaries, summary)
	}

	utils.HTTP.JSONResponse(ctx, fasthttp.StatusOK, summaries)
}

// handleDeploymentNames returns only deployment names from informer cache
func handleDeploymentNames(ctx *fasthttp.RequestCtx) {
	names := informer.GetDeploymentNames()
	utils.HTTP.JSONResponse(ctx, fasthttp.StatusOK, names)
}

// handlePods returns full pod information from informer cache
func handlePods(ctx *fasthttp.RequestCtx) {
	pods := informer.GetPods()

	var summaries []PodSummary
	for _, pod := range pods {
		summary := PodSummary{
			Name:      pod.Name,
			Namespace: pod.Namespace,
			Phase:     string(pod.Status.Phase),
			Ready:     getPodReadyStatus(pod),
			Restarts:  int(utils.K8sHelpers.GetPodRestartCount(pod)),
			Age:       formatAge(pod.CreationTimestamp.Time),
			Image:     utils.K8sHelpers.GetMainPodImage(pod),
			Node:      pod.Spec.NodeName,
			Labels:    pod.Labels,
		}
		summaries = append(summaries, summary)
	}

	utils.HTTP.JSONResponse(ctx, fasthttp.StatusOK, summaries)
}

// handlePodNames returns only pod names from informer cache
func handlePodNames(ctx *fasthttp.RequestCtx) {
	names := informer.GetPodNames()
	utils.HTTP.JSONResponse(ctx, fasthttp.StatusOK, names)
}

// handleDeploymentByName returns a specific deployment by name from informer cache
func handleDeploymentByName(ctx *fasthttp.RequestCtx, name string) {
	deployment, found := informer.GetDeploymentByName(name)
	if !found || deployment == nil {
		utils.HTTP.NotFound(ctx, "Deployment not found")
		return
	}

	summary := DeploymentSummary{
		Name:      deployment.Name,
		Namespace: deployment.Namespace,
		Replicas:  utils.K8sHelpers.GetReplicaCount(deployment.Spec.Replicas),
		Ready:     deployment.Status.ReadyReplicas,
		Updated:   deployment.Status.UpdatedReplicas,
		Available: deployment.Status.AvailableReplicas,
		Age:       formatAge(deployment.CreationTimestamp.Time),
		Image:     utils.K8sHelpers.GetMainContainerImage(deployment),
		Labels:    deployment.Labels,
	}

	utils.HTTP.JSONResponse(ctx, fasthttp.StatusOK, summary)
}

// handlePodByName returns a specific pod by name from informer cache
func handlePodByName(ctx *fasthttp.RequestCtx, name string) {
	pod, found := informer.GetPodByName(name)
	if !found || pod == nil {
		utils.HTTP.NotFound(ctx, "Pod not found")
		return
	}

	summary := PodSummary{
		Name:      pod.Name,
		Namespace: pod.Namespace,
		Phase:     string(pod.Status.Phase),
		Ready:     getPodReadyStatus(pod),
		Restarts:  int(utils.K8sHelpers.GetPodRestartCount(pod)),
		Age:       formatAge(pod.CreationTimestamp.Time),
		Image:     utils.K8sHelpers.GetMainPodImage(pod),
		Node:      pod.Spec.NodeName,
		Labels:    pod.Labels,
	}

	utils.HTTP.JSONResponse(ctx, fasthttp.StatusOK, summary)
}

// Helper functions specific to server package

// getPodReadyStatus returns ready status as "ready/total" string
func getPodReadyStatus(pod *corev1.Pod) string {
	ready := 0
	total := len(pod.Status.ContainerStatuses)

	for _, containerStatus := range pod.Status.ContainerStatuses {
		if containerStatus.Ready {
			ready++
		}
	}

	return fmt.Sprintf("%d/%d", ready, total)
}

// formatAge formats time duration as human-readable age
func formatAge(creationTime time.Time) string {
	age := time.Since(creationTime)

	days := int(age.Hours()) / 24
	hours := int(age.Hours()) % 24
	minutes := int(age.Minutes()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd", days)
	} else if hours > 0 {
		return fmt.Sprintf("%dh", hours)
	} else if minutes > 0 {
		return fmt.Sprintf("%dm", minutes)
	} else {
		return fmt.Sprintf("%ds", int(age.Seconds()))
	}
}
