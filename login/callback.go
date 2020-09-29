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

type callbackBackend interface {
	Token(r *ims.TokenRequest) (*ims.TokenResponse, error)
}

type callbackMiddleware struct {
	client       callbackBackend
	state        string
	clientID     string
	clientSecret string
	scope        []string
	next         http.Handler
}

func (h *callbackMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	values := r.URL.Query()

	urlErr := values.Get("error")
	if urlErr != "" {
		serveError(h.next, w, r, fmt.Errorf("backend error: %s", urlErr))
		return
	}

	state := values.Get("state")
	if state == "" {
		serveError(h.next, w, r, fmt.Errorf("missing state parameter"))
		return
	}

	if h.state != state {
		serveError(h.next, w, r, fmt.Errorf("invalid state parameter"))
		return
	}

	code := values.Get("code")
	if code == "" {
		serveError(h.next, w, r, fmt.Errorf("missing code parameter"))
		return
	}

	res, err := h.client.Token(&ims.TokenRequest{
		Code:         code,
		ClientID:     h.clientID,
		ClientSecret: h.clientSecret,
		Scope:        h.scope,
	})
	if err != nil {
		serveError(h.next, w, r, fmt.Errorf("obtaining access token: %v", err))
		return
	}

	serveResult(h.next, w, r, res)
}
