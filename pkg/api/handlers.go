package api

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
	"k8s.io/client-go/kubernetes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// APIHandler handles API requests
type APIHandler struct {
	clientset kubernetes.Interface
}

// NewAPIHandler creates new API handler
func NewAPIHandler(clientset kubernetes.Interface) *APIHandler {
	return &APIHandler{clientset: clientset}
}

// GetDeployments returns list of deployments
func (h *APIHandler) GetDeployments(c *fiber.Ctx) error {
	namespace := c.Query("namespace", "")
	
	var deployments *appsv1.DeploymentList
	var err error
	
	if namespace != "" {
		deployments, err = h.clientset.AppsV1().Deployments(namespace).List(c.Context(), metav1.ListOptions{})
	} else {
		deployments, err = h.clientset.AppsV1().Deployments(metav1.NamespaceAll).List(c.Context(), metav1.ListOptions{})
	}
	
	if err != nil {
		log.Error().Err(err).Msg("Failed to list deployments")
		return c.Status(500).JSON(ErrorResponse{
			Error:   "Internal Server Error",
			Code:    500,
			Message: err.Error(),
		})
	}

	items := make([]DeploymentResponse, len(deployments.Items))
	for i, dep := range deployments.Items {
		items[i] = FromK8sDeployment(&dep)
	}

	response := DeploymentListResponse{
		Items: items,
		Total: len(items),
	}

	return c.JSON(response)
}

// GetDeployment returns specific deployment
func (h *APIHandler) GetDeployment(c *fiber.Ctx) error {
	namespace := c.Params("namespace")
	name := c.Params("name")

	deployment, err := h.clientset.AppsV1().Deployments(namespace).Get(c.Context(), name, metav1.GetOptions{})
	if err != nil {
		log.Error().Err(err).Str("namespace", namespace).Str("name", name).Msg("Failed to get deployment")
		return c.Status(404).JSON(ErrorResponse{
			Error:   "Not Found",
			Code:    404,
			Message: fmt.Sprintf("Deployment %s/%s not found", namespace, name),
		})
	}

	response := FromK8sDeployment(deployment)
	return c.JSON(response)
}

// CreateDeployment creates new deployment
func (h *APIHandler) CreateDeployment(c *fiber.Ctx) error {
	var req CreateDeploymentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(ErrorResponse{
			Error:   "Bad Request",
			Code:    400,
			Message: "Invalid JSON payload",
		})
	}

	// Create deployment object
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Name,
			Namespace: req.Namespace,
			Labels:    req.Labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &req.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": req.Name},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": req.Name},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  req.Name,
							Image: req.Image,
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("500m"),
									corev1.ResourceMemory: resource.MustParse("512Mi"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("100m"),
									corev1.ResourceMemory: resource.MustParse("128Mi"),
								},
							},
						},
					},
				},
			},
		},
	}

	if req.Port > 0 {
		deployment.Spec.Template.Spec.Containers[0].Ports = []corev1.ContainerPort{
			{
				ContainerPort: req.Port,
				Protocol:      corev1.ProtocolTCP,
			},
		}
	}

	created, err := h.clientset.AppsV1().Deployments(req.Namespace).Create(c.Context(), deployment, metav1.CreateOptions{})
	if err != nil {
		log.Error().Err(err).Msg("Failed to create deployment")
		return c.Status(500).JSON(ErrorResponse{
			Error:   "Internal Server Error",
			Code:    500,
			Message: err.Error(),
		})
	}

	response := FromK8sDeployment(created)
	return c.Status(201).JSON(response)
}

// DeleteDeployment deletes deployment
func (h *APIHandler) DeleteDeployment(c *fiber.Ctx) error {
	namespace := c.Params("namespace")
	name := c.Params("name")

	err := h.clientset.AppsV1().Deployments(namespace).Delete(c.Context(), name, metav1.DeleteOptions{})
	if err != nil {
		log.Error().Err(err).Str("namespace", namespace).Str("name", name).Msg("Failed to delete deployment")
		return c.Status(404).JSON(ErrorResponse{
			Error:   "Not Found",
			Code:    404,
			Message: fmt.Sprintf("Deployment %s/%s not found", namespace, name),
		})
	}

	return c.Status(204).Send(nil)
}
