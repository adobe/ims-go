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

func TestClusterExchange(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("invalid method: %v", r.Method)
		}
		if r.URL.Path != "/ims/token/v3" {
			t.Fatalf("invalid path: %v", r.URL.Path)
		}
		v, ok := r.URL.Query()["client_id"]
		if !ok || v[0] != "client-id" {
			t.Fatalf("invalid client ID: %v", v)
		}

		if err := r.ParseForm(); err != nil {
			t.Fatalf("parse form: %v", err)
		}
		if v := r.PostForm.Get("grant_type"); v != "cluster_at_exchange" {
			t.Fatalf("incorrect grant type")
		}
		if v := r.PostForm.Get("user_token"); v == "" {
			t.Fatalf("missing user token")
		}
		if v := r.PostForm.Get("client_secret"); v != "client-secret" {
			t.Fatalf("invalid client secret: %v", v)
		}
		if v := r.PostForm.Get("owning_org_id"); v != "orgid" {
			t.Fatalf("invalid IMS Org ID: %v", v)
		}
		if v := r.PostForm.Get("scope"); v != "yolo,test" {
			t.Fatalf("invalid scopes: %v", v)
		}

		body := struct {
			TokenType   string `json:"token_type"`
			AccessToken string `json:"access_token"`
			ExpiresIn   int    `json:"expires_in"`
		}{
			TokenType:   "bearer",
			AccessToken: "new-token",
			ExpiresIn:   3600000,
		}

		if err := json.NewEncoder(w).Encode(&body); err != nil {
			t.Fatalf("encode response: %v", err)
		}
	}))
	defer s.Close()

	c, err := ims.NewClient(&ims.ClientConfig{
		URL: s.URL,
	})
	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	r, err := c.ClusterExchange(&ims.ClusterExchangeRequest{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		Scopes:       []string{"yolo", "test"},
		UserToken:    "old-token",
		OrgID:        "orgid",
	})
	if err != nil {
		t.Fatalf("failure exchanging access token: %v", err)
	}
	if r.AccessToken != "new-token" {
		t.Fatalf("invalid access token: %v", r.AccessToken)
	}
	if r.ExpiresIn != 3600*time.Second {
		t.Fatalf("invalid expiration: %v", r.ExpiresIn)
	}
}

func TestClusterExchangeError(t *testing.T) {
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

	c, err := ims.NewClient(&ims.ClientConfig{
		URL: s.URL,
	})
	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	_, err = c.ClusterExchange(&ims.ClusterExchangeRequest{
		ClientID:     "irrelevant",
		ClientSecret: "irrelevant",
		Scopes:       []string{"yolo", "test"},
		UserToken:    "old-token",
		OrgID:        "orgid",
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

func TestClusterExchangeInvalidRequest(t *testing.T) {
	c, err := ims.NewClient(&ims.ClientConfig{
		URL: "http://ims.endpoint",
	})
	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	_, err = c.ClusterExchange(&ims.ClusterExchangeRequest{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		Scopes:       []string{"yolo", "test"},
		UserToken:    "old-token",
		OrgID:        "orgid",
		UserID:       "userid",
	})
	if err == nil {
		t.Fatalf("expected error")
	}
	if err.Error() != "userID and OrgID defined at the same time" {
		t.Fatalf("invalid error: %v", err)
	}
}
