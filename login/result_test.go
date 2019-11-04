// Copyright 2019 Adobe. All rights reserved.
// This file is licensed to you under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License. You may obtain a copy
// of the License at http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software distributed under
// the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR REPRESENTATIONS
// OF ANY KIND, either express or implied. See the License for the specific language
// governing permissions and limitations under the License.

package login

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/adobe/ims-go/ims"
)

func TestResult(t *testing.T) {
	resCh := make(chan *ims.TokenResponse)

	h := &resultHandler{
		resCh: resCh,
	}

	w := httptest.NewRecorder()

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r = r.WithContext(context.WithValue(r.Context(), contextKeyResult, &ims.TokenResponse{}))

	go h.ServeHTTP(w, r)

	if res := <-resCh; res == nil {
		t.Fatalf("expected a response")
	}

	if s := string(w.Body.Bytes()); s != "Success!" {
		t.Fatalf("invalid body: %v", s)
	}
}

func TestResultError(t *testing.T) {
	errCh := make(chan error)

	h := &resultHandler{
		errCh: errCh,
	}

	w := httptest.NewRecorder()

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r = r.WithContext(context.WithValue(r.Context(), contextKeyError, fmt.Errorf("error")))

	go h.ServeHTTP(w, r)

	if err := <-errCh; err == nil {
		t.Fatalf("error expected")
	} else if err.Error() != "error" {
		t.Fatalf("invalid error: %v", err)
	}

	if s := string(w.Body.Bytes()); s != "Error: error" {
		t.Fatalf("invalid body: %v", s)
	}
}

func TestResultSuccessHandler(t *testing.T) {
	resCh := make(chan *ims.TokenResponse)

	h := &resultHandler{
		resCh: resCh,
		successHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, "custom")
		}),
	}

	w := httptest.NewRecorder()

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r = r.WithContext(context.WithValue(r.Context(), contextKeyResult, &ims.TokenResponse{}))

	go h.ServeHTTP(w, r)

	<-resCh

	if s := string(w.Body.Bytes()); s != "custom" {
		t.Fatalf("invalid body: %v", s)
	}
}

func TestResultFailureHandler(t *testing.T) {
	errCh := make(chan error)

	h := &resultHandler{
		errCh: errCh,
		failureHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, "custom")
		}),
	}

	w := httptest.NewRecorder()

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r = r.WithContext(context.WithValue(r.Context(), contextKeyError, fmt.Errorf("error")))

	go h.ServeHTTP(w, r)

	<-errCh

	if s := string(w.Body.Bytes()); s != "custom" {
		t.Fatalf("invalid body: %v", s)
	}
}

func TestResultInvalidContext(t *testing.T) {
	errCh := make(chan error)

	h := &resultHandler{
		errCh: errCh,
	}

	w := httptest.NewRecorder()

	r := httptest.NewRequest(http.MethodGet, "/", nil)

	go h.ServeHTTP(w, r)

	if err := <-errCh; err == nil {
		t.Fatalf("error expected")
	} else if err.Error() != "neither error nor result returned" {
		t.Fatalf("invalid error: %v", err)
	}
}
