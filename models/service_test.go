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

var _ = Describe("Service", func() {
	var (
		clientset   *fake.Clientset
		name        = "test"
		namespace   = "mystack-user"
		username    = "user"
		port        = 80
		targetPort  = 5000
		labelMap    = labels.Set{"mystack/routable": "true"}
		listOptions = v1.ListOptions{
			LabelSelector: labelMap.AsSelector().String(),
			FieldSelector: fields.Everything().String(),
		}
	)

	BeforeEach(func() {
		clientset = fake.NewSimpleClientset()
	})

	Describe("Expose", func() {
		It("should expose a new Service", func() {
			svc := NewService(name, username, port, targetPort)
			Expect(svc.Namespace).To(Equal(namespace))

			svcv1, err := svc.Expose(clientset)
			Expect(err).NotTo(HaveOccurred())
			Expect(svcv1.GetNamespace()).To(Equal(namespace))

			svcs, err := clientset.CoreV1().Services(namespace).List(listOptions)
			Expect(err).NotTo(HaveOccurred())
			Expect(svcs.Items).To(HaveLen(1))
		})

		It("should return error when creating same service twice", func() {
			svc := NewService(name, username, port, targetPort)
			Expect(svc.Namespace).To(Equal(namespace))

			_, err := svc.Expose(clientset)
			Expect(err).NotTo(HaveOccurred())

			_, err = svc.Expose(clientset)
			Expect(err).To(HaveOccurred())
		})
	})
})
