// mystack-controller api
// https://github.com/topfreegames/mystack-controller
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2017 Top Free Games <backend@tfgco.com>

package models

import (
	"database/sql"
)

const (
	//ClientIDEnvVar defines the name of the environmental variable
	ClientIDEnvVar = "MYSTACK_GOOGLE_CLIENT_ID"
	//ClientSecretEnvVar defines the name of the environmental variable
	ClientSecretEnvVar = "MYSTACK_GOOGLE_CLIENT_SECRET"
)

//Credentials is an interface with Get method to get ClientID and ClientSecret
type Credentials interface {
	GetID() string
	GetSecret() string
}

//DB is the mystack-controller db interface
type DB interface {
	MustExec(query string, args ...interface{}) sql.Result
	Get(dest interface{}, query string, args ...interface{}) error
}
