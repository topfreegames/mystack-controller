package models_test

import (
	. "github.com/topfreegames/mystack-controller/models"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("MockCredentials", func() {
	Describe("GetID", func() {
		It("should return correct id", func() {
			credentials := &MockCredentials{ID: "my-id"}

			id := credentials.GetID()

			Expect(id).To(Equal("my-id"))
		})
	})

	Describe("GetSecret", func() {
		It("should return correct secret", func() {
			credentials := &MockCredentials{Key: "my-secret"}

			secret := credentials.GetSecret()

			Expect(secret).To(Equal("my-secret"))
		})
	})
})
