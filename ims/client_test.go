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
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/adobe/ims-go/ims"
)

type errorTestHTTPClient func(*http.Request) (*http.Response, error)

func (f errorTestHTTPClient) Do(r *http.Request) (*http.Response, error) {
	return f(r)
}

type errorTestResponseBody struct {
	io.Reader
	closeErr error
}

func (b errorTestResponseBody) Close() error {
	return b.closeErr
}

func TestTokenErrorCauses(t *testing.T) {
	readErr := errors.New("read failed")
	closeErr := errors.New("close failed")

	tests := []struct {
		name         string
		transportErr error
		readErr      error
		closeErr     error
		wantMessage  string
	}{
		{name: "cancellation", transportErr: context.Canceled, wantMessage: "perform request: context canceled"},
		{name: "deadline", transportErr: context.DeadlineExceeded, wantMessage: "perform request: context deadline exceeded"},
		{name: "read", readErr: readErr, wantMessage: "perform request: read response body: read failed"},
		{name: "close", closeErr: closeErr, wantMessage: "perform request: close body: close failed"},
		{name: "read and close", readErr: readErr, closeErr: closeErr, wantMessage: "perform request: read response body: read failed\nclose body: close failed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := ims.NewClient(&ims.ClientConfig{
				URL: "https://ims.example",
				Client: errorTestHTTPClient(func(*http.Request) (*http.Response, error) {
					if tt.transportErr != nil {
						return nil, tt.transportErr
					}
					var reader io.Reader = strings.NewReader("{}")
					if tt.readErr != nil {
						reader = iotest.ErrReader(tt.readErr)
					}
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       errorTestResponseBody{Reader: reader, closeErr: tt.closeErr},
					}, nil
				}),
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
				t.Fatal("expected error")
			}
			if err.Error() != tt.wantMessage {
				t.Fatalf("unexpected error: %v", err)
			}
			for _, cause := range []error{tt.transportErr, tt.readErr, tt.closeErr} {
				if cause != nil && !errors.Is(err, cause) {
					t.Errorf("expected error to wrap %v, got %v", cause, err)
				}
			}
		})
	}
}

func TestNewClientNoScheme(t *testing.T) {
	_, err := ims.NewClient(&ims.ClientConfig{
		URL: "ims.endpoint",
	})
	if err == nil {
		t.Fatalf("expected error")
	}
	if err.Error() != "missing URL scheme" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewClientNoHost(t *testing.T) {
	_, err := ims.NewClient(&ims.ClientConfig{
		URL: "https:///path",
	})
	if err == nil {
		t.Fatalf("expected error")
	}
	if err.Error() != "missing URL host" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewClientMalformedURL(t *testing.T) {
	_, err := ims.NewClient(&ims.ClientConfig{
		URL: ":",
	})
	if err == nil {
		t.Fatalf("expected error")
	}
	if err.Error() != "malformed URL" {
		t.Fatalf("unexpected error: %v", err)
	}
}
