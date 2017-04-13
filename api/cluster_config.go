// mystack-controller api
// https://github.com/topfreegames/mystack-controller
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2017 Top Free Games <backend@tfgco.com>

package api

import (
	"encoding/json"
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
		break
	case "remove":
		c.remove(w, r)
		break
	case "list":
		c.list(w, r)
		break
	case "info":
		c.info(w, r)
		break
	}
}

func (c *ClusterConfigHandler) create(w http.ResponseWriter, r *http.Request) {
	logger := loggerFromContext(r.Context())
	clusterName := GetClusterName(r)

	log(logger, "Creating cluster config '%s'", clusterName)
	clusterConfig := clusterConfigFromCtx(r.Context())

	err := models.WriteClusterConfig(c.App.DB, clusterName, clusterConfig)
	if err != nil {
		c.App.HandleError(w, Status(err), "writing cluster config error", err)
		return
	}

	Write(w, http.StatusOK, `{"status": "ok"}`)
	log(logger, "Creating cluster config '%s' successfully created", clusterName)
}

func (c *ClusterConfigHandler) remove(w http.ResponseWriter, r *http.Request) {
	logger := loggerFromContext(r.Context())
	clusterName := GetClusterName(r)

	log(logger, "Deleting cluster config '%s'", clusterName)
	err := models.RemoveClusterConfig(c.App.DB, clusterName)
	if err != nil {
		c.App.HandleError(w, Status(err), "removing cluster config error", err)
		return
	}

	Write(w, http.StatusOK, `{"status": "ok"}`)
	log(logger, "Cluster config '%s' successfully deleted", clusterName)
}

func (c *ClusterConfigHandler) list(w http.ResponseWriter, r *http.Request) {
	logger := loggerFromContext(r.Context())

	log(logger, "Getting list of cluster configs")
	names, err := models.ListClusterConfig(c.App.DB)
	if err != nil {
		c.App.HandleError(w, Status(err), "listing cluster configs error", err)
		return
	}

	response := map[string][]string{
		"names": names,
	}
	bts, err := json.Marshal(response)
	if err != nil {
		c.App.HandleError(w, Status(err), "listing cluster configs error", err)
		return
	}
	WriteBytes(w, http.StatusOK, bts)
	log(logger, "Successfully listed Cluster configs")
}

func (c *ClusterConfigHandler) info(w http.ResponseWriter, r *http.Request) {
	logger := loggerFromContext(r.Context())
	clusterName := GetClusterName(r)

	log(logger, "Getting yaml of cluster config '%s'", clusterName)
	yamlStr, err := models.ClusterConfigDetails(c.App.DB, clusterName)
	if err != nil {
		c.App.HandleError(w, Status(err), "cluster configs details error", err)
		return
	}

	response := map[string]string{
		"yaml": yamlStr,
	}
	bts, err := json.Marshal(response)
	if err != nil {
		c.App.HandleError(w, Status(err), "listing cluster configs error", err)
		return
	}
	WriteBytes(w, http.StatusOK, bts)
	log(logger, "Successfully got cluster config")
}
