// mystack-controller api
// https://github.com/topfreegames/mystack-controller
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2017 Top Free Games <backend@tfgco.com>

package models

import (
	runner "gopkg.in/mgutz/dat.v2/sqlx-runner"
	"k8s.io/client-go/kubernetes"
)

//Cluster represents a k8s cluster for a user
type Cluster struct {
	Namespace   string
	Username    string
	Deployments []*Deployment
	Services    []*Service
}

//NewCluster returns a new cluster ready to start
func NewCluster(db runner.Connection, username, clusterName string) (*Cluster, error) {
	namespace := usernameToNamespace(username)

	apps, services, err := LoadClusterConfig(db, clusterName)
	if err != nil {
		return nil, err
	}

	k8sDeployments := buildDeployments(apps, services, username)
	k8sServices := buildServices(k8sDeployments, username)

	cluster := &Cluster{
		Username:    username,
		Namespace:   namespace,
		Deployments: k8sDeployments,
		Services:    k8sServices,
	}

	return cluster, nil
}

func buildDeployments(apps, services map[string]*ClusterAppConfig, username string) []*Deployment {
	deployments := make([]*Deployment, len(apps)+len(services))

	i := 0
	for name, config := range services {
		deployments[i] = NewDeployment(name, username, config.Image, config.Port, config.Environment)
		i = i + 1
	}

	for name, config := range apps {
		deployments[i] = NewDeployment(name, username, config.Image, config.Port, config.Environment)
		i = i + 1
	}

	return deployments
}

func buildServices(deployments []*Deployment, username string) []*Service {
	services := make([]*Service, len(deployments))
	for i, deployment := range deployments {
		services[i] = NewService(deployment.Name, username, 80, deployment.Port)
	}
	return services
}

//Create creates namespace, deployments and services
func (c *Cluster) Create(clientset kubernetes.Interface) error {
	err := CreateNamespace(clientset, c.Username)
	if err != nil {
		return err
	}

	for _, deployment := range c.Deployments {
		_, err = deployment.Deploy(clientset)
		if err != nil {
			//TODO: maybe delete already created deploys?
			return err
		}
	}

	for _, service := range c.Services {
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
	for _, service := range c.Services {
		err = service.Delete(clientset)
		if err != nil {
			//TODO: maybe delete already created deploys and services?
			return err
		}
	}

	for _, deployment := range c.Deployments {
		err = deployment.Delete(clientset)
		if err != nil {
			//TODO: maybe delete already created deploys?
			return err
		}
	}

	err = DeleteNamespace(clientset, c.Username)
	if err != nil {
		return err
	}

	return nil
}
