// mystack-controller api
// +build unit
// https://github.com/topfreegames/mystack-controller
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2017 Top Free Games <backend@tfgco.com>

package models_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/viper"

	"database/sql"
	"testing"

	"github.com/jmoiron/sqlx"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"

	mTest "github.com/topfreegames/mystack-controller/testing"
)

var (
	db     *sql.DB
	sqlxDB *sqlx.DB
	mock   sqlmock.Sqlmock
	err    error
	config *viper.Viper
)

func TestModels(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Models Suite")
}

var _ = BeforeSuite(func() {
	config, err = mTest.GetDefaultConfig()
	Expect(err).NotTo(HaveOccurred())
})

var _ = BeforeEach(func() {
	db, mock, err = sqlmock.New()
	Expect(err).NotTo(HaveOccurred())
	sqlxDB = sqlx.NewDb(db, "postgres")
})

var _ = AfterEach(func() {
	defer db.Close()
	err = mock.ExpectationsWereMet()
	Expect(err).NotTo(HaveOccurred())
})
