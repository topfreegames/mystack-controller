// mystack-controller api
// https://github.com/topfreegames/mystack-controller
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2017 Top Free Games <backend@tfgco.com>

package api

import (
	"net/http"
)

//DeploymentHandler handles deployments on k8s
type DeploymentHandler struct {
	App    *App
	Method string
}

//ServeHTTP method
func (d *DeploymentHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch d.Method {
	case "create":
		d.create(w, r)
	case "delete":
		d.deleteDeploy(w, r)
	}
}

func (d *DeploymentHandler) create(w http.ResponseWriter, r *http.Request) {
}

func (d *DeploymentHandler) deleteDeploy(w http.ResponseWriter, r *http.Request) {
}
