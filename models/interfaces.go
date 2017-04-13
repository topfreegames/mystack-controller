// mystack-controller api
// https://github.com/topfreegames/mystack-controller
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2017 Top Free Games <backend@tfgco.com>

package models

import (
	"database/sql"
	"k8s.io/client-go/kubernetes"
)

const (
	//ClientIDEnvVar defines the name of the environment variable
	ClientIDEnvVar = "MYSTACK_GOOGLE_CLIENT_ID"
	//ClientSecretEnvVar defines the name of the environment variable
	ClientSecretEnvVar = "MYSTACK_GOOGLE_CLIENT_SECRET"
)

//Credentials is an interface with Get method to get ClientID and ClientSecret
type Credentials interface {
	GetID() string
	GetSecret() string
}

//DB is the mystack-controller db interface
type DB interface {
	NamedExec(query string, arg interface{}) (sql.Result, error)
	Get(dest interface{}, query string, args ...interface{}) error
}

//Readiness is the interface that tell how much time to wait until
//the deployment is ready
//and its readiness probe reports Ready
type Readiness interface {
	WaitForCompletion(kubernetes.Interface, interface{}) error
}
