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
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/fields"
	"k8s.io/client-go/pkg/labels"
)

var _ = Describe("Job", func() {
	var (
		username    = "user"
		image       = "setup-img"
		clientset   *fake.Clientset
		namespace   = "mystack-user"
		labelMap    = labels.Set{"mystack/routable": "true"}
		listOptions = v1.ListOptions{
			LabelSelector: labelMap.AsSelector().String(),
			FieldSelector: fields.Everything().String(),
		}
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

			jobs, err := clientset.BatchV1().Jobs(namespace).List(listOptions)
			Expect(err).NotTo(HaveOccurred())
			Expect(jobs.Items).To(HaveLen(1))
		})

		It("should not run job without namespace", func() {
			job := NewJob(username, image, nil)
			_, err = job.Run(clientset)
			Expect(err).To(HaveOccurred())
		})
	})
})
