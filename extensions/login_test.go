// mystack-controller api
// +build unit
// https://github.com/topfreegames/mystack-controller
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2017 Top Free Games <backend@tfgco.com>

package extensions_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/topfreegames/mystack-controller/extensions"

	"github.com/topfreegames/mystack-controller/models"
)

var _ = Describe("Login", func() {
	Describe("Generate Login URL", func() {
		It("should return error for empty ID", func() {
			state := "random"
			credentials := &models.MockCredentials{
				ID:  "",
				Key: "invalid",
			}
			_, err := GenerateLoginURL(state, credentials)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Define your app's OAuth2 Client ID on MYSTACK_GOOGLE_CLIENT_ID environment variable and run again"))
		})

		It("should return error for empty Key", func() {
			state := "random"
			credentials := &models.MockCredentials{
				ID:  "invalid",
				Key: "",
			}
			_, err := GenerateLoginURL(state, credentials)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Define your app's OAuth2 Client Secret on MYSTACK_GOOGLE_CLIENT_SECRET environment variable and run again"))
		})
	})
})
