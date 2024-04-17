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
	"net/url"
	"strings"
)

// GetAdminProfileRequest is the request for GetProfile.
type GetAdminProfileRequest struct {
	Guid         string
	AuthSrc      string
	ServiceToken string
	ApiVersion   string
	ClientID     string
}

// GetAdminProfileResponse is the response for GetProfile.
type GetAdminProfileResponse struct {
	Response
}

// GetAdminProfileWithContext reads the user profile associated to a given access
// token. It returns a non-nil response on success or an error on failure.
func (c *Client) GetAdminProfileWithContext(ctx context.Context, r *GetAdminProfileRequest) (*GetAdminProfileResponse, error) {
	if r.ApiVersion == "" {
		r.ApiVersion = "v1"
	}
	if r.Guid == "" {
		return nil, fmt.Errorf("missing guid")
	}
	if r.AuthSrc == "" {
		return nil, fmt.Errorf("missing auth_src")
	}
	if r.ServiceToken == "" {
		return nil, fmt.Errorf("missing service token")
	}
	if r.ClientID == "" {
		return nil, fmt.Errorf("missing client ID")
	}

	data := url.Values{}
	data.Set("guid", r.Guid)
	data.Set("auth_src", r.AuthSrc)

	req, err := http.NewRequestWithContext(ctx,
		http.MethodPost,
		fmt.Sprintf("%s/ims/admin_profile/%s", c.url, r.ApiVersion),
		strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("create request: %v", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", r.ServiceToken))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-IMS-ClientID", r.ClientID)

	res, err := c.do(req)
	if err != nil {
		return nil, fmt.Errorf("perform request: %v", err)
	}

	if res.StatusCode != http.StatusOK {
		return nil, errorResponse(res)
	}

	return &GetAdminProfileResponse{
		Response: *res,
	}, nil
}

// GetAdminProfile is equivalent to GetAdminProfileWithContext with a background context.
func (c *Client) GetAdminProfile(r *GetAdminProfileRequest) (*GetAdminProfileResponse, error) {
	return c.GetAdminProfileWithContext(context.Background(), r)
}
