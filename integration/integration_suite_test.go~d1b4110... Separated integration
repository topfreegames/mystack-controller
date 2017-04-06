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

	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	oTesting "github.com/topfreegames/mystack-controller/testing"
)

var conn *sqlx.DB
var db *sqlx.Tx

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Suite")
}

var _ = BeforeSuite(func() {
	var err error
	conn, err = oTesting.GetTestDB()
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	err := conn.Close()
	Expect(err).NotTo(HaveOccurred())
})
