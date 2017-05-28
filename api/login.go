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
	log(logger, "Generating log in URL")

	oauthState := r.FormValue("state")
	if len(oauthState) == 0 {
		l.App.HandleError(w, http.StatusBadRequest, "state must not be empty", fmt.Errorf("state must not be empty"))
		return
	}

	url, err := extensions.GenerateLoginURL(oauthState, &models.OSCredentials{})
	if err != nil {
		logger.WithError(err).Errorln("undefined env vars")
		l.App.HandleError(w, http.StatusInternalServerError, "undefined env vars", err)
		return
	}

	bodyResponse := map[string]string{
		"url":            url,
		"controllerHost": fmt.Sprintf("controller.%s", l.App.K8sDomain),
		"loggerHost":     fmt.Sprintf("logger.%s", l.App.K8sDomain),
	}
	bts, err := json.Marshal(bodyResponse)
	if err != nil {
		logger.WithError(err).Errorln("error parsing map")
		l.App.HandleError(w, http.StatusInternalServerError, "error parsing map", err)
		return
	}

	WriteBytes(w, http.StatusOK, bts)
	log(logger, "Login URL generated")
}

func (l *LoginHandler) exchangeAccess(w http.ResponseWriter, r *http.Request) {
	logger := loggerFromContext(r.Context())
	log(logger, "Getting access token")

	authCode := r.FormValue("code")
	if len(authCode) == 0 {
		l.App.HandleError(w, http.StatusBadRequest, "code must not be empty", fmt.Errorf("state must not be empty"))
		return
	}

	token, err := extensions.GetAccessToken(authCode)
	if err != nil {
		l.App.HandleError(w, http.StatusBadRequest, "failed to get access token", fmt.Errorf("failed to get access token"))
		return
	}

	//If the last error didn't occur, then the error from Authenticate method won't happen
	email, _, _ := extensions.Authenticate(token, &models.OSCredentials{}, l.App.DB)
	if !l.App.verifyEmailDomain(email) {
		logger.WithError(err).Error("Invalid email")
		err := errors.NewAccessError(
			"authorization access error",
			fmt.Errorf("the email on OAuth authorization is not from domain %s", l.App.EmailDomain),
		)
		l.App.HandleError(w, http.StatusUnauthorized, "error validating access token", err)
		return
	}

	err = extensions.SaveToken(token, email, token.AccessToken, l.App.DB)
	if err != nil {
		l.App.HandleError(w, http.StatusBadRequest, "", err)
		return
	}

	body := fmt.Sprintf(`{"token": "%s"}`, token.AccessToken)

	Write(w, http.StatusOK, body)
	log(logger, "Returning access token")
}
