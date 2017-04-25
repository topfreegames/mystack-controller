// mystack-controller api
// +build unit
// https://github.com/topfreegames/mystack-controller
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2017 Top Free Games <backend@tfgco.com>

package api_test

import (
	"encoding/json"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/topfreegames/mystack-controller/api"
	"github.com/topfreegames/mystack-controller/models"

	mTest "github.com/topfreegames/mystack-controller/testing"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
	"net/http"
	"net/http/httptest"
)

var _ = Describe("Cluster", func() {

	var (
		recorder       *httptest.ResponseRecorder
		clusterName    = "myCustomApps"
		clusterHandler *ClusterHandler
		yaml1          = `
setup:
  image: setup-img
services:
  test0:
    image: svc1
    port: 5000
apps:
  test1:
    image: app1
    port: 5000
`
		yamlWithoutSetup = `
services:
  test0:
    image: svc1
    port: 5000
apps:
  test1:
    image: app1
    port: 5000
`
	)

	BeforeEach(func() {
		recorder = httptest.NewRecorder()
		clusterHandler = &ClusterHandler{App: app}
	})

	Describe("PUT /clusters/{name}/create", func() {

		var (
			err     error
			request *http.Request
			route   = fmt.Sprintf("/clusters/%s/create", clusterName)
		)

		BeforeEach(func() {
			clusterHandler.Method = "create"
			request, err = http.NewRequest("PUT", route, nil)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			err = mock.ExpectationsWereMet()
			Expect(err).NotTo(HaveOccurred())
		})

		FIt("should create existing clusterName", func() {
			mock.
				ExpectQuery("^SELECT yaml FROM clusters WHERE name = (.+)$").
				WithArgs(clusterName).
				WillReturnRows(sqlmock.NewRows([]string{"yaml"}).AddRow(yaml1))
			mock.
				ExpectQuery("^SELECT yaml FROM clusters WHERE name = (.+)$").
				WithArgs(clusterName).
				WillReturnRows(sqlmock.NewRows([]string{"yaml"}).AddRow(yaml1))

			ctx := NewContextWithEmail(request.Context(), "user@example.com")
			clusterHandler.ServeHTTP(recorder, request.WithContext(ctx))

			Expect(recorder.Header().Get("Content-Type")).To(Equal("application/json"))
			Expect(recorder.Code).To(Equal(http.StatusOK))
			bodyJSON := make(map[string]map[string][]string)
			json.Unmarshal(recorder.Body.Bytes(), &bodyJSON)
			Expect(bodyJSON["domains"]["test0"]).To(Equal([]string{"test0.mystack-user.mystack.com"}))
			Expect(bodyJSON["domains"]["test1"]).To(Equal([]string{"test1.mystack-user.mystack.com"}))
		})

		It("should create existing clusterName without setup", func() {
			mock.
				ExpectQuery("^SELECT yaml FROM clusters WHERE name = (.+)$").
				WithArgs(clusterName).
				WillReturnRows(sqlmock.NewRows([]string{"yaml"}).AddRow(yamlWithoutSetup))
			mock.
				ExpectQuery("^SELECT yaml FROM clusters WHERE name = (.+)$").
				WithArgs(clusterName).
				WillReturnRows(sqlmock.NewRows([]string{"yaml"}).AddRow(yamlWithoutSetup))

			ctx := NewContextWithEmail(request.Context(), "user@example.com")
			clusterHandler.ServeHTTP(recorder, request.WithContext(ctx))

			Expect(recorder.Header().Get("Content-Type")).To(Equal("application/json"))
			Expect(recorder.Code).To(Equal(http.StatusOK))
			bodyJSON := make(map[string]map[string][]string)
			json.Unmarshal(recorder.Body.Bytes(), &bodyJSON)
			Expect(bodyJSON["domains"]["test0"]).To(Equal([]string{"test0.mystack-user.mystack.com"}))
			Expect(bodyJSON["domains"]["test1"]).To(Equal([]string{"test1.mystack-user.mystack.com"}))
		})

		It("should not create cluster twice", func() {
			mock.
				ExpectQuery("^SELECT yaml FROM clusters WHERE name = (.+)$").
				WithArgs(clusterName).
				WillReturnRows(sqlmock.NewRows([]string{"yaml"}).AddRow(yaml1))
			mock.
				ExpectQuery("^SELECT yaml FROM clusters WHERE name = (.+)$").
				WithArgs(clusterName).
				WillReturnRows(sqlmock.NewRows([]string{"yaml"}).AddRow(yaml1))

			ctx := NewContextWithEmail(request.Context(), "user@example.com")
			clusterHandler.ServeHTTP(recorder, request.WithContext(ctx))

			recorder = httptest.NewRecorder()
			request, _ = http.NewRequest("PUT", route, nil)
			ctx = NewContextWithEmail(request.Context(), "user@example.com")
			clusterHandler.ServeHTTP(recorder, request.WithContext(ctx))

			Expect(recorder.Header().Get("Content-Type")).To(Equal("application/json"))
			bodyJSON := make(map[string]string)
			json.Unmarshal(recorder.Body.Bytes(), &bodyJSON)
			Expect(bodyJSON["code"]).To(Equal("OFF-004"))
			Expect(bodyJSON["description"]).To(Equal("namespace for user 'user' already exists"))
			Expect(bodyJSON["error"]).To(Equal("create cluster error"))
			Expect(recorder.Code).To(Equal(http.StatusConflict))
		})

		It("should return error 404 when create non existing clusterName", func() {
			mock.
				ExpectQuery("^SELECT yaml FROM clusters WHERE name = (.+)$").
				WithArgs(clusterName).
				WillReturnError(fmt.Errorf("sql: no rows in result set"))

			ctx := NewContextWithEmail(request.Context(), "user@example.com")
			clusterHandler.ServeHTTP(recorder, request.WithContext(ctx))

			Expect(recorder.Header().Get("Content-Type")).To(Equal("application/json"))
			Expect(recorder.Code).To(Equal(http.StatusNotFound))
			bodyJSON := make(map[string]string)
			json.Unmarshal(recorder.Body.Bytes(), &bodyJSON)
			Expect(bodyJSON["code"]).To(Equal("OFF-003"))
			Expect(bodyJSON["description"]).To(Equal("sql: no rows in result set"))
			Expect(bodyJSON["error"]).To(Equal("database error"))
		})
	})

	Describe("PUT /clusters/{name}/delete", func() {

		var (
			err     error
			request *http.Request
			route   = fmt.Sprintf("/clusters/%s/delete", clusterName)
		)

		BeforeEach(func() {
			clusterHandler.Method = "delete"
			request, err = http.NewRequest("PUT", route, nil)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			err = mock.ExpectationsWereMet()
			Expect(err).NotTo(HaveOccurred())
		})

		It("should delete existing clusterName", func() {
			mock.
				ExpectQuery("^SELECT yaml FROM clusters WHERE name = (.+)$").
				WithArgs(clusterName).
				WillReturnRows(sqlmock.NewRows([]string{"yaml"}).AddRow(yaml1))
			mock.
				ExpectQuery("^SELECT yaml FROM clusters WHERE name = (.+)$").
				WithArgs(clusterName).
				WillReturnRows(sqlmock.NewRows([]string{"yaml"}).AddRow(yaml1))

			cluster, err := models.NewCluster(app.DB, "user", clusterName, &mTest.MockReadiness{}, &mTest.MockReadiness{})
			Expect(err).NotTo(HaveOccurred())
			err = cluster.Create(app.Logger, app.Clientset)
			Expect(err).NotTo(HaveOccurred())

			ctx := NewContextWithEmail(request.Context(), "user@example.com")
			clusterHandler.ServeHTTP(recorder, request.WithContext(ctx))

			Expect(recorder.Header().Get("Content-Type")).To(Equal("application/json"))
			Expect(recorder.Body.String()).To(Equal(`{"status": "ok"}`))
			Expect(recorder.Code).To(Equal(http.StatusOK))
		})

		It("should return error 404 when deleting non existing cluster", func() {
			mock.
				ExpectQuery("^SELECT yaml FROM clusters WHERE name = (.+)$").
				WithArgs(clusterName).
				WillReturnError(fmt.Errorf("sql: no rows in result set"))

			ctx := NewContextWithEmail(request.Context(), "user@example.com")
			clusterHandler.ServeHTTP(recorder, request.WithContext(ctx))

			Expect(recorder.Header().Get("Content-Type")).To(Equal("application/json"))
			bodyJSON := make(map[string]string)
			json.Unmarshal(recorder.Body.Bytes(), &bodyJSON)
			Expect(bodyJSON["description"]).To(Equal("namespace for user 'user' not found"))
			Expect(bodyJSON["error"]).To(Equal("delete cluster error"))
			Expect(bodyJSON["code"]).To(Equal("OFF-004"))
			Expect(recorder.Code).To(Equal(http.StatusNotFound))
		})

		It("should delete cluster even if cluster config doesn't exist anymore", func() {
			mock.
				ExpectQuery("^SELECT yaml FROM clusters WHERE name = (.+)$").
				WithArgs(clusterName).
				WillReturnRows(sqlmock.NewRows([]string{"yaml"}).AddRow(yaml1))
			mock.
				ExpectQuery("^SELECT yaml FROM clusters WHERE name = (.+)$").
				WithArgs(clusterName).
				WillReturnError(fmt.Errorf("sql: no rows in result set"))

			cluster, err := models.NewCluster(app.DB, "user", clusterName, &mTest.MockReadiness{}, &mTest.MockReadiness{})
			Expect(err).NotTo(HaveOccurred())
			err = cluster.Create(app.Logger, app.Clientset)
			Expect(err).NotTo(HaveOccurred())

			ctx := NewContextWithEmail(request.Context(), "user@example.com")
			clusterHandler.ServeHTTP(recorder, request.WithContext(ctx))

			Expect(recorder.Header().Get("Content-Type")).To(Equal("application/json"))
			Expect(recorder.Body.String()).To(Equal(`{"status": "ok"}`))
			Expect(recorder.Code).To(Equal(http.StatusOK))
		})
	})

	Describe("GET /clusters/{name}/apps", func() {
		var (
			err     error
			request *http.Request
			route   = fmt.Sprintf("/clusters/%s/apps", clusterName)
		)

		BeforeEach(func() {
			clusterHandler.Method = "apps"
			request, err = http.NewRequest("GET", route, nil)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			err = mock.ExpectationsWereMet()
			Expect(err).NotTo(HaveOccurred())
		})

		It("should return correct apps", func() {
			mock.
				ExpectQuery("^SELECT yaml FROM clusters WHERE name = (.+)$").
				WithArgs(clusterName).
				WillReturnRows(sqlmock.NewRows([]string{"yaml"}).AddRow(yaml1))
			mock.
				ExpectQuery("^SELECT yaml FROM clusters WHERE name = (.+)$").
				WithArgs(clusterName).
				WillReturnRows(sqlmock.NewRows([]string{"yaml"}).AddRow(yaml1))
			mock.
				ExpectQuery("^SELECT yaml FROM clusters WHERE name = (.+)$").
				WithArgs(clusterName).
				WillReturnRows(sqlmock.NewRows([]string{"yaml"}).AddRow(yaml1))

			cluster, err := models.NewCluster(app.DB, "user", clusterName, &mTest.MockReadiness{}, &mTest.MockReadiness{})
			Expect(err).NotTo(HaveOccurred())
			err = cluster.Create(app.Logger, app.Clientset)
			Expect(err).NotTo(HaveOccurred())

			ctx := NewContextWithEmail(request.Context(), "user@example.com")
			clusterHandler.ServeHTTP(recorder, request.WithContext(ctx))

			Expect(recorder.Header().Get("Content-Type")).To(Equal("application/json"))
			Expect(recorder.Code).To(Equal(http.StatusOK))
			bodyJSON := make(map[string]map[string][]string)
			json.Unmarshal(recorder.Body.Bytes(), &bodyJSON)
			Expect(bodyJSON["domains"]["test0"]).To(Equal([]string{"test0.mystack-user.mystack.com"}))
			Expect(bodyJSON["domains"]["test1"]).To(Equal([]string{"test1.mystack-user.mystack.com"}))
		})

		It("should return status 404 if namespace doesn't exist", func() {
			mock.
				ExpectQuery("^SELECT yaml FROM clusters WHERE name = (.+)$").
				WithArgs(clusterName).
				WillReturnRows(sqlmock.NewRows([]string{"yaml"}).AddRow(yaml1))
			mock.
				ExpectQuery("^SELECT yaml FROM clusters WHERE name = (.+)$").
				WithArgs(clusterName).
				WillReturnRows(sqlmock.NewRows([]string{"yaml"}).AddRow(yaml1))

			_, err := models.NewCluster(app.DB, "user", clusterName, &mTest.MockReadiness{}, &mTest.MockReadiness{})
			Expect(err).NotTo(HaveOccurred())

			ctx := NewContextWithEmail(request.Context(), "user@example.com")
			clusterHandler.ServeHTTP(recorder, request.WithContext(ctx))

			Expect(recorder.Header().Get("Content-Type")).To(Equal("application/json"))
			Expect(recorder.Code).To(Equal(http.StatusNotFound))
			bodyJSON := make(map[string]string)
			json.Unmarshal(recorder.Body.Bytes(), &bodyJSON)
			Expect(bodyJSON["description"]).To(Equal("namespace for user 'user' not found"))
			Expect(bodyJSON["error"]).To(Equal("get apps error"))
			Expect(bodyJSON["code"]).To(Equal("OFF-004"))
		})
	})
})
