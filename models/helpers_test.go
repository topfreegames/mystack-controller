// mystack-controller api
// https://github.com/topfreegames/mystack/mystack-controller
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2017 Top Free Games <backend@tfgco.com>

package models_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/topfreegames/mystack/mystack-controller/models"
)

var _ = Describe("Helpers", func() {
	Describe("GetDB", func() {
		It("should return a DB connection successfully", func() {
			db, err := models.GetDB(
				"localhost", "mystack_controller_test", 8585, "disable",
				"mystack_controller_test", "",
				10, 10, 100,
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(db).NotTo(BeNil())
		})

		It("should succeed if connectionTimeoutMS value is <= 0", func() {
			db, err := models.GetDB(
				"localhost", "mystack_controller_test", 8585, "disable",
				"mystack_controller_test", "",
				10, 10, -50,
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(db).NotTo(BeNil())
		})

		It("should panic if invalid connection information is passed", func() {
			Expect(func() {
				models.GetDB(
					"localhost", "mystack_controller_testtt", 8585, "disable",
					"mystack_controller_test", "password",
					10, 10, 100,
				)
			}).To(Panic())
		})
	})
})
