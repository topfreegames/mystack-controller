// mystack-controller api
// https://github.com/topfreegames/mystack-controller
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2017 Top Free Games <backend@tfgco.com>

package models

//Setup has the job config to run and configure services
type Setup struct {
	Image          string `yaml:"image"`
	PeriodSeconds  int    `yaml:"period-seconds"`
	TimeoutSeconds int    `yaml:"timeout-seconds"`
}
