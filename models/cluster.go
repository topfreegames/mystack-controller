// mystack-controller api
// https://github.com/topfreegames/mystack-controller
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2017 Top Free Games <backend@tfgco.com>

package models

import (
	"fmt"
	"github.com/topfreegames/mystack-controller/errors"
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
func NewCluster(db DB, username, clusterName string) (*Cluster, error) {
	namespace := usernameToNamespace(username)

	apps, services, err := LoadClusterConfig(db, clusterName)
	if err != nil {
		return nil, err
	}

	k8sAppDeployments := buildDeployments(apps, username)
	k8sSvcDeployments := buildDeployments(services, username)
	k8sServices := buildServices(k8sAppDeployments, k8sSvcDeployments, username)

	cluster := &Cluster{
		Username:    username,
		Namespace:   namespace,
		Deployments: append(k8sAppDeployments, k8sSvcDeployments...),
		Services:    k8sServices,
	}

	return cluster, nil
}

func buildDeployments(types map[string]*ClusterAppConfig, username string) []*Deployment {
	deployments := make([]*Deployment, len(types))

	i := 0
	for name, config := range types {
		deployments[i] = NewDeployment(name, username, config.Image, config.Port, config.Environment)
		i = i + 1
	}

	return deployments
}

func buildServices(
	apps []*Deployment,
	svcs []*Deployment,
	username string,
) []*Service {
	services := make([]*Service, len(apps)+len(svcs))
	i := 0
	for _, app := range apps {
		services[i] = NewService(app.Name, username, 80, app.Port)
		i = i + 1
	}
	for _, svc := range svcs {
		services[i] = NewService(svc.Name, username, 80, svc.Port)
		i = i + 1
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
			nsErr := DeleteNamespace(clientset, c.Username)
			if nsErr != nil {
				return errors.NewKubernetesError(
					"create cluster error",
					fmt.Errorf("error during cluster creation and could not rollback: %s", nsErr.Error()),
				)
			}
			return err
		}
	}

	for _, service := range c.Services {
		_, err = service.Expose(clientset)
		if err != nil {
			nsErr := DeleteNamespace(clientset, c.Username)
			if nsErr != nil {
				return errors.NewKubernetesError(
					"create cluster error",
					fmt.Errorf("error during cluster creation and could not rollback: %s", nsErr.Error()),
				)
			}
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
			return err
		}
	}

	for _, deployment := range c.Deployments {
		err = deployment.Delete(clientset)
		if err != nil {
			return err
		}
	}

	err = DeleteNamespace(clientset, c.Username)
	if err != nil {
		return err
	}

	return nil
}
