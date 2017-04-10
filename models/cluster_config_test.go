// mystack-controller api
// +build unit
// https://github.com/topfreegames/mystack-controller
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2017 Top Free Games <backend@tfgco.com>

package models_test

import (
	"fmt"
	. "github.com/topfreegames/mystack-controller/models"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

const (
	yaml1 = `
services:
  postgres:
    image: postgres:1.0
    ports:
      - 8585:5432
  redis:
    image: redis:1.0
    ports:
      - 6379
apps:
  app1:
    image: app1
    ports:
      - 5000:5001
    env:
      - name: DATABASE_URL
        value: postgres://derp:1234@example.com
  app2:
    image: app2
    ports:
      - 5000:5001
`
)

var _ = Describe("ClusterConfig", func() {
	var (
		err         error
		clusterName = "MyCustomApps"
	)

	Describe("ParseYaml", func() {
		It("should build correct struct form yaml", func() {
			apps, services, err := ParseYaml(yaml1)
			Expect(err).NotTo(HaveOccurred())

			Expect(services["postgres"].Image).To(Equal("postgres:1.0"))
			Expect(services["postgres"].Ports).To(BeEquivalentTo([]string{"8585:5432"}))
			Expect(services["redis"].Image).To(Equal("redis:1.0"))
			Expect(services["redis"].Ports).To(BeEquivalentTo([]string{"6379"}))

			Expect(apps["app1"].Image).To(Equal("app1"))
			Expect(apps["app1"].Ports).To(BeEquivalentTo([]string{"5000:5001"}))
			Expect(apps["app1"].Environment).To(BeEquivalentTo([]*EnvVar{
				&EnvVar{
					Name:  "DATABASE_URL",
					Value: "postgres://derp:1234@example.com",
				},
			}))

			Expect(apps["app2"].Image).To(Equal("app2"))
			Expect(apps["app2"].Ports).To(BeEquivalentTo([]string{"5000:5001"}))
			Expect(apps["app2"].Environment).To(BeNil())
		})

		It("should return error with invalid yaml", func() {
			invalidYaml := `
services {
  app1 {
    image: app
}
			`
			_, _, err := ParseYaml(invalidYaml)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("yaml: line 3: mapping values are not allowed in this context"))
			Expect(fmt.Sprintf("%T", err)).To(Equal("*errors.YamlError"))
		})
	})

	Describe("WriteClusterConfig", func() {
		It("should write cluster config", func() {
			mock.
				ExpectExec("INSERT INTO clusters").
				WithArgs(clusterName, yaml1).
				WillReturnResult(sqlmock.NewResult(1, 1))

			err = WriteClusterConfig(sqlxDB, clusterName, yaml1)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should return error when writing invalid yaml", func() {
			invalidYaml := `
services {
  app1 {
    image: app
}
			`
			err := WriteClusterConfig(sqlxDB, clusterName, invalidYaml)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("yaml: line 3: mapping values are not allowed in this context"))
			Expect(fmt.Sprintf("%T", err)).To(Equal("*errors.YamlError"))
		})

		It("should return error when writing cluster with same name", func() {
			mock.
				ExpectExec("INSERT INTO clusters").
				WithArgs(clusterName, yaml1).
				WillReturnResult(sqlmock.NewResult(1, 1))
			mock.
				ExpectExec("INSERT INTO clusters").
				WithArgs(clusterName, yaml1).
				WillReturnError(fmt.Errorf(`pq: duplicate key value violates unique constraint "clusters_name_key"`))

			err = WriteClusterConfig(sqlxDB, clusterName, yaml1)
			Expect(err).NotTo(HaveOccurred())

			err = WriteClusterConfig(sqlxDB, clusterName, yaml1)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(`pq: duplicate key value violates unique constraint "clusters_name_key"`))
			Expect(fmt.Sprintf("%T", err)).To(Equal("*errors.DatabaseError"))
		})

		It("should return error when clusterName is empty", func() {
			err := WriteClusterConfig(sqlxDB, "", yaml1)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("invalid empty cluster name"))
			Expect(fmt.Sprintf("%T", err)).To(Equal("*errors.GenericError"))
		})

		It("should return error with invalid yaml", func() {
			invalidYaml := `
services {
  app1 {
    image: app
}
			`
			err := WriteClusterConfig(sqlxDB, clusterName, invalidYaml)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("yaml: line 3: mapping values are not allowed in this context"))
			Expect(fmt.Sprintf("%T", err)).To(Equal("*errors.YamlError"))
		})

		It("should return error with empty yaml", func() {
			err := WriteClusterConfig(sqlxDB, clusterName, "")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("invalid empty config"))
			Expect(fmt.Sprintf("%T", err)).To(Equal("*errors.YamlError"))
		})
	})

	Describe("LoadClusterConfig", func() {
		It("should load cluster config", func() {
			mock.
				ExpectQuery("^SELECT yaml FROM clusters WHERE name = (.+)$").
				WithArgs(clusterName).
				WillReturnRows(sqlmock.NewRows([]string{"yaml"}).AddRow(yaml1))

			apps, services, err := LoadClusterConfig(sqlxDB, clusterName)
			Expect(err).NotTo(HaveOccurred())
			Expect(services["postgres"].Image).To(Equal("postgres:1.0"))
			Expect(services["postgres"].Ports).To(BeEquivalentTo([]string{"8585:5432"}))
			Expect(services["redis"].Image).To(Equal("redis:1.0"))
			Expect(services["redis"].Ports).To(BeEquivalentTo([]string{"6379"}))

			Expect(apps["app1"].Image).To(Equal("app1"))
			Expect(apps["app1"].Ports).To(BeEquivalentTo([]string{"5000:5001"}))
			Expect(apps["app1"].Environment).To(BeEquivalentTo([]*EnvVar{
				&EnvVar{
					Name:  "DATABASE_URL",
					Value: "postgres://derp:1234@example.com",
				},
			}))

			Expect(apps["app2"].Image).To(Equal("app2"))
			Expect(apps["app2"].Ports).To(BeEquivalentTo([]string{"5000:5001"}))
			Expect(apps["app2"].Environment).To(BeNil())
		})

		It("should return error when loading non existing clusterName", func() {
			mock.
				ExpectQuery("^SELECT yaml FROM clusters WHERE name = (.+)$").
				WithArgs(clusterName).
				WillReturnRows(sqlmock.NewRows([]string{"yaml"}))

			apps, services, err := LoadClusterConfig(sqlxDB, clusterName)
			Expect(apps).To(BeNil())
			Expect(services).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("sql: no rows in result set"))
			Expect(fmt.Sprintf("%T", err)).To(Equal("*errors.DatabaseError"))
		})

		It("should return error when loading empty clusterName", func() {
			apps, services, err := LoadClusterConfig(sqlxDB, "")
			Expect(apps).To(BeNil())
			Expect(services).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("invalid empty cluster name"))
			Expect(fmt.Sprintf("%T", err)).To(Equal("*errors.GenericError"))
		})

		It("should return error if database has invalid yaml", func() {
			invalidYaml := `
services {
  app1 {
    image: app
}
			`
			mock.
				ExpectQuery("^SELECT yaml FROM clusters WHERE name = (.+)$").
				WithArgs(clusterName).
				WillReturnRows(sqlmock.NewRows([]string{"yaml"}).AddRow(invalidYaml))

			apps, services, err := LoadClusterConfig(sqlxDB, clusterName)
			Expect(apps).To(BeNil())
			Expect(services).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("yaml: line 3: mapping values are not allowed in this context"))
			Expect(fmt.Sprintf("%T", err)).To(Equal("*errors.YamlError"))
		})

		It("should return error if database has empty yaml", func() {
			mock.
				ExpectQuery("^SELECT yaml FROM clusters WHERE name = (.+)$").
				WithArgs(clusterName).
				WillReturnRows(sqlmock.NewRows([]string{"yaml"}).AddRow(""))

			apps, services, err := LoadClusterConfig(sqlxDB, clusterName)
			Expect(apps).To(BeNil())
			Expect(services).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("invalid empty config"))
			Expect(fmt.Sprintf("%T", err)).To(Equal("*errors.YamlError"))
		})
	})

	Describe("RemoveClusterConfig", func() {
		It("should delete existing cluster config", func() {
			mock.
				ExpectExec("^DELETE FROM clusters WHERE name=(.+)$").
				WithArgs(clusterName).
				WillReturnResult(sqlmock.NewResult(1, 1))

			err = RemoveClusterConfig(sqlxDB, clusterName)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should return error when deleting non existing cluster config", func() {
			mock.
				ExpectExec("^DELETE FROM clusters WHERE name=(.+)$").
				WithArgs(clusterName).
				WillReturnResult(sqlmock.NewResult(0, 0))

			err = RemoveClusterConfig(sqlxDB, clusterName)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("sql: no rows in result set"))
			Expect(fmt.Sprintf("%T", err)).To(Equal("*errors.DatabaseError"))
		})

		It("should return error when cluster name is empty", func() {
			err = RemoveClusterConfig(sqlxDB, "")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("invalid empty cluster name"))
			Expect(fmt.Sprintf("%T", err)).To(Equal("*errors.GenericError"))
		})
	})
})
