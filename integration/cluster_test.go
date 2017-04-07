// mystack-controller api
// +build integration
// https://github.com/topfreegames/mystack-controller
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2017 Top Free Games <backend@tfgco.com>

package integration_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/topfreegames/mystack-controller/models"
)

var _ = Describe("Cluster", func() {
	const (
		yaml1 = `
apps:
  test1:
    image: app1
    port: 5000
  test2:
    image: app2
    port: 5000
  test3:
    image: app3
    port: 5000
`
		clusterName = "myCustomApps"
		username    = "user"
		port        = 5000
		namespace   = "mystack-user"
	)
	var err error

	mockCluster := func(username string) *Cluster {
		return &Cluster{
			Username:  username,
			Namespace: namespace,
			Deployments: []*Deployment{
				NewDeployment("test1", username, "app1", port, nil),
				NewDeployment("test2", username, "app2", port, nil),
				NewDeployment("test3", username, "app3", port, nil),
			},
			Services: []*Service{
				NewService("test1", username, 80, port),
				NewService("test2", username, 80, port),
				NewService("test3", username, 80, port),
			},
		}
	}

	Describe("NewCluster", func() {
		It("should construct a new cluster", func() {
			err = WriteClusterConfig(db, clusterName, yaml1)
			Expect(err).NotTo(HaveOccurred())

			cluster, err := NewCluster(db, username, clusterName)
			Expect(err).NotTo(HaveOccurred())

			mockedCluster := mockCluster(username)
			Expect(cluster.Deployments).To(ConsistOf(mockedCluster.Deployments))
			Expect(cluster.Services).To(ConsistOf(mockedCluster.Services))
		})

		It("should return error if cluster name not found", func() {
			cluster, err := NewCluster(db, username, clusterName)
			Expect(err).To(HaveOccurred())
			Expect(cluster).To(BeNil())
		})
	})
})
