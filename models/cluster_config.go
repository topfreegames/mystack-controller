// mystack-controller api
// https://github.com/topfreegames/mystack-controller
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2017 Top Free Games <backend@tfgco.com>

package models

import (
	"encoding/json"
	runner "gopkg.in/mgutz/dat.v2/sqlx-runner"
	yaml "gopkg.in/yaml.v2"
)

//ClusterAppConfig defines the configuration of an app
type ClusterAppConfig struct {
	Image       string
	Port        int
	Environment map[string]string
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
	query := `SELECT apps, services FROM clusters WHERE name = $1`

	configJSON, err := db.SQL(query, clusterName).QueryJSON()
	if err != nil {
		return nil, nil, err
	}

	var configCluster map[string]map[string]*ClusterAppConfig
	err = json.Unmarshal(configJSON, configCluster)

	return configCluster["apps"], configCluster["services"], err
}

//WriteClusterConfig writes cluster config on DB
func WriteClusterConfig(
	db runner.Connection,
	clusterName string,
	apps map[string]*ClusterAppConfig,
	services map[string]*ClusterAppConfig,
) error {
	query := `INSERT INTO clusters(name, apps, services)
						VALUES($1, $2, $3)`
	appsJSON, err := json.Marshal(apps)
	if err != nil {
		return err
	}

	servicesJSON, err := json.Marshal(services)
	if err != nil {
		return err
	}

	_, err = db.SQL(query, clusterName, appsJSON, servicesJSON).Exec()
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
	return cluster.Services, cluster.Apps, err
}
