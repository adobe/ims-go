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

	"github.com/adobe/ims-go/ims"
)

func TestDCR(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("invalid method: %v", r.Method)
		}

		if r.URL.Path != "/ims/register" {
			t.Fatalf("invalid path: %v", r.URL.Path)
		}

		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Fatalf("invalid Content-Type: %v", ct)
		}

		var body struct {
			ClientName   string   `json:"client_name"`
			RedirectURIs []string `json:"redirect_uris"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body.ClientName != "my-app" {
			t.Fatalf("invalid client_name: %v", body.ClientName)
		}
		if len(body.RedirectURIs) != 1 || body.RedirectURIs[0] != "https://example.com/callback" {
			t.Fatalf("invalid redirect_uris: %v", body.RedirectURIs)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"client_id":"new-client-id","client_secret":"new-secret"}`))
	}))
	defer s.Close()

	c, err := ims.NewClient(&ims.ClientConfig{URL: s.URL})
	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	resp, err := c.DCR(&ims.DCRRequest{
		ClientName:   "my-app",
		RedirectURIs: []string{"https://example.com/callback"},
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	if string(resp.Body) != `{"client_id":"new-client-id","client_secret":"new-secret"}` {
		t.Fatalf("invalid body: %v", string(resp.Body))
	}
}

func TestDCRWithContext(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("invalid method: %v", r.Method)
		}
		if r.URL.Path != "/ims/register" {
			t.Fatalf("invalid path: %v", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"client_id":"ctx-client"}`))
	}))
	defer s.Close()

	c, err := ims.NewClient(&ims.ClientConfig{URL: s.URL})
	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	resp, err := c.DCRWithContext(context.Background(), &ims.DCRRequest{
		ClientName:   "my-app",
		RedirectURIs: []string{"https://example.com/callback"},
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	if string(resp.Body) != `{"client_id":"ctx-client"}` {
		t.Fatalf("invalid body: %v", string(resp.Body))
	}
}

func TestDCR2xxRange(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"client_id":"created-client"}`))
	}))
	defer s.Close()

	c, err := ims.NewClient(&ims.ClientConfig{URL: s.URL})
	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	resp, err := c.DCR(&ims.DCRRequest{
		ClientName:   "my-app",
		RedirectURIs: []string{"https://example.com/callback"},
	})
	if err != nil {
		t.Fatalf("expected success for 201: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected StatusCode 201, got %d", resp.StatusCode)
	}
}

func TestDCRError(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)

		body := struct {
			ErrorCode    string `json:"error"`
			ErrorMessage string `json:"error_description"`
		}{
			ErrorCode:    "invalid_request",
			ErrorMessage: "client_name is required",
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

	_, err = c.DCR(&ims.DCRRequest{
		ClientName:   "my-app",
		RedirectURIs: []string{"https://example.com/callback"},
	})

	imsErr, ok := ims.IsError(err)
	if !ok {
		t.Fatalf("expected IMS error")
	}
	if imsErr.StatusCode != http.StatusBadRequest {
		t.Fatalf("invalid status code: %v", imsErr.StatusCode)
	}
	if imsErr.ErrorCode != "invalid_request" {
		t.Fatalf("invalid error code: %v", imsErr.ErrorCode)
	}
	if imsErr.ErrorMessage != "client_name is required" {
		t.Fatalf("invalid error message: %v", imsErr.ErrorMessage)
	}
}

