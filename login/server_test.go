// Copyright 2019 Adobe. All rights reserved.
// This file is licensed to you under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License. You may obtain a copy
// of the License at http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software distributed under
// the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR REPRESENTATIONS
// OF ANY KIND, either express or implied. See the License for the specific language
// governing permissions and limitations under the License.

package login_test

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/adobe/ims-go/ims"
	"github.com/adobe/ims-go/login"
)

func TestServerLogin(t *testing.T) {
	lst, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	// Login backend setup

	mux := http.NewServeMux()

	mux.HandleFunc("/ims/authorize/v1", func(w http.ResponseWriter, r *http.Request) {
		v := url.Values{}
		v.Add("code", "code")
		v.Add("state", r.URL.Query().Get("state"))

		u := url.URL{
			Host:     fmt.Sprintf("localhost:%d", port(lst)),
			RawQuery: v.Encode(),
		}

		http.Redirect(w, r, u.String(), http.StatusFound)
	})

	mux.Handle("/ims/token/v2", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{
			"access_token": "access-token",
			"refresh_token": "refresh-token",
			"expires_in": 3600
		}`)
	}))

	backend := httptest.NewServer(mux)
	defer backend.Close()

	// Login server setup

	client, err := ims.NewClient(&ims.ClientConfig{
		URL: backend.URL,
	})
	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	server, err := login.NewServer(&login.ServerConfig{
		Client:       client,
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		Scope:        []string{"a", "b"},
	})
	if err != nil {
		t.Fatalf("create server: %v", err)
	}

	go server.Serve(lst)

	// User flow

	go func() {
		res, err := http.Get(fmt.Sprintf("http://localhost:%d/", port(lst)))
		if err != nil {
			t.Errorf("perform initial request: %v", err)
		}
		defer res.Body.Close()
	}()

	select {
	case res := <-server.Response():
		if res.AccessToken != "access-token" {
			t.Fatalf("invalid access token: %v", res.AccessToken)
		}
		if res.RefreshToken != "refresh-token" {
			t.Fatalf("invalid refresh token: %v", res.RefreshToken)
		}
		if res.ExpiresIn != 3600*time.Second {
			t.Fatalf("invalid expiration time: %v", res.ExpiresIn)
		}
	case err := <-server.Error():
		t.Fatalf("unexpected error: %v", err)
	}

	if err := server.Shutdown(context.Background()); err != nil {
		t.Fatalf("shutdown: %v", err)
	}
}

func TestServerError(t *testing.T) {
	lst, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	// Login backend setup

	mux := http.NewServeMux()

	mux.HandleFunc("/ims/authorize/v1", func(w http.ResponseWriter, r *http.Request) {
		v := url.Values{}
		v.Add("error", "error")

		u := url.URL{
			Host:     fmt.Sprintf("localhost:%d", port(lst)),
			RawQuery: v.Encode(),
		}

		http.Redirect(w, r, u.String(), http.StatusFound)
	})

	backend := httptest.NewServer(mux)
	defer backend.Close()

	// Login server setup

	client, err := ims.NewClient(&ims.ClientConfig{
		URL: backend.URL,
	})
	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	server, err := login.NewServer(&login.ServerConfig{
		Client:       client,
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		Scope:        []string{"a", "b"},
	})
	if err != nil {
		t.Fatalf("create server: %v", err)
	}

	go server.Serve(lst)

	// User flow

	go func() {
		res, err := http.Get(fmt.Sprintf("http://localhost:%d/", port(lst)))
		if err != nil {
			t.Errorf("perform initial request: %v", err)
		}
		defer res.Body.Close()
	}()

	select {
	case <-server.Response():
		t.Fatalf("error expected")
	case err := <-server.Error():
		if err.Error() != "backend error: error" {
			t.Fatalf("invalid error: %v", err)
		}
	}

	if err := server.Shutdown(context.Background()); err != nil {
		t.Fatalf("shutdown: %v", err)
	}
}

func port(lst net.Listener) int {
	return lst.Addr().(*net.TCPAddr).Port
}
