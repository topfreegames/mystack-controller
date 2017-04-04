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

var _ = Describe("Deployment", func() {
	Describe("Create deployment", func() {
		var (
			clientset   *fake.Clientset
			name        = "test"
			namespace   = "mystack-user"
			username    = "user"
			image       = "hello-world"
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

		It("should create a deployment", func() {
			deployment := NewDeployment(name, username, image, port)

			deploy, err := deployment.Deploy(clientset)
			Expect(err).NotTo(HaveOccurred())
			Expect(deploy).NotTo(BeNil())
			Expect(deploy.ObjectMeta.Namespace).To(Equal(namespace))

			deploys, err := clientset.ExtensionsV1beta1().Deployments(namespace).List(listOptions)
			Expect(err).NotTo(HaveOccurred())
			Expect(deploys.Items).To(HaveLen(1))
		})
	})
})
