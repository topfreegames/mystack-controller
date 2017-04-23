// mystack-controller api
// +build unit
// https://github.com/topfreegames/mystack-controller
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2017 Top Free Games <backend@tfgco.com>

package api_test

import (
	"fmt"
	"github.com/topfreegames/mystack-controller/api"
	mTest "github.com/topfreegames/mystack-controller/testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"database/sql"
	"github.com/Sirupsen/logrus"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/viper"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"testing"
)

var app *api.App
var db *sql.DB
var mock sqlmock.Sqlmock
var config *viper.Viper
var clientset kubernetes.Interface

func TestApi(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Api Suite")
}

var _ = BeforeSuite(func() {
	var err error
	l := logrus.New()
	l.Level = logrus.FatalLevel

	clientset = fake.NewSimpleClientset()

	config, err = mTest.GetDefaultConfig()
	Expect(err).NotTo(HaveOccurred())

	app = &api.App{
		Config:              config,
		Address:             fmt.Sprintf("%s:%d", "0.0.0.0", 8889),
		Debug:               false,
		Logger:              l,
		EmailDomain:         config.GetStringSlice("oauth.acceptedDomains"),
		Clientset:           clientset,
		DeploymentReadiness: &mTest.MockReadiness{},
		JobReadiness:        &mTest.MockReadiness{},
	}
	app.ConfigureServer()
})

var _ = BeforeEach(func() {
	var err error
	clientset = fake.NewSimpleClientset()
	app.Clientset = clientset
	db, mock, err = sqlmock.New()
	Expect(err).NotTo(HaveOccurred())
	app.DB = sqlx.NewDb(db, "postgres")
})

var _ = AfterEach(func() {
	defer db.Close()
	err := mock.ExpectationsWereMet()
	Expect(err).NotTo(HaveOccurred())
})
