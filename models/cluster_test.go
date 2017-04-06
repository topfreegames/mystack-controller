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
		port        = 5000
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

	BeforeEach(func() {
		clientset = fake.NewSimpleClientset()
		db, mock, err = sqlmock.New()
		Expect(err).NotTo(HaveOccurred())
		sqlxDB = sqlx.NewDb(db, "postgres")
	})

	Describe("NewCluster", func() {
		It("should return cluster from config on DB", func() {
			defer db.Close()

			mockedCluster := mockCluster(username)

			mock.
				ExpectExec("INSERT INTO clusters").
				WithArgs(clusterName, yaml1).
				WillReturnResult(sqlmock.NewResult(1, 1))
			mock.
				ExpectQuery("^SELECT yaml FROM clusters WHERE name = (.+)$").
				WithArgs(clusterName).
				WillReturnRows(sqlmock.NewRows([]string{"yaml"}).AddRow(yaml1))

			err = WriteClusterConfig(sqlxDB, clusterName, yaml1)
			Expect(err).NotTo(HaveOccurred())

			cluster, err := NewCluster(sqlxDB, username, clusterName)
			Expect(err).NotTo(HaveOccurred())
			Expect(cluster.Deployments).To(ConsistOf(mockedCluster.Deployments))
			Expect(cluster.Services).To(ConsistOf(mockedCluster.Services))

			err = mock.ExpectationsWereMet()
			Expect(err).NotTo(HaveOccurred())
		})

		It("should return error if clusterName doesn't exists on DB", func() {
			defer db.Close()

			mock.
				ExpectQuery("^SELECT yaml FROM clusters WHERE name = (.+)$").
				WithArgs(clusterName).
				WillReturnRows(sqlmock.NewRows([]string{"yaml"}))

			cluster, err := NewCluster(sqlxDB, username, clusterName)
			Expect(cluster).To(BeNil())
			Expect(err).To(HaveOccurred())

			err = mock.ExpectationsWereMet()
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("Create", func() {
		It("should create cluster", func() {
			cluster := mockCluster(username)
			err := cluster.Create(clientset)
			Expect(err).NotTo(HaveOccurred())

			deploys, err := clientset.ExtensionsV1beta1().Deployments(namespace).List(listOptions)
			Expect(err).NotTo(HaveOccurred())
			Expect(deploys.Items).To(HaveLen(3))

			services, err := clientset.CoreV1().Services(namespace).List(listOptions)
			Expect(err).NotTo(HaveOccurred())
			Expect(services.Items).To(HaveLen(3))
		})

		It("should return error if creating same cluster twice", func() {
			cluster := mockCluster(username)
			err := cluster.Create(clientset)
			Expect(err).NotTo(HaveOccurred())

			err = cluster.Create(clientset)
			Expect(err).To(HaveOccurred())
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
			Expect(deploys.Items).To(HaveLen(3))

			services, err = clientset.CoreV1().Services("mystack-user2").List(listOptions)
			Expect(err).NotTo(HaveOccurred())
			Expect(services.Items).To(HaveLen(3))
		})
	})
})