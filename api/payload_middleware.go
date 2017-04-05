// mystack-controller api
// https://github.com/topfreegames/mystack-controller
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2017 Top Free Games <backend@tfgco.com>

package api

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

//PayloadMiddleware handles cluster creation and deletion
type PayloadMiddleware struct {
	App  *App
	next http.Handler
}

//ClusterAppConfig contains information about each app that will run on cluster
type ClusterAppConfig struct {
	Image string
	Port  int
}

const configKey = contextKey("clusterConfigKey")

func newContextWithClusterConfig(ctx context.Context, clusterConfig map[string]*ClusterAppConfig) context.Context {
	c := context.WithValue(ctx, configKey, clusterConfig)
	return c
}

func clusterConfigFromCtx(ctx context.Context) map[string]*ClusterAppConfig {
	clusterConfig := ctx.Value(configKey)
	if clusterConfig == nil {
		return nil
	}
	return clusterConfig.(map[string]*ClusterAppConfig)
}

func (p *PayloadMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	bts, err := ioutil.ReadAll(r.Body)
	if err != nil {
		p.App.HandleError(w, http.StatusBadRequest, "Error reading body", err)
		return
	}

	clusterConfig := make(map[string]*ClusterAppConfig)
	err = json.Unmarshal(bts, clusterConfig)
	if err != nil {
		p.App.HandleError(w, http.StatusBadRequest, "Error reading body", err)
		return
	}

	ctx := newContextWithClusterConfig(r.Context(), clusterConfig)
	p.next.ServeHTTP(w, r.WithContext(ctx))
}

//SetNext handler
func (p *PayloadMiddleware) SetNext(next http.Handler) {
	p.next = next
}
