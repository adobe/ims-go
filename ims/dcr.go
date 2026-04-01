// Copyright 2026 Adobe. All rights reserved.
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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type DCRRequest struct {
	ClientName   string
	RedirectURIs []string
}

type DCRResponse struct {
	Response
}

func (c *Client) validateDCRRequest(r *DCRRequest) error {
	switch {
	case r.ClientName == "":
		return fmt.Errorf("missing client name parameter")
	case len(r.RedirectURIs) == 0:
		return fmt.Errorf("missing redirect URIs parameter")
	default:
		return nil
	}
}

func (c *Client) DCRWithContext(ctx context.Context, r *DCRRequest) (*DCRResponse, error) {
	if err := c.validateDCRRequest(r); err != nil {
		return nil, fmt.Errorf("invalid parameters for client registration: %v", err)
	}

	payload, err := json.Marshal(struct {
		ClientName   string   `json:"client_name"`
		RedirectURIs []string `json:"redirect_uris"`
	}{
		ClientName:   r.ClientName,
		RedirectURIs: r.RedirectURIs,
	})
	if err != nil {
		return nil, fmt.Errorf("error building registration payload: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("%s/ims/register", c.url), bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := c.do(req)
	if err != nil {
		return nil, fmt.Errorf("error performing request: %v", err)
	}

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, errorResponse(res)
	}

	return &DCRResponse{Response: *res}, nil
}

func (c *Client) DCR(r *DCRRequest) (*DCRResponse, error) {
	return c.DCRWithContext(context.Background(), r)
}
