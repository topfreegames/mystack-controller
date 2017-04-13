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

var _ = Describe("Job", func() {
	var (
		username  = "user"
		image     = "setup-img"
		clientset *fake.Clientset
		namespace = "mystack-user"
	)

	BeforeEach(func() {
		clientset = fake.NewSimpleClientset()
	})

	Describe("Run", func() {
		It("should run job", func() {
			err := CreateNamespace(clientset, username)
			Expect(err).NotTo(HaveOccurred())

			job := NewJob(username, image, []*EnvVar{
				&EnvVar{Name: "DATABASE_URL", Value: "postgresql://derp"},
			})

			k8sJob, err := job.Run(clientset)
			Expect(err).NotTo(HaveOccurred())
			Expect(k8sJob).NotTo(BeNil())
			Expect(k8sJob.ObjectMeta.Namespace).To(Equal(namespace))
			Expect(k8sJob.ObjectMeta.Name).To(Equal("setup"))
			Expect(k8sJob.ObjectMeta.Labels["mystack/owner"]).To(Equal(username))
			Expect(k8sJob.ObjectMeta.Labels["app"]).To(Equal("setup"))
			Expect(k8sJob.ObjectMeta.Labels["heritage"]).To(Equal("mystack"))
			Expect(k8sJob.Spec.Template.Spec.Containers[0].Env[0].Name).To(Equal("DATABASE_URL"))
			Expect(k8sJob.Spec.Template.Spec.Containers[0].Env[0].Value).To(Equal("postgresql://derp"))

			k8sJob, err = clientset.BatchV1().Jobs(namespace).Get(job.Name)
			Expect(err).NotTo(HaveOccurred())
			Expect(k8sJob).NotTo(BeNil())
			Expect(k8sJob.ObjectMeta.Namespace).To(Equal(namespace))
			Expect(k8sJob.ObjectMeta.Name).To(Equal("setup"))
			Expect(k8sJob.ObjectMeta.Labels["mystack/owner"]).To(Equal(username))
			Expect(k8sJob.ObjectMeta.Labels["app"]).To(Equal("setup"))
			Expect(k8sJob.ObjectMeta.Labels["heritage"]).To(Equal("mystack"))
			Expect(k8sJob.Spec.Template.Spec.Containers[0].Env[0].Name).To(Equal("DATABASE_URL"))
			Expect(k8sJob.Spec.Template.Spec.Containers[0].Env[0].Value).To(Equal("postgresql://derp"))
			Expect(k8sJob.Spec.Template.Spec.Containers[0].Image).To(Equal("setup-img"))
		})

		It("should not run job without namespace", func() {
			job := NewJob(username, image, nil)
			_, err = job.Run(clientset)
			Expect(err).To(HaveOccurred())
		})
	})
})
