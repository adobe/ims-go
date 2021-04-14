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

// GetProfileRequest is the request for GetProfile.
type GetProfileRequest struct {
	// AccessToken is a valid access token.
	AccessToken string
	ApiVersion  string
}

// GetProfileResponse is the response for GetProfile.
type GetProfileResponse struct {
	Response
}

// GetProfileWithContext reads the user profile associated to a given access
// token. It returns a non-nil response on success or an error on failure.
func (c *Client) GetProfileWithContext(ctx context.Context, r *GetProfileRequest) (*GetProfileResponse, error) {
	if r.ApiVersion == "" {
		r.ApiVersion = "v1"
	}

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/ims/profile/%s", c.url, r.ApiVersion), nil)
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

	return &GetProfileResponse{
		Response: *res,
	}, nil
}

// GetProfile is equivalent to GetProfileWithContext with a background context.
func (c *Client) GetProfile(r *GetProfileRequest) (*GetProfileResponse, error) {
	return c.GetProfileWithContext(context.Background(), r)
}
