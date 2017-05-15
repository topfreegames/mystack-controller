// mystack-controller api
// +build integration
// https://github.com/topfreegames/mystack-controller
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2017 Top Free Games <backend@tfgco.com>

package models_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/topfreegames/mystack-controller/models"

	mTest "github.com/topfreegames/mystack-controller/testing"
)

var _ = Describe("Cluster", func() {
	const (
		yaml1 = `
services:
  test0:
    image: svc1
    ports: 
      - 5000:5001
apps:
  test1:
    image: app1
    ports: 
      - 5000:5001
  test2:
    image: app2
    ports: 
      - 5000:5001
  test3:
    image: app3
    ports: 
      - 5000:5001
`
		clusterName = "myCustomApps"
		username    = "user"
		namespace   = "mystack-user"
	)
	var (
		err     error
		ports   = []int{5001}
		portMap = []*PortMap{
			&PortMap{Port: 5000, TargetPort: 5001},
		}
	)

	mockCluster := func(username string) *Cluster {
		svcDeployment1 := NewDeployment("test0", username, "svc1", ports, nil, nil, nil)
		appDeployment1 := NewDeployment("test1", username, "app1", ports, nil, nil, nil)
		appDeployment2 := NewDeployment("test2", username, "app2", ports, nil, nil, nil)
		appDeployment3 := NewDeployment("test3", username, "app3", ports, nil, nil, nil)

		return &Cluster{
			Username:       username,
			Namespace:      namespace,
			AppDeployments: []*Deployment{appDeployment1, appDeployment2, appDeployment3},
			SvcDeployments: []*Deployment{svcDeployment1},
			K8sServices: map[*Deployment]*Service{
				appDeployment1: NewService("test1", username, portMap, false),
				appDeployment2: NewService("test2", username, portMap, false),
				appDeployment3: NewService("test3", username, portMap, false),
				svcDeployment1: NewService("test0", username, portMap, true),
			},
		}
	}

	Describe("NewCluster", func() {
		It("should construct a new cluster", func() {
			err = WriteClusterConfig(db, clusterName, yaml1)
			Expect(err).NotTo(HaveOccurred())

			cluster, err := NewCluster(db, username, clusterName, &mTest.MockReadiness{}, &mTest.MockReadiness{})
			Expect(err).NotTo(HaveOccurred())

			mockedCluster := mockCluster(username)
			Expect(cluster.AppDeployments).To(ConsistOf(mockedCluster.AppDeployments))
			Expect(cluster.SvcDeployments).To(ConsistOf(mockedCluster.SvcDeployments))
			for dp, svc := range mockedCluster.K8sServices {
				Expect(cluster.K8sServices).To(HaveKeyWithValue(dp, svc))
			}
		})

		It("should return error if cluster name not found", func() {
			cluster, err := NewCluster(db, username, clusterName, &mTest.MockReadiness{}, &mTest.MockReadiness{})
			Expect(err).To(HaveOccurred())
			Expect(cluster).To(BeNil())
		})
	})
})
