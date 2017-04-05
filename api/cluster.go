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
	case "create":
		c.create(w, r)
	case "remove":
		c.remove(w, r)
	}
}

func (c *ClusterHandler) create(w http.ResponseWriter, r *http.Request) {
	email := emailFromCtx(r.Context())
	username := strings.Split(email, "@")[0]
	clusterName := mux.Vars(r)["name"]

	apps, services, err := models.LoadClusterConfig(c.App.DB, clusterName)
	if err != nil {
		c.App.HandleError(w, http.StatusInternalServerError, "Error creating cluster config file", err)
		return
	}

	deployments := make([]*models.Deployment, len(apps)+len(services))

	i := 0
	for name, appConfig := range apps {
		deployments[i] = models.NewDeployment(
			name,
			username,
			appConfig.Image,
			appConfig.Port,
			appConfig.Environment,
		)
		i = i + 1
	}

	for name, serviceConfig := range services {
		deployments[i] = models.NewDeployment(
			name,
			username,
			serviceConfig.Image,
			serviceConfig.Port,
			serviceConfig.Environment,
		)
		i = i + 1
	}

	cluster := models.NewCluster(username, deployments)
	err = cluster.Create(c.App.Clientset)
	if err != nil {
		c.App.HandleError(w, http.StatusInternalServerError, "Error creating cluster", err)
		return
	}

	Write(w, http.StatusOK, `{"success": "true"}`)
}

func (c *ClusterHandler) remove(w http.ResponseWriter, r *http.Request) {
}
