// mystack-controller api
// https://github.com/topfreegames/mystack-controller
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2017 Top Free Games <backend@tfgco.com>

package models

import (
	runner "gopkg.in/mgutz/dat.v2/sqlx-runner"
	yaml "gopkg.in/yaml.v2"
)

//EnvVar has name and value of an environmental value
type EnvVar struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

//ClusterAppConfig defines the configuration of an app
type ClusterAppConfig struct {
	Image       string    `yaml:"image"`
	Port        int       `yaml:"port"`
	Environment []*EnvVar `yaml:"env"`
}

//LoadClusterConfig reads DB and create map with cluster configuration
func LoadClusterConfig(
	db runner.Connection,
	clusterName string,
) (
	map[string]*ClusterAppConfig,
	map[string]*ClusterAppConfig,
	error,
) {
	query := "SELECT yaml FROM clusters WHERE name = $1"
	var yamlStr string
	err := db.SQL(query, clusterName).QueryStruct(&yamlStr)
	if err != nil {
		return nil, nil, err
	}
	apps, services, err := ParseYaml(yamlStr)
	return apps, services, err
}

//WriteClusterConfig writes cluster config on DB
func WriteClusterConfig(
	db runner.Connection,
	clusterName string,
	yamlStr string,
) error {
	query := "INSERT INTO clusters(name, yaml) VALUES($1, $2)"
	_, err := db.SQL(query, clusterName, yamlStr).Exec()
	return err
}

type clusterConfig struct {
	Services map[string]*ClusterAppConfig `yaml:"services"`
	Apps     map[string]*ClusterAppConfig `yaml:"apps"`
}

//ParseYaml convert string to maps
func ParseYaml(yamlStr string) (map[string]*ClusterAppConfig, map[string]*ClusterAppConfig, error) {
	cluster := clusterConfig{}
	err := yaml.Unmarshal([]byte(yamlStr), &cluster)
	return cluster.Apps, cluster.Services, err
}
