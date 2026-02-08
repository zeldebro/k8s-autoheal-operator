package internal

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func ClusterConnect() (*kubernetes.Clientset, error) {
	var config *rest.Config
	var err error

	// try in-cluster first
	config, err = rest.InClusterConfig()
	if err != nil {
		// fallback to kubeconfig (local)
		home := homedir.HomeDir()
		config, err = clientcmd.BuildConfigFromFlags("", home+"/.kube/config")
		if err != nil {
			return nil, err
		}
	}

	// create clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}
