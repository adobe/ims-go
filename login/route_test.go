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
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRoute(t *testing.T) {
	var redirect bool

	m := &routeMiddleware{
		redirect: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			redirect = true
		}),
	}

	m.ServeHTTP(nil, httptest.NewRequest(http.MethodGet, "/", nil))

	if !redirect {
		t.Fatalf("redirect handler not invoked")
	}
}

func TestRouteWithCode(t *testing.T) {
	var callback bool

	m := &routeMiddleware{
		callback: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callback = true
		}),
	}

	target := urlWithParams("/", map[string]string{
		"code": "code",
	})

	m.ServeHTTP(nil, httptest.NewRequest(http.MethodGet, target, nil))

	if !callback {
		t.Fatalf("callback handler not invoked")
	}
}

func TestRouteWithError(t *testing.T) {
	var callback bool

	m := &routeMiddleware{
		callback: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callback = true
		}),
	}

	target := urlWithParams("/", map[string]string{
		"error": "error",
	})

	m.ServeHTTP(nil, httptest.NewRequest(http.MethodGet, target, nil))

	if !callback {
		t.Fatalf("callback handler not invoked")
	}
}
