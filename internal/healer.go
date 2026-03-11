/*
/*
Copyright 2026 k8s-autoheal-operator Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package internal

import (
	"context"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
)

var healerLog = ctrl.Log.WithName("healer")

const (
	// scanInterval is the time between scans for failed deployments.
	scanInterval = 20 * time.Second

	// healCooldown is the minimum time between consecutive heal operations.
	healCooldown = 10 * time.Second

	// restartAnnotationKey is the annotation used to trigger a rolling restart.
	restartAnnotationKey = "k8s-autoheal-operator/restartedAt"
)

// queue is a rate-limited work queue used to process deployments that need healing.
var queue = workqueue.NewRateLimitingQueue(
	workqueue.DefaultControllerRateLimiter(),
)

// restarted tracks deployments that have already been restarted to avoid duplicate restarts.
var restarted = make(map[string]bool)

// FailedDeployments scans the FailedPods list, traces each pod back to its
// parent Deployment (via ReplicaSet owner references), and returns a list
// of deployment keys (namespace/name) that need healing.
func FailedDeployments(clientset *kubernetes.Clientset) []string {
	ctx := context.Background()
	var result []string

	if len(FailedPods) == 0 {
		return result
	}

	healerLog.Info("scanning failed pods for parent deployments")

	for _, item := range FailedPods {
		parts := strings.Split(item, "/")
		if len(parts) != 2 {
			healerLog.Info("invalid pod key format", "key", item)
			continue
		}

		namespace := parts[0]
		podName := parts[1]

		pod, err := clientset.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
		if err != nil {
			healerLog.V(1).Info("failed to get pod", "pod", item, "error", err)
			continue
		}

		// Trace: Pod → ReplicaSet → Deployment
		for _, owner := range pod.OwnerReferences {
			if owner.Kind != "ReplicaSet" {
				continue
			}

			rs, err := clientset.AppsV1().ReplicaSets(namespace).Get(ctx, owner.Name, metav1.GetOptions{})
			if err != nil {
				healerLog.V(1).Info("failed to get ReplicaSet", "namespace", namespace, "replicaset", owner.Name, "error", err)
				continue
			}

			for _, rsOwner := range rs.OwnerReferences {
				if rsOwner.Kind == "Deployment" {
					key := namespace + "/" + rsOwner.Name
					result = append(result, key)
					healerLog.Info("found parent deployment", "deployment", key)
				}
			}
		}
	}

	return result
}

// PushFailedDeploymentsToQueue continuously scans for failed deployments
// and adds them to the rate-limited work queue for processing by the healer.
func PushFailedDeploymentsToQueue(clientset *kubernetes.Clientset) {
	for {
		deployments := FailedDeployments(clientset)

		for _, d := range deployments {
			queue.Add(d)
			healerLog.Info("queued deployment for healing", "deployment", d)
		}

		time.Sleep(scanInterval)
	}
}

// HealFailedPods is the worker that processes the healing queue.
// For each deployment in the queue, it performs a rolling restart by
// updating the pod template annotation with the current timestamp.
// This is equivalent to `kubectl rollout restart deployment/<name>`.
func HealFailedPods(clientset *kubernetes.Clientset) {
	healerLog.Info("auto-healer worker started")

	for {
		item, shutdown := queue.Get()
		if shutdown {
			healerLog.Info("healing queue shut down")
			return
		}

		key := item.(string)
		healerLog.Info("processing deployment", "deployment", key)

		parts := strings.Split(key, "/")
		if len(parts) != 2 {
			healerLog.Info("invalid deployment key format", "key", key)
			queue.Done(item)
			continue
		}

		namespace := parts[0]
		deployName := parts[1]

		// Skip if already restarted
		if restarted[key] {
			healerLog.V(1).Info("deployment already restarted, skipping", "deployment", key)
			queue.Done(item)
			continue
		}

		ctx := context.Background()

		deployObj, err := clientset.AppsV1().Deployments(namespace).Get(ctx, deployName, metav1.GetOptions{})
		if err != nil {
			healerLog.Error(err, "failed to get deployment", "deployment", key)
			queue.AddRateLimited(key)
			queue.Done(item)
			continue
		}

		// Trigger rolling restart via annotation update
		if deployObj.Spec.Template.Annotations == nil {
			deployObj.Spec.Template.Annotations = map[string]string{}
		}
		deployObj.Spec.Template.Annotations[restartAnnotationKey] = time.Now().Format(time.RFC3339)

		_, err = clientset.AppsV1().Deployments(namespace).Update(ctx, deployObj, metav1.UpdateOptions{})
		if err != nil {
			healerLog.Error(err, "failed to restart deployment", "deployment", key)
			queue.AddRateLimited(key)
			queue.Done(item)
			continue
		}

		healerLog.Info("successfully restarted deployment", "deployment", key)
		restarted[key] = true
		queue.Forget(item)
		queue.Done(item)

		time.Sleep(healCooldown)
	}
}
