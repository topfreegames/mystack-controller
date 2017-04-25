// mystack-controller api
// https://github.com/topfreegames/mystack-controller
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2017 Top Free Games <backend@tfgco.com>

package models

import (
	"bytes"
	"fmt"

	"github.com/topfreegames/mystack-controller/errors"
	yaml "gopkg.in/yaml.v2"
)

//ClusterConfig contains the elements of a config file
type ClusterConfig struct {
	Setup    *Setup                       `yaml:"setup"`
	Services map[string]*ClusterAppConfig `yaml:"services"`
	Apps     map[string]*ClusterAppConfig `yaml:"apps"`
}

//LoadClusterConfig reads DB and create map with cluster configuration
func LoadClusterConfig(
	db DB,
	clusterName string,
) (
	*ClusterConfig,
	error,
) {
	if len(clusterName) == 0 {
		return nil, errors.NewGenericError("load cluster config error", fmt.Errorf("invalid empty cluster name"))
	}

	query := "SELECT yaml FROM clusters WHERE name = $1"
	var yamlStr string

	err := db.Get(&yamlStr, query, clusterName)
	if err != nil {
		return nil, errors.NewDatabaseError(err)
	}

	if len(yamlStr) == 0 {
		return nil, errors.NewYamlError("load cluster config error", fmt.Errorf("invalid empty config"))
	}

	clusterConfig, err := ParseYaml(yamlStr)
	if err != nil {
		return nil, errors.NewYamlError("load cluster config error", err)
	}

	return clusterConfig, nil
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
	clusterConfig, err := ParseYaml(yamlStr)
	if err != nil {
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

	if hasInsert, query := BuildQuery(clusterName, clusterConfig); hasInsert {
		res, err = db.NamedExec(query, map[string]interface{}{})
		if err != nil {
			return errors.NewDatabaseError(err)
		}
	}

	return nil
}

func BuildQuery(clusterName string, clusterConfig *ClusterConfig) (bool, string) {
	var buffer bytes.Buffer
	buffer.WriteString("INSERT INTO custom_domains VALUES")
	hasInsert := false

	for name, appConfig := range clusterConfig.Apps {
		if len(appConfig.CustomDomains) > 0 {
			hasInsert = true
			buffer.WriteString("('")
			buffer.WriteString(clusterName)
			buffer.WriteString("', '")
			buffer.WriteString(name)
			buffer.WriteString("', '{")

			for _, domain := range appConfig.CustomDomains {
				buffer.WriteString(`"`)
				buffer.WriteString(domain)
				buffer.WriteString(`", `)
			}

			buffer.Truncate(buffer.Len() - 2)
			buffer.WriteString("}'")
			buffer.WriteString("),")
		}
	}

	buffer.Truncate(buffer.Len() - 1)
	return hasInsert, buffer.String()
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

	query = `DELETE FROM custom_domains WHERE cluster=:name`
	res, err = db.NamedExec(query, values)
	if err != nil {
		return errors.NewDatabaseError(err)
	}

	return nil
}

//ParseYaml convert string to maps
func ParseYaml(yamlStr string) (*ClusterConfig, error) {
	clusterConfig := ClusterConfig{}
	err := yaml.Unmarshal([]byte(yamlStr), &clusterConfig)

	if err != nil {
		return nil, errors.NewYamlError("parse yaml error", err)
	}

	return &clusterConfig, nil
}

//ListClusterConfig return the list of saved cluster configs
func ListClusterConfig(db DB) ([]string, error) {
	names := []string{}
	query := "SELECT name FROM clusters"
	err := db.Select(&names, query)

	if err != nil {
		return nil, errors.NewDatabaseError(err)
	}

	return names, nil
}

//ClusterConfigDetails return the cluster config yaml
func ClusterConfigDetails(db DB, clusterName string) (string, error) {
	var yamlStr string
	query := "SELECT yaml FROM clusters WHERE name = $1"
	err := db.Get(&yamlStr, query, clusterName)
	if err != nil {
		return "", errors.NewDatabaseError(err)
	}

	return yamlStr, nil
}

func ClusterCustomDomains(db DB, clusterName string) (map[string][]string, error) {
	clusterConfig, err := LoadClusterConfig(db, clusterName)
	if err != nil {
		return nil, err
	}

	customDomains := make(map[string][]string)
	for name, app := range clusterConfig.Apps {
		customDomains[name] = app.CustomDomains
	}
	for name, svc := range clusterConfig.Services {
		customDomains[name] = svc.CustomDomains
	}

	return customDomains, nil
}
