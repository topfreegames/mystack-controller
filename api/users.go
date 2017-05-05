// mystack-controller api
// https://github.com/topfreegames/mystack-controller
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2017 Top Free Games <backend@tfgco.com>

package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/topfreegames/mystack-controller/extensions"
	"github.com/topfreegames/mystack-controller/models"
	"github.com/topfreegames/mystack-logger/errors"
)

//UserHandler handles login url requests
type UserHandler struct {
	App *App
}

//ServeHTTP method
func (u *UserHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	u.App.Logger.Info("getting email from access token")
	accessToken := r.FormValue("token")

	token, err := extensions.Token(accessToken, u.App.DB)
	if err != nil {
		u.App.HandleError(w, Status(err), "user access error", err)
		return
	}

	msg, status, err := extensions.Authenticate(token, &models.OSCredentials{})
	if err != nil {
		u.App.HandleError(w, Status(err), "user access error", err)
		return
	}

	if status != 200 {
		u.App.Logger.WithError(err).Error("error validating access token")
		err := errors.NewAccessError("Unauthorized access token", fmt.Errorf(msg))
		u.App.HandleError(w, http.StatusUnauthorized, "Unauthorized access token", err)
		return
	}

	response := map[string]string{
		"email": msg,
	}
	bts, _ := json.Marshal(response)

	WriteBytes(w, status, bts)
	u.App.Logger.Info("successfully got email from access token")
}
