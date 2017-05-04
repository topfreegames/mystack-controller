// Package api mystack-controller
// +build integration
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
	"github.com/topfreegames/mystack-controller/metadata"
	oTesting "github.com/topfreegames/mystack-controller/testing"

	"net/http"
	"net/http/httptest"
)

var _ = Describe("Healthcheck", func() {
	var request *http.Request
	var recorder *httptest.ResponseRecorder

	BeforeEach(func() {
		recorder = httptest.NewRecorder()
	})

	Describe("GET /healthcheck", func() {
		BeforeEach(func() {
			request, _ = http.NewRequest("GET", "/healthcheck", nil)
		})

		Context("when all services healthy", func() {
			It("returns a status code of 200", func() {
				app.Router.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(200))
			})

			It("returns working string", func() {
				app.Router.ServeHTTP(recorder, request)
				Expect(recorder.Body.String()).To(Equal(`{"healthy": true}`))
			})

			It("returns the version as a header", func() {
				app.Router.ServeHTTP(recorder, request)
				Expect(recorder.Header().Get("x-mystack-controller-version")).To(Equal(metadata.Version))
			})
		})

		It("should return 500 if DB is offline", func() {
			dbHolder := app.DB

			conn, err := oTesting.GetTestDB()
			Expect(err).NotTo(HaveOccurred())

			conn.Close()
			app.DB = conn

			app.Router.ServeHTTP(recorder, request)
			Expect(recorder.Code).To(Equal(500))
			bodyJSON := make(map[string]string)
			json.Unmarshal(recorder.Body.Bytes(), &bodyJSON)
			Expect(bodyJSON["code"]).To(Equal("OFF-003"))
			Expect(bodyJSON["description"]).To(Equal("sql: database is closed"))
			Expect(bodyJSON["error"]).To(Equal("database error"))

			app.DB = dbHolder
		})
	})
})
