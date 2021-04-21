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
	"net/url"
	"strings"
	"time"
)

// The cluster_at_exchange is an IMS specific grant type, that is used to exchange access tokens.
// This is useful in the context of T2E compatibility. For more information see the IMS documentation.

type ClusterExchangeRequest struct {
	ClientID     string
	ClientSecret string
	Scopes       []string
	UserToken    string
	UserID       string
	OrgID        string
}

type ClusterExchangeResponse struct {
	Response
	AccessToken string
	ExpiresIn   time.Duration
}

func (c *Client) ClusterExchangeWithContext(ctx context.Context, r *ClusterExchangeRequest) (*ClusterExchangeResponse, error) {

	data := url.Values{}
	data.Set("grant_type", "cluster_at_exchange")
	data.Set("client_secret", r.ClientSecret)
	data.Set("user_token", r.UserToken)
	switch {
	case r.UserID != "":
		if r.OrgID != "" {
			return nil, fmt.Errorf("userID and OrgID defined at the same time")
		}
		data.Set("user_id", r.UserID)
	case r.OrgID != "":
		data.Set("owning_org_id", r.OrgID)
	default:
		return nil, fmt.Errorf("no userID or OrgID parameters to perform the request")
	}
	data.Set("scope", strings.Join(r.Scopes, ","))

	req, err := http.NewRequestWithContext( ctx, http.MethodPost,
		fmt.Sprintf("%s/ims/token/v3?client_id=%s", c.url, r.ClientID),
		strings.NewReader(data.Encode()))
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
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	return &ClusterExchangeResponse{
		Response:    *res,
		AccessToken: body.AccessToken,
		ExpiresIn:   time.Millisecond * time.Duration(body.ExpiresIn),
	}, nil
}

func (c *Client) ClusterExchange(r *ClusterExchangeRequest) (*ClusterExchangeResponse, error) {
	return c.ClusterExchangeWithContext(context.Background(), r)
}
