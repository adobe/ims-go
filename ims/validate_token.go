// Copyright 2021 Adobe. All rights reserved.
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
)

type TokenType string

const (
	AccessToken       TokenType = "access_token"
	RefreshToken      TokenType = "refresh_token"
	DeviceToken       TokenType = "device_token"
	AuthorizationCode TokenType = "authorization_code"
)

// ValidateTokenRequest is the request to ValidateToken.
type ValidateTokenRequest struct {
	// AccessToken is a valid access token.
	Token    string
	Type     TokenType
	ClientID string
}

// ValidateTokenResponse is the response to the ValidateToken request .
type ValidateTokenResponse struct {
	Response
	Valid bool
}

// ValidateTokenWithContext validates a token using the IMS API. It returns a
// non-nil response on success or an error on failure.
func (c *Client) ValidateTokenWithContext(ctx context.Context, r *ValidateTokenRequest) (*ValidateTokenResponse, error) {
	switch r.Type {
	case AccessToken, RefreshToken, DeviceToken, AuthorizationCode:
		// Valid token type.
	default:
		return nil, fmt.Errorf("invalid token type: %v", r.Type)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/ims/validate_token/v1", c.url), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %v", err)
	}

	query := req.URL.Query()

	query.Set("type", string(r.Type))
	query.Set("client_id", r.ClientID)
	query.Set("token", r.Token)

	req.URL.RawQuery = query.Encode()

	// Header X-IMS-ClientID will be mandatory in the future
	req.Header.Set("X-IMS-ClientId", r.ClientID)

	res, err := c.do(req)
	if err != nil {
		return nil, fmt.Errorf("perform request: %v", err)
	}

	if res.StatusCode != http.StatusOK {
		return nil, errorResponse(res)
	}

	var payload struct {
		Valid bool `json:"valid"`
	}

	if err := json.Unmarshal(res.Body, &payload); err != nil {
		return nil, fmt.Errorf("error parsing response: %v", err)
	}

	return &ValidateTokenResponse{
		Response: *res,
		Valid:    payload.Valid,
	}, nil
}

// ValidateToken is equivalent to ValidateTokenWithContext with a background
// context.
func (c *Client) ValidateToken(r *ValidateTokenRequest) (*ValidateTokenResponse, error) {
	return c.ValidateTokenWithContext(context.Background(), r)
}
