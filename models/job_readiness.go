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

func getJobTimes(setup *Setup) (time.Duration, time.Duration) {
	const (
		defaultTimeout = time.Duration(120) * time.Second
		defaultPeriod  = time.Duration(5) * time.Second
	)

	timeout := defaultTimeout
	if setup.TimeoutSeconds != 0 {
		timeout = time.Duration(setup.TimeoutSeconds) * time.Second
	}

	period := defaultPeriod
	if setup.PeriodSeconds != 0 {
		period = time.Duration(setup.PeriodSeconds) * time.Second
	}

	return period, timeout
}

//WaitForCompletion waits until job has completed its task
func (jr *JobReadiness) WaitForCompletion(
	clientset kubernetes.Interface,
	j interface{},
) error {
	if j == nil {
		return nil
	}

	job, ok := j.(*Job)
	if !ok {
		return errors.NewGenericError("wait for job completion error", fmt.Errorf("interface{} is not of type *models.Job"))
	}

	if job == nil {
		return nil
	}

	k8sJob, err := clientset.BatchV1().Jobs(job.Namespace).Get(job.Name)
	if err != nil {
		return errors.NewKubernetesError("setup error", err)
	}

	period, timeout := getJobTimes(job.Setup)
	start := time.Now()

	for k8sJob.Status.Succeeded == 0 {
		if k8sJob.Status.Failed == 1 {
			return errors.NewKubernetesError("setup error", fmt.Errorf("failed to run stup job"))
		}

		time.Sleep(period)
		if time.Now().Sub(start) > timeout {
			return errors.NewKubernetesError("setup error", fmt.Errorf("failed to run stup job due to timeout"))
		}

		k8sJob, err = clientset.BatchV1().Jobs(job.Namespace).Get(job.Name)
		if err != nil {
			return errors.NewKubernetesError("setup error", err)
		}
	}

	return nil
}
