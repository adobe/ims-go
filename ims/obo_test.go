// Copyright 2026 Adobe. All rights reserved.
// This file is licensed to you under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License. You may obtain a copy
// of the License at http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software distributed under
// the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR REPRESENTATIONS
// OF ANY KIND, either express or implied. See the License for the specific language
// governing permissions and limitations under the License.

package ims_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/adobe/ims-go/ims"
)

func TestOBOExchange(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("invalid method: %v", r.Method)
		}

		if r.URL.Path != "/ims/token/v4" {
			t.Fatalf("invalid path: %v", r.URL.Path)
		}

		if err := r.ParseForm(); err != nil {
			t.Fatalf("parse form: %v", err)
		}
		if v := r.PostForm.Get("grant_type"); v != "urn:ietf:params:oauth:grant-type:token-exchange" {
			t.Fatalf("incorrect grant type: %v", v)
		}
		if v := r.PostForm.Get("client_id"); v != "client-id" {
			t.Fatalf("invalid client_id: %v", v)
		}
		if v := r.PostForm.Get("client_secret"); v != "client-secret" {
			t.Fatalf("invalid client_secret: %v", v)
		}
		if v := r.PostForm.Get("subject_token"); v != "user-token" {
			t.Fatalf("invalid subject_token: %v", v)
		}
		if v := r.PostForm.Get("subject_token_type"); v != "urn:ietf:params:oauth:token-type:access_token" {
			t.Fatalf("invalid subject_token_type: %v", v)
		}
		if v := r.PostForm.Get("requested_token_type"); v != "urn:ietf:params:oauth:token-type:access_token" {
			t.Fatalf("invalid requested_token_type: %v", v)
		}
		if v := r.PostForm.Get("scope"); v != "openid,profile" {
			t.Fatalf("invalid scopes: %v", v)
		}

		body := struct {
			AccessToken string `json:"access_token"`
			ExpiresIn   int    `json:"expires_in"`
		}{
			AccessToken: "new-obo-token",
			ExpiresIn:   3600,
		}
		if err := json.NewEncoder(w).Encode(&body); err != nil {
			t.Fatalf("encode response: %v", err)
		}
	}))
	defer s.Close()

	c, err := ims.NewClient(&ims.ClientConfig{URL: s.URL})
	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	r, err := c.OBOExchange(&ims.OBOExchangeRequest{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		SubjectToken: "user-token",
		Scopes:       []string{"openid", "profile"},
	})
	if err != nil {
		t.Fatalf("failure exchanging token: %v", err)
	}
	if r.AccessToken != "new-obo-token" {
		t.Fatalf("invalid access token: %v", r.AccessToken)
	}
	if r.ExpiresIn != 3600*time.Second {
		t.Fatalf("invalid expiration: %v", r.ExpiresIn)
	}
}

func TestOBOExchangeWithContext(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("invalid method: %v", r.Method)
		}
		if r.URL.Path != "/ims/token/v4" {
			t.Fatalf("invalid path: %v", r.URL.Path)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("parse form: %v", err)
		}
		if v := r.PostForm.Get("subject_token"); v != "user-token" {
			t.Fatalf("invalid subject_token: %v", v)
		}
		body := struct {
			AccessToken string `json:"access_token"`
			ExpiresIn   int    `json:"expires_in"`
		}{
			AccessToken: "obo-with-ctx-token",
			ExpiresIn:   1800,
		}
		if err := json.NewEncoder(w).Encode(&body); err != nil {
			t.Fatalf("encode response: %v", err)
		}
	}))
	defer s.Close()

	c, err := ims.NewClient(&ims.ClientConfig{URL: s.URL})
	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	r, err := c.OBOExchangeWithContext(context.Background(), &ims.OBOExchangeRequest{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		SubjectToken: "user-token",
		Scopes:       []string{"openid", "profile"},
	})
	if err != nil {
		t.Fatalf("failure exchanging token: %v", err)
	}
	if r.AccessToken != "obo-with-ctx-token" {
		t.Fatalf("invalid access token: %v", r.AccessToken)
	}
	if r.ExpiresIn != 1800*time.Second {
		t.Fatalf("invalid expiration: %v", r.ExpiresIn)
	}
}

