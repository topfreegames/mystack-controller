// mystack-controller api
// +build integration
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

		It("should create existing clusterName", func() {
			route = fmt.Sprintf("/cluster-configs/%s/create", clusterName)
			createRequest, err := http.NewRequest("PUT", route, nil)
			Expect(err).NotTo(HaveOccurred())

			clusterConfigHandler := &ClusterConfigHandler{App: app, Method: "create"}
			ctx := NewContextWithClusterConfig(createRequest.Context(), yaml1)
			clusterConfigHandler.ServeHTTP(recorder, createRequest.WithContext(ctx))

			recorder = httptest.NewRecorder()
			ctx = NewContextWithEmail(request.Context(), "user@example.com")
			clusterHandler.ServeHTTP(recorder, request.WithContext(ctx))

			Expect(recorder.Header().Get("Content-Type")).To(Equal("application/json"))
			bodyJSON := make(map[string]map[string][]string)
			json.Unmarshal(recorder.Body.Bytes(), &bodyJSON)
			Expect(bodyJSON["domains"]["test0"]).To(Equal([]string{"test0.mystack-user.mystack.com"}))
			Expect(bodyJSON["domains"]["test1"]).To(Equal([]string{"test1.mystack-user.mystack.com"}))
		})

		It("should create clusterName without setup", func() {
			route = fmt.Sprintf("/cluster-configs/%s/create", clusterName)
			createRequest, err := http.NewRequest("PUT", route, nil)
			Expect(err).NotTo(HaveOccurred())

			clusterConfigHandler := &ClusterConfigHandler{App: app, Method: "create"}
			ctx := NewContextWithClusterConfig(createRequest.Context(), yamlWithoutSetup)
			clusterConfigHandler.ServeHTTP(recorder, createRequest.WithContext(ctx))

			recorder = httptest.NewRecorder()
			ctx = NewContextWithEmail(request.Context(), "user@example.com")
			clusterHandler.ServeHTTP(recorder, request.WithContext(ctx))

			Expect(recorder.Header().Get("Content-Type")).To(Equal("application/json"))
			bodyJSON := make(map[string]map[string][]string)
			json.Unmarshal(recorder.Body.Bytes(), &bodyJSON)
			Expect(bodyJSON["domains"]["test0"]).To(Equal([]string{"test0.mystack-user.mystack.com"}))
			Expect(bodyJSON["domains"]["test1"]).To(Equal([]string{"test1.mystack-user.mystack.com"}))
		})

		It("should return error 404 when create non existing clusterName", func() {
			ctx := NewContextWithEmail(request.Context(), "derp@example.com")
			clusterHandler.ServeHTTP(recorder, request.WithContext(ctx))

			Expect(recorder.Header().Get("Content-Type")).To(Equal("application/json"))
			Expect(recorder.Code).To(Equal(http.StatusNotFound))
			bodyJSON := make(map[string]string)
			json.Unmarshal(recorder.Body.Bytes(), &bodyJSON)
			Expect(bodyJSON["code"]).To(Equal("OFF-003"))
			Expect(bodyJSON["description"]).To(Equal("sql: no rows in result set"))
			Expect(bodyJSON["error"]).To(Equal("database error"))
		})

		It("should return status 401 when complete route without access token", func() {
			request, err = http.NewRequest("PUT", route, nil)
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

		It("should delete existing clusterName", func() {
			route = fmt.Sprintf("/cluster-configs/%s/create", clusterName)
			createRequest, _ := http.NewRequest("PUT", route, nil)
			clusterConfigHandler := &ClusterConfigHandler{App: app, Method: "create"}
			ctx := NewContextWithClusterConfig(createRequest.Context(), yaml1)
			clusterConfigHandler.ServeHTTP(recorder, createRequest.WithContext(ctx))

			clusterHandler.Method = "create"
			route = fmt.Sprintf("/clusters/%s/create", clusterName)
			createRequest, _ = http.NewRequest("PUT", route, nil)
			recorder = httptest.NewRecorder()
			ctx = NewContextWithEmail(createRequest.Context(), "user@example.com")
			clusterHandler.ServeHTTP(recorder, createRequest.WithContext(ctx))

			clusterHandler.Method = "delete"
			recorder = httptest.NewRecorder()
			ctx = NewContextWithEmail(request.Context(), "user@example.com")
			clusterHandler.ServeHTTP(recorder, request.WithContext(ctx))

			Expect(recorder.Code).To(Equal(http.StatusOK))
			Expect(recorder.Body.String()).To(Equal(`{"status": "ok"}`))
			Expect(recorder.Header().Get("Content-Type")).To(Equal("application/json"))
		})

		It("should return error 404 when deleting non existing clusterName", func() {
			ctx := NewContextWithEmail(request.Context(), "derp@example.com")
			clusterHandler.ServeHTTP(recorder, request.WithContext(ctx))

			Expect(recorder.Header().Get("Content-Type")).To(Equal("application/json"))
			Expect(recorder.Code).To(Equal(http.StatusNotFound))
			bodyJSON := make(map[string]string)
			json.Unmarshal(recorder.Body.Bytes(), &bodyJSON)
			Expect(bodyJSON["description"]).To(Equal("namespace for user 'derp' not found"))
			Expect(bodyJSON["error"]).To(Equal("delete cluster error"))
			Expect(bodyJSON["code"]).To(Equal("OFF-004"))
		})

		It("should return status 401 when complete route without access token", func() {
			deleteRoute := "/clusters/myCustomApps/delete"
			request, _ = http.NewRequest("DELETE", deleteRoute, nil)
			request.Header.Add("Authorization", "Bearer invalid-token")
			app.Router.ServeHTTP(recorder, request)
			bodyJSON := make(map[string]string)
			json.Unmarshal(recorder.Body.Bytes(), &bodyJSON)
			Expect(bodyJSON["code"]).To(Equal("OFF-002"))
			Expect(bodyJSON["description"]).To(Equal("sql: no rows in result set"))
			Expect(bodyJSON["error"]).To(Equal("Access Token not found (have you logged in?)"))
			Expect(recorder.Code).To(Equal(http.StatusUnauthorized))
		})
	})
})
