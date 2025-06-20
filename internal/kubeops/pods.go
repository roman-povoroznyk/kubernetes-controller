package kubeops

import (
    "context"
    "fmt"
    "time"
    "k8s.io/client-go/kubernetes"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    corev1 "k8s.io/api/core/v1"
    "github.com/rs/zerolog/log"
)

// Timeout for API operations
const defaultTimeout = 10 * time.Second

func ListPods(clientset kubernetes.Interface, namespace string) error {
    log.Debug().Str("namespace", namespace).Msg("Listing pods via API")

    ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
    defer cancel()

    pods, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
    if err != nil {
        return fmt.Errorf("failed to list pods: %w", err)
    }

    if len(pods.Items) == 0 {
        fmt.Println("No pods found.")
        return nil
    }

    for _, pod := range pods.Items {
        fmt.Println(pod.Name)
    }
    return nil
}

func DeletePod(clientset kubernetes.Interface, namespace, podName string) error {
    log.Debug().Str("namespace", namespace).Str("name", podName).Msg("Deleting pod via API")

    ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
    defer cancel()

    err := clientset.CoreV1().Pods(namespace).Delete(ctx, podName, metav1.DeleteOptions{})
    if err != nil {
        return fmt.Errorf("failed to delete pod %s: %w", podName, err)
    }

    return nil
}

func CreatePod(clientset kubernetes.Interface, namespace, podName string) error {
    log.Debug().Str("namespace", namespace).Str("name", podName).Msg("Creating pod via API")

    ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
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
                },
            },
        },
    }

    _, err := clientset.CoreV1().Pods(namespace).Create(ctx, pod, metav1.CreateOptions{})
    if err != nil {
        return fmt.Errorf("failed to create pod %s: %w", podName, err)
    }

    return nil
}
