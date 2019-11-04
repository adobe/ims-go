// Copyright 2019 Adobe. All rights reserved.
// This file is licensed to you under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License. You may obtain a copy
// of the License at http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software distributed under
// the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR REPRESENTATIONS
// OF ANY KIND, either express or implied. See the License for the specific language
// governing permissions and limitations under the License.

package login

import (
	"fmt"
	"net/http"

	"github.com/adobe/ims-go/ims"
)

type redirectBackend interface {
	AuthorizeURL(cfg *ims.AuthorizeURLConfig) (string, error)
}

type redirectMiddleware struct {
	client   redirectBackend
	clientID string
	scope    []string
	state    string
	next     http.Handler
}

func (h *redirectMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	url, err := h.client.AuthorizeURL(&ims.AuthorizeURLConfig{
		ClientID: h.clientID,
		Scope:    h.scope,
		State:    h.state,
	})
	if err != nil {
		serveError(h.next, w, r, fmt.Errorf("generate authorization URL: %v", err))
		return
	}

	http.Redirect(w, r, url, http.StatusFound)
}
