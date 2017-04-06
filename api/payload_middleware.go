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

func newContextWithClusterConfig(ctx context.Context, clusterConfig string) context.Context {
	c := context.WithValue(ctx, configKey, clusterConfig)
	return c
}

func clusterConfigFromCtx(ctx context.Context) string {
	clusterConfig := ctx.Value(configKey)
	if clusterConfig == nil {
		return ""
	}
	return clusterConfig.(string)
}

func (p *PayloadMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	bts, err := ioutil.ReadAll(r.Body)
	if err != nil {
		p.App.HandleError(w, http.StatusBadRequest, "Error reading body", err)
		return
	}

	bodyJSON := make(map[string]string)
	err = json.Unmarshal(bts, &bodyJSON)
	if err != nil {
		p.App.HandleError(w, http.StatusInternalServerError, "Error reading body", err)
		return
	}

	ctx := newContextWithClusterConfig(r.Context(), bodyJSON["yaml"])
	p.next.ServeHTTP(w, r.WithContext(ctx))
}

//SetNext handler
func (p *PayloadMiddleware) SetNext(next http.Handler) {
	p.next = next
}
