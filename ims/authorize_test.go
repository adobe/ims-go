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
	"net/url"
	"testing"

	"github.com/adobe/ims-go/ims"
)

func TestAuthorizeURL(t *testing.T) {
	c, err := ims.NewClient(&ims.ClientConfig{
		URL: "http://ims.endpoint",
	})
	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	u, err := c.AuthorizeURL(&ims.AuthorizeURLConfig{
		ClientID:    "clientID",
		Scope:       []string{"one", "two"},
		RedirectURI: "http://redirect.uri",
		State:       "state-value",
	})
	if err != nil {
		t.Fatalf("authorize: %v", err)
	}

	url, err := url.Parse(u)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	if url.Scheme != "http" {
		t.Errorf("invalid scheme: %v", url.Scheme)
	}
	if url.Host != "ims.endpoint" {
		t.Errorf("invalid host: %v", url.Host)
	}

	q := url.Query()

	if v := q.Get("client_id"); v != "clientID" {
		t.Errorf("invalid client ID: %v", v)
	}
	if v := q.Get("scope"); v != "one,two" {
		t.Errorf("invalid scope: %v", v)
	}
	if v := q.Get("redirect_uri"); v != "http://redirect.uri" {
		t.Errorf("invalid redirect URI: %v", v)
	}
	if v := q.Get("state"); v != "state-value" {
		t.Errorf("invalid state: %v", v)
	}
}

func TestAuthorizeURLNoClientID(t *testing.T) {
	c, err := ims.NewClient(&ims.ClientConfig{
		URL: "http://ims.endpoint",
	})
	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	if _, err := c.AuthorizeURL(&ims.AuthorizeURLConfig{
		Scope: []string{"one", "two"},
	}); err == nil {
		t.Fatalf("expected error")
	} else if err.Error() != "missing client ID" {
		t.Fatalf("invalid error: %v", err)
	}
}

func TestAuthorizeURLNoScope(t *testing.T) {
	c, err := ims.NewClient(&ims.ClientConfig{
		URL: "http://ims.endpoint",
	})
	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	if _, err := c.AuthorizeURL(&ims.AuthorizeURLConfig{
		ClientID: "clientID",
	}); err == nil {
		t.Fatalf("expected error")
	} else if err.Error() != "missing scope" {
		t.Fatalf("invalid error: %v", err)
	}
}
