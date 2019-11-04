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
	"net/url"
	"testing"

	"github.com/adobe/ims-go/ims"
)

type testCallbackBackend func(r *ims.TokenRequest) (*ims.TokenResponse, error)

func (b testCallbackBackend) Token(r *ims.TokenRequest) (*ims.TokenResponse, error) {
	return b(r)
}

func urlWithParams(path string, params map[string]string) string {
	values := url.Values{}

	for k, v := range params {
		values.Set(k, v)
	}

	u := url.URL{
		Path:     path,
		RawQuery: values.Encode(),
	}

	return u.String()
}

func TestCallback(t *testing.T) {
	m := &callbackMiddleware{
		state:        "state",
		clientID:     "client-id",
		clientSecret: "client-secret",
		scope:        []string{"a", "b"},

		client: testCallbackBackend(func(r *ims.TokenRequest) (*ims.TokenResponse, error) {
			if r.Code != "code" {
				t.Fatalf("invalid code: %v", r.Code)
			}
			if r.ClientID != "client-id" {
				t.Fatalf("invalid client ID: %v", r.ClientID)
			}
			if r.ClientSecret != "client-secret" {
				t.Fatalf("invalid client secret: %v", r.ClientSecret)
			}
			if len(r.Scope) != 2 && r.Scope[0] != "a" && r.Scope[1] != "b" {
				t.Fatalf("invalid scope: %v", r.Scope)
			}
			return &ims.TokenResponse{}, nil
		}),

		next: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			res, ok := r.Context().Value(contextKeyResult).(*ims.TokenResponse)
			if !ok {
				t.Fatalf("invalid context value")
			}
			if res == nil {
				t.Fatalf("no response returned")
			}
		}),
	}

	target := urlWithParams("/", map[string]string{
		"code":  "code",
		"state": "state",
	})

	m.ServeHTTP(nil, httptest.NewRequest(http.MethodGet, target, nil))
}

func TestCallbackError(t *testing.T) {
	m := &callbackMiddleware{
		next: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err, ok := r.Context().Value(contextKeyError).(error)
			if !ok {
				t.Fatalf("invalid context value")
			}
			if err == nil {
				t.Fatalf("no error returned")
			}
			if err.Error() != "backend error: error" {
				t.Fatalf("invalid error: %v", err)
			}
		}),
	}

	target := urlWithParams("/", map[string]string{
		"error": "error",
	})

	m.ServeHTTP(nil, httptest.NewRequest(http.MethodGet, target, nil))
}

func TestCallbackNoState(t *testing.T) {
	m := &callbackMiddleware{
		next: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err, ok := r.Context().Value(contextKeyError).(error)
			if !ok {
				t.Fatalf("invalid context value")
			}
			if err == nil {
				t.Fatalf("no error returned")
			}
			if err.Error() != "missing state parameter" {
				t.Fatalf("invalid error: %v", err)
			}
		}),
	}

	m.ServeHTTP(nil, httptest.NewRequest(http.MethodGet, "/", nil))
}

func TestCallbackInvalidState(t *testing.T) {
	m := &callbackMiddleware{
		state: "state",
		next: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err, ok := r.Context().Value(contextKeyError).(error)
			if !ok {
				t.Fatalf("invalid context value")
			}
			if err == nil {
				t.Fatalf("no error returned")
			}
			if err.Error() != "invalid state parameter" {
				t.Fatalf("invalid error: %v", err)
			}
		}),
	}

	target := urlWithParams("/", map[string]string{
		"state": "invalid-state",
	})

	m.ServeHTTP(nil, httptest.NewRequest(http.MethodGet, target, nil))
}

func TestCallbackMissingCode(t *testing.T) {
	m := &callbackMiddleware{
		state: "state",
		next: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err, ok := r.Context().Value(contextKeyError).(error)
			if !ok {
				t.Fatalf("invalid context value")
			}
			if err == nil {
				t.Fatalf("no error returned")
			}
			if err.Error() != "missing code parameter" {
				t.Fatalf("invalid error: %v", err)
			}
		}),
	}

	target := urlWithParams("/", map[string]string{
		"state": "state",
	})

	m.ServeHTTP(nil, httptest.NewRequest(http.MethodGet, target, nil))
}

func TestCallbackBackendError(t *testing.T) {
	m := &callbackMiddleware{
		state: "state",
		client: testCallbackBackend(func(r *ims.TokenRequest) (*ims.TokenResponse, error) {
			return nil, fmt.Errorf("error")
		}),
		next: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err, ok := r.Context().Value(contextKeyError).(error)
			if !ok {
				t.Fatalf("invalid context value")
			}
			if err == nil {
				t.Fatalf("no error returned")
			}
			if err.Error() != "obtaining access token: error" {
				t.Fatalf("invalid error: %v", err)
			}
		}),
	}

	target := urlWithParams("/", map[string]string{
		"state": "state",
		"code":  "code",
	})

	m.ServeHTTP(nil, httptest.NewRequest(http.MethodGet, target, nil))
}
