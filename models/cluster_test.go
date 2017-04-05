// mystack-controller api
// https://github.com/topfreegames/mystack-controller
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2017 Top Free Games <backend@tfgco.com>

package models_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/topfreegames/mystack-controller/models"

	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/fields"
	"k8s.io/client-go/pkg/labels"
)

var _ = Describe("Cluster", func() {
	var (
		clientset   *fake.Clientset
		username    = "user"
		namespace   = "mystack-user"
		port        = 5000
		labelMap    = labels.Set{"mystack/routable": "true"}
		listOptions = v1.ListOptions{
			LabelSelector: labelMap.AsSelector().String(),
			FieldSelector: fields.Everything().String(),
		}
	)

	BeforeEach(func() {
		clientset = fake.NewSimpleClientset()
	})

	Describe("Create", func() {
		It("should create cluster", func() {
			deployments := []*Deployment{
				NewDeployment("app1", username, "image1", port),
				NewDeployment("app2", username, "image2", port),
				NewDeployment("app3", username, "image3", port),
			}

			cluster := NewCluster(username, deployments)
			err := cluster.Create(clientset)
			Expect(err).NotTo(HaveOccurred())

			deploys, err := clientset.ExtensionsV1beta1().Deployments(namespace).List(listOptions)
			Expect(err).NotTo(HaveOccurred())
			Expect(deploys.Items).To(HaveLen(3))

			services, err := clientset.CoreV1().Services(namespace).List(listOptions)
			Expect(err).NotTo(HaveOccurred())
			Expect(services.Items).To(HaveLen(3))
		})
	})

	Describe("Delete", func() {

	})
})
