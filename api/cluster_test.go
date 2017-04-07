package api_test

import (
	"encoding/json"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/topfreegames/mystack-controller/api"
	"github.com/topfreegames/mystack-controller/models"

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

	Describe("PUT /clusters/{name}/run", func() {

		var (
			err     error
			request *http.Request
			route   = fmt.Sprintf("/cluster/%s/run", clusterName)
		)

		BeforeEach(func() {
			clusterHandler.Method = "run"
			request, err = http.NewRequest("PUT", route, nil)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			err = mock.ExpectationsWereMet()
			Expect(err).NotTo(HaveOccurred())
		})

		It("should run existing clusterName", func() {
			mock.
				ExpectQuery("^SELECT yaml FROM clusters WHERE name = (.+)$").
				WithArgs(clusterName).
				WillReturnRows(sqlmock.NewRows([]string{"yaml"}).AddRow(yaml1))

			ctx := NewContextWithEmail(request.Context(), "derp@example.com")
			clusterHandler.ServeHTTP(recorder, request.WithContext(ctx))

			Expect(recorder.Header().Get("Content-Type")).To(Equal("application/json"))
			Expect(recorder.Body.String()).To(Equal(`{"status": "ok"}`))
			Expect(recorder.Code).To(Equal(http.StatusOK))
		})

		It("should return error 422 when run non existing clusterName", func() {
			mock.
				ExpectQuery("^SELECT yaml FROM clusters WHERE name = (.+)$").
				WithArgs(clusterName).
				WillReturnError(fmt.Errorf("sql: no rows in result set"))

			ctx := NewContextWithEmail(request.Context(), "derp@example.com")
			clusterHandler.ServeHTTP(recorder, request.WithContext(ctx))

			Expect(recorder.Header().Get("Content-Type")).To(Equal("application/json"))
			//TODO: change this to 422 (THIS IS URGENT)
			Expect(recorder.Code).To(Equal(http.StatusInternalServerError))
			bodyJSON := make(map[string]string)
			json.Unmarshal(recorder.Body.Bytes(), &bodyJSON)
			Expect(bodyJSON["code"]).To(Equal("OFF-001"))
			Expect(bodyJSON["description"]).To(Equal("sql: no rows in result set"))
			Expect(bodyJSON["error"]).To(Equal("Error creating cluster"))
		})
	})

	Describe("PUT /clusters/{name}/delete", func() {

		var (
			err     error
			request *http.Request
			route   = fmt.Sprintf("/cluster/%s/delete", clusterName)
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

			cluster, err := models.NewCluster(app.DB, "user", clusterName)
			Expect(err).NotTo(HaveOccurred())
			err = cluster.Create(app.Clientset)
			Expect(err).NotTo(HaveOccurred())

			ctx := NewContextWithEmail(request.Context(), "user@example.com")
			clusterHandler.ServeHTTP(recorder, request.WithContext(ctx))

			Expect(recorder.Header().Get("Content-Type")).To(Equal("application/json"))
			Expect(recorder.Body.String()).To(Equal(`{"status": "ok"}`))
			Expect(recorder.Code).To(Equal(http.StatusOK))
		})

		It("should return error 422 when deleting non existing clusterName", func() {
			mock.
				ExpectQuery("^SELECT yaml FROM clusters WHERE name = (.+)$").
				WithArgs(clusterName).
				WillReturnError(fmt.Errorf("sql: no rows in result set"))

			ctx := NewContextWithEmail(request.Context(), "derp@example.com")
			clusterHandler.ServeHTTP(recorder, request.WithContext(ctx))

			Expect(recorder.Header().Get("Content-Type")).To(Equal("application/json"))
			//TODO: change this to 422 (THIS IS URGENT)
			Expect(recorder.Code).To(Equal(http.StatusInternalServerError))
			bodyJSON := make(map[string]string)
			json.Unmarshal(recorder.Body.Bytes(), &bodyJSON)
			Expect(bodyJSON["code"]).To(Equal("OFF-001"))
			Expect(bodyJSON["description"]).To(Equal("sql: no rows in result set"))
			Expect(bodyJSON["error"]).To(Equal("Error retrieving cluster"))
		})
	})
})
