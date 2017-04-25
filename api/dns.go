// mystack-controller
// https://github.com/topfreegames/mystack-controller
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2017 Top Free Games <backend@tfgco.com>

package api

import (
	"encoding/json"
	"net/http"
)

//DNSHandler handler
type DNSHandler struct {
	App *App
}

//ServeHTTP method
func (d *DNSHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logger := loggerFromContext(r.Context())
	clusterName := GetClusterName(r)

	log(logger, "Getting domains of cluster config '%s'", clusterName)
	customDomains := map[string]map[string][]string{
		"domains": map[string][]string{
			"app1": []string{"i am temporay"},
		},
	}
	var err error
	if err != nil {
		d.App.HandleError(w, Status(err), "cluster custom domains error", err)
		return
	}

	bts, err := json.Marshal(customDomains)
	if err != nil {
		d.App.HandleError(w, Status(err), "cluster custom domains error", err)
		return
	}
	WriteBytes(w, http.StatusOK, bts)
	log(logger, "Successfully got cluster custom domains")
}
