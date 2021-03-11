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
	"testing"

	"github.com/adobe/ims-go/ims"
)

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
