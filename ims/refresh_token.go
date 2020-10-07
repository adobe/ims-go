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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"strings"
	"time"
)

// RefreshTokenRequest is the request for refreshing an access token.
type RefreshTokenRequest struct {
	// RefreshToken is the refresh token obtained during the first request for
	// an access token. This field is required.
	RefreshToken string
	// ClientID is the client ID. This field is required.
	ClientID string
	// ClientSecret is the client secret. This field is required.
	ClientSecret string
	// Scope is the scope list in the refresh token. This field is optional. If
	// provided, it must be a subset of the scopes in the request token.
	Scope []string
}

// RefreshTokenResponse is the response of an access token refresh.
type RefreshTokenResponse struct {
	// Body is the raw response body.
	Body []byte
	// AccessToken is the new access token.
	AccessToken string
	// RefreshToken is a new refresh token.
	RefreshToken string
	// ExpiresIn is the expiration time for the access token.
	ExpiresIn time.Duration
}

// RefreshToken refreshes an access token.
func (c *Client) RefreshToken(r *RefreshTokenRequest) (*RefreshTokenResponse, error) {
	if r.RefreshToken == "" {
		return nil, fmt.Errorf("missing refresh token")
	}

	if r.ClientID == "" {
		return nil, fmt.Errorf("missing client ID")
	}

	if r.ClientSecret == "" {
		return nil, fmt.Errorf("missing client secret")
	}

	data := url.Values{}

	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", r.RefreshToken)
	data.Set("client_id", r.ClientID)
	data.Set("client_secret", r.ClientSecret)

	if len(r.Scope) > 0 {
		data.Set("scope", strings.Join(r.Scope, ","))
	}

	res, err := c.client.PostForm(fmt.Sprintf("%s/ims/token/v2", c.url), data)
	if err != nil {
		return nil, fmt.Errorf("perform request: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, errorResponse(res)
	}

	raw, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %v", err)
	}

	var payload struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
	}

	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, fmt.Errorf("decode response: %v", err)
	}

	return &RefreshTokenResponse{
		Body:         raw,
		AccessToken:  payload.AccessToken,
		RefreshToken: payload.RefreshToken,
		ExpiresIn:    time.Second * time.Duration(payload.ExpiresIn),
	}, nil
}
