// mystack-controller api
// +build unit
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
		deployments = []*Deployment{
			NewDeployment("app1", username, "image1", port, nil),
			NewDeployment("app2", username, "image2", port, nil),
			NewDeployment("app3", username, "image3", port, nil),
		}
	)

	BeforeEach(func() {
		clientset = fake.NewSimpleClientset()
	})

	Describe("Create", func() {
		It("should create cluster", func() {
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

		It("should return error if creating same cluster twice", func() {
			cluster := NewCluster(username, deployments)
			err := cluster.Create(clientset)
			Expect(err).NotTo(HaveOccurred())

			err = cluster.Create(clientset)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Delete", func() {
		It("should delete cluster", func() {
			cluster := NewCluster(username, deployments)
			err := cluster.Create(clientset)
			Expect(err).NotTo(HaveOccurred())

			err = cluster.Delete(clientset)
			Expect(err).NotTo(HaveOccurred())

			Expect(NamespaceExists(clientset, namespace)).To(BeFalse())

			deploys, err := clientset.ExtensionsV1beta1().Deployments(namespace).List(listOptions)
			Expect(err).NotTo(HaveOccurred())
			Expect(deploys.Items).To(BeEmpty())

			services, err := clientset.CoreV1().Services(namespace).List(listOptions)
			Expect(err).NotTo(HaveOccurred())
			Expect(services.Items).To(BeEmpty())
		})

		It("should delete only specified cluster", func() {
			deployments1 := []*Deployment{
				NewDeployment("app1", "user1", "image1", port, nil),
				NewDeployment("app2", "user1", "image2", port, nil),
				NewDeployment("app3", "user1", "image3", port, nil),
			}
			cluster1 := NewCluster("user1", deployments1)
			err := cluster1.Create(clientset)
			Expect(err).NotTo(HaveOccurred())

			deployments2 := []*Deployment{
				NewDeployment("app1", "user2", "image1", port, nil),
				NewDeployment("app2", "user2", "image2", port, nil),
				NewDeployment("app3", "user2", "image3", port, nil),
			}
			cluster2 := NewCluster("user2", deployments2)
			err = cluster2.Create(clientset)
			Expect(err).NotTo(HaveOccurred())

			err = cluster1.Delete(clientset)
			Expect(err).NotTo(HaveOccurred())

			Expect(NamespaceExists(clientset, "mystack-user1")).To(BeFalse())
			Expect(NamespaceExists(clientset, "mystack-user2")).To(BeTrue())

			deploys, err := clientset.ExtensionsV1beta1().Deployments("mystack-user1").List(listOptions)
			Expect(err).NotTo(HaveOccurred())
			Expect(deploys.Items).To(BeEmpty())

			services, err := clientset.CoreV1().Services("mystack-user1").List(listOptions)
			Expect(err).NotTo(HaveOccurred())
			Expect(services.Items).To(BeEmpty())

			deploys, err = clientset.ExtensionsV1beta1().Deployments("mystack-user2").List(listOptions)
			Expect(err).NotTo(HaveOccurred())
			Expect(deploys.Items).To(HaveLen(3))

			services, err = clientset.CoreV1().Services("mystack-user2").List(listOptions)
			Expect(err).NotTo(HaveOccurred())
			Expect(services.Items).To(HaveLen(3))
		})
	})
})
