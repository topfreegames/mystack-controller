// mystack-controller api
// https://github.com/topfreegames/mystack-controller
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2017 Top Free Games <backend@tfgco.com>

package api

import (
	"github.com/gorilla/mux"
	"github.com/topfreegames/mystack-controller/errors"
	"net/http"
	"strings"
)

//Write to the response and with the status code
func Write(w http.ResponseWriter, status int, text string) {
	WriteBytes(w, status, []byte(text))
}

//WriteBytes to the response and with the status code
func WriteBytes(w http.ResponseWriter, status int, text []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(text)
}

//Status return the HTTP status code of an error
func Status(err error) int {
	if err == nil {
		return http.StatusOK
	}

	switch err.(type) {
	case *errors.DatabaseError:
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			return http.StatusConflict
		} else if strings.Contains(err.Error(), "no rows in result set") {
			return http.StatusNotFound
		}
	case *errors.YamlError:
		if strings.Contains(err.Error(), "empty") {
			return http.StatusUnprocessableEntity
		}
		return http.StatusBadRequest
	case *errors.GenericError:
		return http.StatusUnprocessableEntity
	case *errors.KubernetesError:
		if strings.Contains(err.Error(), "not found") {
			return http.StatusNotFound
		}
	}

	return http.StatusInternalServerError
}

//GetClusterName gets the name from URL from request
func GetClusterName(r *http.Request) string {
	clusterName := mux.Vars(r)["name"]

	if len(clusterName) == 0 {
		parts := strings.Split(r.URL.String(), "/")
		clusterName = parts[2]
	}

	return clusterName
}

func usernameFromEmail(email string) string {
	username := strings.Split(email, "@")[0]
	username = strings.Replace(username, ".", "-", -1)
	return username
}
