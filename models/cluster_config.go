// mystack-controller api
// https://github.com/topfreegames/mystack-controller
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2017 Top Free Games <backend@tfgco.com>

package models

import (
	"fmt"
	"github.com/topfreegames/mystack-controller/errors"
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
	db DB,
	clusterName string,
) (
	map[string]*ClusterAppConfig,
	map[string]*ClusterAppConfig,
	error,
) {
	if len(clusterName) == 0 {
		return nil, nil, errors.NewGenericError("load cluster config error", fmt.Errorf("invalid empty cluster name"))
	}

	query := "SELECT yaml FROM clusters WHERE name = $1"
	var yamlStr string

	err := db.Get(&yamlStr, query, clusterName)
	if err != nil {
		return nil, nil, errors.NewDatabaseError(err)
	}

	if len(yamlStr) == 0 {
		return nil, nil, errors.NewYamlError("load cluster config error", fmt.Errorf("invalid empty config"))
	}

	apps, services, err := ParseYaml(yamlStr)
	if err != nil {
		return nil, nil, errors.NewYamlError("load cluster config error", err)
	}

	return apps, services, nil
}

//WriteClusterConfig writes cluster config on DB
func WriteClusterConfig(
	db DB,
	clusterName string,
	yamlStr string,
) error {
	if len(clusterName) == 0 {
		return errors.NewGenericError("write cluster config error", fmt.Errorf("invalid empty cluster name"))
	}
	if _, _, err := ParseYaml(yamlStr); err != nil {
		return errors.NewYamlError("write cluster config error", err)
	}
	if len(yamlStr) == 0 {
		return errors.NewYamlError("write cluster config error", fmt.Errorf("invalid empty config"))
	}

	query := `INSERT INTO clusters(name, yaml) VALUES(:name, :yaml)`
	values := map[string]interface{}{
		"name": clusterName,
		"yaml": yamlStr,
	}
	res, err := db.NamedExec(query, values)
	if err != nil {
		return errors.NewDatabaseError(err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return errors.NewDatabaseError(fmt.Errorf("couldn't insert on database"))
	}
	return nil
}

//RemoveClusterConfig writes cluster config on DB
func RemoveClusterConfig(
	db DB,
	clusterName string,
) error {
	if len(clusterName) == 0 {
		return errors.NewGenericError("remove cluster config error", fmt.Errorf("invalid empty cluster name"))
	}

	query := `DELETE FROM clusters WHERE name=:name`
	values := map[string]interface{}{
		"name": clusterName,
	}
	res, err := db.NamedExec(query, values)
	if err != nil {
		return errors.NewDatabaseError(err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		err = fmt.Errorf("sql: no rows in result set")
		return errors.NewDatabaseError(err)
	}
	return nil
}

type clusterConfig struct {
	Services map[string]*ClusterAppConfig `yaml:"services"`
	Apps     map[string]*ClusterAppConfig `yaml:"apps"`
}

//ParseYaml convert string to maps
func ParseYaml(yamlStr string) (map[string]*ClusterAppConfig, map[string]*ClusterAppConfig, error) {
	cluster := clusterConfig{}
	err := yaml.Unmarshal([]byte(yamlStr), &cluster)

	if err != nil {
		return nil, nil, errors.NewYamlError("parse yaml error", err)
	}

	return cluster.Apps, cluster.Services, nil
}
