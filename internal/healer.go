package internal

import (
	"context"
	"fmt"
	Metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"strings"
)

func HealFailedPods(clientset *kubernetes.Clientset) {
	ctx := context.Background()

	if len(FailedPods) == 0 {
		return
	}

	fmt.Println("🩺 Healer scanning failed pods...")

	var remaining []string

	for _, item := range FailedPods {

		parts := strings.Split(item, "/")
		if len(parts) != 2 {
			fmt.Println("Invalid pod key:", item)
			continue
		}

		namespace := parts[0]
		podName := parts[1]

		fmt.Printf("Healing pod: %s/%s\n", namespace, podName)

		err := clientset.CoreV1().
			Pods(namespace).
			Delete(ctx, podName, Metav1.DeleteOptions{})

		if err != nil {
			fmt.Println("Error deleting pod:", err)
			remaining = append(remaining, item) // keep in list
		} else {
			fmt.Printf("Deleted pod: %s/%s\n", namespace, podName)
		}
	}

	// update list after healing
	FailedPods = remaining
}
