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
				ExpectExec("INSERT INTO clusters").
				WithArgs(clusterName, yaml1).
				WillReturnResult(sqlmock.NewResult(1, 1))

			ctx := NewContextWithClusterConfig(request.Context(), yaml1)
			clusterConfigHandler.ServeHTTP(recorder, request.WithContext(ctx))

			Expect(recorder.Header().Get("Content-Type")).To(Equal("application/json"))
			Expect(recorder.Body.String()).To(Equal(`{"status": "ok"}`))
			Expect(recorder.Code).To(Equal(http.StatusOK))
		})

		It("should return status 409 when creating cluster config with known name", func() {
			mock.
				ExpectExec("INSERT INTO clusters").
				WithArgs(clusterName, yaml1).
				WillReturnResult(sqlmock.NewResult(1, 1))
			mock.
				ExpectExec("INSERT INTO clusters").
				WithArgs(clusterName, yaml1).
				WillReturnError(fmt.Errorf(`pq: duplicate key value violates unique constraint "clusters_name_key"`))

			ctx := NewContextWithClusterConfig(request.Context(), yaml1)
			clusterConfigHandler.ServeHTTP(recorder, request.WithContext(ctx))

			recorder = httptest.NewRecorder()
			yamlReader := mTest.JSONFor(map[string]interface{}{
				"yaml": yaml1,
			})
			request, err = http.NewRequest("PUT", route, yamlReader)

			ctx = NewContextWithClusterConfig(request.Context(), yaml1)
			clusterConfigHandler.ServeHTTP(recorder, request.WithContext(ctx))

			Expect(recorder.Header().Get("Content-Type")).To(Equal("application/json"))
			//TODO: change to return 409 (THIS IS CRITICAL)
			Expect(recorder.Code).To(Equal(http.StatusInternalServerError))
			bodyJSON := make(map[string]string)
			json.Unmarshal(recorder.Body.Bytes(), &bodyJSON)
			Expect(bodyJSON["code"]).To(Equal("OFF-001"))
			Expect(bodyJSON["description"]).To(Equal("pq: duplicate key value violates unique constraint \"clusters_name_key\""))
			Expect(bodyJSON["error"]).To(Equal("Error writing cluster config"))
		})

		It("should return status 422 when creating invalid cluster config", func() {
			invalidYaml := "iam {invalid: 123}"
			mock.
				ExpectExec("INSERT INTO clusters").
				WithArgs(clusterName, invalidYaml).
				WillReturnError(fmt.Errorf(`yaml: line 3: mapping values are not allowed in this context`))

			ctx := NewContextWithClusterConfig(request.Context(), invalidYaml)
			clusterConfigHandler.ServeHTTP(recorder, request.WithContext(ctx))

			Expect(recorder.Header().Get("Content-Type")).To(Equal("application/json"))
			//TODO: change to return 422 (THIS IS CRITICAL)
			Expect(recorder.Code).To(Equal(http.StatusInternalServerError))
			bodyJSON := make(map[string]string)
			json.Unmarshal(recorder.Body.Bytes(), &bodyJSON)
			Expect(bodyJSON["code"]).To(Equal("OFF-001"))
			Expect(bodyJSON["description"]).To(Equal("yaml: line 3: mapping values are not allowed in this context"))
			Expect(bodyJSON["error"]).To(Equal("Error writing cluster config"))
		})

		It("should return status 422 when creating empty cluster config", func() {
			yamlReader := mTest.JSONFor(map[string]interface{}{
				"yaml": "",
			})
			request, err = http.NewRequest("PUT", route, yamlReader)
			ctx := NewContextWithClusterConfig(request.Context(), "")
			clusterConfigHandler.ServeHTTP(recorder, request.WithContext(ctx))

			Expect(recorder.Header().Get("Content-Type")).To(Equal("application/json"))
			//TODO: change to return 422 (THIS IS CRITICAL)
			Expect(recorder.Code).To(Equal(http.StatusInternalServerError))
			bodyJSON := make(map[string]string)
			json.Unmarshal(recorder.Body.Bytes(), &bodyJSON)
			Expect(bodyJSON["code"]).To(Equal("OFF-001"))
			Expect(bodyJSON["description"]).To(Equal("yaml: invalid empty yaml"))
			Expect(bodyJSON["error"]).To(Equal("Error writing cluster config"))
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
			removeRoute = fmt.Sprintf("/cluster-configs/%s/remove", clusterName)
			createRoute = fmt.Sprintf("/cluster-configs/%s/create", clusterName)
		)

		AfterEach(func() {
			err = mock.ExpectationsWereMet()
			Expect(err).NotTo(HaveOccurred())
		})

		It("should return 200 when removing existing cluster", func() {
			mock.
				ExpectExec("INSERT INTO clusters").
				WithArgs(clusterName, yaml1).
				WillReturnResult(sqlmock.NewResult(1, 1))
			mock.
				ExpectExec("DELETE FROM clusters").
				WithArgs(clusterName).
				WillReturnResult(sqlmock.NewResult(1, 1))

			clusterConfigHandler.Method = "create"
			yamlReader := mTest.JSONFor(map[string]interface{}{
				"yaml": yaml1,
			})
			request, err = http.NewRequest("PUT", createRoute, yamlReader)
			Expect(err).NotTo(HaveOccurred())
			ctx := NewContextWithClusterConfig(request.Context(), yaml1)
			clusterConfigHandler.ServeHTTP(recorder, request.WithContext(ctx))

			clusterConfigHandler.Method = "remove"
			recorder = httptest.NewRecorder()
			request, err = http.NewRequest("PUT", removeRoute, nil)
			Expect(err).NotTo(HaveOccurred())
			clusterConfigHandler.ServeHTTP(recorder, request)

			Expect(recorder.Header().Get("Content-Type")).To(Equal("application/json"))
			Expect(recorder.Body.String()).To(Equal(`{"status": "ok"}`))
			Expect(recorder.Code).To(Equal(http.StatusOK))
		})

		It("should return 422 when removing non existing cluster", func() {
			mock.
				ExpectExec("DELETE FROM clusters").
				WithArgs(clusterName).
				WillReturnError(fmt.Errorf("Error removing cluster config"))

			clusterConfigHandler.Method = "remove"
			recorder = httptest.NewRecorder()
			request, err = http.NewRequest("PUT", removeRoute, nil)
			Expect(err).NotTo(HaveOccurred())
			clusterConfigHandler.ServeHTTP(recorder, request)

			Expect(recorder.Header().Get("Content-Type")).To(Equal("application/json"))
			//TODO: change to return 422 (THIS IS CRITICAL)
			Expect(recorder.Code).To(Equal(http.StatusInternalServerError))
			bodyJSON := make(map[string]string)
			json.Unmarshal(recorder.Body.Bytes(), &bodyJSON)
			Expect(bodyJSON["code"]).To(Equal("OFF-001"))
			Expect(bodyJSON["description"]).To(Equal("Error removing cluster config"))
			Expect(bodyJSON["error"]).To(Equal("Error removing cluster config"))
		})

		It("should return status 401 when complete route without access token", func() {
			request, err = http.NewRequest("PUT", removeRoute, nil)
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
})
