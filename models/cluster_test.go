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
	. "github.com/topfreegames/mystack-controller/models"

	"database/sql"
	"github.com/jmoiron/sqlx"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/fields"
	"k8s.io/client-go/pkg/labels"
)

var _ = Describe("Cluster", func() {
	const (
		yaml1 = `
setup:
  image: setup-img
services:
  test0:
    image: svc1
    ports: 
      - "5000"
      - "5001:5002"
apps:
  test1:
    image: app1
    ports: 
      - "5000"
      - "5001:5002"
  test2:
    image: app2
    ports: 
      - "5000"
      - "5001:5002"
  test3:
    image: app3
    ports: 
      - "5000"
      - "5001:5002"
`
	)
	var (
		db          *sql.DB
		sqlxDB      *sqlx.DB
		mock        sqlmock.Sqlmock
		err         error
		clusterName = "MyCustomApps"
		clientset   *fake.Clientset
		username    = "user"
		namespace   = "mystack-user"
		ports       = []int{5000, 5002}
		portMaps    = []*PortMap{
			&PortMap{Port: 5000, TargetPort: 5000},
			&PortMap{Port: 5001, TargetPort: 5002},
		}
		labelMap    = labels.Set{"mystack/routable": "true"}
		listOptions = v1.ListOptions{
			LabelSelector: labelMap.AsSelector().String(),
			FieldSelector: fields.Everything().String(),
		}
	)

	mockCluster := func(username string) *Cluster {
		return &Cluster{
			Username:  username,
			Namespace: namespace,
			Deployments: []*Deployment{
				NewDeployment("test0", username, "svc1", ports, nil),
				NewDeployment("test1", username, "app1", ports, nil),
				NewDeployment("test2", username, "app2", ports, nil),
				NewDeployment("test3", username, "app3", ports, nil),
			},
			Services: []*Service{
				NewService("test0", username, portMaps),
				NewService("test1", username, portMaps),
				NewService("test2", username, portMaps),
				NewService("test3", username, portMaps),
			},
			Setup: NewJob(username, "setup-img"),
		}
	}

	BeforeEach(func() {
		clientset = fake.NewSimpleClientset()
	})

	Describe("NewCluster", func() {
		BeforeEach(func() {
			db, mock, err = sqlmock.New()
			Expect(err).NotTo(HaveOccurred())
			sqlxDB = sqlx.NewDb(db, "postgres")
		})

		AfterEach(func() {
			err = mock.ExpectationsWereMet()
			Expect(err).NotTo(HaveOccurred())
			db.Close()
		})

		It("should return cluster from config on DB", func() {
			mockedCluster := mockCluster(username)

			mock.
				ExpectQuery("^SELECT yaml FROM clusters WHERE name = (.+)$").
				WithArgs(clusterName).
				WillReturnRows(sqlmock.NewRows([]string{"yaml"}).AddRow(yaml1))

			cluster, err := NewCluster(sqlxDB, username, clusterName)
			Expect(err).NotTo(HaveOccurred())
			Expect(cluster.Deployments).To(ConsistOf(mockedCluster.Deployments))
			Expect(cluster.Services).To(ConsistOf(mockedCluster.Services))
			Expect(cluster.Setup).To(Equal(mockedCluster.Setup))
		})

		It("should return error if clusterName doesn't exists on DB", func() {
			mock.
				ExpectQuery("^SELECT yaml FROM clusters WHERE name = (.+)$").
				WithArgs(clusterName).
				WillReturnRows(sqlmock.NewRows([]string{"yaml"}))

			cluster, err := NewCluster(sqlxDB, username, clusterName)
			Expect(cluster).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("sql: no rows in result set"))
		})

		It("should return error if empty clusterName", func() {
			mock.
				ExpectQuery("^SELECT yaml FROM clusters WHERE name = (.+)$").
				WithArgs(clusterName).
				WillReturnRows(sqlmock.NewRows([]string{"yaml"}))

			cluster, err := NewCluster(sqlxDB, username, clusterName)
			Expect(cluster).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("sql: no rows in result set"))
		})
	})

	Describe("Create", func() {
		It("should create cluster", func() {
			cluster := mockCluster(username)
			err := cluster.Create(clientset)
			Expect(err).NotTo(HaveOccurred())

			deploys, err := clientset.ExtensionsV1beta1().Deployments(namespace).List(listOptions)
			Expect(err).NotTo(HaveOccurred())
			Expect(deploys.Items).To(HaveLen(4))

			services, err := clientset.CoreV1().Services(namespace).List(listOptions)
			Expect(err).NotTo(HaveOccurred())
			Expect(services.Items).To(HaveLen(4))

			jobs, err := clientset.BatchV1().Jobs(namespace).List(listOptions)
			Expect(err).NotTo(HaveOccurred())
			Expect(jobs.Items).To(HaveLen(1))
		})

		It("should return error if creating same cluster twice", func() {
			cluster := mockCluster(username)
			err := cluster.Create(clientset)
			Expect(err).NotTo(HaveOccurred())

			err = cluster.Create(clientset)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Namespace \"mystack-user\" already exists"))
		})

		It("should run without setup image", func() {
			cluster := mockCluster(username)
			cluster.Setup = nil
			err := cluster.Create(clientset)
			Expect(err).NotTo(HaveOccurred())

			deploys, err := clientset.ExtensionsV1beta1().Deployments(namespace).List(listOptions)
			Expect(err).NotTo(HaveOccurred())
			Expect(deploys.Items).To(HaveLen(4))

			services, err := clientset.CoreV1().Services(namespace).List(listOptions)
			Expect(err).NotTo(HaveOccurred())
			Expect(services.Items).To(HaveLen(4))

			jobs, err := clientset.BatchV1().Jobs(namespace).List(listOptions)
			Expect(err).NotTo(HaveOccurred())
			Expect(jobs.Items).To(BeEmpty())
		})
	})

	Describe("Delete", func() {
		It("should delete cluster", func() {
			cluster := mockCluster(username)
			err := cluster.Create(clientset)
			Expect(err).NotTo(HaveOccurred())

			err = cluster.Delete(clientset)
			Expect(err).NotTo(HaveOccurred())

			Expect(NamespaceExists(clientset, namespace)).To(BeFalse())

			deploys, err := clientset.ExtensionsV1beta1().Deployments(namespace).List(listOptions)
			Expect(err).NotTo(HaveOccurred())
			Expect(deploys.Items).To(BeEmpty())

			services, err := clientset.CoreV1().Services(namespace).List(listOptions)
			Expect(err).NotTo(HaveOccurred())
			Expect(services.Items).To(BeEmpty())
		})

		It("should delete only specified cluster", func() {
			cluster1 := mockCluster("user1")
			err := cluster1.Create(clientset)
			Expect(err).NotTo(HaveOccurred())

			cluster2 := mockCluster("user2")
			err = cluster2.Create(clientset)
			Expect(err).NotTo(HaveOccurred())

			err = cluster1.Delete(clientset)
			Expect(err).NotTo(HaveOccurred())

			Expect(NamespaceExists(clientset, "mystack-user1")).To(BeFalse())
			Expect(NamespaceExists(clientset, "mystack-user2")).To(BeTrue())

			deploys, err := clientset.ExtensionsV1beta1().Deployments("mystack-user1").List(listOptions)
			Expect(err).NotTo(HaveOccurred())
			Expect(deploys.Items).To(BeEmpty())

			services, err := clientset.CoreV1().Services("mystack-user1").List(listOptions)
			Expect(err).NotTo(HaveOccurred())
			Expect(services.Items).To(BeEmpty())

			deploys, err = clientset.ExtensionsV1beta1().Deployments("mystack-user2").List(listOptions)
			Expect(err).NotTo(HaveOccurred())
			Expect(deploys.Items).To(HaveLen(4))

			services, err = clientset.CoreV1().Services("mystack-user2").List(listOptions)
			Expect(err).NotTo(HaveOccurred())
			Expect(services.Items).To(HaveLen(4))
		})

		It("should return error when deleting non existing cluster", func() {
			cluster := mockCluster(username)

			err = cluster.Delete(clientset)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Service \"test0\" not found"))
		})
	})
})
