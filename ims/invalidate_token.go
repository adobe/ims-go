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
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// InvalidateTokenRequest is the request to InvalidateToken.
type InvalidateTokenRequest struct {
	Token        string
	Type         TokenType
	ClientID     string
	Cascading    bool
	ClientSecret string
}

// InvalidateTokenWithContext invalidates a token using the IMS API. It returns a
// non-nil response on success or an error on failure.
func (c *Client) InvalidateTokenWithContext(ctx context.Context, r *InvalidateTokenRequest) error {

	switch {
	case r.ClientID == "":
		return fmt.Errorf("missing client ID parameter")
	case r.Type == "":
		return fmt.Errorf("missing token type parameter")
	case r.Token == "":
		return fmt.Errorf("missing token parameter")
	}

	switch r.Type {
	case AccessToken, RefreshToken, DeviceToken:
	case ServiceToken:
		if r.ClientSecret == "" {
			return fmt.Errorf("service token invalidation needs client secret parameter")
		}
	default:
		return fmt.Errorf("invalid token type: %v", r.Type)
	}

	data := url.Values{}
	data.Set("token_type", string(r.Type))
	data.Set("token", r.Token)
	if r.Cascading {
		data.Set("cascading", "all")
	}
	data.Set("client_id", r.ClientID)
	data.Set("client_secret", r.ClientSecret)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/ims/invalidate_token/v2", c.url),
		strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("create request: %v", err)
	}

	// Header X-IMS-ClientID will be mandatory in the future
	req.Header.Set("X-IMS-ClientId", r.ClientID)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := c.do(req)
	if err != nil {
		return fmt.Errorf("perform request: %v", err)
	}

	if res.StatusCode != http.StatusOK {
		return errorResponse(res)
	}

	return nil
}

// InvalidateToken is equivalent to InvalidateTokenWithContext with a background
// context.
func (c *Client) InvalidateToken(r *InvalidateTokenRequest) error {
	return c.InvalidateTokenWithContext(context.Background(), r)
}
