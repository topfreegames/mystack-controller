// mystack-controller api
// +build integration
// https://github.com/topfreegames/mystack-controller
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2017 Top Free Games <backend@tfgco.com>

package api_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/topfreegames/mystack-controller/api"

	"encoding/json"
	"fmt"
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
			request, err = http.NewRequest("PUT", route, nil)
			Expect(err).NotTo(HaveOccurred())
			clusterConfigHandler.Method = "create"
		})

		It("should return status 200 when creating valid cluster config", func() {
			ctx := NewContextWithClusterConfig(request.Context(), yaml1)
			clusterConfigHandler.ServeHTTP(recorder, request.WithContext(ctx))

			Expect(recorder.Header().Get("Content-Type")).To(Equal("application/json"))
			Expect(recorder.Body.String()).To(Equal(`{"status": "ok"}`))
			Expect(recorder.Code).To(Equal(http.StatusOK))
		})

		It("should return status 409 when creating cluster config with known name", func() {
			ctx := NewContextWithClusterConfig(request.Context(), yaml1)
			clusterConfigHandler.ServeHTTP(recorder, request.WithContext(ctx))

			recorder = httptest.NewRecorder()
			request, err = http.NewRequest("PUT", route, nil)

			ctx = NewContextWithClusterConfig(request.Context(), yaml1)
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
iam {
  invalid: 123
}`
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

		It("should return status 401 when complete route without access token", func() {
			deleteRoute := "/clusters/myCustomApps/delete"
			request, err = http.NewRequest("DELETE", deleteRoute, nil)
			request.Header.Add("Authorization", "Bearer invalid-token")
			app.Router.ServeHTTP(recorder, request)
			Expect(recorder.Header().Get("Content-Type")).To(Equal("application/json"))
			Expect(recorder.Code).To(Equal(http.StatusUnauthorized))
			bodyJSON := make(map[string]string)
			json.Unmarshal(recorder.Body.Bytes(), &bodyJSON)
			Expect(bodyJSON["code"]).To(Equal("OFF-002"))
			Expect(bodyJSON["description"]).To(Equal("sql: no rows in result set"))
			Expect(bodyJSON["error"]).To(Equal("Access Token not found (have you logged in?)"))
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

		It("should return 200 when removing existing cluster", func() {
			clusterConfigHandler.Method = "create"
			request, err = http.NewRequest("PUT", createRoute, nil)
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

		It("should return 404 when removing non existing cluster", func() {
			clusterConfigHandler.Method = "remove"
			recorder = httptest.NewRecorder()
			request, err = http.NewRequest("PUT", removeRoute, nil)
			Expect(err).NotTo(HaveOccurred())
			clusterConfigHandler.ServeHTTP(recorder, request)

			Expect(recorder.Header().Get("Content-Type")).To(Equal("application/json"))
			Expect(recorder.Code).To(Equal(http.StatusNotFound))
			bodyJSON := make(map[string]string)
			json.Unmarshal(recorder.Body.Bytes(), &bodyJSON)
			Expect(bodyJSON["code"]).To(Equal("OFF-003"))
			Expect(bodyJSON["description"]).To(Equal("sql: no rows in result set"))
			Expect(bodyJSON["error"]).To(Equal("database error"))
		})

		It("should return status 401 when complete route without access token", func() {
			request, err = http.NewRequest("DELETE", removeRoute, nil)
			request.Header.Add("Authorization", "Bearer invalid-token")
			app.Router.ServeHTTP(recorder, request)
			Expect(recorder.Header().Get("Content-Type")).To(Equal("application/json"))
			Expect(recorder.Code).To(Equal(http.StatusUnauthorized))
			bodyJSON := make(map[string]string)
			json.Unmarshal(recorder.Body.Bytes(), &bodyJSON)
			Expect(bodyJSON["code"]).To(Equal("OFF-002"))
			Expect(bodyJSON["description"]).To(Equal("sql: no rows in result set"))
			Expect(bodyJSON["error"]).To(Equal("Access Token not found (have you logged in?)"))
		})
	})
})
