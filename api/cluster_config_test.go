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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/topfreegames/mystack-controller/api"

	"fmt"
	mTest "github.com/topfreegames/mystack-controller/testing"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
	"net/http"
	"net/http/httptest"
)

var _ = Describe("ClusterConfig", func() {
	var recorder *httptest.ResponseRecorder
	var clusterConfigHandler *ClusterConfigHandler
	var yaml1 = `
services:
  test0:
    image: svc1
    port: 5000
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
	var yamlWithVolume = `
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

	BeforeEach(func() {
		recorder = httptest.NewRecorder()
		clusterConfigHandler = &ClusterConfigHandler{App: app}
	})

	Describe("PUT /cluster-configs/{name}/create", func() {
		var (
			request     *http.Request
			err         error
			clusterName = "myCustomApps"
			route       = fmt.Sprintf("/cluster-configs/%s/create", clusterName)
		)

		BeforeEach(func() {
			yamlReader := mTest.JSONFor(map[string]interface{}{
				"yaml": yaml1,
			})
			request, err = http.NewRequest("PUT", route, yamlReader)
			Expect(err).NotTo(HaveOccurred())
			clusterConfigHandler.Method = "create"
		})

		AfterEach(func() {
			err = mock.ExpectationsWereMet()
			Expect(err).NotTo(HaveOccurred())
		})

		It("should return status 200 when creating valid cluster config", func() {
			mock.
				ExpectExec("^INSERT INTO clusters\\(name, yaml\\) VALUES\\((.+)\\)$").
				WithArgs(clusterName, yaml1).
				WillReturnResult(sqlmock.NewResult(1, 1))

			ctx := NewContextWithClusterConfig(request.Context(), yaml1)
			clusterConfigHandler.ServeHTTP(recorder, request.WithContext(ctx))

			Expect(recorder.Header().Get("Content-Type")).To(Equal("application/json"))
			Expect(recorder.Body.String()).To(Equal(`{"status": "ok"}`))
			Expect(recorder.Code).To(Equal(http.StatusOK))
		})

		It("should return status 200 when creating valid cluster config with volume", func() {
			mock.
				ExpectExec("^INSERT INTO clusters\\(name, yaml\\) VALUES\\((.+)\\)$").
				WithArgs(clusterName, yamlWithVolume).
				WillReturnResult(sqlmock.NewResult(1, 1))

			ctx := NewContextWithClusterConfig(request.Context(), yamlWithVolume)
			clusterConfigHandler.ServeHTTP(recorder, request.WithContext(ctx))

			Expect(recorder.Header().Get("Content-Type")).To(Equal("application/json"))
			Expect(recorder.Body.String()).To(Equal(`{"status": "ok"}`))
			Expect(recorder.Code).To(Equal(http.StatusOK))
		})

		It("should return status 409 when creating cluster config with known name", func() {
			mock.
				ExpectExec("^INSERT INTO clusters\\(name, yaml\\) VALUES\\((.+)\\)$").
				WithArgs(clusterName, yaml1).
				WillReturnError(fmt.Errorf(`pq: duplicate key value violates unique constraint "clusters_name_key"`))

			ctx := NewContextWithClusterConfig(request.Context(), yaml1)
			clusterConfigHandler.ServeHTTP(recorder, request.WithContext(ctx))

			Expect(recorder.Header().Get("Content-Type")).To(Equal("application/json"))
			Expect(recorder.Code).To(Equal(http.StatusConflict))
			bodyJSON := make(map[string]string)
			json.Unmarshal(recorder.Body.Bytes(), &bodyJSON)
			Expect(bodyJSON["code"]).To(Equal("OFF-003"))
			Expect(bodyJSON["description"]).To(Equal("pq: duplicate key value violates unique constraint \"clusters_name_key\""))
			Expect(bodyJSON["error"]).To(Equal("database error"))
		})

		It("should return status 400 when creating invalid cluster config", func() {
			invalidYaml := `
iam 
  {invalid: 123}
`

			ctx := NewContextWithClusterConfig(request.Context(), invalidYaml)
			clusterConfigHandler.ServeHTTP(recorder, request.WithContext(ctx))

			Expect(recorder.Header().Get("Content-Type")).To(Equal("application/json"))
			Expect(recorder.Code).To(Equal(http.StatusBadRequest))
			bodyJSON := make(map[string]string)
			json.Unmarshal(recorder.Body.Bytes(), &bodyJSON)
			Expect(bodyJSON["code"]).To(Equal("OFF-004"))
			Expect(bodyJSON["description"]).To(Equal("yaml: line 2: mapping values are not allowed in this context"))
			Expect(bodyJSON["error"]).To(Equal("parse yaml error"))
		})

		It("should return status 422 when creating empty cluster config", func() {
			yamlReader := mTest.JSONFor(map[string]interface{}{
				"yaml": "",
			})
			request, err = http.NewRequest("PUT", route, yamlReader)
			ctx := NewContextWithClusterConfig(request.Context(), "")
			clusterConfigHandler.ServeHTTP(recorder, request.WithContext(ctx))

			Expect(recorder.Header().Get("Content-Type")).To(Equal("application/json"))
			Expect(recorder.Code).To(Equal(http.StatusUnprocessableEntity))
			bodyJSON := make(map[string]string)
			json.Unmarshal(recorder.Body.Bytes(), &bodyJSON)
			Expect(bodyJSON["code"]).To(Equal("OFF-004"))
			Expect(bodyJSON["description"]).To(Equal("invalid empty config"))
			Expect(bodyJSON["error"]).To(Equal("write cluster config error"))
		})

		It("should return status 422 when creating empty cluster name", func() {
			yamlReader := mTest.JSONFor(map[string]interface{}{
				"yaml": "",
			})
			route := "/cluster-configs//create"
			request, _ := http.NewRequest("PUT", route, yamlReader)
			ctx := NewContextWithClusterConfig(request.Context(), "")
			clusterConfigHandler.ServeHTTP(recorder, request.WithContext(ctx))

			Expect(recorder.Header().Get("Content-Type")).To(Equal("application/json"))
			Expect(recorder.Code).To(Equal(http.StatusUnprocessableEntity))
			bodyJSON := make(map[string]string)
			json.Unmarshal(recorder.Body.Bytes(), &bodyJSON)
			Expect(bodyJSON["code"]).To(Equal("OFF-001"))
			Expect(bodyJSON["description"]).To(Equal("invalid empty cluster name"))
			Expect(bodyJSON["error"]).To(Equal("write cluster config error"))
		})

		It("should return status 401", func() {
			request.Header.Add("Authorization", "Bearer invalid-token")
			app.Router.ServeHTTP(recorder, request)
			Expect(recorder.Header().Get("Content-Type")).To(Equal("application/json"))
			Expect(recorder.Code).To(Equal(http.StatusUnauthorized))
			bodyJSON := make(map[string]string)
			json.Unmarshal(recorder.Body.Bytes(), &bodyJSON)
			Expect(bodyJSON["code"]).To(Equal("OFF-002"))
			Expect(bodyJSON["description"]).To(Equal("{\n \"error\": \"invalid_token\",\n \"error_description\": \"Invalid Value\"\n}\n"))
			Expect(bodyJSON["error"]).To(Equal("Unauthorized access token"))
		})
	})

	Describe("PUT /cluster-configs/{name}/remove", func() {
		var (
			request     *http.Request
			err         error
			clusterName = "myCustomApps"
			route       = fmt.Sprintf("/cluster-configs/%s/remove", clusterName)
		)

		BeforeEach(func() {
			clusterConfigHandler.Method = "remove"
			request, err = http.NewRequest("PUT", route, nil)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			err = mock.ExpectationsWereMet()
			Expect(err).NotTo(HaveOccurred())
		})

		It("should return 200 when removing existing cluster", func() {
			mock.
				ExpectExec("DELETE FROM clusters").
				WithArgs(clusterName).
				WillReturnResult(sqlmock.NewResult(1, 1))

			Expect(err).NotTo(HaveOccurred())
			clusterConfigHandler.ServeHTTP(recorder, request)

			Expect(recorder.Header().Get("Content-Type")).To(Equal("application/json"))
			Expect(recorder.Body.String()).To(Equal(`{"status": "ok"}`))
			Expect(recorder.Code).To(Equal(http.StatusOK))
		})

		It("should return 404 when removing non existing cluster", func() {
			mock.
				ExpectExec("DELETE FROM clusters").
				WithArgs(clusterName).
				WillReturnError(fmt.Errorf("sql: no rows in result set"))

			clusterConfigHandler.ServeHTTP(recorder, request)

			Expect(recorder.Header().Get("Content-Type")).To(Equal("application/json"))
			Expect(recorder.Code).To(Equal(http.StatusNotFound))
			bodyJSON := make(map[string]string)
			json.Unmarshal(recorder.Body.Bytes(), &bodyJSON)
			Expect(bodyJSON["code"]).To(Equal("OFF-003"))
			Expect(bodyJSON["description"]).To(Equal("sql: no rows in result set"))
			Expect(bodyJSON["error"]).To(Equal("database error"))
		})
	})

	Describe("GET /cluster-configs", func() {
		var (
			request *http.Request
			err     error
			route   = "/cluster-configs"
		)

		BeforeEach(func() {
			request, err = http.NewRequest("GET", route, nil)
			Expect(err).NotTo(HaveOccurred())
			clusterConfigHandler.Method = "list"
		})

		AfterEach(func() {
			err = mock.ExpectationsWereMet()
			Expect(err).NotTo(HaveOccurred())
		})

		It("should return list of cluster configs", func() {
			mock.
				ExpectQuery("^SELECT name FROM clusters$").
				WillReturnRows(sqlmock.NewRows([]string{"yaml"}).AddRow("cluster1").AddRow("cluster2"))

			clusterConfigHandler.ServeHTTP(recorder, request)

			Expect(recorder.Header().Get("Content-Type")).To(Equal("application/json"))
			Expect(recorder.Code).To(Equal(http.StatusOK))

			bodyJSON := make(map[string][]string)
			json.Unmarshal(recorder.Body.Bytes(), &bodyJSON)
			Expect(bodyJSON["names"]).To(ConsistOf("cluster1", "cluster2"))
		})

		It("should not return error is list is empty", func() {
			mock.
				ExpectQuery("^SELECT name FROM clusters$").
				WillReturnRows(sqlmock.NewRows([]string{"yaml"}))

			clusterConfigHandler.ServeHTTP(recorder, request)

			Expect(recorder.Header().Get("Content-Type")).To(Equal("application/json"))
			Expect(recorder.Code).To(Equal(http.StatusOK))

			bodyJSON := make(map[string][]string)
			json.Unmarshal(recorder.Body.Bytes(), &bodyJSON)
			Expect(bodyJSON["names"]).To(BeEmpty())
		})
	})

	Describe("GET /cluster-configs/{name}", func() {
		var (
			request     *http.Request
			err         error
			clusterName = "myCustomApps"
			route       = fmt.Sprintf("/cluster-configs/%s", clusterName)
		)

		BeforeEach(func() {
			request, err = http.NewRequest("GET", route, nil)
			Expect(err).NotTo(HaveOccurred())
			clusterConfigHandler.Method = "info"
		})

		AfterEach(func() {
			err = mock.ExpectationsWereMet()
			Expect(err).NotTo(HaveOccurred())
		})

		It("should return info from cluster configs", func() {
			mock.
				ExpectQuery("^SELECT yaml FROM clusters WHERE name(.+)$").
				WillReturnRows(sqlmock.NewRows([]string{"yaml"}).AddRow(yaml1))

			clusterConfigHandler.ServeHTTP(recorder, request)

			Expect(recorder.Header().Get("Content-Type")).To(Equal("application/json"))
			Expect(recorder.Code).To(Equal(http.StatusOK))

			bodyJSON := make(map[string]string)
			json.Unmarshal(recorder.Body.Bytes(), &bodyJSON)
			Expect(bodyJSON["yaml"]).To(Equal(yaml1))
		})

		It("should return error if cluster name doesn't exist", func() {
			mock.
				ExpectQuery("^SELECT yaml FROM clusters WHERE name(.+)$").
				WithArgs(clusterName).
				WillReturnError(fmt.Errorf("sql: no rows in result set"))

			clusterConfigHandler.ServeHTTP(recorder, request)

			Expect(recorder.Header().Get("Content-Type")).To(Equal("application/json"))
			Expect(recorder.Code).To(Equal(http.StatusNotFound))
		})
	})
})
