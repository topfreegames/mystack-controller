// Package api mystack-controller
// https://github.com/topfreegames/mystack-controller
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2017 Top Free Games <backend@tfgco.com>
package api

import (
	"github.com/topfreegames/mystack-controller/metadata"
	"net/http"
)

// VersionMiddleware adds the version to the request
type VersionMiddleware struct {
	next http.Handler
}

//ServeHTTP method
func (m *VersionMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-Kubecos-Version", metadata.Version)
	m.next.ServeHTTP(w, r)
}

//SetNext handler
func (m *VersionMiddleware) SetNext(next http.Handler) {
	m.next = next
}
