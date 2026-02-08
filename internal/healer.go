package internal

import (
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/workqueue"
	"strings"
	"time"
)

// global queue
var queue = workqueue.NewRateLimitingQueue(
	workqueue.DefaultControllerRateLimiter(),
)

// track restarted deployments (avoid duplicate restart)
var restarted = make(map[string]bool)

// ----------------------------------------------------
// FIND FAILED PODS → RETURN DEPLOYMENTS
// ----------------------------------------------------
func FailedDeployments(clientset *kubernetes.Clientset) []string {

	ctx := context.Background()
	var result []string

	if len(FailedPods) == 0 {
		return result
	}

	fmt.Println("🩺 Scanning failed pods...")

	for _, item := range FailedPods {

		parts := strings.Split(item, "/")
		if len(parts) != 2 {
			fmt.Println("Invalid pod key:", item)
			continue
		}

		namespace := parts[0]
		podName := parts[1]

		pod, err := clientset.CoreV1().
			Pods(namespace).
			Get(ctx, podName, metav1.GetOptions{})
		if err != nil {
			continue
		}

		// find ReplicaSet
		for _, owner := range pod.OwnerReferences {
			if owner.Kind == "ReplicaSet" {

				rs, err := clientset.AppsV1().
					ReplicaSets(namespace).
					Get(ctx, owner.Name, metav1.GetOptions{})
				if err != nil {
					continue
				}

				// find Deployment
				for _, rsOwner := range rs.OwnerReferences {
					if rsOwner.Kind == "Deployment" {

						key := namespace + "/" + rsOwner.Name
						result = append(result, key)
						fmt.Println("Found deployment:", key)
					}
				}
			}
		}
	}

	return result
}

// ----------------------------------------------------
// PUSH DEPLOYMENTS TO QUEUE (runs forever)
// ----------------------------------------------------
func PushFailedDeploymentsToQueue(clientset *kubernetes.Clientset) {

	for {
		deployments := FailedDeployments(clientset)

		for _, d := range deployments {
			queue.Add(d)
			fmt.Println("Added to queue:", d)
		}

		time.Sleep(20 * time.Second) // scan interval
	}
}

// WORKER: RESTART DEPLOYMENT

func HealFailedPods(clientset *kubernetes.Clientset) {

	fmt.Println("Auto-healer worker started")

	for {

		item, shutdown := queue.Get()
		if shutdown {
			fmt.Println("Queue shutdown")
			return
		}

		key := item.(string)
		fmt.Println("Processing:", key)

		parts := strings.Split(key, "/")
		if len(parts) != 2 {
			queue.Done(item)
			continue
		}

		namespace := parts[0]
		deploy := parts[1]

		// avoid duplicate restart
		if restarted[key] {
			fmt.Println("Already restarted:", key)
			queue.Done(item)
			continue
		}

		ctx := context.Background()

		deployObj, err := clientset.AppsV1().
			Deployments(namespace).
			Get(ctx, deploy, metav1.GetOptions{})

		if err != nil {
			fmt.Println("Deployment not found:", err)
			queue.AddRateLimited(key)
			queue.Done(item)
			continue
		}

		// rollout restart using annotation
		if deployObj.Spec.Template.Annotations == nil {
			deployObj.Spec.Template.Annotations = map[string]string{}
		}

		deployObj.Spec.Template.Annotations["k8s.autoheal.operator/restartedAt"] =
			time.Now().Format(time.RFC3339)

		_, err = clientset.AppsV1().
			Deployments(namespace).
			Update(ctx, deployObj, metav1.UpdateOptions{})

		if err != nil {
			fmt.Println("Restart failed:", err)
			queue.AddRateLimited(key)
			queue.Done(item)
			continue
		}

		fmt.Println("Restarted deployment:", key)

		restarted[key] = true

		queue.Forget(item)
		queue.Done(item)

		time.Sleep(10 * time.Second)
	}
}
