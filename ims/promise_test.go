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

func TestPromiseToken(t *testing.T) {
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
		if v := r.PostForm.Get("grant_type"); v != "promise" {
			t.Fatalf("incorrect grant type: %v", v)
		}
		if v := r.PostForm.Get("client_id"); v != "client-id" {
			t.Fatalf("invalid client_id: %v", v)
		}
		if v := r.PostForm.Get("client_secret"); v != "client-secret" {
			t.Fatalf("invalid client_secret: %v", v)
		}
		if v := r.PostForm.Get("promise_definition_id"); v != "promise-def" {
			t.Fatalf("invalid promise_definition_id: %v", v)
		}
		if v := r.PostForm.Get("authenticating_token"); v != "user-token" {
			t.Fatalf("invalid authenticating_token: %v", v)
		}
		if v := r.PostForm.Get("scope"); v != "openid,AdobeID" {
			t.Fatalf("invalid scopes: %v", v)
		}

		body := struct {
			PromiseToken string `json:"promise_token"`
			TokenType    string `json:"token_type"`
			Scope        string `json:"scope"`
			ExpiresIn    int    `json:"expires_in"`
		}{
			PromiseToken: "new-promise-token",
			TokenType:    "promise_token",
			Scope:        "openid,AdobeID",
			ExpiresIn:    2591999,
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

	r, err := c.PromiseToken(&ims.PromiseTokenRequest{
		ClientID:            "client-id",
		ClientSecret:        "client-secret",
		PromiseDefinitionID: "promise-def",
		AuthenticatingToken: "user-token",
		Scopes:              []string{"openid", "AdobeID"},
	})
	if err != nil {
		t.Fatalf("failure exchanging token: %v", err)
	}
	if r.PromiseToken != "new-promise-token" {
		t.Fatalf("invalid promise token: %v", r.PromiseToken)
	}
	if r.TokenType != "promise_token" {
		t.Fatalf("invalid token type: %v", r.TokenType)
	}
	if r.Scope != "openid,AdobeID" {
		t.Fatalf("invalid scope: %v", r.Scope)
	}
	if r.ExpiresIn != 2591999*time.Second {
		t.Fatalf("invalid expiration: %v", r.ExpiresIn)
	}
}

func TestPromiseTokenWithContext(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/ims/token/v4" {
			t.Fatalf("invalid path: %v", r.URL.Path)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("parse form: %v", err)
		}
		if v := r.PostForm.Get("authenticating_token"); v != "user-token" {
			t.Fatalf("invalid authenticating_token: %v", v)
		}
		body := struct {
			PromiseToken string `json:"promise_token"`
			ExpiresIn    int    `json:"expires_in"`
		}{
			PromiseToken: "promise-with-ctx-token",
			ExpiresIn:    1800,
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

	r, err := c.PromiseTokenWithContext(context.Background(), &ims.PromiseTokenRequest{
		ClientID:            "client-id",
		ClientSecret:        "client-secret",
		PromiseDefinitionID: "promise-def",
		AuthenticatingToken: "user-token",
		Scopes:              []string{"openid"},
	})
	if err != nil {
		t.Fatalf("failure exchanging token: %v", err)
	}
	if r.PromiseToken != "promise-with-ctx-token" {
		t.Fatalf("invalid promise token: %v", r.PromiseToken)
	}
	if r.ExpiresIn != 1800*time.Second {
		t.Fatalf("invalid expiration: %v", r.ExpiresIn)
	}
}

func TestPromiseTokenError(t *testing.T) {
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

	_, err = c.PromiseToken(&ims.PromiseTokenRequest{
		ClientID:            "irrelevant",
		ClientSecret:        "irrelevant",
		PromiseDefinitionID: "irrelevant",
		AuthenticatingToken: "irrelevant",
		Scopes:              []string{"openid"},
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

func TestPromiseTokenInvalidRequest(t *testing.T) {
	c, err := ims.NewClient(&ims.ClientConfig{URL: "http://ims.endpoint"})
	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	tests := []struct {
		name    string
		request *ims.PromiseTokenRequest
		wantErr string
	}{
		{
			name: "missing ClientID",
			request: &ims.PromiseTokenRequest{
				ClientSecret:        "client-secret",
				PromiseDefinitionID: "promise-def",
				AuthenticatingToken: "user-token",
				Scopes:              []string{"openid"},
			},
			wantErr: "invalid parameters for promise token exchange: missing client ID parameter",
		},
		{
			name: "missing ClientSecret",
			request: &ims.PromiseTokenRequest{
				ClientID:            "client-id",
				PromiseDefinitionID: "promise-def",
				AuthenticatingToken: "user-token",
				Scopes:              []string{"openid"},
			},
			wantErr: "invalid parameters for promise token exchange: missing client secret parameter",
		},
		{
			name: "missing PromiseDefinitionID",
			request: &ims.PromiseTokenRequest{
				ClientID:            "client-id",
				ClientSecret:        "client-secret",
				AuthenticatingToken: "user-token",
				Scopes:              []string{"openid"},
			},
			wantErr: "invalid parameters for promise token exchange: missing promise definition ID parameter",
		},
		{
			name: "missing AuthenticatingToken",
			request: &ims.PromiseTokenRequest{
				ClientID:            "client-id",
				ClientSecret:        "client-secret",
				PromiseDefinitionID: "promise-def",
				Scopes:              []string{"openid"},
			},
			wantErr: "invalid parameters for promise token exchange: missing authenticating token parameter",
		},
		{
			name: "empty Scopes",
			request: &ims.PromiseTokenRequest{
				ClientID:            "client-id",
				ClientSecret:        "client-secret",
				PromiseDefinitionID: "promise-def",
				AuthenticatingToken: "user-token",
				Scopes:              nil,
			},
			wantErr: "invalid parameters for promise token exchange: scopes are required for promise token exchange",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := c.PromiseToken(tt.request)
			if err == nil {
				t.Fatal("expected error")
			}
			if got := err.Error(); got != tt.wantErr {
				t.Fatalf("error = %q, want %q", got, tt.wantErr)
			}
		})
	}
}

func TestPromiseExchange(t *testing.T) {
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
		if v := r.PostForm.Get("grant_type"); v != "promise_exchange" {
			t.Fatalf("incorrect grant type: %v", v)
		}
		if v := r.PostForm.Get("client_id"); v != "client-id" {
			t.Fatalf("invalid client_id: %v", v)
		}
		if v := r.PostForm.Get("client_secret"); v != "client-secret" {
			t.Fatalf("invalid client_secret: %v", v)
		}
		if v := r.PostForm.Get("promise_token"); v != "promise-token" {
			t.Fatalf("invalid promise_token: %v", v)
		}
		if v := r.PostForm.Get("scope"); v != "openid,AdobeID" {
			t.Fatalf("invalid scopes: %v", v)
		}

		body := struct {
			AccessToken           string `json:"access_token"`
			PromiseToken          string `json:"promise_token"`
			PromiseTokenID        string `json:"promise_token_id"`
			TokenType             string `json:"token_type"`
			Scope                 string `json:"scope"`
			ExpiresIn             int    `json:"expires_in"`
			PromiseTokenExpiresIn int    `json:"promise_token_expires_in"`
		}{
			AccessToken:           "new-access-token",
			PromiseToken:          "rotated-promise-token",
			PromiseTokenID:        "promise-token-id",
			TokenType:             "access_token",
			Scope:                 "openid,AdobeID",
			ExpiresIn:             299,
			PromiseTokenExpiresIn: 2591927,
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

	r, err := c.PromiseExchange(&ims.PromiseExchangeRequest{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		PromiseToken: "promise-token",
		Scopes:       []string{"openid", "AdobeID"},
	})
	if err != nil {
		t.Fatalf("failure exchanging token: %v", err)
	}
	if r.AccessToken != "new-access-token" {
		t.Fatalf("invalid access token: %v", r.AccessToken)
	}
	if r.PromiseToken != "rotated-promise-token" {
		t.Fatalf("invalid promise token: %v", r.PromiseToken)
	}
	if r.PromiseTokenID != "promise-token-id" {
		t.Fatalf("invalid promise token id: %v", r.PromiseTokenID)
	}
	if r.TokenType != "access_token" {
		t.Fatalf("invalid token type: %v", r.TokenType)
	}
	if r.Scope != "openid,AdobeID" {
		t.Fatalf("invalid scope: %v", r.Scope)
	}
	if r.ExpiresIn != 299*time.Second {
		t.Fatalf("invalid expiration: %v", r.ExpiresIn)
	}
	if r.PromiseTokenExpiresIn != 2591927*time.Second {
		t.Fatalf("invalid promise token expiration: %v", r.PromiseTokenExpiresIn)
	}
}

func TestPromiseExchangeWithContext(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/ims/token/v4" {
			t.Fatalf("invalid path: %v", r.URL.Path)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("parse form: %v", err)
		}
		if v := r.PostForm.Get("promise_token"); v != "promise-token" {
			t.Fatalf("invalid promise_token: %v", v)
		}
		body := struct {
			AccessToken string `json:"access_token"`
			ExpiresIn   int    `json:"expires_in"`
		}{
			AccessToken: "access-with-ctx-token",
			ExpiresIn:   299,
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

	r, err := c.PromiseExchangeWithContext(context.Background(), &ims.PromiseExchangeRequest{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		PromiseToken: "promise-token",
		Scopes:       []string{"openid"},
	})
	if err != nil {
		t.Fatalf("failure exchanging token: %v", err)
	}
	if r.AccessToken != "access-with-ctx-token" {
		t.Fatalf("invalid access token: %v", r.AccessToken)
	}
	if r.ExpiresIn != 299*time.Second {
		t.Fatalf("invalid expiration: %v", r.ExpiresIn)
	}
}

func TestPromiseExchangeError(t *testing.T) {
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

	_, err = c.PromiseExchange(&ims.PromiseExchangeRequest{
		ClientID:     "irrelevant",
		ClientSecret: "irrelevant",
		PromiseToken: "irrelevant",
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

func TestPromiseExchangeInvalidRequest(t *testing.T) {
	c, err := ims.NewClient(&ims.ClientConfig{URL: "http://ims.endpoint"})
	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	tests := []struct {
		name    string
		request *ims.PromiseExchangeRequest
		wantErr string
	}{
		{
			name: "missing ClientID",
			request: &ims.PromiseExchangeRequest{
				ClientSecret: "client-secret",
				PromiseToken: "promise-token",
				Scopes:       []string{"openid"},
			},
			wantErr: "invalid parameters for promise exchange: missing client ID parameter",
		},
		{
			name: "missing ClientSecret",
			request: &ims.PromiseExchangeRequest{
				ClientID:     "client-id",
				PromiseToken: "promise-token",
				Scopes:       []string{"openid"},
			},
			wantErr: "invalid parameters for promise exchange: missing client secret parameter",
		},
		{
			name: "missing PromiseToken",
			request: &ims.PromiseExchangeRequest{
				ClientID:     "client-id",
				ClientSecret: "client-secret",
				Scopes:       []string{"openid"},
			},
			wantErr: "invalid parameters for promise exchange: missing promise token parameter",
		},
		{
			name: "empty Scopes",
			request: &ims.PromiseExchangeRequest{
				ClientID:     "client-id",
				ClientSecret: "client-secret",
				PromiseToken: "promise-token",
				Scopes:       nil,
			},
			wantErr: "invalid parameters for promise exchange: scopes are required for promise exchange",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := c.PromiseExchange(tt.request)
			if err == nil {
				t.Fatal("expected error")
			}
			if got := err.Error(); got != tt.wantErr {
				t.Fatalf("error = %q, want %q", got, tt.wantErr)
			}
		})
	}
}
