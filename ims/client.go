// Copyright 2019 Adobe. All rights reserved.
// This file is licensed to you under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License. You may obtain a copy
// of the License at http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software distributed under
// the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR REPRESENTATIONS
// OF ANY KIND, either express or implied. See the License for the specific language
// governing permissions and limitations under the License.

package ims

import (
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// ClientConfig is the configuration for a Client.
type ClientConfig struct {
	// URL is the endpoint for the IMS API.
	URL string
	// Client is the HTTP client to use when performing requests. If not
	// provided, the default HTTP client is used.
	Client *http.Client
}

// Client is the client for the IMS API.
type Client struct {
	url    string
	client *http.Client
}

// NewClient creates a new Client for the given configuration.
func NewClient(cfg *ClientConfig) (*Client, error) {
	client := cfg.Client

	if client == nil {
		client = &http.Client{
			Timeout: 30 * time.Second,
		}
	}

	endpointURL, err := url.Parse(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("malformed URL")
	}

	if endpointURL.Scheme == "" {
		return nil, fmt.Errorf("missing URL scheme")
	}

	if endpointURL.Host == "" {
		return nil, fmt.Errorf("missing URL host")
	}

	endpointURL.User = nil
	endpointURL.RawQuery = ""
	endpointURL.Fragment = ""

	return &Client{
		url:    endpointURL.String(),
		client: client,
	}, nil
}
