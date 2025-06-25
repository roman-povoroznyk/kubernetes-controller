package kubernetes

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// ListDeployments lists deployments in the given namespace
func ListDeployments(clientset kubernetes.Interface, namespace string) error {
	log.Debug().Str("namespace", namespace).Msg("Listing deployments")

	deployments, err := clientset.AppsV1().Deployments(namespace).List(context.Background(), metav1.ListOptions{})
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
