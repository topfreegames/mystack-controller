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
	"net/http"
	"net/http/httptest"

	. "github.com/topfreegames/mystack-controller/api"
	sqlmock "gopkg.in/DATA-DOG/go-sqlmock.v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DNS", func() {
	var recorder *httptest.ResponseRecorder
	var dnsHandler *DNSHandler
	var clusterConfigHandler *ClusterConfigHandler
	var yaml1 = `
apps:
  test1:
    image: app1
    port: 5000
    customDomains:
      - app1.example.com
      - app1.another.com
      - app1.org
`

	BeforeEach(func() {
		recorder = httptest.NewRecorder()
		dnsHandler = &DNSHandler{App: app}
		clusterConfigHandler = &ClusterConfigHandler{App: app, Method: "create"}
	})

	Describe("GET /cluster-configs/{name}/domains", func() {
		var (
			request     *http.Request
			err         error
			clusterName = "myCustomApps"
			route       = fmt.Sprintf("/cluster-configs/%s/domains", clusterName)
		)

		BeforeEach(func() {
			request, err = http.NewRequest("GET", route, nil)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			err = mock.ExpectationsWereMet()
			Expect(err).NotTo(HaveOccurred())
		})

		It("should return custom domains", func() {
			request, err = http.NewRequest("PUT", route, nil)
			Expect(err).NotTo(HaveOccurred())

			ctx := NewContextWithClusterConfig(request.Context(), yaml1)
			clusterConfigHandler.Method = "create"
			clusterConfigHandler.ServeHTTP(recorder, request.WithContext(ctx))

			recorder = httptest.NewRecorder()

			mock.
				ExpectQuery("^SELECT yaml FROM clusters WHERE name(.+)$").
				WillReturnRows(sqlmock.NewRows([]string{"yaml"}).AddRow(yaml1))

			dnsHandler.ServeHTTP(recorder, request)

			Expect(recorder.Header().Get("Content-Type")).To(Equal("application/json"))
			Expect(recorder.Code).To(Equal(http.StatusOK))

			bodyJSON := make(map[string][]string)
			json.Unmarshal(recorder.Body.Bytes(), &bodyJSON)
			Expect(bodyJSON["test1"]).To(ConsistOf("app1.example.com", "app1.another.com", "app1.org"))
		})

		It("should return 404 if cluster not found", func() {
			mock.
				ExpectQuery("^SELECT yaml FROM clusters WHERE name(.+)$").
				WillReturnError(fmt.Errorf("sql: no rows in result set"))

			dnsHandler.ServeHTTP(recorder, request)

			Expect(recorder.Header().Get("Content-Type")).To(Equal("application/json"))
			Expect(recorder.Code).To(Equal(http.StatusNotFound))

			bodyJSON := make(map[string]string)
			json.Unmarshal(recorder.Body.Bytes(), &bodyJSON)
			Expect(bodyJSON["description"]).To(Equal("sql: no rows in result set"))
			Expect(bodyJSON["error"]).To(Equal("database error"))
			Expect(bodyJSON["code"]).To(Equal("OFF-003"))
		})

		It("should return 400 if config was invalid", func() {
			mock.
				ExpectQuery("^SELECT yaml FROM clusters WHERE name(.+)$").
				WillReturnRows(sqlmock.NewRows([]string{"yaml"}).AddRow("i am invalid"))

			dnsHandler.ServeHTTP(recorder, request)

			Expect(recorder.Header().Get("Content-Type")).To(Equal("application/json"))
			Expect(recorder.Code).To(Equal(http.StatusBadRequest))

			bodyJSON := make(map[string]string)
			json.Unmarshal(recorder.Body.Bytes(), &bodyJSON)
			Expect(bodyJSON["description"]).To(Equal("yaml: unmarshal errors:\n  line 1: cannot unmarshal !!str `i am in...` into models.ClusterConfig"))
			Expect(bodyJSON["error"]).To(Equal("parse yaml error"))
			Expect(bodyJSON["code"]).To(Equal("OFF-004"))
		})
	})
})
