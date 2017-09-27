// mystack-controller
// https://github.com/topfreegames/mystack-controller
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2017 Top Free Games <backend@tfgco.com>

package models

// Resources holds how much CPU and memory is allocated to the pod
type Resources struct {
	Limits   *MemoryAndCPUResource `yaml:"limits"`
	Requests *MemoryAndCPUResource `yaml:"requests"`
}

// MemoryAndCPUResource has information about CPU and memory to be allocated to a pod
type MemoryAndCPUResource struct {
	CPU    string `yaml:"cpu"`
	Memory string `yaml:"memory"`
}
