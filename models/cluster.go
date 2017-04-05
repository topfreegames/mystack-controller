// mystack-controller api
// https://github.com/topfreegames/mystack-controller
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2017 Top Free Games <backend@tfgco.com>

package models

import (
	"k8s.io/client-go/kubernetes"
)

//Cluster represents a k8s cluster for a user
type Cluster struct {
	Namespace   string
	Deployments []*Deployment
	Services    []*Service
}

//NewCluster returns a new cluster ready to start
func NewCluster(username string, deployments []*Deployment) *Cluster {
	namespace := usernameToNamespace(username)
	services := servicesFromDeployment(deployments)
	return &Cluster{
		Namespace:   namespace,
		Deployments: deployments,
		Services:    services,
	}
}

func servicesFromDeployment(deployments []*Deployment) []*Service {
	services := make([]*Service, len(deployments))
	for i, deployment := range deployments {
		services[i] = &Service{
			Name:       deployment.Name,
			Namespace:  deployment.Namespace,
			Port:       80,
			TargetPort: deployment.Port,
		}
	}
	return services
}

//Create creates namespace, deployments and services
func (c *Cluster) Create(clientset kubernetes.Interface) error {
	return nil
}

//Delete deletes namespace and all deployments and services
func (c *Cluster) Delete(clientset kubernetes.Interface) error {
	return nil
}
