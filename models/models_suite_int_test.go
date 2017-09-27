// mystack-controller api
// +build integration
// https://github.com/topfreegames/mystack-controller
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2017 Top Free Games <backend@tfgco.com>

package models_test

import (
	"github.com/Sirupsen/logrus"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/viper"

	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	oTesting "github.com/topfreegames/mystack-controller/testing"
)

var conn *sqlx.DB
var db *sqlx.Tx
var err error
var config *viper.Viper

func TestModels(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Models Integration Suite")
}

var _ = BeforeSuite(func() {
	l := logrus.New()
	l.Level = logrus.FatalLevel

	conn, err = oTesting.GetTestDB()
	Expect(err).NotTo(HaveOccurred())

	config, err = oTesting.GetDefaultConfig()
	Expect(err).NotTo(HaveOccurred())
})

var _ = BeforeEach(func() {
	db, err = conn.Beginx()
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterEach(func() {
	err = db.Rollback()
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	if conn != nil {
		err := conn.Close()
		Expect(err).NotTo(HaveOccurred())
		db = nil
	}
})
