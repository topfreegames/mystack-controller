// mystack-controller api
// +build integration
// https://github.com/topfreegames/mystack-controller
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2017 Top Free Games <backend@tfgco.com>

package integration_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/topfreegames/mystack-controller/models"
)

var _ = Describe("Probe", func() {
	It("should return status code 0 if postgres is running", func() {
		probe := Probe{
			Command: []string{"pg_isready", "-p", "8585", "-U", "mystack_controller_test", "-h", "localhost"},
		}

		Expect(probe.Status()).To(BeZero())
	})

	It("should return status code 1 if postgres is not running", func() {
		probe := Probe{
			Command: []string{"pg_isready", "-p", "63122", "-U", "mystack_controller_test", "-h", "localhost"},
		}

		Expect(probe.Status()).NotTo(BeZero())
	})
})
