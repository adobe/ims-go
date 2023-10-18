// Copyright 2021 Adobe. All rights reserved.
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
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/adobe/ims-go/ims"
)

func TestValidateToken(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only POST accepted
		if r.Method != http.MethodPost {
			t.Fatalf("invalid method: %v", r.Method)
		}

		// Test X-IMS-ClientId header
		if h := r.Header.Get("X-IMS-ClientId"); h != "test_client_id" {
			t.Fatalf("invalid X-IMS-ClientId header: %v", h)
		}

		if err := r.ParseForm(); err != nil {
			t.Fatalf("parse form: %v", err)
		}

		if v := r.PostForm.Get("client_id"); v != "test_client_id" {
			t.Fatalf("missing client ID: %v", v)
		}

		if v := r.PostForm.Get("type"); v == "" {
			t.Fatalf("missing type: %v", v)
		}

		var tokenType = ims.TokenType(r.PostForm.Get("type"))

		switch tokenType {
		case ims.AccessToken, ims.RefreshToken, ims.DeviceToken, ims.AuthorizationCode:
		default:
			t.Fatalf("incorrect type of token")
		}
		fmt.Fprint(w, `{"foo":"bar"}`)
	}))
	defer s.Close()

	c, err := ims.NewClient(&ims.ClientConfig{
		URL: s.URL,
	})
	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	res, err := c.ValidateToken(&ims.ValidateTokenRequest{
		Token:    "YOLO",
		Type:     ims.AccessToken,
		ClientID: "test_client_id",
	})
	if err != nil {
		t.Fatalf("validate token: %v", err)
	}

	if body := string(res.Body); body != `{"foo":"bar"}` {
		t.Fatalf("invalid body: %v", body)
	}
}

func TestValidateTokenEmptyErrorResponse(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer s.Close()

	c, err := ims.NewClient(&ims.ClientConfig{
		URL: s.URL,
	})
	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	res, err := c.ValidateToken(&ims.ValidateTokenRequest{
		Type: ims.AccessToken,
	})
	if res != nil {
		t.Fatalf("non-nil response returned")
	}
	if err == nil {
		t.Fatalf("nil error returned")
	}
	if _, ok := ims.IsError(err); !ok {
		t.Fatalf("invalid error type: %v", err)
	}
}
