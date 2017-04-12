// mystack-controller api
// https://github.com/topfreegames/mystack-controller
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2017 Top Free Games <backend@tfgco.com>

package api

import (
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
	case "create":
		c.create(w, r)
	case "delete":
		c.deleteCluster(w, r)
	}
}

func (c *ClusterHandler) create(w http.ResponseWriter, r *http.Request) {
	email := emailFromCtx(r.Context())
	username := usernameFromEmail(email)
	clusterName := GetClusterName(r)

	cluster, err := models.NewCluster(c.App.DB, username, clusterName)
	if err != nil {
		c.App.HandleError(w, Status(err), "create cluster error", err)
		return
	}

	err = cluster.Create(c.App.Clientset)
	if err != nil {
		c.App.HandleError(w, Status(err), "create cluster error", err)
		return
	}

	Write(w, http.StatusOK, `{"status": "ok"}`)
}

func (c *ClusterHandler) deleteCluster(w http.ResponseWriter, r *http.Request) {
	email := emailFromCtx(r.Context())
	username := usernameFromEmail(email)
	clusterName := GetClusterName(r)

	cluster, err := models.NewCluster(c.App.DB, username, clusterName)
	if err != nil && strings.Contains(err.Error(), "no rows in result set") {
		cluster = &models.Cluster{Username: username}
	} else if err != nil {
		c.App.HandleError(w, Status(err), "retrieve cluster error", err)
		return
	}

	err = cluster.Delete(c.App.Clientset)
	if err != nil {
		c.App.HandleError(w, Status(err), "delete cluster error", err)
		return
	}

	Write(w, http.StatusOK, `{"status": "ok"}`)
}
