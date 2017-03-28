// mystack-controller api
// https://github.com/topfreegames/mystack
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2017 Top Free Games <backend@tfgco.com>

package api

import (
	"fmt"
	"github.com/topfreegames/mystack/mystack-controller/extensions"
	"net/http"
)

//LoginHandler handles login url requests
type LoginHandler struct {
	App *App
}

//ServeHTTP method
func (l *LoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	oauthState := r.FormValue("state")
	if len(oauthState) == 0 {
		l.App.HandleError(w, http.StatusBadRequest, "state must not be empty", fmt.Errorf("state must not be empty"))
		return
	}

	url := extensions.GenerateLoginURL(oauthState)
	body := fmt.Sprintf(`{"url": %s}`, url)

	Write(w, http.StatusOK, body)
}
