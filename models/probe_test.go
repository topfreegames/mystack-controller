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
)

var _ = Describe("Probe", func() {
	It("should return status error if postgres is not running", func() {
		probe := Probe{
			Command: []string{"pg_isready", "-p", "61242", "-U", "postgres", "-h", "localhost"},
		}

		Expect(probe.Status()).NotTo(BeZero())
	})
})
