// mystack
// https://github.com/topfreegames/mystack-controller
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package cmd

import (
	"log"
	"os/user"
	"path"

	"github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/topfreegames/mystack-controller/api"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var host string
var port int
var debug, quiet, out bool

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "starts the mystack-api server",
	Long: `Starts mystack-api server with the specified arguments. You can use
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

		clientset, err := getClientset()
		if err != nil {
			cmdL.WithError(err).Fatal("Failed to start kubernetes clientset.")
		}

		cmdL.Info("Creating application...")
		app, err := api.NewApp(
			host,
			port,
			config,
			debug,
			log,
			clientset,
		)
		if err != nil {
			cmdL.WithError(err).Fatal("Failed to start application.")
		}
		cmdL.Info("Application created successfully.")

		cmdL.Info("Starting application...")
		closer, err := app.ListenAndServe()
		if closer != nil {
			defer closer.Close()
		}
		if err != nil {
			cmdL.WithError(err).Fatal("Error running application.")
		}
	},
}

func getClientset() (kubernetes.Interface, error) {
	var config *rest.Config
	var err error

	if out {
		usr, err := user.Current()
		if err != nil {
			log.Fatal(err)
		}
		kubeconfig := path.Join(usr.HomeDir, ".kube", "config")

		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, err
		}
	} else {
		config, err = rest.InClusterConfig()
	}

	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)

	return clientset, err
}

func init() {
	RootCmd.AddCommand(startCmd)

	startCmd.Flags().StringVarP(&host, "bind", "b", "0.0.0.0", "Host to bind mystack to")
	startCmd.Flags().IntVarP(&port, "port", "p", 8080, "Port to bind mystack to")
	startCmd.Flags().BoolVarP(&debug, "debug", "d", false, "Debug mode")
	startCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "Quiet mode (log level error)")
	startCmd.Flags().BoolVarP(&out, "out", "o", false, "Run controller out-of-cluster")
}
