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
	"net/http/httptest"
	"testing"

	"github.com/adobe/ims-go/ims"
)

type testRedirectBackend func(cfg *ims.AuthorizeURLConfig) (string, error)

func (b testRedirectBackend) AuthorizeURL(cfg *ims.AuthorizeURLConfig) (string, error) {
	return b(cfg)
}

func TestRedirect(t *testing.T) {
	m := &redirectMiddleware{
		clientID: "client-id",
		scope:    []string{"a", "b"},
		state:    "state",
		client: testRedirectBackend(func(cfg *ims.AuthorizeURLConfig) (string, error) {
			if cfg.ClientID != "client-id" {
				t.Fatalf("invalid client ID: %v", cfg.ClientID)
			}
			if cfg.State != "state" {
				t.Fatalf("invalid state: %v", cfg.State)
			}
			if len(cfg.Scope) != 2 && cfg.Scope[0] != "a" && cfg.Scope[1] != "b" {
				t.Fatalf("invalid scope: %v", cfg.Scope)
			}
			return "http://acme.com/login", nil
		}),
	}

	w := httptest.NewRecorder()

	m.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))

	r := w.Result()

	if r.StatusCode != http.StatusFound {
		t.Fatalf("invalid status code: %v", r.StatusCode)
	}
	if h := r.Header.Get("location"); h != "http://acme.com/login" {
		t.Fatalf("invalid location: %v", h)
	}
}

func TestRedirectBackendError(t *testing.T) {
	m := &redirectMiddleware{
		clientID: "client-id",
		scope:    []string{"a", "b"},
		state:    "state",
		client: testRedirectBackend(func(cfg *ims.AuthorizeURLConfig) (string, error) {
			return "", fmt.Errorf("error")
		}),
		next: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err, ok := r.Context().Value(contextKeyError).(error)
			if !ok {
				t.Fatalf("invalid context value")
			}
			if err == nil {
				t.Fatalf("expected error")
			}
			if err.Error() != "generate authorization URL: error" {
				t.Fatalf("invalid error: %v", err)
			}
		}),
	}

	m.ServeHTTP(nil, httptest.NewRequest(http.MethodGet, "/", nil))
}
