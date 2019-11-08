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
	"net/url"
	"strings"
	"time"
)

// TokenRequest is the request for obtaining an access token.
type TokenRequest struct {
	// Code is the authorization code obtained via the authorization workflow.
	// This field is required.
	Code string
	// ClientID is the client ID. This field is required.
	ClientID string
	// ClientSecret is the client secret. This field is required.
	ClientSecret string
	// Scope is the scope of list for the access token. This field is optional.
	// If not provided, the scopes will be bound to the ones requested during
	// the authorization workflow.
	Scope []string
}

// TokenResponse is the response returned after an access token request.
type TokenResponse struct {
	// AccessToken is the access token.
	AccessToken string
	// RefreshToken is the refresh token.
	RefreshToken string
	// ExpiresIn is the expiration time of the access token.
	ExpiresIn time.Duration
	// Raw ims response
	IMSRawResponse map[string]interface{}
}

// Token requests an access token.
func (c *Client) Token(r *TokenRequest) (*TokenResponse, error) {
	if r.Code == "" {
		return nil, fmt.Errorf("missing code")
	}

	if r.ClientID == "" {
		return nil, fmt.Errorf("missing client ID")
	}

	if r.ClientSecret == "" {
		return nil, fmt.Errorf("missing client secret")
	}

	data := url.Values{}

	data.Set("grant_type", "authorization_code")
	data.Set("code", r.Code)
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

	tokenMapResponse := make(map[string]interface{})

	if err := json.NewDecoder(res.Body).Decode(&tokenMapResponse); err != nil {
		return nil, fmt.Errorf("decode response: %v", err)
	}

	return &TokenResponse{
		AccessToken:    tokenMapResponse["access_token"].(string),
		RefreshToken:   tokenMapResponse["refresh_token"].(string),
		ExpiresIn:      time.Second * time.Duration(tokenMapResponse["expires_in"].(float64)),
		IMSRawResponse: tokenMapResponse,
	}, nil
}
