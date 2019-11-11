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

func TestToken(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("invalid method: %v", r.Method)
		}
		if r.URL.Path != "/ims/token/v2" {
			t.Fatalf("invalid path: %v", r.URL.Path)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("parse form: %v", err)
		}
		if v := r.PostForm.Get("grant_type"); v != "authorization_code" {
			t.Fatalf("invalid grant type: %v", v)
		}
		if v := r.PostForm.Get("code"); v != "code" {
			t.Fatalf("invalid code: %v", v)
		}
		if v := r.PostForm.Get("client_id"); v != "clientID" {
			t.Fatalf("invalid client ID: %v", v)
		}
		if v := r.PostForm.Get("client_secret"); v != "clientSecret" {
			t.Fatalf("invalid client secret: %v", v)
		}
		if v := r.PostForm.Get("scope"); v != "a,b" {
			t.Fatalf("invalid scope: %v", v)
		}

		body := struct {
			ExpiresIn    int    `json:"expires_in"`
			RefreshToken string `json:"refresh_token"`
			AccessToken  string `json:"access_token"`
			UserId       string `json:"userId"`
		}{
			ExpiresIn:    3600,
			RefreshToken: "refreshToken",
			AccessToken:  "accessToken",
			UserId:       "user-id",
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

	r, err := c.Token(&ims.TokenRequest{
		Code:         "code",
		ClientID:     "clientID",
		ClientSecret: "clientSecret",
		Scope:        []string{"a", "b"},
	})
	if err != nil {
		t.Fatalf("token: %v", err)
	}
	if r.AccessToken != "accessToken" {
		t.Errorf("invalid access token: %v", r.AccessToken)
	}
	if r.RefreshToken != "refreshToken" {
		t.Errorf("invalid refresh token: %v", r.RefreshToken)
	}
	if r.ExpiresIn != 3600*time.Second {
		t.Errorf("invalid expiration: %v", r.ExpiresIn)
	}
	if r.UserID != "user-id" {
		t.Errorf("invalid userId: %v", r.UserID)
	}
}

func TestTokenError(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)

		body := struct {
			ErrorCode    string `json:"error"`
			ErrorMessage string `json:"error_description"`
		}{
			ErrorCode:    "errorCode",
			ErrorMessage: "errorMessage",
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

	_, err = c.Token(&ims.TokenRequest{
		Code:         "code",
		ClientID:     "clientID",
		ClientSecret: "clientSecret",
	})
	if err == nil {
		t.Fatalf("no error returned")
	}

	imsErr, ok := ims.IsError(err)
	if !ok {
		t.Fatalf("invalid error: %v", err)
	}
	if imsErr.StatusCode != http.StatusBadRequest {
		t.Fatalf("invalid status code: %v", imsErr.StatusCode)
	}
	if imsErr.ErrorCode != "errorCode" {
		t.Fatalf("invalid error code: %v", imsErr.ErrorCode)
	}
	if imsErr.ErrorMessage != "errorMessage" {
		t.Fatalf("invalid error message: %v", imsErr.ErrorMessage)
	}
}

func TestTokenNoCode(t *testing.T) {
	c, err := ims.NewClient(&ims.ClientConfig{
		URL: "http://ims.endpoint",
	})
	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	if _, err := c.Token(&ims.TokenRequest{
		ClientID:     "clientID",
		ClientSecret: "clientSecret",
	}); err == nil {
		t.Fatalf("no error returned")
	} else if err.Error() != "missing code" {
		t.Fatalf("invalid error: %v", err)
	}
}

func TestTokenNoClientID(t *testing.T) {
	c, err := ims.NewClient(&ims.ClientConfig{
		URL: "http://ims.endpoint",
	})
	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	if _, err := c.Token(&ims.TokenRequest{
		Code:         "code",
		ClientSecret: "clientSecret",
	}); err == nil {
		t.Fatalf("no error returned")
	} else if err.Error() != "missing client ID" {
		t.Fatalf("invalid error: %v", err)
	}
}

func TestTokenNoClientSecret(t *testing.T) {
	c, err := ims.NewClient(&ims.ClientConfig{
		URL: "http://ims.endpoint",
	})
	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	if _, err := c.Token(&ims.TokenRequest{
		Code:     "code",
		ClientID: "clientID",
	}); err == nil {
		t.Fatalf("no error returned")
	} else if err.Error() != "missing client secret" {
		t.Fatalf("invalid error: %v", err)
	}
}
