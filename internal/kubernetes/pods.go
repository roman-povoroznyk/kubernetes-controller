package kubernetes

import (
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const defaultTimeout = 10 * time.Second

// CreatePod creates a new pod with the given name in the specified namespace
func CreatePod(clientset kubernetes.Interface, namespace, podName string) error {
	log.Debug().Str("namespace", namespace).Str("name", podName).Msg("Creating pod via API")

	ctx, cancel := DefaultContext()
	defer cancel()

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: podName,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "nginx",
					Image: "nginx:alpine",
					Ports: []corev1.ContainerPort{
						{
							ContainerPort: 80,
							Protocol:      corev1.ProtocolTCP,
						},
					},
				},
			},
		},
	}

	_, err := clientset.CoreV1().Pods(namespace).Create(ctx, pod, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create pod %s: %w", podName, err)
	}

	// kubectl-style output
	fmt.Printf("pod/%s created\n", podName)
	return nil
}

// DeletePod deletes the pod with the given name in the specified namespace
func DeletePod(clientset kubernetes.Interface, namespace, podName string) error {
	log.Debug().Str("namespace", namespace).Str("name", podName).Msg("Deleting pod via API")
	ctx, cancel := DefaultContext()
	defer cancel()

	err := clientset.CoreV1().Pods(namespace).Delete(ctx, podName, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete pod %s: %w", podName, err)
	}

	// kubectl-style output
	fmt.Printf("pod \"%s\" deleted\n", podName)
	return nil
}

// ListPods lists all pods in the specified namespace
func ListPods(clientset kubernetes.Interface, namespace string) error {
	log.Debug().Str("namespace", namespace).Msg("Listing pods via API")

	ctx, cancel := DefaultContext()
	defer cancel()

	pods, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list pods: %w", err)
	}

	if len(pods.Items) == 0 {
		fmt.Printf("No resources found in %s namespace.\n", namespace)
		return nil
	}

	// Properly formatted header like kubectl
	fmt.Printf("%-50s %-7s %-10s %-10s %s\n", "NAME", "READY", "STATUS", "RESTARTS", "AGE")

	for _, pod := range pods.Items {
		ready := fmt.Sprintf("%d/%d", countReadyContainers(pod), len(pod.Spec.Containers))
		status := string(pod.Status.Phase)
		restarts := countRestarts(pod)
		age := formatAge(pod.CreationTimestamp.Time)

		fmt.Printf("%-50s %-7s %-10s %-10d %s\n",
			pod.Name, ready, status, restarts, age)
	}

	return nil
}

// countReadyContainers returns the number of ready containers in a pod
func countReadyContainers(pod corev1.Pod) int {
	ready := 0
	for _, condition := range pod.Status.Conditions {
		if condition.Type == corev1.PodReady && condition.Status == corev1.ConditionTrue {
			ready++
			break
		}
	}
	return ready
}

// countRestarts returns the total number of restarts across all containers in a pod
func countRestarts(pod corev1.Pod) int32 {
	var restarts int32
	for _, containerStatus := range pod.Status.ContainerStatuses {
		restarts += containerStatus.RestartCount
	}
	return restarts
}
