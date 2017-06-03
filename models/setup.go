// mystack-controller api
// https://github.com/topfreegames/mystack-controller
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2017 Top Free Games <backend@tfgco.com>

package models

//Setup has the job config to run and configure services
type Setup struct {
	Image          string    `yaml:"image"`
	Command        []string  `yaml:"command"`
	PeriodSeconds  int       `yaml:"periodSeconds"`
	Environment    []*EnvVar `yaml:"env"`
	TimeoutSeconds int       `yaml:"timeoutSeconds"`
}
