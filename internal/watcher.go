package internal

import (
	"fmt"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"slices"
	"time"
)

var FailedPods []string

func GetPodStatus(clientset *kubernetes.Clientset) {

	factory := informers.NewSharedInformerFactory(clientset, time.Minute*10)
	podInformer := factory.Core().V1().Pods()

	podInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(oldObj, newObj interface{}) {

			newPod := newObj.(*v1.Pod)

			for _, cs := range newPod.Status.ContainerStatuses {

				// CrashLoopBackOff
				if cs.State.Waiting != nil {
					reason := cs.State.Waiting.Reason

					if reason == "CrashLoopBackOff" {
						podKey := fmt.Sprintf("%s/%s", newPod.Namespace, newPod.Name)

						if !slices.Contains(FailedPods, podKey) {
							FailedPods = append(FailedPods, podKey)
							fmt.Println("Added failed pod:", podKey)
						}
					}
				}

				// OOMKilled / Error
				if cs.State.Terminated != nil {
					reason := cs.State.Terminated.Reason

					if reason == "OOMKilled" || reason == "Error" {
						podKey := fmt.Sprintf("%s/%s", newPod.Namespace, newPod.Name)

						if !slices.Contains(FailedPods, podKey) {
							FailedPods = append(FailedPods, podKey)
							fmt.Println("Added terminated pod:", podKey)
						}
					}
				}
			}
		},
	})

	factory.Start(wait.NeverStop)
	factory.WaitForCacheSync(wait.NeverStop)

	fmt.Println("qPod watcher started...")
	select {}
}
