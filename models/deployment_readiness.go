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

const (
	readinessProbeInitialDelay = 5 * time.Second
	readinessProbePeriod       = 5 * time.Second
	readinessProbeTimeout      = 3 * time.Minute
)

//WaitForCompletion waits until job has completed its task
func (dr *DeploymentReadiness) WaitForCompletion(clientset kubernetes.Interface, d interface{}) error {
	deployments, ok := d.([]*Deployment)
	if !ok {
		return errors.NewGenericError("wait for deployment completion error", fmt.Errorf("interface{} is not of type []*Deployment"))
	}

	for _, deploy := range deployments {
		start := time.Now()
		k8sDeploy, err := clientset.ExtensionsV1beta1().Deployments(deploy.Namespace).Get(deploy.Name)
		if err != nil {
			return err
		}
		desiredNumberReplicas := *k8sDeploy.Spec.Replicas

		for desiredNumberReplicas > k8sDeploy.Status.AvailableReplicas {
			k8sDeploy, err = clientset.ExtensionsV1beta1().Deployments(deploy.Namespace).Get(deploy.Name)
			if err != nil {
				return err
			}
			time.Sleep(readinessProbePeriod)
			if time.Now().Sub(start) > readinessProbeTimeout {
				return errors.NewKubernetesError(
					"wait for deployment completion error",
					fmt.Errorf("timeout"),
				)
			}
		}
	}

	return nil
}