func TestDCRTooManyRequests(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "30")
		w.Header().Set("x-debug-id", "debug-xyz")
		w.WriteHeader(http.StatusTooManyRequests)

		body := struct {
			ErrorCode    string `json:"error"`
			ErrorMessage string `json:"error_description"`
		}{
			ErrorCode:    "rate_limit",
			ErrorMessage: "too many requests",
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

	res, err := c.DCR(&ims.DCRRequest{
		ClientName:   "my-app",
		RedirectURIs: []string{"https://example.com/callback"},
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
	if imsErr.RetryAfter != "30" {
		t.Fatalf("invalid retry-after header: %v", imsErr.RetryAfter)
	}
	if imsErr.XDebugID != "debug-xyz" {
		t.Fatalf("invalid x-debug-id header: %v", imsErr.XDebugID)
	}
}

func TestDCRInvalidRequest(t *testing.T) {
	c, err := ims.NewClient(&ims.ClientConfig{URL: "http://ims.endpoint"})
	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	tests := []struct {
		name    string
		request *ims.DCRRequest
		wantErr string
	}{
		{
			name:    "missing ClientName",
			request: &ims.DCRRequest{ClientName: "", RedirectURIs: []string{"https://example.com/cb"}},
			wantErr: "invalid parameters for client registration: missing client name parameter",
		},
		{
			name:    "nil RedirectURIs",
			request: &ims.DCRRequest{ClientName: "my-app", RedirectURIs: nil},
			wantErr: "invalid parameters for client registration: missing redirect URIs parameter",
		},
		{
			name:    "empty RedirectURIs",
			request: &ims.DCRRequest{ClientName: "my-app", RedirectURIs: []string{}},
			wantErr: "invalid parameters for client registration: missing redirect URIs parameter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := c.DCR(tt.request)
			if err == nil {
				t.Fatal("expected error")
			}
			if got := err.Error(); got != tt.wantErr {
				t.Fatalf("error = %q, want %q", got, tt.wantErr)
			}
		})
	}
}

func TestDCRMultipleRedirectURIs(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			ClientName   string   `json:"client_name"`
			RedirectURIs []string `json:"redirect_uris"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if len(body.RedirectURIs) != 2 {
			t.Fatalf("expected 2 redirect URIs, got %d", len(body.RedirectURIs))
		}
		if body.RedirectURIs[0] != "https://example.com/cb1" || body.RedirectURIs[1] != "https://example.com/cb2" {
			t.Fatalf("invalid redirect_uris: %v", body.RedirectURIs)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"client_id":"multi-uri-client"}`))
	}))
	defer s.Close()

	c, err := ims.NewClient(&ims.ClientConfig{URL: s.URL})
	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	resp, err := c.DCR(&ims.DCRRequest{
		ClientName:   "my-app",
		RedirectURIs: []string{"https://example.com/cb1", "https://example.com/cb2"},
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	if string(resp.Body) != `{"client_id":"multi-uri-client"}` {
		t.Fatalf("invalid body: %v", string(resp.Body))
	}
}

func TestDCRWithScopes(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			ClientName   string   `json:"client_name"`
			RedirectURIs []string `json:"redirect_uris"`
			Scope        string   `json:"scope"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body.Scope != "openid profile" {
			t.Fatalf("invalid scope: %v", body.Scope)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"client_id":"scoped-client"}`))
	}))
	defer s.Close()

	c, err := ims.NewClient(&ims.ClientConfig{URL: s.URL})
	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	resp, err := c.DCR(&ims.DCRRequest{
		ClientName:   "my-app",
		RedirectURIs: []string{"https://example.com/callback"},
		Scopes:       []string{"openid", "profile"},
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	if string(resp.Body) != `{"client_id":"scoped-client"}` {
		t.Fatalf("invalid body: %v", string(resp.Body))
	}
}

func TestDCRWithoutScopes(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if _, ok := body["scope"]; ok {
			t.Fatalf("expected no scope field in payload, got: %v", body["scope"])
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"client_id":"no-scope-client"}`))
	}))
	defer s.Close()

	c, err := ims.NewClient(&ims.ClientConfig{URL: s.URL})
	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	resp, err := c.DCR(&ims.DCRRequest{
		ClientName:   "my-app",
		RedirectURIs: []string{"https://example.com/callback"},
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	if string(resp.Body) != `{"client_id":"no-scope-client"}` {
		t.Fatalf("invalid body: %v", string(resp.Body))
	}
}
