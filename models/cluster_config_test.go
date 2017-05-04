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
setup:
  image: setup-img
  periodSeconds: 10
  timeoutSeconds: 180
services:
  postgres:
    image: postgres:1.0
    ports:
      - 8585:5432
    readinessProbe:
      command:
        - pg_isready
        - -h
        - localhost
        - -p
        - 5432
        - -U
        - postgres
      periodSeconds: 10
      startDeploymentTimeoutSeconds: 180
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
        value: postgresql://derp:1234@example.com
      - name: USERNAME
        value: derp
  app2:
    image: app2
    ports:
      - 5000:5001
`
	yamlWithoutSetup = `
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
        value: postgresql://derp:1234@example.com
      - name: USERNAME
        value: derp
  app2:
    image: app2
    ports:
      - 5000:5001
`
	yamlWithVolume = `
volumes:
  - name: postgres-volume
    storage: 1Gi
services:
  postgres:
    image: postgres:1.0
    ports:
      - 8585:5432
    env:
      - name: PGDATA
        value: /var/lib/postgresql/data/pgdata
    volumeMount:
      name: postgres-volume
      mountPath: /var/lib/postgresql/data
apps:
  app1:
    image: app1
    ports:
      - 5000:5001
    env:
      - name: DATABASE_URL
        value: postgresql://derp:1234@example.com
      - name: USERNAME
        value: derp
`
	yamlWith2Volumes = `
volumes:
  - name: postgres-volume
    storage: 1Gi
  - name: redis-volume
    storage: 1Gi
services:
  postgres:
    image: postgres:1.0
    ports:
      - 8585:5432
    env:
      - name: PGDATA
        value: /var/lib/postgresql/data/pgdata
    volumeMount:
      name: postgres-volume
      mountPath: /var/lib/postgresql/data
  redis:
    image: redis:1.0
    ports:
      - 6333
    volumeMount:
      name: redis-volume
      mountPath: /data
apps:
  app1:
    image: app1
    ports:
      - 5000
