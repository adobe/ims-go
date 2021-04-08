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
	"fmt"
	"io/ioutil"
	"net/http"
)

type TokenType string

const (
	AccessToken       TokenType = "access_token"
	RefreshToken                = "refresh_token"
	DeviceToken                 = "device_token"
	AuthorizationCode           = "authorization_code"
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
	// Body is the raw response body.
	Body []byte
}

// ValidateToken validates a token using the IMS API.
// It returns a non-nil response on success or an error on failure.
func (c *Client) ValidateToken(r *ValidateTokenRequest) (*ValidateTokenResponse, error) {

	// The token type is a mandatory parameter and should be validated.
	switch r.Type {
	case AccessToken, RefreshToken, DeviceToken, AuthorizationCode:
	default:
		return nil, fmt.Errorf("invalid token type: %v", r.Type)
	}

	req, err := http.NewRequest(http.MethodGet,
		fmt.Sprintf("%s/ims/validate_token/v1?type=%s&client_id=%s&token=%s",
			c.url, r.Type, r.ClientID, r.Token), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %v", err)
	}

	// Setting this header is recommended in the documentation
	// but it is not working, using client_id in the URL in the meantime
	//req.Header.Set("X-IMS-ClientId", r.ClientID)

	res, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("perform request: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, errorResponse(res)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %v", err)
	}

	return &ValidateTokenResponse{
		Body: body,
	}, nil
}
