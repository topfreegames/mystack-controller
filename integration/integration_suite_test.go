// mystack-controller api
// +build integration
// https://github.com/topfreegames/mystack-controller
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2017 Top Free Games <backend@tfgco.com>

package integration_test

import (
	"io"

	"github.com/Sirupsen/logrus"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/viper"

	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/topfreegames/mystack-controller/api"
	oTesting "github.com/topfreegames/mystack-controller/testing"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

var clientset kubernetes.Interface
var app *api.App
var conn *sqlx.DB
var db *sqlx.Tx
var closer io.Closer
var config *viper.Viper

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Suite")
}

var _ = BeforeSuite(func() {
	l := logrus.New()
	l.Level = logrus.FatalLevel

	var err error
	conn, err = oTesting.GetTestDB()
	Expect(err).NotTo(HaveOccurred())

	clientset := fake.NewSimpleClientset()

	config, err = oTesting.GetDefaultConfig()
	Expect(err).NotTo(HaveOccurred())
	app, err = api.NewApp("0.0.0.0", 8889, config, false, l, clientset)
	Expect(err).NotTo(HaveOccurred())
})

var _ = BeforeEach(func() {
	var err error
	db, err = conn.Beginx()
	Expect(err).NotTo(HaveOccurred())
	app.DB = db
})

var _ = AfterEach(func() {
	err := db.Rollback()
	Expect(err).NotTo(HaveOccurred())
	db = nil
	app.DB = conn
})

var _ = AfterSuite(func() {
	if conn != nil {
		err := conn.Close()
		Expect(err).NotTo(HaveOccurred())
		db = nil
	}

	if closer != nil {
		closer.Close()
		closer = nil
	}
})