`
)

var _ = Describe("ClusterConfig", func() {
	var (
		err         error
		clusterName = "MyCustomApps"
	)

	Describe("ParseYaml", func() {
		It("should build correct struct from yaml", func() {
			clusterConfig, err := ParseYaml(yaml1)
			Expect(err).NotTo(HaveOccurred())

			Expect(clusterConfig.Services["postgres"].Image).To(Equal("postgres:1.0"))
			Expect(clusterConfig.Services["postgres"].Ports).To(BeEquivalentTo([]string{"8585:5432"}))
			Expect(clusterConfig.Services["postgres"].ReadinessProbe).To(BeEquivalentTo(&Probe{
				Command:        []string{"pg_isready", "-h", "localhost", "-p", "5432", "-U", "postgres"},
				TimeoutSeconds: 180,
				PeriodSeconds:  10,
			}))

			Expect(clusterConfig.Services["redis"].Image).To(Equal("redis:1.0"))
			Expect(clusterConfig.Services["redis"].Ports).To(BeEquivalentTo([]string{"6379"}))

			Expect(clusterConfig.Apps["app1"].Image).To(Equal("app1"))
			Expect(clusterConfig.Apps["app1"].Ports).To(BeEquivalentTo([]string{"5000:5001"}))
			Expect(clusterConfig.Apps["app1"].Environment).To(BeEquivalentTo([]*EnvVar{
				&EnvVar{
					Name:  "DATABASE_URL",
					Value: "postgresql://derp:1234@example.com",
				},
				&EnvVar{
					Name:  "USERNAME",
					Value: "derp",
				},
			}))

			Expect(clusterConfig.Apps["app2"].Image).To(Equal("app2"))
			Expect(clusterConfig.Apps["app2"].Ports).To(BeEquivalentTo([]string{"5000:5001"}))
			Expect(clusterConfig.Apps["app2"].Environment).To(BeNil())

			Expect(clusterConfig.Setup.Image).To(Equal("setup-img"))
			Expect(clusterConfig.Setup.TimeoutSeconds).To(Equal(180))
			Expect(clusterConfig.Setup.PeriodSeconds).To(Equal(10))
		})

		It("should build correct struct from yaml with volume", func() {
			clusterConfig, err := ParseYaml(yamlWithVolume)
			Expect(err).NotTo(HaveOccurred())

			Expect(clusterConfig.Services["postgres"].Image).To(Equal("postgres:1.0"))
			Expect(clusterConfig.Services["postgres"].Ports).To(BeEquivalentTo([]string{"8585:5432"}))
			Expect(clusterConfig.Services["postgres"].ReadinessProbe).To(BeNil())
			Expect(clusterConfig.Services["postgres"].Environment).To(BeEquivalentTo([]*EnvVar{
				&EnvVar{
					Name:  "PGDATA",
					Value: "/var/lib/postgresql/data/pgdata",
				},
			}))
		})

		It("should build correct struct from yaml with two volumes", func() {
			clusterConfig, err := ParseYaml(yamlWith2Volumes)
			Expect(err).NotTo(HaveOccurred())

			Expect(clusterConfig.Services["postgres"].Image).To(Equal("postgres:1.0"))
			Expect(clusterConfig.Services["postgres"].Ports).To(BeEquivalentTo([]string{"8585:5432"}))
			Expect(clusterConfig.Services["postgres"].ReadinessProbe).To(BeNil())
			Expect(clusterConfig.Services["postgres"].Environment).To(BeEquivalentTo([]*EnvVar{
				&EnvVar{
					Name:  "PGDATA",
					Value: "/var/lib/postgresql/data/pgdata",
				},
			}))

			volumeMount := clusterConfig.Services["postgres"].VolumeMount
			Expect(volumeMount.Name).To(Equal("postgres-volume"))
			Expect(volumeMount.MountPath).To(Equal("/var/lib/postgresql/data"))

			volumeMount = clusterConfig.Services["redis"].VolumeMount
			Expect(volumeMount.Name).To(Equal("redis-volume"))
			Expect(volumeMount.MountPath).To(Equal("/data"))

			pvcs := clusterConfig.Volumes
			Expect(pvcs).To(HaveLen(2))
			Expect(pvcs[0].Name).To(Equal("postgres-volume"))
			Expect(pvcs[0].Namespace).To(Equal(""))
			Expect(pvcs[0].Storage).To(Equal("1Gi"))

			Expect(pvcs[1].Name).To(Equal("redis-volume"))
			Expect(pvcs[1].Namespace).To(Equal(""))
			Expect(pvcs[1].Storage).To(Equal("1Gi"))
		})

		It("should return error with invalid syntax yaml", func() {
			invalidYaml := `
