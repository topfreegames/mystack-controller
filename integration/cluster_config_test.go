// mystack-controller api
// +build integration
// https://github.com/topfreegames/mystack-controller
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2017 Top Free Games <backend@tfgco.com>

package integration_test

import (
	. "github.com/topfreegames/mystack-controller/models"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ClusterConfig", func() {
	const (
		yaml1 = `
services:
  postgres:
    image: postgres:1.0
  redis:
    image: redis:1.0
apps:
  app1:
    image: app1
    port: 5000
    env:
      - name: DATABASE_URL
        value: postgres://derp:1234@example.com
  app2:
    image: app2
    port: 5001
`
		clusterName = "myCustomApps"
	)

	var (
		err      error
		services = map[string]*ClusterAppConfig{
			"postgres": &ClusterAppConfig{Image: "postgres:1.0"},
			"redis":    &ClusterAppConfig{Image: "redis:1.0"},
		}
		apps = map[string]*ClusterAppConfig{
			"app1": &ClusterAppConfig{
				Image: "app1",
				Port:  5000,
				Environment: []*EnvVar{
					&EnvVar{
						Name:  "DATABASE_URL",
						Value: "postgres://derp:1234@example.com",
					},
				},
			},
			"app2": &ClusterAppConfig{
				Image: "app2",
				Port:  5001,
			},
		}
	)

	BeforeEach(func() {
		db, err = conn.Beginx()
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		err = db.Rollback()
		Expect(err).NotTo(HaveOccurred())
		db = nil
	})

	Describe("WriteClusterConfig", func() {
		It("should write cluster config", func() {
			err = WriteClusterConfig(db, clusterName, yaml1)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("LoadClusterConfig", func() {
		It("should load cluster config", func() {
			err = WriteClusterConfig(db, clusterName, yaml1)
			Expect(err).NotTo(HaveOccurred())

			returnApps, returnServices, err := LoadClusterConfig(db, clusterName)
			Expect(err).NotTo(HaveOccurred())
			Expect(returnServices).To(BeEquivalentTo(services))
			Expect(returnApps).To(BeEquivalentTo(apps))
		})

		It("should return error if clusterName doesn' exist on DB", func() {
			apps, services, err := LoadClusterConfig(db, clusterName)
			Expect(apps).To(BeNil())
			Expect(services).To(BeNil())
			Expect(err).To(HaveOccurred())
		})
	})
})
