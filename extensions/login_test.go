// mystack-controller api
// +build integration
// https://github.com/topfreegames/mystack-controller
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2017 Top Free Games <backend@tfgco.com>

package extensions_test

import (
	. "github.com/topfreegames/mystack-controller/extensions"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Login", func() {
	Describe("Generate Login URL", func() {
		It("should return an valid URL", func() {
			state := "random"
			url, err := GenerateLoginURL(state)
			Expect(err).NotTo(HaveOccurred())
			Expect(url).To(ContainSubstring(state))
		})
	})
})
