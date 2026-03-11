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
	"fmt"
	"slices"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	ctrl "sigs.k8s.io/controller-runtime"
)

var watcherLog = ctrl.Log.WithName("watcher")

// FailedPods holds the list of pod keys (namespace/name) that are in a failure state.
// It is populated by the pod watcher and consumed by the healer.
var FailedPods []string

// waitingFailureReasons contains the pod waiting reasons that trigger auto-healing.
var waitingFailureReasons = []string{
	"CrashLoopBackOff",
}

// terminatedFailureReasons contains the pod termination reasons that trigger auto-healing.
var terminatedFailureReasons = []string{
	"OOMKilled",
	"Error",
}

// GetPodStatus starts a SharedInformer that watches all pods in the cluster
// and detects containers in failure states (CrashLoopBackOff, OOMKilled, Error).
// Detected pod keys are added to the FailedPods slice for processing by the healer.
func GetPodStatus(clientset *kubernetes.Clientset) {
	factory := informers.NewSharedInformerFactory(clientset, 10*time.Minute)
	podInformer := factory.Core().V1().Pods()

	podInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(oldObj, newObj interface{}) {
			newPod := newObj.(*v1.Pod)
			checkPodHealth(newPod)
		},
	})

	factory.Start(wait.NeverStop)
	factory.WaitForCacheSync(wait.NeverStop)

	watcherLog.Info("pod watcher started")
	select {}
}

// checkPodHealth inspects a pod's container statuses and adds it to the
// FailedPods list if any container is in a recognized failure state.
func checkPodHealth(pod *v1.Pod) {
	podKey := fmt.Sprintf("%s/%s", pod.Namespace, pod.Name)

	for _, cs := range pod.Status.ContainerStatuses {
		// Check for waiting failure states (e.g., CrashLoopBackOff)
		if cs.State.Waiting != nil {
			if slices.Contains(waitingFailureReasons, cs.State.Waiting.Reason) {
				addFailedPod(podKey, cs.State.Waiting.Reason)
				return
			}
		}

		// Check for terminated failure states (e.g., OOMKilled, Error)
		if cs.State.Terminated != nil {
			if slices.Contains(terminatedFailureReasons, cs.State.Terminated.Reason) {
				addFailedPod(podKey, cs.State.Terminated.Reason)
				return
			}
		}
	}
}

// addFailedPod adds a pod key to the FailedPods list if it's not already present.
func addFailedPod(podKey, reason string) {
	if !slices.Contains(FailedPods, podKey) {
		FailedPods = append(FailedPods, podKey)
		watcherLog.Info("detected failed pod", "pod", podKey, "reason", reason)
	}
}
