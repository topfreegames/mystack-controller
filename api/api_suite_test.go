package api_test

import (
	"io"

	"github.com/Sirupsen/logrus"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/viper"
	runner "gopkg.in/mgutz/dat.v2/sqlx-runner"

	"testing"

	"github.com/topfreegames/mystack/mystack-controller/api"
	oTesting "github.com/topfreegames/mystack/mystack-controller/testing"
)

var app *api.App
var db runner.Connection
var closer io.Closer
var config *viper.Viper

func TestApi(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Offers API - API Suite")
}

var _ = BeforeSuite(func() {
	l := logrus.New()
	l.Level = logrus.FatalLevel

	var err error
	db, err = oTesting.GetTestDB()
	Expect(err).NotTo(HaveOccurred())

	config, err = oTesting.GetDefaultConfig()
	Expect(err).NotTo(HaveOccurred())
	app, err = api.NewApp("0.0.0.0", 8889, config, false, l)
	Expect(err).NotTo(HaveOccurred())
})

var _ = BeforeEach(func() {
	tx, err := db.Begin()
	Expect(err).NotTo(HaveOccurred())
	app.DB = tx
})

var _ = AfterEach(func() {
	err := app.DB.(*runner.Tx).Rollback()
	Expect(err).NotTo(HaveOccurred())
	app.DB = db
})

var _ = AfterSuite(func() {
	if db != nil {
		err := db.(*runner.DB).DB.Close()
		Expect(err).NotTo(HaveOccurred())
		db = nil
	}

	if closer != nil {
		closer.Close()
		closer = nil
	}
})
