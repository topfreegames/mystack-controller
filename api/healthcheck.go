// mystack-controller api
// https://github.com/topfreegames/mystack-controller
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2017 Top Free Games <backend@tfgco.com>

package api

import (
	"net/http"

	"github.com/topfreegames/mystack-controller/errors"
)

//HealthcheckHandler handler
type HealthcheckHandler struct {
	App *App
}

//ServeHTTP method
func (h *HealthcheckHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	l := loggerFromContext(r.Context())

	l.Debug("Performing healthcheck...")

	_, err := h.App.DB.Exec("select 1")

	if err != nil {
		l.WithError(err).Error("Database is offline")
		vErr := errors.NewDatabaseError(err)
		WriteBytes(w, http.StatusInternalServerError, vErr.Serialize())
		return
	}

	Write(w, http.StatusOK, `{"healthy": true}`)
	l.Debug("Healthcheck done.")
}
