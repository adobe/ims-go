// Copyright 2019 Adobe. All rights reserved.
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
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/adobe/ims-go/ims"
)

func TestOBOExchange(t *testing.T) {
	// Spin up a fake IMS server that validates what OBOExchange sends it
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// OBO requires POST, just like the other exchanges
		if r.Method != http.MethodPost {
			t.Fatalf("invalid method: %v", r.Method)
		}

		// This is the key difference from cluster — OBO needs /ims/token/v4, not v3
		if r.URL.Path != "/ims/token/v4" {
			t.Fatalf("invalid path: %v", r.URL.Path)
		}

		// client_id must also appear as a query parameter (your original logic does this)
		v, ok := r.URL.Query()["client_id"]
		if !ok || v[0] != "client-id" {
			t.Fatalf("invalid client ID in query: %v", v)
		}

		// Now check all the POST body fields
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

		// Return a successful response
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

func TestOBOExchangeError(t *testing.T) {
	// Fake server that simulates IMS rejecting the request
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

	// The library has a standard error type — use IsError just like the cluster tests do
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
	// No fake server needed — validation should fail before any HTTP call is made
	c, err := ims.NewClient(&ims.ClientConfig{URL: "http://ims.endpoint"})
	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	// Test with a missing required field — for example, empty SubjectToken
	_, err = c.OBOExchange(&ims.OBOExchangeRequest{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		SubjectToken: "", // missing — this is the subject token, OBO can't work without it
		Scopes:       []string{"openid"},
	})
	if err == nil {
		t.Fatalf("expected error for missing SubjectToken")
	}
}
