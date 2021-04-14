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
	"context"
	"encoding/json"
	"fmt"
	"net/http"
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
	Response
	// AccessToken is the access token.
	AccessToken string
	// RefreshToken is the refresh token.
	RefreshToken string
	// ExpiresIn is the expiration time of the access token.
	ExpiresIn time.Duration
	// User id received from IMS token
	UserID string
}

// TokenWithContext requests an access token.
func (c *Client) TokenWithContext(ctx context.Context, r *TokenRequest) (*TokenResponse, error) {
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

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/ims/token/v2", c.url), strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := c.do(req)
	if err != nil {
		return nil, fmt.Errorf("perform request: %v", err)
	}

	if res.StatusCode != http.StatusOK {
		return nil, errorResponse(res)
	}

	var payload struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
		UserID       string `json:"userId"`
	}

	if err := json.Unmarshal(res.Body, &payload); err != nil {
		return nil, fmt.Errorf("decode response: %v", err)
	}

	return &TokenResponse{
		Response:     *res,
		AccessToken:  payload.AccessToken,
		RefreshToken: payload.RefreshToken,
		ExpiresIn:    time.Second * time.Duration(payload.ExpiresIn),
		UserID:       payload.UserID,
	}, nil
}

// Token is equivalent to TokenWithContext with a background context.
func (c *Client) Token(r *TokenRequest) (*TokenResponse, error) {
	return c.TokenWithContext(context.Background(), r)
}
