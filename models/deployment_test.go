// mystack-controller api
// https://github.com/topfreegames/mystack-controller
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2017 Top Free Games <backend@tfgco.com>

package models_test

import (
	. "github.com/onsi/ginkgo"
	//	. "github.com/onsi/gomega"
	//	"github.com/topfreegames/mystack-controller/models"
	//
	//	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	//	"k8s.io/client-go/kubernetes/fake"
)

var _ = Describe("Deployment", func() {
	Describe("Create deployment", func() {
		It("should create and delete a deployment", func() {
			//deployment := &models.Deployment{
			//	Name:      "test",
			//	Namespace: "test",
			//	Image:     "hello-world",
			//}

			//clientset := fake.NewSimpleClientset()

			//_, err := deployment.Deploy(clientset)
			//Expect(err).NotTo(HaveOccurred())
			//pods, err := clientset.CoreV1().Pods("").List(metav1.ListOptions{})
			//Expect(err).NotTo(HaveOccurred())
			//Expect(pods.Items).To(HaveLen(1))
		})
	})
})
