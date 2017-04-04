//THIS FILE IS TEMPORARY AND MUST NOT BE COMMITED

package main

import (
	"github.com/topfreegames/mystack-controller/models"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func main() {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	deployment := models.NewDeployment("hello", "henrique.rodrigues", "hello-world")
	_, err = deployment.Deploy(clientset)
	if err != nil {
		panic(err.Error())
	}
}
