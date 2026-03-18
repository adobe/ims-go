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

const defaultOBOGrantType = "urn:ietf:params:oauth:grant-type:token-exchange"

type OBOExchangeRequest struct {
	ClientID string

	ClientSecret string

	SubjectToken string

	Scopes []string
}

type OBOExchangeResponse struct {
	Response
	AccessToken string
	ExpiresIn   time.Duration
}

func (c *Client) validateOBOExchangeRequest(r *OBOExchangeRequest) error {
	switch {
	case r.ClientID == "":
		return fmt.Errorf("missing client ID parameter")
	case r.ClientSecret == "":
		return fmt.Errorf("missing client secret parameter")
	case r.SubjectToken == "":
		return fmt.Errorf("missing subject token parameter (only access tokens are accepted)")
	case len(r.Scopes) == 0 || (len(r.Scopes) == 1 && r.Scopes[0] == ""):
		return fmt.Errorf("scopes are required for On-Behalf-Of exchange")
	default:
		return nil
	}
}

func (c *Client) OBOExchangeWithContext(ctx context.Context, r *OBOExchangeRequest) (*OBOExchangeResponse, error) {
	if err := c.validateOBOExchangeRequest(r); err != nil {
		return nil, fmt.Errorf("invalid parameters for On-Behalf-Of exchange: %v", err)
	}

	data := url.Values{}
	data.Set("grant_type", defaultOBOGrantType)
	data.Set("client_id", r.ClientID)
	data.Set("client_secret", r.ClientSecret)
	data.Set("subject_token", r.SubjectToken)
	data.Set("subject_token_type", "urn:ietf:params:oauth:token-type:access_token")
	data.Set("requested_token_type", "urn:ietf:params:oauth:token-type:access_token")
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
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.Unmarshal(res.Body, &body); err != nil {
		return nil, fmt.Errorf("decode response: %v", err)
	}

	return &OBOExchangeResponse{
		Response:    *res,
		AccessToken: body.AccessToken,
		ExpiresIn:   time.Second * time.Duration(body.ExpiresIn),
	}, nil
}

func (c *Client) OBOExchange(r *OBOExchangeRequest) (*OBOExchangeResponse, error) {
	return c.OBOExchangeWithContext(context.Background(), r)
}