services {
  app1 {
    image: app
}
      `
			_, err := ParseYaml(invalidYaml)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("yaml: line 3: mapping values are not allowed in this context"))
			Expect(fmt.Sprintf("%T", err)).To(Equal("*errors.YamlError"))
		})
	})

	Describe("WriteClusterConfig", func() {
		It("should write cluster config", func() {
			mock.
				ExpectExec("^INSERT INTO clusters\\(name, yaml\\) VALUES\\((.+)\\)$").
				WithArgs(clusterName, yaml1).
				WillReturnResult(sqlmock.NewResult(1, 1))

			err = WriteClusterConfig(sqlxDB, clusterName, yaml1)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should write cluster config without setup", func() {
			mock.
				ExpectExec("^INSERT INTO clusters\\(name, yaml\\) VALUES\\((.+)\\)$").
				WithArgs(clusterName, yamlWithoutSetup).
				WillReturnResult(sqlmock.NewResult(1, 1))

			err = WriteClusterConfig(sqlxDB, clusterName, yamlWithoutSetup)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should write cluster config with volumes", func() {
			mock.
				ExpectExec("^INSERT INTO clusters\\(name, yaml\\) VALUES\\((.+)\\)$").
				WithArgs(clusterName, yamlWithVolume).
				WillReturnResult(sqlmock.NewResult(1, 1))

			err = WriteClusterConfig(sqlxDB, clusterName, yamlWithVolume)
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
				ExpectExec("^INSERT INTO clusters\\(name, yaml\\) VALUES\\((.+)\\)$").
				WithArgs(clusterName, yaml1).
				WillReturnResult(sqlmock.NewResult(1, 1))
			mock.
				ExpectExec("^INSERT INTO clusters\\(name, yaml\\) VALUES\\((.+)\\)$").
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

		It("should write object as string on env var", func() {
			validYaml := `
apps:
  app1:
    image: img
    env:
      - name: OBJ
        value: "{\"key\": \"value\"}"
      `
			mock.
				ExpectExec("^INSERT INTO clusters\\(name, yaml\\) VALUES\\((.+)\\)$").
				WithArgs(clusterName, validYaml).
				WillReturnResult(sqlmock.NewResult(1, 1))
			err := WriteClusterConfig(sqlxDB, clusterName, validYaml)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("LoadClusterConfig", func() {
		It("should load cluster config", func() {
			mock.
				ExpectQuery("^SELECT yaml FROM clusters WHERE name = (.+)$").
				WithArgs(clusterName).
				WillReturnRows(sqlmock.NewRows([]string{"yaml"}).AddRow(yaml1))

			clusterConfig, err := LoadClusterConfig(sqlxDB, clusterName)
			Expect(err).NotTo(HaveOccurred())
			Expect(clusterConfig.Services["postgres"].Image).To(Equal("postgres:1.0"))
			Expect(clusterConfig.Services["postgres"].Ports).To(BeEquivalentTo([]string{"8585:5432"}))
			Expect(clusterConfig.Services["postgres"].ReadinessProbe).To(BeEquivalentTo(&Probe{
				Command:        []string{"pg_isready", "-h", "localhost", "-p", "5432", "-U", "postgres"},
				TimeoutSeconds: 180,
				PeriodSeconds:  10,
			}))

			Expect(clusterConfig.Services["redis"].Image).To(Equal("redis:1.0"))
			Expect(clusterConfig.Services["redis"].Ports).To(BeEquivalentTo([]string{"6379"}))

			Expect(clusterConfig.Apps["app1"].Image).To(Equal("app1"))
			Expect(clusterConfig.Apps["app1"].Ports).To(BeEquivalentTo([]string{"5000:5001"}))
			Expect(clusterConfig.Apps["app1"].Environment).To(BeEquivalentTo([]*EnvVar{
				&EnvVar{
					Name:  "DATABASE_URL",
					Value: "postgresql://derp:1234@example.com",
				},
				&EnvVar{
					Name:  "USERNAME",
					Value: "derp",
				},
			}))

			Expect(clusterConfig.Apps["app2"].Image).To(Equal("app2"))
			Expect(clusterConfig.Apps["app2"].Ports).To(BeEquivalentTo([]string{"5000:5001"}))
			Expect(clusterConfig.Apps["app2"].Environment).To(BeNil())

			Expect(clusterConfig.Setup.Image).To(Equal("setup-img"))
			Expect(clusterConfig.Setup.TimeoutSeconds).To(Equal(180))
			Expect(clusterConfig.Setup.PeriodSeconds).To(Equal(10))
		})

		It("should load cluster config with volumes", func() {
			mock.
				ExpectQuery("^SELECT yaml FROM clusters WHERE name = (.+)$").
				WithArgs(clusterName).
				WillReturnRows(sqlmock.NewRows([]string{"yaml"}).AddRow(yamlWithVolume))

			clusterConfig, err := LoadClusterConfig(sqlxDB, clusterName)
			Expect(err).NotTo(HaveOccurred())
			Expect(clusterConfig.Services["postgres"].Image).To(Equal("postgres:1.0"))
			Expect(clusterConfig.Services["postgres"].Ports).To(BeEquivalentTo([]string{"8585:5432"}))
			Expect(clusterConfig.Services["postgres"].ReadinessProbe).To(BeNil())
			Expect(clusterConfig.Services["postgres"].Environment).To(BeEquivalentTo([]*EnvVar{
				&EnvVar{
					Name:  "PGDATA",
					Value: "/var/lib/postgresql/data/pgdata",
				},
			}))

			volumeMount := clusterConfig.Services["postgres"].VolumeMount
			Expect(volumeMount.Name).To(Equal("postgres-volume"))
			Expect(volumeMount.MountPath).To(Equal("/var/lib/postgresql/data"))

			pvcs := clusterConfig.Volumes
			Expect(pvcs).To(HaveLen(1))
			Expect(pvcs[0].Name).To(Equal("postgres-volume"))
			Expect(pvcs[0].Namespace).To(Equal(""))
			Expect(pvcs[0].Storage).To(Equal("1Gi"))
		})

		It("should return error when loading non existing clusterName", func() {
			mock.
				ExpectQuery("^SELECT yaml FROM clusters WHERE name = (.+)$").
				WithArgs(clusterName).
				WillReturnRows(sqlmock.NewRows([]string{"yaml"}))

			clusterConfig, err := LoadClusterConfig(sqlxDB, clusterName)
			Expect(clusterConfig).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("sql: no rows in result set"))
			Expect(fmt.Sprintf("%T", err)).To(Equal("*errors.DatabaseError"))
		})

		It("should return error when loading empty clusterName", func() {
			clusterConfig, err := LoadClusterConfig(sqlxDB, "")
			Expect(clusterConfig).To(BeNil())
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

			clusterConfig, err := LoadClusterConfig(sqlxDB, clusterName)
			Expect(clusterConfig).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("yaml: line 3: mapping values are not allowed in this context"))
			Expect(fmt.Sprintf("%T", err)).To(Equal("*errors.YamlError"))
		})

		It("should return error if database has empty yaml", func() {
			mock.
				ExpectQuery("^SELECT yaml FROM clusters WHERE name = (.+)$").
				WithArgs(clusterName).
				WillReturnRows(sqlmock.NewRows([]string{"yaml"}).AddRow(""))

			clusterConfig, err := LoadClusterConfig(sqlxDB, clusterName)
			Expect(clusterConfig).To(BeNil())
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

	Describe("ListClusterConfig", func() {
		It("should list cluster configs", func() {
			mock.
				ExpectQuery("^SELECT name FROM clusters$").
				WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("cluster1").AddRow("cluster2"))

			names, err := ListClusterConfig(sqlxDB)
			Expect(err).NotTo(HaveOccurred())
			Expect(names).To(ConsistOf("cluster1", "cluster2"))
		})

		It("should return error if list is empty", func() {
			mock.
				ExpectQuery("^SELECT name FROM clusters$").
				WillReturnRows(sqlmock.NewRows([]string{"name"}))

			names, err := ListClusterConfig(sqlxDB)
			Expect(err).NotTo(HaveOccurred())
			Expect(names).To(BeEmpty())
		})
	})

	Describe("ClusterConfigDetails", func() {
		It("should return cluster config yaml", func() {
			mock.
				ExpectQuery("^SELECT yaml FROM clusters WHERE name(.+)$").
				WithArgs(clusterName).
				WillReturnRows(sqlmock.NewRows([]string{"yaml"}).AddRow(yaml1))

			config, err := ClusterConfigDetails(sqlxDB, clusterName)
			Expect(err).NotTo(HaveOccurred())
			Expect(config).To(Equal(yaml1))
		})

		It("should return error if name doesn't exist", func() {
			mock.
				ExpectQuery("^SELECT yaml FROM clusters WHERE name(.+)$").
				WithArgs(clusterName).
				WillReturnError(fmt.Errorf(`pq: no rows in result set`))

			config, err := ClusterConfigDetails(sqlxDB, clusterName)
			Expect(err).To(HaveOccurred())
			Expect(config).To(BeEmpty())
		})
	})
})
