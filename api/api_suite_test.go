// mystack-controller api
// https://github.com/topfreegames/mystack-controller
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2017 Top Free Games <backend@tfgco.com>

package api_test

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
var db *sqlx.DB
var closer io.Closer
var config *viper.Viper

func TestApi(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "MyStack Controller API - API Suite")
}

var _ = BeforeSuite(func() {
	l := logrus.New()
	l.Level = logrus.FatalLevel

	var err error
	db, err = oTesting.GetTestDB()
	Expect(err).NotTo(HaveOccurred())

	clientset := fake.NewSimpleClientset()

	config, err = oTesting.GetDefaultConfig()
	Expect(err).NotTo(HaveOccurred())
	app, err = api.NewApp("0.0.0.0", 8889, config, false, l, clientset)
	Expect(err).NotTo(HaveOccurred())
})

var _ = BeforeEach(func() {
	tx, err := db.Beginx()
	Expect(err).NotTo(HaveOccurred())
	app.DB = tx
})

var _ = AfterEach(func() {
	err := app.DB.(*sqlx.Tx).Rollback()
	Expect(err).NotTo(HaveOccurred())
	app.DB = db
})

var _ = AfterSuite(func() {
	if db != nil {
		err := db.Close()
		Expect(err).NotTo(HaveOccurred())
		db = nil
	}

	if closer != nil {
		closer.Close()
		closer = nil
	}
})
