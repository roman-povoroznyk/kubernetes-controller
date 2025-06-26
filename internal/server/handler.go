package server

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/roman-povoroznyk/kubernetes-controller/internal/informer"
	"github.com/valyala/fasthttp"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// HandleRequests is the main HTTP request handler
func HandleRequests(ctx *fasthttp.RequestCtx) {
	path := string(ctx.Path())
	method := string(ctx.Method())

	// Check for individual resource endpoints first
	if strings.HasPrefix(path, "/deployments/") && path != "/deployments/names" {
		if method != fasthttp.MethodGet {
			ctx.SetStatusCode(fasthttp.StatusMethodNotAllowed)
			fmt.Fprintf(ctx, "Method Not Allowed")
			return
		}
		deploymentName := strings.TrimPrefix(path, "/deployments/")
		handleDeploymentByName(ctx, deploymentName)
		return
	}

	if strings.HasPrefix(path, "/pods/") && path != "/pods/names" {
		if method != fasthttp.MethodGet {
			ctx.SetStatusCode(fasthttp.StatusMethodNotAllowed)
			fmt.Fprintf(ctx, "Method Not Allowed")
			return
		}
		podName := strings.TrimPrefix(path, "/pods/")
		handlePodByName(ctx, podName)
		return
	}

	switch path {
	case "/health":
		if method != fasthttp.MethodGet {
			ctx.SetStatusCode(fasthttp.StatusMethodNotAllowed)
			fmt.Fprintf(ctx, "Method Not Allowed")
			return
		}
		handleHealth(ctx)
	case "/deployments/names":
		if method != fasthttp.MethodGet {
			ctx.SetStatusCode(fasthttp.StatusMethodNotAllowed)
			fmt.Fprintf(ctx, "Method Not Allowed")
			return
		}
		handleDeploymentNames(ctx)
	case "/deployments":
		if method != fasthttp.MethodGet {
			ctx.SetStatusCode(fasthttp.StatusMethodNotAllowed)
			fmt.Fprintf(ctx, "Method Not Allowed")
			return
		}
		handleDeployments(ctx)
	case "/pods/names":
		if method != fasthttp.MethodGet {
			ctx.SetStatusCode(fasthttp.StatusMethodNotAllowed)
			fmt.Fprintf(ctx, "Method Not Allowed")
			return
		}
		handlePodNames(ctx)
	case "/pods":
		if method != fasthttp.MethodGet {
			ctx.SetStatusCode(fasthttp.StatusMethodNotAllowed)
			fmt.Fprintf(ctx, "Method Not Allowed")
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
	ctx.SetStatusCode(fasthttp.StatusOK)
	fmt.Fprintf(ctx, "OK")
}

// handleDefault handles the root route
func handleDefault(ctx *fasthttp.RequestCtx) {
	ctx.SetStatusCode(fasthttp.StatusOK)
	fmt.Fprintf(ctx, "Hello from k8s-ctrl FastHTTP server!")
}

// handleNotFound handles requests to non-existent routes
func handleNotFound(ctx *fasthttp.RequestCtx) {
	ctx.SetStatusCode(fasthttp.StatusNotFound)
	fmt.Fprintf(ctx, "404 Not Found: %s", ctx.Path())
}

// Deployment response structures
type DeploymentSummary struct {
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace"`
	Replicas    int32             `json:"replicas"`
	Ready       int32             `json:"ready"`
	Updated     int32             `json:"updated"`
	Available   int32             `json:"available"`
	Age         string            `json:"age"`
	Image       string            `json:"image"`
	Labels      map[string]string `json:"labels,omitempty"`
}

type PodSummary struct {
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace"`
	Phase       string            `json:"phase"`
	Ready       string            `json:"ready"`
	Restarts    int               `json:"restarts"`
	Age         string            `json:"age"`
	Image       string            `json:"image"`
	Node        string            `json:"node,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
}

// handleDeployments returns full deployment information from informer cache
func handleDeployments(ctx *fasthttp.RequestCtx) {
	deployments := informer.GetDeployments()

	var summaries []DeploymentSummary
	for _, dep := range deployments {
		summary := DeploymentSummary{
			Name:      dep.Name,
			Namespace: dep.Namespace,
			Replicas:  getReplicaCount(dep.Spec.Replicas),
			Ready:     dep.Status.ReadyReplicas,
			Updated:   dep.Status.UpdatedReplicas,
			Available: dep.Status.AvailableReplicas,
			Age:       formatAge(dep.CreationTimestamp.Time),
			Image:     getMainContainerImage(dep),
			Labels:    dep.Labels,
		}
		summaries = append(summaries, summary)
	}

	ctx.SetContentType("application/json")
	ctx.SetStatusCode(fasthttp.StatusOK)

	if err := json.NewEncoder(ctx).Encode(summaries); err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		fmt.Fprintf(ctx, `{"error": "Failed to encode response"}`)
	}
}

// handleDeploymentNames returns only deployment names from informer cache
func handleDeploymentNames(ctx *fasthttp.RequestCtx) {
	names := informer.GetDeploymentNames()

	ctx.SetContentType("application/json")
	ctx.SetStatusCode(fasthttp.StatusOK)

	if err := json.NewEncoder(ctx).Encode(names); err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		fmt.Fprintf(ctx, `{"error": "Failed to encode response"}`)
	}
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
			Restarts:  getPodRestartCount(pod),
			Age:       formatAge(pod.CreationTimestamp.Time),
			Image:     getMainPodImage(pod),
			Node:      pod.Spec.NodeName,
			Labels:    pod.Labels,
		}
		summaries = append(summaries, summary)
	}

	ctx.SetContentType("application/json")
	ctx.SetStatusCode(fasthttp.StatusOK)

	if err := json.NewEncoder(ctx).Encode(summaries); err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		fmt.Fprintf(ctx, `{"error": "Failed to encode response"}`)
	}
}

// handlePodNames returns only pod names from informer cache
func handlePodNames(ctx *fasthttp.RequestCtx) {
	names := informer.GetPodNames()

	ctx.SetContentType("application/json")
	ctx.SetStatusCode(fasthttp.StatusOK)

	if err := json.NewEncoder(ctx).Encode(names); err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		fmt.Fprintf(ctx, `{"error": "Failed to encode response"}`)
	}
}

// handleDeploymentByName returns a specific deployment by name from informer cache
func handleDeploymentByName(ctx *fasthttp.RequestCtx, name string) {
	deployment, found := informer.GetDeploymentByName(name)
	if !found || deployment == nil {
		ctx.SetStatusCode(fasthttp.StatusNotFound)
		fmt.Fprintf(ctx, `{"error": "Deployment not found"}`)
		return
	}

	summary := DeploymentSummary{
		Name:      deployment.Name,
		Namespace: deployment.Namespace,
		Replicas:  getReplicaCount(deployment.Spec.Replicas),
		Ready:     deployment.Status.ReadyReplicas,
		Updated:   deployment.Status.UpdatedReplicas,
		Available: deployment.Status.AvailableReplicas,
		Age:       formatAge(deployment.CreationTimestamp.Time),
		Image:     getMainContainerImage(deployment),
		Labels:    deployment.Labels,
	}

	ctx.SetContentType("application/json")
	ctx.SetStatusCode(fasthttp.StatusOK)

	if err := json.NewEncoder(ctx).Encode(summary); err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		fmt.Fprintf(ctx, `{"error": "Failed to encode response"}`)
	}
}

// handlePodByName returns a specific pod by name from informer cache
func handlePodByName(ctx *fasthttp.RequestCtx, name string) {
	pod, found := informer.GetPodByName(name)
	if !found || pod == nil {
		ctx.SetStatusCode(fasthttp.StatusNotFound)
		fmt.Fprintf(ctx, `{"error": "Pod not found"}`)
		return
	}

	summary := PodSummary{
		Name:      pod.Name,
		Namespace: pod.Namespace,
		Phase:     string(pod.Status.Phase),
		Ready:     getPodReadyStatus(pod),
		Restarts:  getPodRestartCount(pod),
		Age:       formatAge(pod.CreationTimestamp.Time),
		Image:     getMainPodImage(pod),
		Node:      pod.Spec.NodeName,
		Labels:    pod.Labels,
	}

	ctx.SetContentType("application/json")
	ctx.SetStatusCode(fasthttp.StatusOK)

	if err := json.NewEncoder(ctx).Encode(summary); err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		fmt.Fprintf(ctx, `{"error": "Failed to encode response"}`)
	}
}

// Helper functions

// getReplicaCount safely gets replica count from pointer
func getReplicaCount(replicas *int32) int32 {
	if replicas == nil {
		return 0
	}
	return *replicas
}

// getMainContainerImage gets the image of the first container in deployment
func getMainContainerImage(deployment *appsv1.Deployment) string {
	if len(deployment.Spec.Template.Spec.Containers) > 0 {
		return deployment.Spec.Template.Spec.Containers[0].Image
	}
	return "unknown"
}

// getMainPodImage gets the image of the first container in pod
func getMainPodImage(pod *corev1.Pod) string {
	if len(pod.Spec.Containers) > 0 {
		return pod.Spec.Containers[0].Image
	}
	return "unknown"
}

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

// getPodRestartCount gets total restart count for all containers
func getPodRestartCount(pod *corev1.Pod) int {
	totalRestarts := 0
	for _, containerStatus := range pod.Status.ContainerStatuses {
		totalRestarts += int(containerStatus.RestartCount)
	}
	return totalRestarts
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
