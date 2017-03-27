// kubecos
// https://github.com/topfreegames/kubecos
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package cmd

import (
	"github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/topfreegames/kubecos/kubecos-controller/api"
)

var host string
var port int
var debug bool
var quiet bool

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "starts the kubecos-api server",
	Long: `Starts kubecos-api server with the specified arguments. You can use
environment variables to override configuration keys.`,
	Run: func(cmd *cobra.Command, args []string) {
		InitConfig()
		ll := logrus.InfoLevel
		switch Verbose {
		case 0:
			ll = logrus.ErrorLevel
		case 1:
			ll = logrus.WarnLevel
		case 3:
			ll = logrus.DebugLevel
		}

		var log = logrus.New()
		log.Formatter = new(logrus.JSONFormatter)
		log.Level = ll

		cmdL := log.WithFields(logrus.Fields{
			"source":    "startCmd",
			"operation": "Run",
			"host":      host,
			"port":      port,
			"debug":     debug,
		})

		cmdL.Debug("Creating application...")
		app, err := api.NewApp(
			host,
			port,
			config,
			debug,
			log,
		)
		if err != nil {
			cmdL.WithError(err).Fatal("Failed to start application.")
		}
		cmdL.Debug("Application created successfully.")

		cmdL.Debug("Starting application...")
		closer, err := app.ListenAndServe()
		if closer != nil {
			defer closer.Close()
		}
		if err != nil {
			cmdL.WithError(err).Fatal("Error running application.")
		}
	},
}

func init() {
	RootCmd.AddCommand(startCmd)

	startCmd.Flags().StringVarP(&host, "bind", "b", "0.0.0.0", "Host to bind kubecos to")
	startCmd.Flags().IntVarP(&port, "port", "p", 8888, "Port to bind kubecos to")
	startCmd.Flags().BoolVarP(&debug, "debug", "d", false, "Debug mode")
	startCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "Quiet mode (log level error)")
}
