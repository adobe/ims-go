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

// GetProfileRequest is the request for GetProfile.
type GetProfileRequest struct {
	// AccessToken is a valid access token.
	AccessToken string
}

// GetProfileResponse is the response for GetProfile.
type GetProfileResponse struct {
	// Body is the raw body of the response returned when reading the profile.
	Body []byte
}

// GetProfile reads the user profile associated to a given access token. It
// returns a non-nil response on success or an error on failure.
func (c *Client) GetProfile(r *GetProfileRequest) (*GetProfileResponse, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/ims/profile/v1", c.url), nil)
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

	return &GetProfileResponse{
		Body: body,
	}, nil
}
