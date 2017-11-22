package main

import (
	"fmt"
	"os"
	"path"

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type Clientset struct {
	kubernetes.Interface
}

func NewKubeClient(configPath string, context string, cluster string) (*Clientset, error) {
	configOverrides := &clientcmd.ConfigOverrides{}
	if configPath == "" {
		configPath = path.Join(os.Getenv("HOME"), ".kube", "config")
	}

	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.ExplicitPath = configPath

	if context != "" {
		configOverrides.CurrentContext = context
	}

	if cluster != "" {
		configOverrides.Context.Cluster = cluster
	}

	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides).ClientConfig()
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	return &Clientset{Interface: clientset}, err
}

func (clientset *Clientset) GetEndpoints(namespace string, enpointName string) ([]string, error) {
	tillers := []string{}

	endpoints, err := clientset.CoreV1().Endpoints(namespace).Get(enpointName, meta_v1.GetOptions{})
	if err != nil {
		return nil, err
	}
	for _, sub := range endpoints.Subsets {
		for _, port := range sub.Ports {
			for _, addr := range sub.Addresses {
				tillers = append(tillers, fmt.Sprintf("%s:%d", addr.IP, port.Port))
			}
		}
	}
	return tillers, nil
}
