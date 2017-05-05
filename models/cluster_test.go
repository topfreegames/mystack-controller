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

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/topfreegames/mystack-controller/models"

	"database/sql"

	"github.com/jmoiron/sqlx"
	mTest "github.com/topfreegames/mystack-controller/testing"
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
    readinessProbe:
      command:
        - echo
        - ready
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
    env:
      - name: VARIABLE_1
        value: 100
`
		yaml2 = `
setup:
  image: setup-img
  timeoutSeconds: 180
  periodSeconds: 10
services:
  test0:
    image: svc1
    ports: 
      - "5000"
      - "5001:5002"
    readinessProbe:
      command:
        - echo
        - ready
      periodSeconds: 10
      startDeploymentTimeoutSeconds: 180
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
    env:
      - name: VARIABLE_1
        value: 100
`
		yamlWithVolume = `
setup:
  image: setup-img
volumes:
  - name: svc-volume
    storage: 1Gi
services:
  svc1:
    image: svc1
    ports: 
      - "5000"
    volumeMount:
      name: svc-volume
      mountPath: /data
apps:
  app1:
    image: app1
    ports: 
      - "5000"
`
		invalidYaml1 = `
services:
  postgres:
    image: postgres:1.0
    ports:
      - 8!asd
apps:
  app1:
    image: app1
    ports:
      - 5000:5001
`
		invalidYaml2 = `
services:
  postgres:
    image: postgres:1.0
    ports:
      - 8585:8!asd
apps:
  app1:
    image: app1
    ports:
      - 5000:5001
