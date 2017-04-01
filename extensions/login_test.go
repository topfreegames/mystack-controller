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
