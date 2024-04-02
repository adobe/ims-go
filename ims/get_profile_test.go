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

func TestGetProfile(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("invalid method: %v", r.Method)
		}
		if v := r.Header.Get("authorization"); v != "Bearer accessToken" {
			t.Fatalf("invalid authorization header: %v", v)
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

	res, err := c.GetProfile(&ims.GetProfileRequest{
		AccessToken: "accessToken",
	})
	if err != nil {
		t.Fatalf("get profile: %v", err)
	}

	if body := string(res.Body); body != `{"foo":"bar"}` {
		t.Fatalf("invalid body: %v", body)
	}
}

func TestGetProfileEmptyErrorResponse(t *testing.T) {
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

	res, err := c.GetProfile(&ims.GetProfileRequest{})
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

func TestGetProfileTooManyRequests(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("invalid method: %v", r.Method)
		}
		if v := r.Header.Get("authorization"); v != "Bearer accessToken" {
			t.Fatalf("invalid authorization header: %v", v)
		}

		w.Header().Set("Retry-After", "77")
		w.Header().Set("x-debug-id", "banana")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer s.Close()

	c, err := ims.NewClient(&ims.ClientConfig{
		URL: s.URL,
	})
	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	res, err := c.GetProfile(&ims.GetProfileRequest{
		AccessToken: "accessToken",
	})

	if err == nil {
		t.Fatalf("expected error in get profile")
	}

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
