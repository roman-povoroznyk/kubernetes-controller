package kubernetes

import (
	"fmt"

	"github.com/rs/zerolog/log"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// CreateDeployment creates a simple nginx deployment
func CreateDeployment(clientset kubernetes.Interface, namespace, name string) error {
	log.Debug().Str("namespace", namespace).Str("name", name).Msg("Creating deployment via API")

	replicas := int32(1)
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": name},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": name},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "nginx",
							Image: "nginx:alpine",
							Ports: []corev1.ContainerPort{
								{ContainerPort: 80},
							},
						},
					},
				},
			},
		},
	}

	ctx, cancel := DefaultContext()
	defer cancel()

	_, err := clientset.AppsV1().Deployments(namespace).Create(ctx, deployment, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create deployment %s: %w", name, err)
	}
	fmt.Printf("deployment/%s created\n", name)
	return nil
}

// ListDeployments lists deployments in the given namespace
func ListDeployments(clientset kubernetes.Interface, namespace string) error {
	log.Debug().Str("namespace", namespace).Msg("Listing deployments")

	ctx, cancel := DefaultContext()
	defer cancel()

	deployments, err := clientset.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list deployments: %w", err)
	}

	if len(deployments.Items) == 0 {
		fmt.Println("No deployments found in namespace:", namespace)
		return nil
	}

	// Print header
	fmt.Printf("%-30s %-10s %-12s %-12s %-12s\n", "NAME", "READY", "UP-TO-DATE", "AVAILABLE", "AGE")

	// Print each deployment
	for _, d := range deployments.Items {
		age := formatAge(d.CreationTimestamp.Time)

		var desiredReplicas int32 = 0
		if d.Spec.Replicas != nil {
			desiredReplicas = *d.Spec.Replicas
		}

		ready := fmt.Sprintf("%d/%d", d.Status.ReadyReplicas, desiredReplicas)
		fmt.Printf("%-30s %-10s %-12d %-12d %-12s\n",
			d.Name,
			ready,
			d.Status.UpdatedReplicas,
			d.Status.AvailableReplicas,
			age)
	}

	return nil
}

// DeleteDeployment deletes a deployment
func DeleteDeployment(clientset kubernetes.Interface, namespace, name string) error {
	log.Debug().Str("namespace", namespace).Str("name", name).Msg("Deleting deployment via API")

	ctx, cancel := DefaultContext()
	defer cancel()

	err := clientset.AppsV1().Deployments(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete deployment %s: %w", name, err)
	}
	fmt.Printf("deployment \"%s\" deleted\n", name)
	return nil
}
