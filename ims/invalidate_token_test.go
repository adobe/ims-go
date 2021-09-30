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

func TestInvalidateToken(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only POST accepted
		if r.Method != http.MethodPost {
			t.Fatalf("invalid method: %v", r.Method)
		}

		if r.URL.Path != "/ims/invalidate_token/v2" {
			t.Fatalf("invalid path: %v", r.URL.Path)
		}

		// Test X-IMS-ClientId header
		if h := r.Header.Get("X-IMS-ClientId"); h != "test_client_id" {
			t.Fatalf("invalid X-IMS-ClientId header: %v", h)
		}

		if err := r.ParseForm(); err != nil {
			t.Fatalf("parse form: %v", err)
		}

		fmt.Printf("Form\n")
		for key, value := range r.Form {
			fmt.Printf("%s = %s\n", key, value)
		}
		fmt.Printf("PostForm\n")
		for key, value := range r.PostForm {
			fmt.Printf("%s = %s\n", key, value)
		}

		// Test form mandatory values
		if v := r.PostForm.Get("client_id"); v != "test_client_id" {
			t.Fatalf("incorrect client id: %v", v)
		}
		if v := r.PostForm.Get("token"); v != "YOLO" {
			t.Fatalf("incorrect token: %v", v)
		}

		// Test specific parameters for specific token types
		tokenType := r.PostForm.Get("token_type")
		switch tokenType {
		case string(ims.AccessToken), string(ims.RefreshToken), string(ims.DeviceToken):
		case string(ims.ServiceToken):
			if v := r.PostForm.Get("client_secret"); v != "SECRET" {
				t.Fatalf("incorrect client_secret: %v", v)
			}
		default:
			t.Fatalf("incorrect type of token: %v", tokenType)
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

	// Test Access Token
	err = c.InvalidateToken(&ims.InvalidateTokenRequest{
		Token:    "YOLO",
		Type:     ims.AccessToken,
		ClientID: "test_client_id",
	})
	if err != nil {
		t.Fatalf("invalidate access token: %v", err)
	}

	// Test Refresh Token, use cascade parameter
	err = c.InvalidateToken(&ims.InvalidateTokenRequest{
		Token:     "YOLO",
		Type:      ims.RefreshToken,
		ClientID:  "test_client_id",
		Cascading: true,
	})
	if err != nil {
		t.Fatalf("invalidate refresh token: %v", err)
	}

	// Test Device Token, don't use cascade parameter
	err = c.InvalidateToken(&ims.InvalidateTokenRequest{
		Token:    "YOLO",
		Type:     ims.DeviceToken,
		ClientID: "test_client_id",
	})
	if err != nil {
		t.Fatalf("invalidate refresh token: %v", err)
	}

	// Test Service Token, client_secret is mandatory
	err = c.InvalidateToken(&ims.InvalidateTokenRequest{
		Token:        "YOLO",
		Type:         ims.ServiceToken,
		ClientID:     "test_client_id",
		ClientSecret: "SECRET",
	})
	if err != nil {
		t.Fatalf("invalidate service token: %v", err)
	}
}

func TestInvalidateTokenEmptyErrorResponse(t *testing.T) {
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

	err = c.InvalidateToken(&ims.InvalidateTokenRequest{
		Type: ims.AccessToken,
	})
	if err == nil {
		t.Fatalf("nil error returned, when error is expected")
	}
	if err.Error() != "missing client ID parameter" {
		t.Fatalf("invalid error: %v", err)
	}
}
