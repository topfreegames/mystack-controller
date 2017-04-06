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

	"fmt"
	mTest "github.com/topfreegames/mystack-controller/testing"
	"net/http"
	"net/http/httptest"
)

var _ = Describe("ClusterConfig", func() {
	var recorder *httptest.ResponseRecorder

	BeforeEach(func() {
		recorder = httptest.NewRecorder()
	})

	Describe("PUT /cluster-configs/{name}/create", func() {
		var (
			request     *http.Request
			err         error
			clusterName = "myCustomApps"
			route       = fmt.Sprintf("/cluster-configs/%s/create", clusterName)
			yaml1       = `
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
			yamlReader = mTest.JSONFor(map[string]interface{}{
				"yaml": yaml1,
			})
		)

		BeforeEach(func() {
			request, err = http.NewRequest("PUT", route, yamlReader)
			Expect(err).NotTo(HaveOccurred())
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
			request, err = http.NewRequest("PUT", route, nil)
			Expect(err).NotTo(HaveOccurred())

			request.Header.Add("Authorization", "Bearer invalid-token")
		})

		It("should return status 401", func() {
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
