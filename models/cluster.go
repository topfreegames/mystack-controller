// mystack-controller api
// https://github.com/topfreegames/mystack-controller
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2017 Top Free Games <backend@tfgco.com>

package models

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/topfreegames/mystack-controller/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/fields"
	"k8s.io/client-go/pkg/labels"
)

//Cluster represents a k8s cluster for a user
type Cluster struct {
	Namespace              string
	Username               string
	AppDeployments         []*Deployment
	SvcDeployments         []*Deployment
	K8sServices            map[*Deployment]*Service
	Job, PostJob           *Job
	PersistentVolumeClaims []*PersistentVolumeClaim
	DeploymentReadiness    Readiness
	JobReadiness           Readiness
}

//NewCluster returns a new cluster ready to start
func NewCluster(
	db DB,
	username, clusterName string,
	deploymentReadiness, jobReadiness Readiness,
	config *viper.Viper,
) (*Cluster, error) {
	namespace := usernameToNamespace(username)

	clusterConfig, err := LoadClusterConfig(db, clusterName)
	if err != nil {
		return nil, err
	}

	portMap := make(map[string][]*PortMap)
	environment := []*EnvVar{}
	k8sAppDeployments, environment, err := buildDeployments(clusterConfig.Apps, username, portMap, environment, config)
	if err != nil {
		return nil, errors.NewYamlError("parse yaml error", err)
	}
	k8sSvcDeployments, environment, err := buildDeployments(clusterConfig.Services, username, portMap, environment, config)
	if err != nil {
		return nil, errors.NewYamlError("parse yaml error", err)
	}

	clusterServices := make(map[*Deployment]*Service)
	clusterServices = buildServices(k8sAppDeployments, clusterConfig.Apps, username, portMap, false, clusterServices)
	clusterServices = buildServices(k8sSvcDeployments, clusterConfig.Services, username, portMap, true, clusterServices)

	k8sJob := NewJob("setup", username, clusterConfig.Setup, environment)
	k8sPostJob := NewJob("postSetup", username, clusterConfig.PostSetup, environment)

	k8sPersistentVolumeClaims := buildPersistentVolumeClaims(clusterConfig.Volumes, username)

	cluster := &Cluster{
		Username:               username,
		Namespace:              namespace,
		AppDeployments:         k8sAppDeployments,
		SvcDeployments:         k8sSvcDeployments,
		K8sServices:            clusterServices,
		Job:                    k8sJob,
		PostJob:                k8sPostJob,
		DeploymentReadiness:    deploymentReadiness,
		JobReadiness:           jobReadiness,
		PersistentVolumeClaims: k8sPersistentVolumeClaims,
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

func hasSatisfiedDependencies(links []string, createdDeployments map[string]*Deployment) bool {
	for _, link := range links {
		_, ok := createdDeployments[link]
		if !ok {
			return false
		}
	}

	return true
}

func buildDeployments(
	types map[string]*ClusterAppConfig,
	username string,
	portMap map[string][]*PortMap,
	environment []*EnvVar,
	appConfig *viper.Viper,
) ([]*Deployment, []*EnvVar, error) {
	deployments := make([]*Deployment, len(types))
	createdDeployments := make(map[string]*Deployment)
	notCreatedDeployments := make(map[string]bool)

	for name := range types {
		notCreatedDeployments[name] = true
	}

	i := 0
	for len(notCreatedDeployments) > 0 {
		for name := range notCreatedDeployments {
			config := types[name]
			if hasSatisfiedDependencies(config.Links, createdDeployments) {
				ports, err := getPorts(name, config.Ports, portMap)
				if err != nil {
					return nil, environment, err
				}
				deployment := NewDeployment(name, username, config.Image, ports, config.Environment, config.ReadinessProbe, config.VolumeMount, config.Command, config.Resources, appConfig)
				for _, link := range config.Links {
					deployment.Links = append(deployment.Links, createdDeployments[link])
				}

				createdDeployments[name] = deployment

				deployments[i] = deployment
				environment = append(environment, config.Environment...)
				i = i + 1
				delete(notCreatedDeployments, name)
			}
		}
	}

	return deployments, environment, nil
}

func buildServices(
	deploys []*Deployment,
	appConfigs map[string]*ClusterAppConfig,
	username string,
	portMap map[string][]*PortMap,
	isMystackSvc bool,
	services map[*Deployment]*Service,
) map[*Deployment]*Service {
	for _, deploy := range deploys {
		services[deploy] = NewService(deploy.Name, username, portMap[deploy.Name], isMystackSvc, appConfigs[deploy.Name].IsSocket)
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

func buildPersistentVolumeClaims(persistentVolumeClaims []*PersistentVolumeClaim, username string) []*PersistentVolumeClaim {
	for _, pvc := range persistentVolumeClaims {
		pvc.Namespace = usernameToNamespace(username)
	}

	return persistentVolumeClaims
}

func log(logger logrus.FieldLogger, msg string) {
	if logger != nil {
		logger.Info(msg)
	}
}

//Create creates namespace, deployments and services
func (c *Cluster) Create(logger logrus.FieldLogger, clientset kubernetes.Interface) error {
	if NamespaceExists(clientset, c.Namespace) {
		return errors.NewKubernetesError(
			"create cluster error",
			fmt.Errorf("namespace for user '%s' already exists", c.Username),
		)
	}

	log(logger, "creating namespace")
	err := CreateNamespace(clientset, c.Username)
	if err != nil {
		return rollback(clientset, c.Username, err)
	}
	log(logger, "done creating namespace")

	log(logger, "creating svc volume")
	for _, pvc := range c.PersistentVolumeClaims {
		_, err = pvc.Start(clientset)
		if err != nil {
			logger.WithError(err).Error("failed to create PVC")
			return rollback(clientset, c.Username, err)
		}
	}
	log(logger, "done creating svc volume")

	err = c.startDeploymentsAndItsServicesWithLinks(logger, clientset, c.SvcDeployments)
	if err != nil {
		return rollback(clientset, c.Username, err)
	}

	err = c.runJob(logger, clientset, c.Job)
	if err != nil {
		return rollback(clientset, c.Username, err)
	}

	err = c.startDeploymentsAndItsServicesWithLinks(logger, clientset, c.AppDeployments)
	if err != nil {
		return rollback(clientset, c.Username, err)
	}

	log(logger, "creating post-setup job")
	_, err = c.PostJob.Run(clientset)
	if err != nil {
		logger.WithError(err).Error("failed to run post job")
		return rollback(clientset, c.Username, err)
	}

	return nil
}

func (c *Cluster) runJob(
	logger logrus.FieldLogger,
	clientset kubernetes.Interface,
	job *Job,
) error {
	log(logger, "creating job")

	_, err := job.Run(clientset)
	if err != nil {
		logger.WithError(err).Error("failed to run job")
		return rollback(clientset, c.Username, err)
	}

	log(logger, "waiting for job completion")
	err = c.JobReadiness.WaitForCompletion(clientset, job)
	if err != nil {
		logger.WithError(err).Error("failed to run job")
		return err
	}

	log(logger, "finished job")

	return nil
}

func (c *Cluster) startDeploymentsAndItsServicesWithLinks(
	logger logrus.FieldLogger,
	clientset kubernetes.Interface,
	deployments []*Deployment,
) error {
	log(logger, "Creating linked services")
	deploymentsNotReady := make(map[*Deployment]bool)
	for _, deployment := range deployments {
		deploymentsNotReady[deployment] = true
	}

	for len(deploymentsNotReady) > 0 {
		runnableDeployments := []*Deployment{}
		for deployment := range deploymentsNotReady {
			if canRun(deployment, deploymentsNotReady) {
				runnableDeployments = append(runnableDeployments, deployment)
			}
		}

		err := c.startDeploymentsAndItsServices(logger, clientset, runnableDeployments, c.K8sServices)
		if err != nil {
			logger.WithError(err).Error("failed to create deployment and service")
			return err
		}

		for _, deployment := range runnableDeployments {
			delete(deploymentsNotReady, deployment)
		}
	}

	return nil
}

//A deployment can start if all its dependencies (links) are already running
func canRun(deployment *Deployment, deploymentsNotReady map[*Deployment]bool) bool {
	for _, dependencie := range deployment.Links {
		if _, contains := deploymentsNotReady[dependencie]; contains {
			return false
		}
	}
	return true
}

func (c *Cluster) startDeploymentsAndItsServices(
	logger logrus.FieldLogger,
	clientset kubernetes.Interface,
	deployments []*Deployment,
	services map[*Deployment]*Service,
) error {
	log(logger, "creating deployments")
	for _, deployment := range deployments {
		_, err := deployment.Deploy(clientset)
		if err != nil {
			logger.WithError(err).Errorf("failed to create deployment: %s", deployment)
			return rollback(clientset, c.Username, err)
		}
	}

	log(logger, "waiting deployment completion")
	err := c.DeploymentReadiness.WaitForCompletion(clientset, deployments)
	if err != nil {
		return rollback(clientset, c.Username, err)
	}
	log(logger, "done creating deployments")

	log(logger, "creating services")
	for _, deployment := range deployments {
		service := services[deployment]
		_, err = service.Expose(clientset)
		if err != nil {
			logger.WithError(err).Errorf("failed to create service: %s", deployment)
			return rollback(clientset, c.Username, err)
		}
	}
	log(logger, "done creating services")

	return nil
}

//Delete deletes namespace and all deployments and services
func (c *Cluster) Delete(clientset kubernetes.Interface) error {
	if !NamespaceExists(clientset, c.Namespace) {
		return errors.NewKubernetesError(
			"delete cluster error",
			fmt.Errorf("namespace for user '%s' not found", c.Username),
		)
	}

	for _, service := range c.K8sServices {
		service.Delete(clientset)
	}

	for _, deployment := range c.AppDeployments {
		deployment.Delete(clientset)
	}

	for _, deployment := range c.SvcDeployments {
		deployment.Delete(clientset)
	}

	for _, pvc := range c.PersistentVolumeClaims {
		pvc.Delete(clientset)
	}

	err := DeleteNamespace(clientset, c.Username)
	if err != nil {
		return err
	}

	timeout := time.Duration(2) * time.Minute
	start := time.Now()
	for NamespaceExists(clientset, c.Namespace) {
		if time.Now().Sub(start) > timeout {
			return errors.NewKubernetesError(
				"delete cluster error",
				fmt.Errorf("delete cluster reached timeout: %s", c.Username),
			)
		}
		time.Sleep(5)
	}

	return nil
}

//Apps returns a list of cluster apps
func (c *Cluster) Apps(
	config *viper.Viper,
	clientset kubernetes.Interface,
	k8sDomain string,
) (map[string][]string, error) {
	if !NamespaceExists(clientset, c.Namespace) {
		return nil, errors.NewKubernetesError(
			"get apps error",
			fmt.Errorf("namespace for user '%s' not found", c.Username),
		)
	}

	labelMap := labels.Set{
		"mystack/routable": "true",
		"mystack/service":  "false",
	}
	listOptions := v1.ListOptions{
		LabelSelector: labelMap.AsSelector().String(),
		FieldSelector: fields.Everything().String(),
	}

	services, err := clientset.CoreV1().Services(c.Namespace).List(listOptions)
	if err != nil {
		return nil, errors.NewKubernetesError(
			"get apps error",
			fmt.Errorf("couldn't retrieve services"),
		)
	}

	//Return array for further improvements (e.g. custom domains)
	domains := make(map[string][]string)

	for _, service := range services.Items {
		if service.GetLabels()["mystack/socket"] == "true" {
			url := config.GetString("kubernetes.service-domain-suffix")
			port := service.GetLabels()["mystack/socketPorts"]
			domains[service.Name] = []string{fmt.Sprintf("controller.%s:%s", url, port)}
			continue
		}
		domains[service.Name] = []string{fmt.Sprintf("%s.%s.%s", service.Name, service.Namespace, k8sDomain)}
	}

	return domains, nil
}

//Services returns a list of cluster services
func (c *Cluster) Services(clientset kubernetes.Interface) ([]string, error) {
	if !NamespaceExists(clientset, c.Namespace) {
		return nil, errors.NewKubernetesError(
			"get apps error",
			fmt.Errorf("namespace for user '%s' not found", c.Username),
		)
	}

	labelMap := labels.Set{
		"mystack/routable": "true",
		"mystack/service":  "true",
	}
	listOptions := v1.ListOptions{
		LabelSelector: labelMap.AsSelector().String(),
		FieldSelector: fields.Everything().String(),
	}

	services, err := clientset.CoreV1().Services(c.Namespace).List(listOptions)
	if err != nil {
		return nil, errors.NewKubernetesError(
			"get apps error",
			fmt.Errorf("couldn't retrieve services"),
		)
	}

	serviceNames := []string{}
	for _, service := range services.Items {
		serviceNames = append(serviceNames, service.Name)
	}

	return serviceNames, nil
}
