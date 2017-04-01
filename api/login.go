// mystack-controller api
// https://github.com/topfreegames/mystack-controller
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2017 Top Free Games <backend@tfgco.com>

package api

import (
	"fmt"
	"github.com/topfreegames/mystack-controller/extensions"
	"net/http"
)

//LoginHandler handles login url requests
type LoginHandler struct {
	App    *App
	Method string
}

//ServeHTTP method
func (l *LoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch l.Method {
	case "login":
		l.generateURL(w, r)
	case "access":
		l.exchangeAccess(w, r)
	}
}

func (l *LoginHandler) generateURL(w http.ResponseWriter, r *http.Request) {
	logger := loggerFromContext(r.Context())
	oauthState := r.FormValue("state")
	if len(oauthState) == 0 {
		l.App.HandleError(w, http.StatusBadRequest, "state must not be empty", fmt.Errorf("state must not be empty"))
		return
	}

	url, err := extensions.GenerateLoginURL(oauthState)
	if err != nil {
		logger.WithError(err).Errorln("undefined env vars")
		l.App.HandleError(w, http.StatusInternalServerError, "undefined env vars", err)
		return
	}

	body := fmt.Sprintf(`{"url": "%s"}`, url)

	Write(w, http.StatusOK, body)
}

func (l *LoginHandler) exchangeAccess(w http.ResponseWriter, r *http.Request) {
	authCode := r.FormValue("code")
	if len(authCode) == 0 {
		l.App.HandleError(w, http.StatusBadRequest, "state must not be empty", fmt.Errorf("state must not be empty"))
		return
	}

	token, err := extensions.GetAccessToken(authCode)
	if err != nil {
		l.App.HandleError(w, http.StatusBadRequest, "failed to get access token", fmt.Errorf("failed to get access token"))
		return
	}

	body := fmt.Sprintf(`{"token": "%s"}`, token)

	Write(w, http.StatusOK, body)
}
