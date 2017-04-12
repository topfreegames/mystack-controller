// mystack-controller
// https://github.com/topfreegames/mystack-controller
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2017 Top Free Games <backend@tfgco.com>

package api

import (
	"net/http"
)

//DNSHandler handler
type DNSHandler struct {
	App *App
}

//ServeHTTP method
func (d *DNSHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	l := loggerFromContext(r.Context())

	l.Debug("Performing DNS...")

	Write(w, http.StatusOK, `{"domains": ["test.example.test", "test2.example.test"]}`)
	l.Debug("DNS done.")
}
