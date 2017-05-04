// mystack-controller api
// https://github.com/topfreegames/mystack-controller
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2017 Top Free Games <backend@tfgco.com>

package models

//EnvVar has name and value of an environment value
type EnvVar struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

//VolumeMount helps getting PersistentVolume from yaml
type VolumeMount struct {
	Name      string `yaml:"name"`
	MountPath string `yaml:"mountPath"`
}

//ClusterAppConfig defines the configuration of an app and service
type ClusterAppConfig struct {
	Image          string       `yaml:"image"`
	Ports          []string     `yaml:"ports"`
	Environment    []*EnvVar    `yaml:"env,flow"`
	ReadinessProbe *Probe       `yaml:"readinessProbe"`
	VolumeMount    *VolumeMount `yaml:"volumeMount"`
}
