package kubeops

import (
    "context"
    "fmt"
    "k8s.io/client-go/kubernetes"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    corev1 "k8s.io/api/core/v1"
)

func ListPods(clientset kubernetes.Interface, namespace string) error {
    pods, err := clientset.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{})
    if err != nil {
        return err
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
    return clientset.CoreV1().Pods(namespace).Delete(context.Background(), podName, metav1.DeleteOptions{})
}

func CreatePod(clientset kubernetes.Interface, namespace, podName string) error {
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
    _, err := clientset.CoreV1().Pods(namespace).Create(context.Background(), pod, metav1.CreateOptions{})
    return err
}
