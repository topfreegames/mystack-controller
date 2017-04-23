// mystack-controller api
// +build integration
// https://github.com/topfreegames/mystack-controller
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2017 Top Free Games <backend@tfgco.com>

package models_test

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
      - 5001
`
		clusterName = "myCustomApps"
	)

	var (
		err      error
		services = map[string]*ClusterAppConfig{
			"postgres": &ClusterAppConfig{
				Image: "postgres:1.0",
				Ports: []string{"8585:5432"},
			},
			"redis": &ClusterAppConfig{
				Image: "redis:1.0",
				Ports: []string{"6379"},
			},
		}
		apps = map[string]*ClusterAppConfig{
			"app1": &ClusterAppConfig{
				Image: "app1",
				Ports: []string{"5000:5001"},
				Environment: []*EnvVar{
					&EnvVar{
						Name:  "DATABASE_URL",
						Value: "postgres://derp:1234@example.com",
					},
				},
			},
			"app2": &ClusterAppConfig{
				Image: "app2",
				Ports: []string{"5001"},
			},
		}
	)

	Describe("WriteClusterConfig", func() {
		It("should write cluster config", func() {
			err = WriteClusterConfig(db, clusterName, yaml1)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should return error when writing cluster config with same name", func() {
			err = WriteClusterConfig(db, clusterName, yaml1)
			Expect(err).NotTo(HaveOccurred())

			err = WriteClusterConfig(db, clusterName, yaml1)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("LoadClusterConfig", func() {
		It("should load cluster config", func() {
			err = WriteClusterConfig(db, clusterName, yaml1)
			Expect(err).NotTo(HaveOccurred())

			clusterConfig, err := LoadClusterConfig(db, clusterName)
			Expect(err).NotTo(HaveOccurred())
			Expect(clusterConfig.Services).To(BeEquivalentTo(services))
			Expect(clusterConfig.Apps).To(BeEquivalentTo(apps))
		})

		It("should return error if clusterName doesn't exist on DB", func() {
			clusterConfig, err := LoadClusterConfig(db, clusterName)
			Expect(clusterConfig).To(BeNil())
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("RemoveClusterConfig", func() {
		It("should delete existing cluster config", func() {
			err = WriteClusterConfig(db, clusterName, yaml1)
			Expect(err).NotTo(HaveOccurred())

			err = RemoveClusterConfig(db, clusterName)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should not return error when deleting non existing cluster config", func() {
			err = RemoveClusterConfig(db, clusterName)
			Expect(err).To(HaveOccurred())
		})
	})
})
