// mystack-controller api
// +build unit
// https://github.com/topfreegames/mystack-controller
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2017 Top Free Games <backend@tfgco.com>

package models_test

import (
	"fmt"

	. "github.com/topfreegames/mystack-controller/models"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"k8s.io/client-go/kubernetes/fake"
)

var _ = Describe("DeploymentReadiness", func() {

	var (
		clientset *fake.Clientset
	)

	BeforeEach(func() {
		clientset = fake.NewSimpleClientset()
	})

	It("should reach non default timeout", func() {
		CreateNamespace(clientset, "user")
		probe := &Probe{
			PeriodSeconds:  1,
			TimeoutSeconds: 1,
		}
		deploy := NewDeployment("app", "user", "image", nil, nil, probe)
		_, err := deploy.Deploy(clientset)
		Expect(err).NotTo(HaveOccurred())

		deployments := []*Deployment{deploy}

		readiness := &DeploymentReadiness{}
		err = readiness.WaitForCompletion(clientset, deployments)
		Expect(fmt.Sprintf("%T", err)).To(Equal("*errors.KubernetesError"))
		Expect(err.Error()).To(Equal("wait for deployment completion error due to timeout"))
	})

	It("should return error for non existing deployment", func() {
		probe := &Probe{
			PeriodSeconds:  1,
			TimeoutSeconds: 1,
		}
		deployments := []*Deployment{
			NewDeployment("app", "user", "image", nil, nil, probe),
		}

		readiness := &DeploymentReadiness{}
		err := readiness.WaitForCompletion(clientset, deployments)
		Expect(err.Error()).To(Equal("Deployment.extensions \"app\" not found"))
	})
})
