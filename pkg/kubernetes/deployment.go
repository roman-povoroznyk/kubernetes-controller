package kubernetes

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DeploymentList lists deployments in the specified namespace
func (c *Client) DeploymentList(namespace string) (*appsv1.DeploymentList, error) {
	return c.clientset.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{})
}

// DeploymentCreate creates a new deployment
func (c *Client) DeploymentCreate(namespace, name, image string, replicas int32) error {
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  name,
							Image: image,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 80,
								},
							},
						},
					},
				},
			},
		},
	}

	_, err := c.clientset.AppsV1().Deployments(namespace).Create(context.TODO(), deployment, metav1.CreateOptions{})
	return err
}

// DeploymentDelete deletes a deployment
func (c *Client) DeploymentDelete(namespace, name string) error {
	return c.clientset.AppsV1().Deployments(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
}

// DeploymentPrint prints deployments in kubectl-like format
func DeploymentPrint(deployments []appsv1.Deployment, showNamespace bool) {
	if len(deployments) == 0 {
		fmt.Println("No resources found.")
		return
	}

	// Create tabwriter for aligned output
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	// Print header
	if showNamespace {
		fmt.Fprintln(w, "NAMESPACE\tNAME\tREADY\tUP-TO-DATE\tAVAILABLE\tAGE")
	} else {
		fmt.Fprintln(w, "NAME\tREADY\tUP-TO-DATE\tAVAILABLE\tAGE")
	}

	// Print each deployment
	for _, deploy := range deployments {
		ready := fmt.Sprintf("%d/%d", deploy.Status.ReadyReplicas, deploy.Status.Replicas)
		upToDate := fmt.Sprintf("%d", deploy.Status.UpdatedReplicas)
		available := fmt.Sprintf("%d", deploy.Status.AvailableReplicas)
		age := FormatAge(deploy.CreationTimestamp.Time)

		if showNamespace {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
				deploy.Namespace, deploy.Name, ready, upToDate, available, age)
		} else {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				deploy.Name, ready, upToDate, available, age)
		}
	}
}

// FormatAge formats age like kubectl
func FormatAge(t time.Time) string {
	now := time.Now()
	age := now.Sub(t)

	days := int(age.Hours() / 24)
	hours := int(age.Hours()) % 24
	minutes := int(age.Minutes()) % 60
	seconds := int(age.Seconds()) % 60

	if days >= 7 {
		// 7d+ - show only days
		return fmt.Sprintf("%dd", days)
	} else if days >= 2 {
		// 2-7d - show days and hours
		if hours > 0 {
			return fmt.Sprintf("%dd%dh", days, hours)
		}
		return fmt.Sprintf("%dd", days)
	} else if days >= 1 {
		// 1-2d - show days and hours
		if hours > 0 {
			return fmt.Sprintf("%dd%dh", days, hours)
		}
		return fmt.Sprintf("%dd", days)
	} else if age.Hours() >= 1 {
		// 1h+ - show hours and minutes
		if minutes > 0 {
			return fmt.Sprintf("%dh%dm", hours, minutes)
		}
		return fmt.Sprintf("%dh", hours)
	} else if age.Minutes() >= 1 {
		// 1m+ - show minutes and seconds
		if seconds > 0 {
			return fmt.Sprintf("%dm%ds", minutes, seconds)
		}
		return fmt.Sprintf("%dm", minutes)
	} else {
		// <1m - show seconds
		return fmt.Sprintf("%ds", seconds)
	}
}
