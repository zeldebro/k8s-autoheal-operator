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
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// ClusterConnect establishes a connection to the Kubernetes cluster.
// It first attempts in-cluster configuration (for running inside a pod),
// then falls back to the local kubeconfig file (~/.kube/config).
func ClusterConnect() (*kubernetes.Clientset, error) {
	var config *rest.Config
	var err error

	// Try in-cluster config first (when running as a pod)
	config, err = rest.InClusterConfig()
	if err != nil {
		// Fallback to kubeconfig for local development
		home := homedir.HomeDir()
		kubeconfig := filepath.Join(home, ".kube", "config")
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, err
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}
