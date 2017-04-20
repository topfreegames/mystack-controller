// mystack-controller api
// https://github.com/topfreegames/mystack-controller
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2017 Top Free Games <backend@tfgco.com>

package testing

import (
	"strings"

	"github.com/spf13/viper"
)

//GetDefaultConfig returns the configuration at ./config/test.yaml
func GetDefaultConfig() (*viper.Viper, error) {
	config := viper.New()
	config.SetConfigFile("../config/test.yaml")
	config.SetConfigType("yaml")
	config.SetEnvPrefix("mystack")
	config.AddConfigPath(".")
	config.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	config.AutomaticEnv()

	// If a config file is found, read it in.
	if err := config.ReadInConfig(); err != nil {
		return nil, err
	}

	return config, nil
}
