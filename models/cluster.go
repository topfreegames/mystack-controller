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
	namespace   string
	username    string
	deployments []*Deployment
	services    []*Service
}

//NewCluster returns a new cluster ready to start
func NewCluster(username string, deployments []*Deployment) *Cluster {
	namespace := usernameToNamespace(username)
	services := servicesFromDeployment(deployments)
	return &Cluster{
		username:    username,
		namespace:   namespace,
		deployments: deployments,
		services:    services,
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
	err := CreateNamespace(clientset, c.username)
	if err != nil {
		return err
	}

	for _, deployment := range c.deployments {
		_, err = deployment.Deploy(clientset)
		if err != nil {
			//TODO: maybe delete already created deploys?
			return err
		}
	}

	for _, service := range c.services {
		_, err = service.Expose(clientset)
		if err != nil {
			//TODO: maybe delete already created deploys and services?
			return err
		}
	}

	return nil
}

//Delete deletes namespace and all deployments and services
func (c *Cluster) Delete(clientset kubernetes.Interface) error {
	var err error
	for _, service := range c.services {
		err = service.Delete(clientset)
		if err != nil {
			//TODO: maybe delete already created deploys and services?
			return err
		}
	}

	for _, deployment := range c.deployments {
		err = deployment.Delete(clientset)
		if err != nil {
			//TODO: maybe delete already created deploys?
			return err
		}
	}

	err = DeleteNamespace(clientset, c.username)
	if err != nil {
		return err
	}

	return nil
}
