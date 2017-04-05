// mystack-controller api
// +build unit
// https://github.com/topfreegames/mystack-controller
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2017 Top Free Games <backend@tfgco.com>

package models_test

import (
	. "github.com/topfreegames/mystack-controller/models"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/kubernetes/fake"
)

var _ = Describe("Namespace", func() {
	var (
		clientset *fake.Clientset
		username  = "user"
		namespace = "mystack-user"
	)

	BeforeEach(func() {
		clientset = fake.NewSimpleClientset()
	})

	Describe("CreateNamespace", func() {
		It("should create a namespace", func() {
			err := CreateNamespace(clientset, username)
			Expect(err).NotTo(HaveOccurred())

			ns, err := ListNamespaces(clientset)
			Expect(err).NotTo(HaveOccurred())
			Expect(ns.Items).To(HaveLen(1))
		})

		It("should return error when creating existing namespace", func() {
			err := CreateNamespace(clientset, username)
			Expect(err).NotTo(HaveOccurred())

			err = CreateNamespace(clientset, username)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("NamespaceExists", func() {
		It("should return false if namespace does not exist", func() {
			exist := NamespaceExists(clientset, namespace)
			Expect(exist).To(BeFalse())
		})

		It("should return true after creating namespace", func() {
			err := CreateNamespace(clientset, username)
			Expect(err).NotTo(HaveOccurred())

			exist := NamespaceExists(clientset, namespace)
			Expect(exist).To(BeTrue())
		})
	})

	Describe("DeleteNamespace", func() {
		It("should return error when deleting non-exiting namespace", func() {
			err := DeleteNamespace(clientset, username)
			Expect(err).To(HaveOccurred())
		})

		It("should delete namespace if exists", func() {
			err := CreateNamespace(clientset, username)
			Expect(err).NotTo(HaveOccurred())

			err = DeleteNamespace(clientset, username)
			Expect(err).NotTo(HaveOccurred())

			exist := NamespaceExists(clientset, namespace)
			Expect(exist).To(BeFalse())
		})
	})
})
