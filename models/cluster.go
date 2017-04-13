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
	Namespace           string
	Username            string
	AppDeployments      []*Deployment
	SvcDeployments      []*Deployment
	AppServices         []*Service
	SvcServices         []*Service
	Setup               *Job
	DeploymentReadiness Readiness
	JobReadiness        Readiness
}

//NewCluster returns a new cluster ready to start
func NewCluster(
	db DB,
	username, clusterName string,
	deploymentReadiness, jobReadiness Readiness,
) (*Cluster, error) {
	namespace := usernameToNamespace(username)

	clusterConfig, err := LoadClusterConfig(db, clusterName)
	if err != nil {
		return nil, err
	}

	portMap := make(map[string][]*PortMap)
	environment := []*EnvVar{}
	k8sAppDeployments, environment, err := buildDeployments(clusterConfig.Apps, username, portMap, environment)
	k8sSvcDeployments, environment, err := buildDeployments(clusterConfig.Services, username, portMap, environment)
	if err != nil {
		return nil, errors.NewYamlError("parse yaml error", err)
	}
	if err != nil {
		return nil, errors.NewYamlError("parse yaml error", err)
	}
	k8sAppServices := buildServices(k8sAppDeployments, username, portMap)
	k8sSvcServices := buildServices(k8sSvcDeployments, username, portMap)

	k8sJob := NewJob(username, clusterConfig.Setup["image"], environment)

	cluster := &Cluster{
		Username:            username,
		Namespace:           namespace,
		AppDeployments:      k8sAppDeployments,
		SvcDeployments:      k8sSvcDeployments,
		AppServices:         k8sAppServices,
		SvcServices:         k8sSvcServices,
		Setup:               k8sJob,
		DeploymentReadiness: deploymentReadiness,
		JobReadiness:        jobReadiness,
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
	environment []*EnvVar,
) ([]*Deployment, []*EnvVar, error) {
	deployments := make([]*Deployment, len(types))

	i := 0
	for name, config := range types {
		ports, err := getPorts(name, config.Ports, portMap)
		if err != nil {
			return nil, environment, err
		}
		deployments[i] = NewDeployment(name, username, config.Image, ports, config.Environment, config.ReadinessProbe)
		environment = append(environment, config.Environment...)
		i = i + 1
	}

	return deployments, environment, nil
}

func buildServices(
	deploys []*Deployment,
	username string,
	portMap map[string][]*PortMap,
) []*Service {
	services := make([]*Service, len(deploys))
	i := 0
	for _, deploy := range deploys {
		services[i] = NewService(deploy.Name, username, portMap[deploy.Name])
		i = i + 1
	}
	return services
}

func rollback(clientset kubernetes.Interface, username string, err error) error {
	nsErr := DeleteNamespace(clientset, username)
	if nsErr != nil {
		return errors.NewKubernetesError(
			"create cluster error",
			fmt.Errorf("error during cluster creation and could not rollback: %s", nsErr.Error()),
		)
	}
	return err
}

//Create creates namespace, deployments and services
func (c *Cluster) Create(clientset kubernetes.Interface) error {
	err := CreateNamespace(clientset, c.Username)
	if err != nil {
		return err
	}

	for _, deployment := range c.SvcDeployments {
		_, err := deployment.Deploy(clientset)
		if err != nil {
			return rollback(clientset, c.Username, err)
		}
	}

	err = c.DeploymentReadiness.WaitForCompletion(clientset, c.SvcDeployments)
	if err != nil {
		return rollback(clientset, c.Username, err)
	}

	for _, service := range c.SvcServices {
		_, err = service.Expose(clientset)
		if err != nil {
			return rollback(clientset, c.Username, err)
		}
	}

	_, err = c.Setup.Run(clientset)
	if err != nil {
		return rollback(clientset, c.Username, err)
	}
	err = c.JobReadiness.WaitForCompletion(clientset, c.Setup)
	if err != nil {
		return rollback(clientset, c.Username, err)
	}

	for _, deployment := range c.AppDeployments {
		_, err := deployment.Deploy(clientset)
		if err != nil {
			return rollback(clientset, c.Username, err)
		}
	}

	err = c.DeploymentReadiness.WaitForCompletion(clientset, c.AppDeployments)
	if err != nil {
		return rollback(clientset, c.Username, err)
	}

	for _, service := range c.AppServices {
		_, err = service.Expose(clientset)
		if err != nil {
			return rollback(clientset, c.Username, err)
		}
	}

	return nil
}

//Delete deletes namespace and all deployments and services
func (c *Cluster) Delete(clientset kubernetes.Interface) error {
	var err error
	for _, service := range c.AppServices {
		err = service.Delete(clientset)
		if err != nil {
			return err
		}
	}

	for _, deployment := range c.AppDeployments {
		err = deployment.Delete(clientset)
		if err != nil {
			return err
		}
	}

	for _, service := range c.SvcServices {
		err = service.Delete(clientset)
		if err != nil {
			return err
		}
	}

	for _, deployment := range c.SvcDeployments {
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
