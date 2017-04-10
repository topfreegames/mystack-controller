// mystack-controller api
// +build unit
// https://github.com/topfreegames/mystack-controller
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2017 Top Free Games <backend@tfgco.com>

package api_test

import (
	. "github.com/topfreegames/mystack-controller/api"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	mTest "github.com/topfreegames/mystack-controller/testing"
	"net/http"
	"net/http/httptest"
)

type DummyHandler struct{}

func (d *DummyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {}

var _ = Describe("PayloadMiddleware", func() {
	const validYaml = `
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
`
	const validYamlWithTabs = `
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
`
	const yamlFromFile = "services:\n  test0:\n    image: svc1\n    ports:\n        - \"5000\""

	It("should decode valid yaml", func() {
		payloadMiddleware := &PayloadMiddleware{}
		request, err := http.NewRequest("PUT", "", mTest.JSONFor(map[string]interface{}{
			"yaml": validYaml,
		}))
		Expect(err).NotTo(HaveOccurred())

		recorder := httptest.NewRecorder()

		payloadMiddleware.SetNext(&DummyHandler{})
		payloadMiddleware.ServeHTTP(recorder, request)

		Expect(recorder.Code).To(Equal(http.StatusOK))
	})

	It("should decode valid yaml from file", func() {
		payloadMiddleware := &PayloadMiddleware{}
		request, err := http.NewRequest("PUT", "", mTest.JSONFor(map[string]interface{}{
			"yaml": yamlFromFile,
		}))
		Expect(err).NotTo(HaveOccurred())

		recorder := httptest.NewRecorder()

		payloadMiddleware.SetNext(&DummyHandler{})
		payloadMiddleware.ServeHTTP(recorder, request)

		Expect(recorder.Code).To(Equal(http.StatusOK))
	})

	It("should decode valid yaml with tabs", func() {
		payloadMiddleware := &PayloadMiddleware{}
		request, err := http.NewRequest("PUT", "", mTest.JSONFor(map[string]interface{}{
			"yaml": yamlFromFile,
		}))
		Expect(err).NotTo(HaveOccurred())

		recorder := httptest.NewRecorder()

		payloadMiddleware.SetNext(&DummyHandler{})
		payloadMiddleware.ServeHTTP(recorder, request)

		Expect(recorder.Code).To(Equal(http.StatusOK))
	})

	It("should return error if valid inyaml", func() {
		payloadMiddleware := &PayloadMiddleware{}
		request, err := http.NewRequest("PUT", "", mTest.JSONFor(map[string]interface{}{
			"yaml": validYamlWithTabs,
		}))
		Expect(err).NotTo(HaveOccurred())

		recorder := httptest.NewRecorder()

		payloadMiddleware.SetNext(&DummyHandler{})
		payloadMiddleware.ServeHTTP(recorder, request)

		Expect(recorder.Code).To(Equal(http.StatusOK))
	})
})
