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

// The user token is used to authorize the request and to define which user's profile is requested.
type UserToken struct {
	string
}

type UserProfile struct{
	string
}

func (c *Client) GetProfile(t *UserToken) (*UserProfile, error){

	up := &UserProfile{}
	// Create request
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/ims/profile/v1", c.url), nil)

	// Add the user token as Bearer token
	bearer := fmt.Sprintf("Bearer %v", *t)
	req.Header.Add("Authorization", bearer )

	// Perform request
	res, err := c.client.Do(req)
	if err != nil {
		return up, fmt.Errorf("error requesting profile: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return up, errorResponse(res)
	}

	bodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return up, fmt.Errorf("error reading request body")
	}
	up = &UserProfile{string(bodyBytes)}

	return up, nil
}