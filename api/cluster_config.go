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
)

//ClusterConfigHandler handles cluster creation and deletion
type ClusterConfigHandler struct {
	App    *App
	Method string
}

func (c *ClusterConfigHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch c.Method {
	case "create":
		c.create(w, r)
	case "remove":
		c.remove(w, r)
	}
}

func (c *ClusterConfigHandler) create(w http.ResponseWriter, r *http.Request) {
	clusterName := mux.Vars(r)["name"]
	clusterConfig := clusterConfigFromCtx(r.Context())

	err := models.WriteClusterConfig(c.App.DB, clusterName, clusterConfig)
	if err != nil {
		c.App.HandleError(w, http.StatusInternalServerError, "Error writing cluster config", err)
		return
	}

	Write(w, http.StatusOK, `{"status": "ok"}`)
}

func (c *ClusterConfigHandler) remove(w http.ResponseWriter, r *http.Request) {
	clusterName := mux.Vars(r)["name"]

	err := models.RemoveClusterConfig(c.App.DB, clusterName)
	if err != nil {
		c.App.HandleError(w, http.StatusInternalServerError, "Error removing cluster config", err)
		return
	}

	Write(w, http.StatusOK, `{"status": "ok"}`)
}
