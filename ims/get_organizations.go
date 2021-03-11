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
	"fmt"
	"io/ioutil"
	"net/http"
)

// GetOrganizationsRequest is the request for GetOrganizations.
type GetOrganizationsRequest struct {
	// AccessToken is a valid access token.
	AccessToken string
	ApiVersion string
}

// GetOrganizationsResponse is the response for GetOrganizations.
type GetOrganizationsResponse struct {
	// Body is the raw response body.
	Body []byte
}

// GetOrganizations reads the user organizations associated to a given access token. It
// returns a non-nil response on success or an error on failure.
func (c *Client) GetOrganizations(r *GetOrganizationsRequest) (*GetOrganizationsResponse, error) {
	if r.ApiVersion == "" {
		r.ApiVersion = "v5"
	}
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/ims/organizations/%s", c.url, r.ApiVersion), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %v", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", r.AccessToken))

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

	return &GetOrganizationsResponse{
		Body: body,
	}, nil
}
