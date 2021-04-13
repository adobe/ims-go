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
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// MetaScope is a meta-scope that can be optionally added to a JWT token.
//
// Deprecated: use explicit claims in ExchangeJWTRequest.
type MetaScope int

const (
	// MetaScopeCloudManager is the meta-scope for Cloud Manager.
	//
	// Deprecated: use explicit claims in ExchangeJWTRequest.
	MetaScopeCloudManager MetaScope = iota
	// MetaScopeAdobeIO is the meta-scope for Adobe IO.
	//
	// Deprecated: use explicit claims in ExchangeJWTRequest.
	MetaScopeAdobeIO
	// MetaScopeAnalyticsBulkIngest is the meta-scope for Analytics Bulk Ingest.
	//
	// Deprecated: use explicit claims in ExchangeJWTRequest.
	MetaScopeAnalyticsBulkIngest
)

// ExchangeJWTRequest contains the data for exchanging a JWT token with an
// access token.
type ExchangeJWTRequest struct {
	// The private key for signing the JWT token. This field is required.
	PrivateKey []byte
	// The expiration time for the access token. This field is required.
	Expiration time.Time
	// The issuer of the JWT token. It represents the identity of the
	// organization issuing the token. This field is required.
	Issuer string
	// The subject of the JWT token. It represents the identity of the technical
	// account.
	Subject string
	// The client ID.
	ClientID string
	// The client secret.
	ClientSecret string
	// The additional meta-scopes to add to the JWT token.
	//
	// Deprecated: use explicit claims in ExchangeJWTRequest.
	MetaScope []MetaScope
	// Additional claims to add to the JWT token.
	Claims map[string]interface{}
}

// ExchangeJWTResponse contains the response of a successful exchange of a JWT
// token.
type ExchangeJWTResponse struct {
	// Body is the raw response body.
	Body []byte
	// AccessToken is the access token.
	AccessToken string
	// ExpiresIn is the expiration for the token.
	ExpiresIn time.Duration
}

// ExchangeJWTWithContext exchanges a JWT token for an access token.
func (c *Client) ExchangeJWTWithContext(ctx context.Context, r *ExchangeJWTRequest) (*ExchangeJWTResponse, error) {
	key, err := jwt.ParseRSAPrivateKeyFromPEM(r.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("parse key: %v", err)
	}

	claims := jwt.MapClaims{
		"exp": r.Expiration.Unix(),
		"iss": r.Issuer,
		"sub": r.Subject,
		"aud": fmt.Sprintf("%s/c/%s", c.url, r.ClientID),
	}

	for _, ms := range r.MetaScope {
		switch ms {
		case MetaScopeCloudManager:
			claims[fmt.Sprintf("%v/s/ent_cloudmgr_sdk", c.url)] = true
		case MetaScopeAdobeIO:
			claims[fmt.Sprintf("%v/s/ent_adobeio_sdk", c.url)] = true
		case MetaScopeAnalyticsBulkIngest:
			claims[fmt.Sprintf("%v/s/ent_analytics_bulk_ingest_sdk", c.url)] = true
		default:
			return nil, fmt.Errorf("invalid meta-scope: %v", ms)
		}
	}

	for k, v := range r.Claims {
		claims[k] = v
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	signed, err := token.SignedString(key)
	if err != nil {
		return nil, fmt.Errorf("sign token: %v", err)
	}

	data := url.Values{}

	data.Set("client_id", r.ClientID)
	data.Set("client_secret", r.ClientSecret)
	data.Set("jwt_token", signed)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/ims/exchange/v1/jwt", c.url), strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("perform request: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, errorResponse(res)
	}

	raw, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %v", err)
	}

	var body struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}

	if err := json.Unmarshal(raw, &body); err != nil {
		return nil, fmt.Errorf("decode response: %v", err)
	}

	return &ExchangeJWTResponse{
		Body:        raw,
		AccessToken: body.AccessToken,
		ExpiresIn:   time.Millisecond * time.Duration(body.ExpiresIn),
	}, nil
}

// ExchangeJWT is quivalent to ExchangeJWTWithContext with a background context.
func (c *Client) ExchangeJWT(r *ExchangeJWTRequest) (*ExchangeJWTResponse, error) {
	return c.ExchangeJWTWithContext(context.Background(), r)
}
