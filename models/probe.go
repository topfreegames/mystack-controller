// mystack-controller api
// https://github.com/topfreegames/mystack-controller
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2017 Top Free Games <backend@tfgco.com>

package models

import (
	"os/exec"
)

//Probe refers to the Kubernetes readiness probe
type Probe struct {
	Command []string `yaml:"command"`
}

//Status executes the probe command
//Return 0 if exit code was 0 (successfully executed)
//Return 1 otherwise
func (p *Probe) Status() int {
	cmd := p.Command[0]
	args := p.Command[1:]

	err := exec.Command(cmd, args...).Run()
	if err != nil {
		return 1
	}

	return 0
}
