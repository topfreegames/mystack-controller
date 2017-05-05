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
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/fields"
	"k8s.io/client-go/pkg/labels"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("PersistentVolumeClaim", func() {
	var (
		clientset   *fake.Clientset
		name        = "pvc"
		username    = "user"
		namespace   = "mystack-user"
		capacity    = "2Gi"
		pvc         = NewPVC(name, username, capacity)
		labelMap    = labels.Set{"mystack/routable": "true"}
		listOptions = v1.ListOptions{
			LabelSelector: labelMap.AsSelector().String(),
			FieldSelector: fields.Everything().String(),
		}
	)

	BeforeEach(func() {
		clientset = fake.NewSimpleClientset()
	})

	Describe("Start", func() {
		It("should start pvc correctly", func() {
			err := CreateNamespace(clientset, username)
			Expect(err).NotTo(HaveOccurred())

			k8sPVC, err := pvc.Start(clientset)
			Expect(err).NotTo(HaveOccurred())
			Expect(k8sPVC.ObjectMeta.Name).To(Equal(name))
			Expect(k8sPVC.ObjectMeta.Namespace).To(Equal(namespace))
			Expect(k8sPVC.ObjectMeta.Annotations["volume.alpha.kubernetes.io/storage-class"]).To(Equal("gp2"))
			Expect(k8sPVC.ObjectMeta.Labels["app"]).To(Equal(name))
			Expect(k8sPVC.ObjectMeta.Labels["mystack/routable"]).To(Equal("true"))
			Expect(k8sPVC.Spec.AccessModes).To(Equal([]v1.PersistentVolumeAccessMode{"ReadWriteOnce"}))
		})

		It("should return error if namespace doesn't exist", func() {
			_, err := pvc.Start(clientset)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Start", func() {
		It("should delete pvc correctly", func() {
			err := CreateNamespace(clientset, username)
			Expect(err).NotTo(HaveOccurred())

			_, err = pvc.Start(clientset)
			Expect(err).NotTo(HaveOccurred())

			err = pvc.Delete(clientset)
			Expect(err).NotTo(HaveOccurred())

			volumes, err := clientset.CoreV1().PersistentVolumeClaims(namespace).List(listOptions)
			Expect(err).NotTo(HaveOccurred())
			Expect(volumes.Items).To(BeEmpty())
		})

		It("should return error if pvc was not started", func() {
			err := pvc.Delete(clientset)
			Expect(err).To(HaveOccurred())
		})
	})
})
