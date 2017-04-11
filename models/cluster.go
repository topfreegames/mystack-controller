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
	"strconv"
	"strings"
)

//Cluster represents a k8s cluster for a user
type Cluster struct {
	Namespace   string
	Username    string
	Deployments []*Deployment
	Services    []*Service
	Setup       *Job
}

//NewCluster returns a new cluster ready to start
func NewCluster(db DB, username, clusterName string) (*Cluster, error) {
	namespace := usernameToNamespace(username)

	clusterConfig, err := LoadClusterConfig(db, clusterName)
	if err != nil {
		return nil, err
	}

	portMap := make(map[string][]*PortMap)
	k8sAppDeployments, err := buildDeployments(clusterConfig.Apps, username, portMap)
	k8sSvcDeployments, err := buildDeployments(clusterConfig.Services, username, portMap)
	if err != nil {
		return nil, errors.NewYamlError("parse yaml error", err)
	}
	if err != nil {
		return nil, errors.NewYamlError("parse yaml error", err)
	}
	k8sServices := buildServices(k8sAppDeployments, k8sSvcDeployments, username, portMap)

	k8sJob := NewJob(username, clusterConfig.Setup["image"])

	cluster := &Cluster{
		Username:    username,
		Namespace:   namespace,
		Deployments: append(k8sAppDeployments, k8sSvcDeployments...),
		Services:    k8sServices,
		Setup:       k8sJob,
	}

	return cluster, nil
}

func getPorts(name string, ports []string, portMap map[string][]*PortMap) ([]int, error) {
	var err error
	containerPorts := make([]int, len(ports))
	portMap[name] = make([]*PortMap, len(ports))
	for i, port := range ports {
		splitedPorts := strings.Split(port, ":")
		if containerPorts[i], err = strconv.Atoi(splitedPorts[0]); err != nil {
			return nil, err
		}

		portMap[name][i] = &PortMap{
			Port:       containerPorts[i],
			TargetPort: containerPorts[i],
		}

		if len(splitedPorts) == 2 {
			if containerPorts[i], err = strconv.Atoi(splitedPorts[1]); err != nil {
				return nil, err
			}
			portMap[name][i].TargetPort = containerPorts[i]
		}
	}

	return containerPorts, nil
}

func buildDeployments(
	types map[string]*ClusterAppConfig,
	username string,
	portMap map[string][]*PortMap,
) ([]*Deployment, error) {
	deployments := make([]*Deployment, len(types))

	i := 0
	for name, config := range types {
		ports, err := getPorts(name, config.Ports, portMap)
		if err != nil {
			return nil, err
		}
		deployments[i] = NewDeployment(name, username, config.Image, ports, config.Environment)
		i = i + 1
	}

	return deployments, nil
}

func buildServices(
	apps []*Deployment,
	svcs []*Deployment,
	username string,
	portMap map[string][]*PortMap,
) []*Service {
	services := make([]*Service, len(apps)+len(svcs))
	i := 0
	for _, svc := range svcs {
		services[i] = NewService(svc.Name, username, portMap[svc.Name])
		i = i + 1
	}
	for _, app := range apps {
		services[i] = NewService(app.Name, username, portMap[app.Name])
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

	_, err = c.Setup.Run(clientset)
	if err != nil {
		return err
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
