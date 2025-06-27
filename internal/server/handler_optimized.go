package server

import (
	"fmt"

	"github.com/roman-povoroznyk/kubernetes-controller/internal/informer"
	"github.com/roman-povoroznyk/kubernetes-controller/internal/utils"
	"github.com/valyala/fasthttp"
)

// Response structures for optimized handlers
type DeploymentResponse struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Replicas  int32  `json:"replicas"`
	Image     string `json:"image"`
	Age       string `json:"age"`
	Ready     string `json:"ready"`
}

type PodResponse struct {
	Name         string `json:"name"`
	Namespace    string `json:"namespace"`
	Phase        string `json:"phase"`
	Node         string `json:"node"`
	Image        string `json:"image"`
	Age          string `json:"age"`
	Ready        bool   `json:"ready"`
	RestartCount int32  `json:"restart_count"`
}

// RefactoredHandleRequests shows how the handler could be improved
func RefactoredHandleRequests(ctx *fasthttp.RequestCtx) {
	path := string(ctx.Path())

	// Use path extractor for efficient path parsing
	if deploymentName, ok := utils.PathExt.ExtractResourceName(path, "/deployments"); ok && path != "/deployments/names" {
		if !utils.ReqMatch.IsMethodAllowed(ctx, fasthttp.MethodGet) {
			utils.HTTP.MethodNotAllowed(ctx)
			return
		}
		handleDeploymentByNameOptimized(ctx, deploymentName)
		return
	}

	if podName, ok := utils.PathExt.ExtractResourceName(path, "/pods"); ok && path != "/pods/names" {
		if !utils.ReqMatch.IsMethodAllowed(ctx, fasthttp.MethodGet) {
			utils.HTTP.MethodNotAllowed(ctx)
			return
		}
		handlePodByNameOptimized(ctx, podName)
		return
	}

	// Route table approach instead of switch
	routes := map[string]func(*fasthttp.RequestCtx){
		"/health":           handleHealthOptimized,
		"/deployments/names": handleDeploymentNamesOptimized,
		"/deployments":      handleDeploymentsOptimized,
		"/pods/names":       handlePodNamesOptimized,
		"/pods":            handlePodsOptimized,
	}

	if handler, exists := routes[path]; exists {
		if !utils.ReqMatch.IsMethodAllowed(ctx, fasthttp.MethodGet) {
			utils.HTTP.MethodNotAllowed(ctx)
			return
		}
		handler(ctx)
		return
	}

	utils.HTTP.PlainResponse(ctx, fasthttp.StatusOK, "Hello from k8s-ctrl FastHTTP server!")
}

// Example optimized handlers using new utilities
func handleDeploymentByNameOptimized(ctx *fasthttp.RequestCtx, name string) {
	// Check cache first
	cacheKey := fmt.Sprintf("deployment:%s", name)
	if cached, found := utils.Cache.Get(cacheKey); found {
		utils.HTTP.JSONResponse(ctx, fasthttp.StatusOK, cached)
		return
	}

	deployment, found := informer.GetDeploymentByName(name)
	if !found {
		utils.HTTP.NotFound(ctx, fmt.Sprintf("Deployment '%s'", name))
		return
	}

	response := DeploymentResponse{
		Name:      deployment.Name,
		Namespace: deployment.Namespace,
		Replicas:  utils.K8sHelpers.GetReplicaCount(deployment.Spec.Replicas),
		Image:     utils.K8sHelpers.GetMainContainerImage(deployment),
		Age:       formatAge(deployment.CreationTimestamp.Time),
		Ready:     fmt.Sprintf("%d/%d", deployment.Status.ReadyReplicas, utils.K8sHelpers.GetReplicaCount(deployment.Spec.Replicas)),
	}

	// Cache the response
	utils.Cache.Set(cacheKey, response)
	
	utils.HTTP.JSONResponse(ctx, fasthttp.StatusOK, response)
}

func handleHealthOptimized(ctx *fasthttp.RequestCtx) {
	utils.HTTP.PlainResponse(ctx, fasthttp.StatusOK, "OK")
}

func handleDeploymentNamesOptimized(ctx *fasthttp.RequestCtx) {
	// Check cache first
	if cached, found := utils.Cache.Get("deployment_names"); found {
		utils.HTTP.JSONResponse(ctx, fasthttp.StatusOK, cached)
		return
	}

	names := informer.GetDeploymentNames()
	utils.Cache.Set("deployment_names", names)
	utils.HTTP.JSONResponse(ctx, fasthttp.StatusOK, names)
}

func handleDeploymentsOptimized(ctx *fasthttp.RequestCtx) {
	// Implementation with caching...
	deployments := informer.GetDeployments()
	
	response := make([]DeploymentResponse, 0, len(deployments))
	for _, dep := range deployments {
		response = append(response, DeploymentResponse{
			Name:      dep.Name,
			Namespace: dep.Namespace,
			Replicas:  utils.K8sHelpers.GetReplicaCount(dep.Spec.Replicas),
			Image:     utils.K8sHelpers.GetMainContainerImage(dep),
			Age:       formatAge(dep.CreationTimestamp.Time),
			Ready:     fmt.Sprintf("%d/%d", dep.Status.ReadyReplicas, utils.K8sHelpers.GetReplicaCount(dep.Spec.Replicas)),
		})
	}
	
	utils.HTTP.JSONResponse(ctx, fasthttp.StatusOK, response)
}

func handlePodNamesOptimized(ctx *fasthttp.RequestCtx) {
	names := informer.GetPodNames()
	utils.HTTP.JSONResponse(ctx, fasthttp.StatusOK, names)
}

func handlePodsOptimized(ctx *fasthttp.RequestCtx) {
	pods := informer.GetPods()
	
	response := make([]PodResponse, 0, len(pods))
	for _, pod := range pods {
		response = append(response, PodResponse{
			Name:         pod.Name,
			Namespace:    pod.Namespace,
			Phase:        string(pod.Status.Phase),
			Node:         pod.Spec.NodeName,
			Image:        utils.K8sHelpers.GetMainPodImage(pod),
			Age:          formatAge(pod.CreationTimestamp.Time),
			Ready:        utils.K8sHelpers.GetPodReadyCondition(pod),
			RestartCount: utils.K8sHelpers.GetPodRestartCount(pod),
		})
	}
	
	utils.HTTP.JSONResponse(ctx, fasthttp.StatusOK, response)
}

func handlePodByNameOptimized(ctx *fasthttp.RequestCtx, name string) {
	pod, found := informer.GetPodByName(name)
	if !found {
		utils.HTTP.NotFound(ctx, fmt.Sprintf("Pod '%s'", name))
		return
	}

	response := PodResponse{
		Name:         pod.Name,
		Namespace:    pod.Namespace,
		Phase:        string(pod.Status.Phase),
		Node:         pod.Spec.NodeName,
		Image:        utils.K8sHelpers.GetMainPodImage(pod),
		Age:          formatAge(pod.CreationTimestamp.Time),
		Ready:        utils.K8sHelpers.GetPodReadyCondition(pod),
		RestartCount: utils.K8sHelpers.GetPodRestartCount(pod),
	}
	
	utils.HTTP.JSONResponse(ctx, fasthttp.StatusOK, response)
}
