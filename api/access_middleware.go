// https://github.com/topfreegames/mystack-controller
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2017 Top Free Games <backend@tfgco.com>

package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/topfreegames/mystack-controller/errors"
	"io/ioutil"
	"net/http"
	"strings"
)

//AccessMiddleware guarantees that the user is logged
type AccessMiddleware struct {
	App  *App
	next http.Handler
}

const emailKey = contextKey("emailKey")

func newContextWithEmail(ctx context.Context, email string) context.Context {
	c := context.WithValue(ctx, emailKey, email)
	return c
}

func emailFromCtx(ctx context.Context) string {
	email := ctx.Value(emailKey)
	if email == nil {
		return ""
	}
	return email.(string)
}

//ServeHTTP methods
func (m *AccessMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logger := loggerFromContext(r.Context())
	token := r.Header.Get("Authorization")
	token = strings.TrimPrefix(token, "Bearer ")
	url := fmt.Sprintf("https://www.googleapis.com/oauth2/v1/tokeninfo?access_token=%s", token)
	resp, err := http.Get(url)

	if err != nil {
		logger.WithError(err).Error("Error fetching googleapis")
		m.App.HandleError(w, http.StatusInternalServerError, "Error fetching googleapis", err)
		return
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.WithError(err).Error("Error parsing response")
		m.App.HandleError(w, http.StatusInternalServerError, "Error parsing response", err)
		return
	}

	if resp.StatusCode == http.StatusBadRequest {
		logger.WithError(err).Error("Error validating access token")
		err := errors.NewAccessError("Unauthorized access token", fmt.Errorf(string(body)))
		m.App.HandleError(w, http.StatusUnauthorized, "Unauthorized access token", err)
		return
	}

	if resp.StatusCode != 200 {
		logger.WithError(err).Error("Invalid access token")
		err := errors.NewAccessError("Invalid access token", fmt.Errorf(string(body)))
		m.App.HandleError(w, resp.StatusCode, "Error validating access token", err)
		return
	}

	var bodyObj map[string]interface{}
	json.Unmarshal(body, &bodyObj)
	email := bodyObj["email"].(string)
	if !m.verifyEmailDomain(email) {
		logger.WithError(err).Error("Invalid email")
		err := errors.NewAccessError(
			fmt.Sprintf("The email on OAuth authorization is not from domain %s", m.App.EmailDomain),
			fmt.Errorf("Invalid email"),
		)
		m.App.HandleError(w, http.StatusUnauthorized, "Error validating access token", err)
		return
	}
	ctx := newContextWithEmail(r.Context(), email)
	m.next.ServeHTTP(w, r.WithContext(ctx))
}

func (m *AccessMiddleware) verifyEmailDomain(email string) bool {
	for _, domain := range m.App.EmailDomain {
		if strings.HasSuffix(email, domain) {
			return true
		}
	}
	return false
}

//SetNext handler
func (m *AccessMiddleware) SetNext(next http.Handler) {
	m.next = next
}
