// Copyright 2024 Adobe. All rights reserved.
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

// GetAdminOrganizationsRequest is the request for GetOrganizations.
type GetAdminOrganizationsRequest struct {
	Guid         string
	AuthSrc      string
	ServiceToken string
	ApiVersion   string
	ClientID     string
}

// GetAdminOrganizationsResponse is the response for GetOrganizations.
type GetAdminOrganizationsResponse struct {
	Response
}

// GetAdminOrganizationsWithContext reads the user organizations associated to a
// given access token. It returns a non-nil response on success or an error on
// failure.
func (c *Client) GetAdminOrganizationsWithContext(ctx context.Context, r *GetAdminOrganizationsRequest) (*GetAdminOrganizationsResponse, error) {
	if r.ApiVersion == "" {
		r.ApiVersion = "v5"
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
		fmt.Sprintf("%s/ims/admin_organizations/%s", c.url, r.ApiVersion),
		strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("create request: %v", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", r.ServiceToken))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-IMS-ClientId", r.ClientID)

	res, err := c.do(req)
	if err != nil {
		return nil, fmt.Errorf("perform request: %v", err)
	}

	if res.StatusCode != http.StatusOK {
		return nil, errorResponse(res)
	}

	return &GetAdminOrganizationsResponse{
		Response: *res,
	}, nil
}

// GetAdminOrganizations is equivalent to GetAdminOrganizationsWithContext with a
// background context.
func (c *Client) GetAdminOrganizations(r *GetAdminOrganizationsRequest) (*GetAdminOrganizationsResponse, error) {
	return c.GetAdminOrganizationsWithContext(context.Background(), r)
}