func TestOBOExchangeError(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)

		body := struct {
			ErrorCode    string `json:"error"`
			ErrorMessage string `json:"error_description"`
		}{
			ErrorCode:    "error-code",
			ErrorMessage: "error-message",
		}
		if err := json.NewEncoder(w).Encode(&body); err != nil {
			t.Fatalf("encode response: %v", err)
		}
	}))
	defer s.Close()

	c, err := ims.NewClient(&ims.ClientConfig{URL: s.URL})
	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	_, err = c.OBOExchange(&ims.OBOExchangeRequest{
		ClientID:     "irrelevant",
		ClientSecret: "irrelevant",
		SubjectToken: "irrelevant",
		Scopes:       []string{"openid"},
	})

	imsErr, ok := ims.IsError(err)
	if !ok {
		t.Fatalf("expected IMS error")
	}
	if imsErr.StatusCode != http.StatusBadRequest {
		t.Fatalf("invalid status code: %v", imsErr.StatusCode)
	}
	if imsErr.ErrorCode != "error-code" {
		t.Fatalf("invalid error code: %v", imsErr.ErrorCode)
	}
	if imsErr.ErrorMessage != "error-message" {
		t.Fatalf("invalid error message: %v", imsErr.ErrorMessage)
	}
}

func TestOBOExchangeTooManyRequests(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "77")
		w.Header().Set("x-debug-id", "banana")
		w.WriteHeader(http.StatusTooManyRequests)

		body := struct {
			ErrorCode    string `json:"error"`
			ErrorMessage string `json:"error_description"`
		}{
			ErrorCode:    "error-code",
			ErrorMessage: "error-message",
		}
		if err := json.NewEncoder(w).Encode(&body); err != nil {
			t.Fatalf("encode response: %v", err)
		}
	}))
	defer s.Close()

	c, err := ims.NewClient(&ims.ClientConfig{URL: s.URL})
	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	res, err := c.OBOExchange(&ims.OBOExchangeRequest{
		ClientID:     "irrelevant",
		ClientSecret: "irrelevant",
		SubjectToken: "irrelevant",
		Scopes:       []string{"openid"},
	})
	if res != nil {
		t.Fatalf("expected nil response because of error")
	}

	imsErr, ok := ims.IsError(err)
	if !ok {
		t.Fatalf("expected IMS error")
	}
	if imsErr.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("invalid status code: %v", imsErr.StatusCode)
	}
	if imsErr.RetryAfter != "77" {
		t.Fatalf("invalid retry-after header: %v", imsErr.RetryAfter)
	}
	if imsErr.XDebugID != "banana" {
		t.Fatalf("invalid x-debug-id header: %v", imsErr.XDebugID)
	}
}

func TestOBOExchangeInvalidRequest(t *testing.T) {
	c, err := ims.NewClient(&ims.ClientConfig{URL: "http://ims.endpoint"})
	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	tests := []struct {
		name    string
		request *ims.OBOExchangeRequest
		wantErr string
	}{
		{
			name: "missing ClientID",
			request: &ims.OBOExchangeRequest{
				ClientID:     "",
				ClientSecret: "client-secret",
				SubjectToken: "user-token",
				Scopes:       []string{"openid"},
			},
			wantErr: "invalid parameters for On-Behalf-Of exchange: missing client ID parameter",
		},
		{
			name: "missing ClientSecret",
			request: &ims.OBOExchangeRequest{
				ClientID:     "client-id",
				ClientSecret: "",
				SubjectToken: "user-token",
				Scopes:       []string{"openid"},
			},
			wantErr: "invalid parameters for On-Behalf-Of exchange: missing client secret parameter",
		},
		{
			name: "missing SubjectToken",
			request: &ims.OBOExchangeRequest{
				ClientID:     "client-id",
				ClientSecret: "client-secret",
				SubjectToken: "",
				Scopes:       []string{"openid"},
			},
			wantErr: "invalid parameters for On-Behalf-Of exchange: missing subject token parameter (only access tokens are accepted)",
		},
		{
			name: "empty Scopes",
			request: &ims.OBOExchangeRequest{
				ClientID:     "client-id",
				ClientSecret: "client-secret",
				SubjectToken: "user-token",
				Scopes:       nil,
			},
			wantErr: "invalid parameters for On-Behalf-Of exchange: scopes are required for On-Behalf-Of exchange",
		},
		{
			name: "single empty scope string",
			request: &ims.OBOExchangeRequest{
				ClientID:     "client-id",
				ClientSecret: "client-secret",
				SubjectToken: "user-token",
				Scopes:       []string{""},
			},
			wantErr: "invalid parameters for On-Behalf-Of exchange: scopes are required for On-Behalf-Of exchange",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := c.OBOExchange(tt.request)
			if err == nil {
				t.Fatal("expected error")
			}
			if got := err.Error(); got != tt.wantErr {
				t.Fatalf("error = %q, want %q", got, tt.wantErr)
			}
		})
	}
}
