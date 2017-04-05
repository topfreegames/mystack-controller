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
	case "remove":
		c.remove(w, r)
	}
}

func (c *ClusterHandler) create(w http.ResponseWriter, r *http.Request) {
	email := emailFromCtx(r.Context())
	username := strings.Split(email, "@")[0]
	clusterConfig := clusterConfigFromCtx(r.Context())
	deployments := make([]*models.Deployment, len(clusterConfig))

	i := 0
	for name, appConfig := range clusterConfig {
		deployments[i] = models.NewDeployment(
			name,
			username,
			appConfig.Image,
			appConfig.Port,
		)
		i = i + 1
	}
}

func (c *ClusterHandler) remove(w http.ResponseWriter, r *http.Request) {
}
