// mystack-controller api
// https://github.com/topfreegames/mystack-controller
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2017 Top Free Games <backend@tfgco.com>

package api

import (
	"fmt"
	"io"
	"net"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/viper"
	"github.com/topfreegames/mystack-controller/errors"
	"github.com/topfreegames/mystack-controller/metadata"
	"github.com/topfreegames/mystack-controller/models"
	"k8s.io/client-go/kubernetes"
)

//App is our API application
type App struct {
	Address             string
	Config              *viper.Viper
	DB                  models.DB
	Debug               bool
	Logger              logrus.FieldLogger
	Router              *mux.Router
	Server              *http.Server
	EmailDomain         []string
	AppsRoutesDomain    string
	Clientset           kubernetes.Interface
	DeploymentReadiness models.Readiness
	JobReadiness        models.Readiness
}

//NewApp ctor
func NewApp(
	host string,
	port int,
	config *viper.Viper,
	debug bool,
	logger logrus.FieldLogger,
	clientset kubernetes.Interface,
) (*App, error) {
	a := &App{
		Config:              config,
		Address:             fmt.Sprintf("%s:%d", host, port),
		Debug:               debug,
		Logger:              logger,
		EmailDomain:         config.GetStringSlice("oauth.acceptedDomains"),
		AppsRoutesDomain:    config.GetString("kubernetes.appsDomain"),
		Clientset:           clientset,
		DeploymentReadiness: &models.DeploymentReadiness{},
		JobReadiness:        &models.JobReadiness{},
	}
	err := a.configureApp()
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (a *App) getRouter() *mux.Router {
	r := mux.NewRouter()

	r.Handle("/healthcheck", Chain(
		&HealthcheckHandler{App: a},
		&LoggingMiddleware{App: a},
		&VersionMiddleware{},
	)).Methods("GET").Name("healthcheck")

	r.Handle("/login", Chain(
		&LoginHandler{App: a, Method: "login"},
		&LoggingMiddleware{App: a},
		&VersionMiddleware{},
	)).Methods("GET").Name("oauth")

	r.Handle("/access", Chain(
		&LoginHandler{App: a, Method: "access"},
		&LoggingMiddleware{App: a},
		&VersionMiddleware{},
	)).Methods("GET").Name("oauth")

	r.Handle("/clusters/{name}/create", Chain(
		&ClusterHandler{App: a, Method: "create"},
		&LoggingMiddleware{App: a},
		&VersionMiddleware{},
		&AccessMiddleware{App: a},
	)).Methods("PUT").Name("cluster")

	r.Handle("/clusters/{name}/delete", Chain(
		&ClusterHandler{App: a, Method: "delete"},
		&LoggingMiddleware{App: a},
		&VersionMiddleware{},
		&AccessMiddleware{App: a},
	)).Methods("DELETE").Name("cluster")

	r.Handle("/clusters/{name}/routes", Chain(
		&ClusterHandler{App: a, Method: "routes"},
		&LoggingMiddleware{App: a},
		&VersionMiddleware{},
		&AccessMiddleware{App: a},
	)).Methods("GET").Name("cluster")

	r.Handle("/cluster-configs/{name}/create", Chain(
		&ClusterConfigHandler{App: a, Method: "create"},
		&VersionMiddleware{},
		&LoggingMiddleware{App: a},
		&AccessMiddleware{App: a},
		&PayloadMiddleware{App: a},
	)).Methods("PUT").Name("cluster-config")

	r.Handle("/cluster-configs/{name}/remove", Chain(
		&ClusterConfigHandler{App: a, Method: "remove"},
		&LoggingMiddleware{App: a},
		&VersionMiddleware{},
		&AccessMiddleware{App: a},
	)).Methods("DELETE").Name("cluster-config")

	r.Handle("/cluster-configs", Chain(
		&ClusterConfigHandler{App: a, Method: "list"},
		&LoggingMiddleware{App: a},
		&VersionMiddleware{},
		&AccessMiddleware{App: a},
	)).Methods("GET").Name("cluster-config")

	r.Handle("/cluster-configs/{name}", Chain(
		&ClusterConfigHandler{App: a, Method: "info"},
		&LoggingMiddleware{App: a},
		&VersionMiddleware{},
		&AccessMiddleware{App: a},
	)).Methods("GET").Name("cluster-config")

	r.Handle("/dns", Chain(
		&DNSHandler{App: a},
		&LoggingMiddleware{App: a},
		&AccessMiddleware{App: a},
		&VersionMiddleware{},
	)).Methods("GET").Name("dns")

	return r
}

func (a *App) configureApp() error {
	a.configureLogger()

	err := a.configureDatabase()
	if err != nil {
		return err
	}

	a.ConfigureServer()
	return nil
}

func (a *App) configureDatabase() error {
	db, err := a.getDB()
	if err != nil {
		return err
	}

	a.DB = db
	return nil
}

func (a *App) getDB() (*sqlx.DB, error) {
	host := a.Config.GetString("postgres.host")
	user := a.Config.GetString("postgres.user")
	dbName := a.Config.GetString("postgres.dbname")
	password := a.Config.GetString("postgres.password")
	port := a.Config.GetInt("postgres.port")
	sslMode := a.Config.GetString("postgres.sslMode")
	maxIdleConns := a.Config.GetInt("postgres.maxIdleConns")
	maxOpenConns := a.Config.GetInt("postgres.maxOpenConns")
	connectionTimeoutMS := a.Config.GetInt("postgres.connectionTimeoutMS")

	l := a.Logger.WithFields(logrus.Fields{
		"postgres.host":    host,
		"postgres.user":    user,
		"postgres.dbName":  dbName,
		"postgres.port":    port,
		"postgres.sslMode": sslMode,
	})

	l.Debug("Connecting to DB...")
	db, err := models.GetDB(
		host, user, port, sslMode, dbName,
		password, maxIdleConns, maxOpenConns,
		connectionTimeoutMS,
	)
	if err != nil {
		l.WithError(err).Error("Connection to database failed.")
		return nil, err
	}
	l.Debug("Successful connection to database.")
	return db, nil
}

func (a *App) configureLogger() {
	a.Logger = a.Logger.WithFields(logrus.Fields{
		"source":    "api/app.go",
		"operation": "initializeApp",
		"version":   metadata.Version,
	})
}

//ConfigureServer construct the routes
func (a *App) ConfigureServer() {
	a.Router = a.getRouter()
	a.Server = &http.Server{Addr: a.Address, Handler: a.Router}
}

//HandleError writes an error response with message and status
func (a *App) HandleError(w http.ResponseWriter, status int, msg string, err interface{}) {
	w.WriteHeader(status)
	var sErr errors.SerializableError
	val, ok := err.(errors.SerializableError)
	if ok {
		sErr = val
	} else {
		sErr = errors.NewGenericError(msg, err.(error))
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(sErr.Serialize())
}

//ListenAndServe requests
func (a *App) ListenAndServe() (io.Closer, error) {
	listener, err := net.Listen("tcp", a.Address)
	if err != nil {
		return nil, err
	}

	err = a.Server.Serve(listener)
	if err != nil {
		listener.Close()
		return nil, err
	}

	return listener, nil
}
