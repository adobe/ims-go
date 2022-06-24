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
	"fmt"
	"net/http"
)

// GetUserInfoRequest is the request for GetUserInfo.
type GetUserInfoRequest struct {
	// AccessToken is a valid access token.
	AccessToken string
	ApiVersion  string
}

// GetUserInfoResponse is the response for GetUserInfo.
type GetUserInfoResponse struct {
	Response
}

// GetUserInfoWithContext reads the user profile associated to a given access
// token. It returns a non-nil response on success or an error on failure.
func (c *Client) GetUserInfoWithContext(ctx context.Context, r *GetUserInfoRequest) (*GetUserInfoResponse, error) {
	if r.ApiVersion == "" {
		r.ApiVersion = "v1"
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/ims/userinfo/%s", c.url, r.ApiVersion), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %v", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", r.AccessToken))

	res, err := c.do(req)
	if err != nil {
		return nil, fmt.Errorf("perform request: %v", err)
	}

	if res.StatusCode != http.StatusOK {
		return nil, errorResponse(res)
	}

	return &GetUserInfoResponse{
		Response: *res,
	}, nil
}

// GetUserInfo is equivalent to GetUserInfoWithContext with a background context.
func (c *Client) GetUserInfo(r *GetUserInfoRequest) (*GetUserInfoResponse, error) {
	return c.GetUserInfoWithContext(context.Background(), r)
}
