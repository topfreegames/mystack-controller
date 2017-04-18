// mystack-controller api
// https://github.com/topfreegames/mystack-controller
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2017 Top Free Games <backend@tfgco.com>

package models

//Probe refers to the Kubernetes readiness probe
type Probe struct {
	Command       []string `yaml:"command"`
	PeriodSeconds int      `yaml:"period-seconds"`
	//This timeout is different from the kubernetes object api timeoutSecond.
	//This timeout represents the total wait time until stop deployment initiation and rollback
	TimeoutSeconds int `yaml:"start-deployment-timeout-seconds"`
}
