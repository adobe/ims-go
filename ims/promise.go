// Copyright 2026 Adobe. All rights reserved.
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

const (
	grantTypePromise         = "promise"
	grantTypePromiseExchange = "promise_exchange"
)

// PromiseTokenRequest contains the data for exchanging an access token for a
// promise token.
type PromiseTokenRequest struct {
	ClientID            string
	ClientSecret        string
	PromiseDefinitionID string
	AuthenticatingToken string
	Scopes              []string
}

// PromiseTokenResponse contains the response of a successful exchange of an
// access token for a promise token.
type PromiseTokenResponse struct {
	Response
	PromiseToken string
	TokenType    string
	Scope        string
	ExpiresIn    time.Duration
}

func (c *Client) validatePromiseTokenRequest(r *PromiseTokenRequest) error {
	switch {
	case r.ClientID == "":
		return fmt.Errorf("missing client ID parameter")
	case r.ClientSecret == "":
		return fmt.Errorf("missing client secret parameter")
	case r.PromiseDefinitionID == "":
		return fmt.Errorf("missing promise definition ID parameter")
	case r.AuthenticatingToken == "":
		return fmt.Errorf("missing authenticating token parameter")
	case len(r.Scopes) == 0 || (len(r.Scopes) == 1 && r.Scopes[0] == ""):
		return fmt.Errorf("scopes are required for promise token exchange")
	default:
		return nil
	}
}

// PromiseTokenWithContext exchanges an access token for a promise token.
func (c *Client) PromiseTokenWithContext(ctx context.Context, r *PromiseTokenRequest) (*PromiseTokenResponse, error) {
	if err := c.validatePromiseTokenRequest(r); err != nil {
		return nil, fmt.Errorf("invalid parameters for promise token exchange: %v", err)
	}

	data := url.Values{}
	data.Set("grant_type", grantTypePromise)
	data.Set("client_id", r.ClientID)
	data.Set("client_secret", r.ClientSecret)
	data.Set("promise_definition_id", r.PromiseDefinitionID)
	data.Set("authenticating_token", r.AuthenticatingToken)
	data.Set("scope", strings.Join(r.Scopes, ","))

	tokenURL := fmt.Sprintf("%s/ims/token/v4", c.url)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := c.do(req)
	if err != nil {
		return nil, fmt.Errorf("error performing request: %v", err)
	}

	if res.StatusCode != http.StatusOK {
		return nil, errorResponse(res)
	}

	var body struct {
		PromiseToken string `json:"promise_token"`
		TokenType    string `json:"token_type"`
		Scope        string `json:"scope"`
		ExpiresIn    int    `json:"expires_in"`
	}
	if err := json.Unmarshal(res.Body, &body); err != nil {
		return nil, fmt.Errorf("decode response: %v", err)
	}

	return &PromiseTokenResponse{
		Response:     *res,
		PromiseToken: body.PromiseToken,
		TokenType:    body.TokenType,
		Scope:        body.Scope,
		ExpiresIn:    time.Second * time.Duration(body.ExpiresIn),
	}, nil
}

// PromiseToken is equivalent to PromiseTokenWithContext with a background
// context.
func (c *Client) PromiseToken(r *PromiseTokenRequest) (*PromiseTokenResponse, error) {
	return c.PromiseTokenWithContext(context.Background(), r)
}

// PromiseExchangeRequest contains the data for exchanging a promise token for a
// new access token.
type PromiseExchangeRequest struct {
	ClientID     string
	ClientSecret string
	PromiseToken string
	Scopes       []string
}

// PromiseExchangeResponse contains the response of a successful exchange of a
// promise token for an access token.
type PromiseExchangeResponse struct {
	Response
	AccessToken           string
	PromiseToken          string
	PromiseTokenID        string
	TokenType             string
	Scope                 string
	ExpiresIn             time.Duration
	PromiseTokenExpiresIn time.Duration
}

func (c *Client) validatePromiseExchangeRequest(r *PromiseExchangeRequest) error {
	switch {
	case r.ClientID == "":
		return fmt.Errorf("missing client ID parameter")
	case r.ClientSecret == "":
		return fmt.Errorf("missing client secret parameter")
	case r.PromiseToken == "":
		return fmt.Errorf("missing promise token parameter")
	case len(r.Scopes) == 0 || (len(r.Scopes) == 1 && r.Scopes[0] == ""):
		return fmt.Errorf("scopes are required for promise exchange")
	default:
		return nil
	}
}

// PromiseExchangeWithContext exchanges a promise token for a new access token.
func (c *Client) PromiseExchangeWithContext(ctx context.Context, r *PromiseExchangeRequest) (*PromiseExchangeResponse, error) {
	if err := c.validatePromiseExchangeRequest(r); err != nil {
		return nil, fmt.Errorf("invalid parameters for promise exchange: %v", err)
	}

	data := url.Values{}
	data.Set("grant_type", grantTypePromiseExchange)
	data.Set("client_id", r.ClientID)
	data.Set("client_secret", r.ClientSecret)
	data.Set("promise_token", r.PromiseToken)
	data.Set("scope", strings.Join(r.Scopes, ","))

	tokenURL := fmt.Sprintf("%s/ims/token/v4", c.url)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := c.do(req)
	if err != nil {
		return nil, fmt.Errorf("error performing request: %v", err)
	}

	if res.StatusCode != http.StatusOK {
		return nil, errorResponse(res)
	}

	var body struct {
		AccessToken           string `json:"access_token"`
		PromiseToken          string `json:"promise_token"`
		PromiseTokenID        string `json:"promise_token_id"`
		TokenType             string `json:"token_type"`
		Scope                 string `json:"scope"`
		ExpiresIn             int    `json:"expires_in"`
		PromiseTokenExpiresIn int    `json:"promise_token_expires_in"`
	}
	if err := json.Unmarshal(res.Body, &body); err != nil {
		return nil, fmt.Errorf("decode response: %v", err)
	}

	return &PromiseExchangeResponse{
		Response:              *res,
		AccessToken:           body.AccessToken,
		PromiseToken:          body.PromiseToken,
		PromiseTokenID:        body.PromiseTokenID,
		TokenType:             body.TokenType,
		Scope:                 body.Scope,
		ExpiresIn:             time.Second * time.Duration(body.ExpiresIn),
		PromiseTokenExpiresIn: time.Second * time.Duration(body.PromiseTokenExpiresIn),
	}, nil
}

// PromiseExchange is equivalent to PromiseExchangeWithContext with a background
// context.
func (c *Client) PromiseExchange(r *PromiseExchangeRequest) (*PromiseExchangeResponse, error) {
	return c.PromiseExchangeWithContext(context.Background(), r)
}
