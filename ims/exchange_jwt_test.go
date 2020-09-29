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
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/adobe/ims-go/ims"
)

func newPrivateKey(t *testing.T) []byte {
	t.Helper()

	k, err := rsa.GenerateKey(rand.Reader, 512)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	return pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(k),
	})
}

func TestExchangeJWT(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("invalid method: %v", r.Method)
		}
		if r.URL.Path != "/ims/exchange/v1/jwt" {
			t.Fatalf("invalid path: %v", r.URL.Path)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("parse form: %v", err)
		}
		if v := r.PostForm.Get("jwt_token"); v == "" {
			t.Fatalf("missing JWT token")
		}
		if v := r.PostForm.Get("client_id"); v != "client-id" {
			t.Fatalf("invalid client ID: %v", v)
		}
		if v := r.PostForm.Get("client_secret"); v != "client-secret" {
			t.Fatalf("invalid client secret: %v", v)
		}

		body := struct {
			AccessToken string `json:"access_token"`
			ExpiresIn   int    `json:"expires_in"`
		}{
			AccessToken: "access-token",
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

	r, err := c.ExchangeJWT(&ims.ExchangeJWTRequest{
		PrivateKey:   newPrivateKey(t),
		Expiration:   time.Now().Add(24 * time.Hour),
		Issuer:       "organization",
		Subject:      "technical-user",
		ClientID:     "client-id",
		ClientSecret: "client-secret",
	})
	if err != nil {
		t.Fatalf("exchange JWT: %v", err)
	}
	if r.AccessToken != "access-token" {
		t.Fatalf("invalid access token: %v", r.AccessToken)
	}
	if r.ExpiresIn != 3600*time.Second {
		t.Fatalf("invalid expiration: %v", r.ExpiresIn)
	}
}

func TestExchangeJWTError(t *testing.T) {
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

	_, err = c.ExchangeJWT(&ims.ExchangeJWTRequest{
		PrivateKey:   newPrivateKey(t),
		Expiration:   time.Now().Add(24 * time.Hour),
		Issuer:       "organization",
		Subject:      "technical-user",
		ClientID:     "client-id",
		ClientSecret: "client-secret",
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

func TestExchangeJWTInvalidMetaScope(t *testing.T) {
	c, err := ims.NewClient(&ims.ClientConfig{
		URL: "http://ims.endpoint",
	})
	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	_, err = c.ExchangeJWT(&ims.ExchangeJWTRequest{
		PrivateKey:   newPrivateKey(t),
		Expiration:   time.Now().Add(24 * time.Hour),
		Issuer:       "organization",
		Subject:      "technical-user",
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		MetaScope:    []ims.MetaScope{-1},
	})
	if err == nil {
		t.Fatalf("expected error")
	}
	if err.Error() != "invalid meta-scope: -1" {
		t.Fatalf("invalid error: %v", err)
	}
}