`
	)
	var (
		db          *sql.DB
		sqlxDB      *sqlx.DB
		mock        sqlmock.Sqlmock
		err         error
		clusterName = "MyCustomApps"
		domain      = "mystack.com"
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

	mockCluster := func(period, timeout int, username string) *Cluster {
		namespace := fmt.Sprintf("mystack-%s", username)
		return &Cluster{
			Username:  username,
			Namespace: namespace,
			AppDeployments: []*Deployment{
				NewDeployment("test1", username, "app1", ports, nil, nil, nil),
				NewDeployment("test2", username, "app2", ports, nil, nil, nil),
				NewDeployment("test3", username, "app3", ports, []*EnvVar{
					&EnvVar{Name: "VARIABLE_1", Value: "100"},
				}, nil, nil),
			},
			SvcDeployments: []*Deployment{
				NewDeployment(
					"test0",
					username,
					"svc1",
					ports,
					nil,
					&Probe{
						Command:        []string{"echo", "ready"},
						TimeoutSeconds: timeout,
						PeriodSeconds:  period,
					}, nil,
				),
			},
			AppServices: []*Service{
				NewService("test1", username, portMaps),
				NewService("test2", username, portMaps),
				NewService("test3", username, portMaps),
			},
			SvcServices: []*Service{
				NewService("test0", username, portMaps),
			},
			Job: NewJob(
				username,
				&Setup{
					Image:          "setup-img",
					PeriodSeconds:  period,
					TimeoutSeconds: timeout,
				},
				[]*EnvVar{
					&EnvVar{Name: "VARIABLE_1", Value: "100"},
				},
			),
			DeploymentReadiness: &mTest.MockReadiness{},
			JobReadiness:        &mTest.MockReadiness{},
		}
	}

	mockedClusterWithVolume := &Cluster{
		Username:  username,
		Namespace: namespace,
		PersistentVolumeClaims: []*PersistentVolumeClaim{
			&PersistentVolumeClaim{Name: "svc-volume", Storage: "1Gi", Namespace: namespace},
		},
		AppDeployments: []*Deployment{
			NewDeployment("app1", username, "app1", []int{5000}, nil, nil, nil),
		},
		SvcDeployments: []*Deployment{
			NewDeployment("svc1", username, "svc1", []int{5000}, nil, nil, &VolumeMount{Name: "svc-volume", MountPath: "/data"}),
		},
		AppServices: []*Service{
			NewService("app1", username, []*PortMap{
				&PortMap{Port: 5000, TargetPort: 5000},
			}),
		},
		SvcServices: []*Service{
			NewService("svc1", username, []*PortMap{
				&PortMap{Port: 5000, TargetPort: 5000},
			}),
		},
		Job: NewJob(username, &Setup{
			Image:          "setup-img",
			PeriodSeconds:  0,
			TimeoutSeconds: 0,
		}, []*EnvVar{}),
		DeploymentReadiness: &mTest.MockReadiness{},
		JobReadiness:        &mTest.MockReadiness{},
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
			mockedCluster := mockCluster(0, 0, username)

			mock.
				ExpectQuery("^SELECT yaml FROM clusters WHERE name = (.+)$").
				WithArgs(clusterName).
				WillReturnRows(sqlmock.NewRows([]string{"yaml"}).AddRow(yaml1))

			cluster, err := NewCluster(sqlxDB, username, clusterName, &mTest.MockReadiness{}, &mTest.MockReadiness{})
			Expect(err).NotTo(HaveOccurred())
			Expect(cluster.AppDeployments).To(ConsistOf(mockedCluster.AppDeployments))
			Expect(cluster.SvcDeployments).To(ConsistOf(mockedCluster.SvcDeployments))
			Expect(cluster.SvcServices).To(ConsistOf(mockedCluster.SvcServices))
			Expect(cluster.AppServices).To(ConsistOf(mockedCluster.AppServices))
			Expect(cluster.Job).To(Equal(mockedCluster.Job))
		})

		It("should return cluster with non default times from DB", func() {
			mockedCluster := mockCluster(10, 180, username)

			mock.
				ExpectQuery("^SELECT yaml FROM clusters WHERE name = (.+)$").
				WithArgs(clusterName).
				WillReturnRows(sqlmock.NewRows([]string{"yaml"}).AddRow(yaml2))

			cluster, err := NewCluster(sqlxDB, username, clusterName, &mTest.MockReadiness{}, &mTest.MockReadiness{})
			Expect(err).NotTo(HaveOccurred())
			Expect(cluster.AppDeployments).To(ConsistOf(mockedCluster.AppDeployments))
			Expect(cluster.SvcDeployments).To(ConsistOf(mockedCluster.SvcDeployments))
			Expect(cluster.SvcServices).To(ConsistOf(mockedCluster.SvcServices))
			Expect(cluster.AppServices).To(ConsistOf(mockedCluster.AppServices))
			Expect(cluster.Job).To(Equal(mockedCluster.Job))
		})

		It("should return cluster with volume from DB", func() {
			mock.
				ExpectQuery("^SELECT yaml FROM clusters WHERE name = (.+)$").
				WithArgs(clusterName).
				WillReturnRows(sqlmock.NewRows([]string{"yaml"}).AddRow(yamlWithVolume))

			cluster, err := NewCluster(sqlxDB, username, clusterName, &mTest.MockReadiness{}, &mTest.MockReadiness{})
			Expect(err).NotTo(HaveOccurred())
			Expect(cluster.AppDeployments).To(ConsistOf(mockedClusterWithVolume.AppDeployments))
			Expect(cluster.SvcDeployments).To(ConsistOf(mockedClusterWithVolume.SvcDeployments))
			Expect(cluster.SvcServices).To(ConsistOf(mockedClusterWithVolume.SvcServices))
			Expect(cluster.AppServices).To(ConsistOf(mockedClusterWithVolume.AppServices))
			Expect(cluster.Job).To(Equal(mockedClusterWithVolume.Job))
			Expect(cluster.PersistentVolumeClaims).To(Equal(mockedClusterWithVolume.PersistentVolumeClaims))
		})

		It("should return error if clusterName doesn't exists on DB", func() {
			mock.
				ExpectQuery("^SELECT yaml FROM clusters WHERE name = (.+)$").
				WithArgs(clusterName).
				WillReturnRows(sqlmock.NewRows([]string{"yaml"}))

			cluster, err := NewCluster(sqlxDB, username, clusterName, &mTest.MockReadiness{}, &mTest.MockReadiness{})
			Expect(cluster).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("sql: no rows in result set"))
		})

		It("should return error if empty clusterName", func() {
			mock.
				ExpectQuery("^SELECT yaml FROM clusters WHERE name = (.+)$").
				WithArgs(clusterName).
				WillReturnRows(sqlmock.NewRows([]string{"yaml"}))

			cluster, err := NewCluster(sqlxDB, username, clusterName, &mTest.MockReadiness{}, &mTest.MockReadiness{})
			Expect(cluster).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("sql: no rows in result set"))
		})

		It("should return error with invalid yaml", func() {
			mock.
				ExpectQuery("^SELECT yaml FROM clusters WHERE name = (.+)$").
				WithArgs(clusterName).
				WillReturnRows(sqlmock.NewRows([]string{"yaml"}).AddRow(invalidYaml1))

			_, err := NewCluster(sqlxDB, username, clusterName, &mTest.MockReadiness{}, &mTest.MockReadiness{})
			Expect(fmt.Sprintf("%T", err)).To(Equal("*errors.YamlError"))
			Expect(err.Error()).To(Equal("strconv.Atoi: parsing \"8!asd\": invalid syntax"))
		})

		It("should return error with invalid yaml 2", func() {
			mock.
				ExpectQuery("^SELECT yaml FROM clusters WHERE name = (.+)$").
				WithArgs(clusterName).
				WillReturnRows(sqlmock.NewRows([]string{"yaml"}).AddRow(invalidYaml2))

			_, err := NewCluster(sqlxDB, username, clusterName, &mTest.MockReadiness{}, &mTest.MockReadiness{})
			Expect(fmt.Sprintf("%T", err)).To(Equal("*errors.YamlError"))
			Expect(err.Error()).To(Equal("strconv.Atoi: parsing \"8!asd\": invalid syntax"))
		})
	})

	Describe("Create", func() {
		It("should create cluster", func() {
			cluster := mockCluster(0, 0, username)
			err := cluster.Create(nil, clientset)
			Expect(err).NotTo(HaveOccurred())

			deploys, err := clientset.ExtensionsV1beta1().Deployments(namespace).List(listOptions)
			Expect(err).NotTo(HaveOccurred())
			Expect(deploys.Items).To(HaveLen(4))

			services, err := clientset.CoreV1().Services(namespace).List(listOptions)
			Expect(err).NotTo(HaveOccurred())
			Expect(services.Items).To(HaveLen(4))

			k8sJob, err := clientset.BatchV1().Jobs(namespace).Get("setup")
			Expect(err).NotTo(HaveOccurred())
			Expect(k8sJob).NotTo(BeNil())
			Expect(k8sJob.ObjectMeta.Namespace).To(Equal(namespace))
			Expect(k8sJob.ObjectMeta.Name).To(Equal("setup"))
			Expect(k8sJob.ObjectMeta.Labels["mystack/owner"]).To(Equal(username))
			Expect(k8sJob.ObjectMeta.Labels["app"]).To(Equal("setup"))
			Expect(k8sJob.ObjectMeta.Labels["heritage"]).To(Equal("mystack"))
		})

		It("should create cluster with volumes", func() {
			cluster := mockedClusterWithVolume
			err := cluster.Create(nil, clientset)
			Expect(err).NotTo(HaveOccurred())

			deploys, err := clientset.ExtensionsV1beta1().Deployments(namespace).List(listOptions)
			Expect(err).NotTo(HaveOccurred())
			Expect(deploys.Items).To(HaveLen(2))

			services, err := clientset.CoreV1().Services(namespace).List(listOptions)
			Expect(err).NotTo(HaveOccurred())
			Expect(services.Items).To(HaveLen(2))

			k8sJob, err := clientset.BatchV1().Jobs(namespace).Get("setup")
			Expect(err).NotTo(HaveOccurred())
			Expect(k8sJob).NotTo(BeNil())
			Expect(k8sJob.ObjectMeta.Namespace).To(Equal(namespace))
			Expect(k8sJob.ObjectMeta.Name).To(Equal("setup"))
			Expect(k8sJob.ObjectMeta.Labels["mystack/owner"]).To(Equal(username))
			Expect(k8sJob.ObjectMeta.Labels["app"]).To(Equal("setup"))
			Expect(k8sJob.ObjectMeta.Labels["heritage"]).To(Equal("mystack"))

			volumes, err := clientset.CoreV1().PersistentVolumeClaims(namespace).List(listOptions)
			Expect(err).NotTo(HaveOccurred())
			Expect(volumes.Items).To(HaveLen(1))

			volume := volumes.Items[0]
			Expect(volume.ObjectMeta.Name).To(Equal("svc-volume"))
			Expect(volume.ObjectMeta.Namespace).To(Equal(namespace))
			Expect(volume.ObjectMeta.Annotations["volume.alpha.kubernetes.io/storage-class"]).To(Equal("gp2"))
			Expect(volume.ObjectMeta.Labels["app"]).To(Equal("svc-volume"))
			Expect(volume.ObjectMeta.Labels["mystack/routable"]).To(Equal("true"))
			Expect(volume.Spec.AccessModes).To(Equal([]v1.PersistentVolumeAccessMode{"ReadWriteOnce"}))
		})

		It("should return error if creating same cluster twice", func() {
			cluster := mockCluster(0, 0, username)
			err := cluster.Create(nil, clientset)
			Expect(err).NotTo(HaveOccurred())

			err = cluster.Create(nil, clientset)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("namespace for user 'user' already exists"))
		})

		It("should run without setup image", func() {
			cluster := mockCluster(0, 0, username)
			cluster.Job = nil
			err := cluster.Create(nil, clientset)
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

		It("should run with env var as object", func() {
			obj := "{\\\"key\\\": \\\"value\\\"}"
			cluster := &Cluster{
				Username:  username,
				Namespace: namespace,
				AppDeployments: []*Deployment{
					NewDeployment("test1", username, "app1", ports, []*EnvVar{
						&EnvVar{Name: "VARIABLE_1", Value: obj},
					}, nil, nil),
				},
				AppServices: []*Service{
					NewService("test1", username, portMaps),
				},
				DeploymentReadiness: &mTest.MockReadiness{},
				JobReadiness:        &mTest.MockReadiness{},
			}
			err := cluster.Create(nil, clientset)
			Expect(err).NotTo(HaveOccurred())

			deploys, err := clientset.ExtensionsV1beta1().Deployments(namespace).List(listOptions)
			Expect(err).NotTo(HaveOccurred())
			deploy := deploys.Items[0]
			Expect(deploy.Spec.Template.Spec.Containers[0].Env[0].Value).To(Equal("{\"key\": \"value\"}"))
		})
	})

	Describe("Delete", func() {
		It("should delete cluster", func() {
			cluster := mockCluster(0, 0, username)
			err := cluster.Create(nil, clientset)
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
			cluster1 := mockCluster(0, 0, "user1")
			err := cluster1.Create(nil, clientset)
			Expect(err).NotTo(HaveOccurred())

			cluster2 := mockCluster(0, 0, "user2")
			err = cluster2.Create(nil, clientset)
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

		It("should delete cluster with volumes", func() {
			cluster := mockedClusterWithVolume
			err := cluster.Create(nil, clientset)
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

			volumes, err := clientset.CoreV1().PersistentVolumeClaims(namespace).List(listOptions)
			Expect(err).NotTo(HaveOccurred())
			Expect(volumes.Items).To(BeEmpty())
		})

		It("should return error when deleting non existing cluster", func() {
			cluster := mockCluster(0, 0, username)

			err = cluster.Delete(clientset)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("namespace for user 'user' not found"))
		})
	})

	Describe("Apps", func() {
		It("should return correct apps if cluster is running", func() {
			cluster := mockCluster(0, 0, "user")
			err := cluster.Create(nil, clientset)

			domains, err := cluster.Apps(clientset, domain)

			Expect(err).NotTo(HaveOccurred())
			Expect(domains["test0"]).To(Equal([]string{"test0.mystack-user.mystack.com"}))
			Expect(domains["test1"]).To(Equal([]string{"test1.mystack-user.mystack.com"}))
			Expect(domains["test2"]).To(Equal([]string{"test2.mystack-user.mystack.com"}))
			Expect(domains["test3"]).To(Equal([]string{"test3.mystack-user.mystack.com"}))
		})

		It("should return error if cluster is not runnig", func() {
			cluster := mockCluster(0, 0, "user")
			_, err := cluster.Apps(clientset, domain)
			Expect(err).To(HaveOccurred())
		})
	})
})
