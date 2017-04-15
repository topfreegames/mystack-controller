package models_test

import (
	"fmt"

	. "github.com/topfreegames/mystack-controller/models"
	"k8s.io/client-go/kubernetes/fake"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("JobReadiness", func() {

	var (
		clientset *fake.Clientset
	)

	BeforeEach(func() {
		clientset = fake.NewSimpleClientset()
	})

	It("should reach non default timeout", func() {
		CreateNamespace(clientset, "user")
		setup := &Setup{
			Image:          "image",
			PeriodSeconds:  1,
			TimeoutSeconds: 2,
		}
		job := NewJob("user", setup, nil)
		_, err := job.Run(clientset)
		Expect(err).NotTo(HaveOccurred())

		readiness := &JobReadiness{}
		err = readiness.WaitForCompletion(clientset, job)
		Expect(fmt.Sprintf("%T", err)).To(Equal("*errors.KubernetesError"))
		Expect(err.Error()).To(Equal("failed to run stup job due to timeout"))
	})

	It("should return error for non existing deployment", func() {
		setup := &Setup{
			Image:          "image",
			PeriodSeconds:  1,
			TimeoutSeconds: 1,
		}
		job := NewJob("user", setup, nil)
		readiness := &JobReadiness{}
		err = readiness.WaitForCompletion(clientset, job)

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("Job.batch \"setup\" not found"))
	})

	It("should return error if interface{} is not job", func() {
		readiness := &JobReadiness{}
		err = readiness.WaitForCompletion(clientset, "not a job")
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("interface{} is not of type *models.Job"))
	})
})
