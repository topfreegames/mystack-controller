// mystack-controller api
// https://github.com/topfreegames/mystack-controller
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2017 Top Free Games <backend@tfgco.com>

package api

import (
	"github.com/gorilla/mux"
	"github.com/topfreegames/mystack-controller/models"
	"net/http"
	"strings"
)

//ClusterHandler handles cluster creation and deletion
type ClusterHandler struct {
	App    *App
	Method string
}

func (c *ClusterHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch c.Method {
	case "run":
		c.run(w, r)
	case "delete":
		c.deleteCluster(w, r)
	}
}

func (c *ClusterHandler) run(w http.ResponseWriter, r *http.Request) {
	email := emailFromCtx(r.Context())
	username := strings.Split(email, "@")[0]
	clusterName := mux.Vars(r)["name"]

	if len(clusterName) == 0 {
		parts := strings.Split(r.URL.String(), "/")
		clusterName = parts[2]
	}

	cluster, err := models.NewCluster(c.App.DB, username, clusterName)
	if err != nil {
		c.App.HandleError(w, http.StatusInternalServerError, "Error creating cluster", err)
		return
	}

	err = cluster.Create(c.App.Clientset)
	if err != nil {
		c.App.HandleError(w, http.StatusInternalServerError, "Error creating cluster", err)
		return
	}

	Write(w, http.StatusOK, `{"status": "ok"}`)
}

func (c *ClusterHandler) deleteCluster(w http.ResponseWriter, r *http.Request) {
	email := emailFromCtx(r.Context())
	username := strings.Split(email, "@")[0]
	clusterName := mux.Vars(r)["name"]

	if len(clusterName) == 0 {
		parts := strings.Split(r.URL.String(), "/")
		clusterName = parts[2]
	}

	cluster, err := models.NewCluster(c.App.DB, username, clusterName)
	if err != nil {
		c.App.HandleError(w, http.StatusInternalServerError, "Error retrieving cluster", err)
		return
	}

	err = cluster.Delete(c.App.Clientset)
	if err != nil {
		c.App.HandleError(w, http.StatusInternalServerError, "Error deleting cluster", err)
		return
	}

	Write(w, http.StatusOK, `{"status": "ok"}`)
}
