// mystack-controller api
// https://github.com/topfreegames/mystack-controller
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2017 Top Free Games <backend@tfgco.com>

package models

//MockCredentials implements Credentials interface
type MockCredentials struct {
	ID  string
	Key string
}

//GetID gets ID from environmental variable
func (m *MockCredentials) GetID() string {
	return m.ID
}

//GetSecret gets secret from environmental variable
func (m *MockCredentials) GetSecret() string {
	return m.Key
}
