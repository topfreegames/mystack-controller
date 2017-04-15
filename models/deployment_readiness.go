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
	"time"
)

//DeploymentReadiness implements Readiness interface
type DeploymentReadiness struct{}

func getDeployTimes(probe *Probe) (time.Duration, time.Duration) {
	const (
		defaultPeriod  = time.Duration(5) * time.Second
		defaultTimeout = time.Duration(120) * time.Second
	)

	if probe == nil {
		return defaultPeriod, defaultTimeout
	}

	timeout := defaultTimeout
	if probe.TimeoutSeconds != 0 {
		timeout = time.Duration(probe.TimeoutSeconds) * time.Second
	}

	period := defaultPeriod
	if probe.PeriodSeconds != 0 {
		period = time.Duration(probe.PeriodSeconds) * time.Second
	}

	return period, timeout
}

//WaitForCompletion waits until job has completed its task
func (dr *DeploymentReadiness) WaitForCompletion(clientset kubernetes.Interface, d interface{}) error {
	deployments, ok := d.([]*Deployment)
	if !ok {
		return errors.NewGenericError("wait for deployment completion error", fmt.Errorf("interface{} is not of type []*models.Deployment"))
	}

	for _, deploy := range deployments {
		period, timeout := getDeployTimes(deploy.ReadinessProbe)
		k8sDeploy, err := clientset.ExtensionsV1beta1().Deployments(deploy.Namespace).Get(deploy.Name)
		if err != nil {
			return err
		}

		start := time.Now()
		desiredNumberReplicas := *k8sDeploy.Spec.Replicas

		for desiredNumberReplicas > k8sDeploy.Status.AvailableReplicas {
			time.Sleep(period)
			k8sDeploy, err = clientset.ExtensionsV1beta1().Deployments(deploy.Namespace).Get(deploy.Name)
			if err != nil {
				return err
			}
			if time.Now().Sub(start) > timeout {
				return errors.NewKubernetesError(
					"wait for deployment completion error",
					fmt.Errorf("wait for deployment completion error due to timeout"),
				)
			}
		}
	}

	return nil
}
