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

//JobReadiness implements Readiness interface
type JobReadiness struct{}

//WaitForCompletion waits until job has completed its task
func (*JobReadiness) WaitForCompletion(clientset kubernetes.Interface, j interface{}) error {
	if j == nil {
		return nil
	}

	job := j.(*Job)
	if job == nil {
		return nil
	}

	k8sJob, err := clientset.BatchV1().Jobs(job.Namespace).Get(job.Name)
	if err != nil {
		return errors.NewKubernetesError("setup error", err)
	}

	start := time.Now()

	for k8sJob.Status.Succeeded == 0 {
		if k8sJob.Status.Failed == 1 {
			return errors.NewKubernetesError("setup error", fmt.Errorf("failed to run stup job"))
		}

		time.Sleep(readinessProbePeriod)
		if time.Now().Sub(start) > readinessProbeTimeout {
			return errors.NewKubernetesError("setup error", fmt.Errorf("failed to run stup job due to timeout"))
		}

		k8sJob, err = clientset.BatchV1().Jobs(job.Namespace).Get(job.Name)
		if err != nil {
			return errors.NewKubernetesError("setup error", err)
		}
	}

	return nil
}
