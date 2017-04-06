// mystack-controller api
// https://github.com/topfreegames/mystack-controller
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2017 Top Free Games <backend@tfgco.com>

package models

import "os"

//OSCredentials implements Credentials interface
type OSCredentials struct{}

//GetID gets ID from environmental variable
func (o *OSCredentials) GetID() string {
	return os.Getenv(ClientIDEnvVar)
}

//GetSecret gets secret from environmental variable
func (o *OSCredentials) GetSecret() string {
	return os.Getenv(ClientSecretEnvVar)
}
