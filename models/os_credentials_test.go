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

	"os"
)

var _ = Describe("OsCredentials", func() {
	Describe("GetID", func() {
		It("should return correct id", func() {
			os.Setenv(ClientIDEnvVar, "my-id")
			credentials := &OSCredentials{}

			id := credentials.GetID()

			Expect(id).To(Equal("my-id"))

			os.Unsetenv(ClientIDEnvVar)
		})
	})

	Describe("GetSecret", func() {
		It("should return correct secret", func() {
			os.Setenv(ClientSecretEnvVar, "my-secret")
			credentials := &OSCredentials{}

			secret := credentials.GetSecret()

			Expect(secret).To(Equal("my-secret"))

			os.Unsetenv(ClientSecretEnvVar)
		})
	})
})
