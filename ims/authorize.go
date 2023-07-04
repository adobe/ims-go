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
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"log"
	"net/url"
	"strings"
)

// GrantType is the grant type specified when building an authorization URL.
type GrantType int

const (
	// GrantTypeDefault is the default grant type as specified by IMS.
	GrantTypeDefault GrantType = iota
	// GrantTypeCode is the authorization code grant type.
	GrantTypeCode
	// GrantTypeImplicit is the implicit grant type.
	GrantTypeImplicit
	// GrantTypeDevice is the device token grant type.
	GrantTypeDevice
)

// AuthorizeURLConfig is the configuration for building an authorization URL.
type AuthorizeURLConfig struct {
	ClientID     string
	GrantType    GrantType
	Scope        []string
	RedirectURI  string
	State        string
	CodeVerifier string
}

// AuthorizeURL builds an authorization URL according to the provided configuration.
func (c *Client) AuthorizeURL(cfg *AuthorizeURLConfig) (string, error) {
	if cfg.ClientID == "" {
		return "", fmt.Errorf("missing client ID")
	}

	if len(cfg.Scope) == 0 {
		return "", fmt.Errorf("missing scope")
	}

	apiURL, err := url.Parse(fmt.Sprintf("%s/ims/authorize/v1", c.url))
	if err != nil {
		return "", fmt.Errorf("parse URL: %v", err)
	}

	q := apiURL.Query()

	q.Set("client_id", cfg.ClientID)
	q.Set("scope", strings.Join(cfg.Scope, ","))

	switch cfg.GrantType {
	case GrantTypeCode:
		q.Set("response_type", "code")
	case GrantTypeImplicit:
		q.Set("response_type", "token")
	case GrantTypeDevice:
		q.Set("response_type", "device")
	default:
		// Default to Authz Code Grant for backward compatibility reasons
		q.Set("response_type", "code")
	}

	if cfg.RedirectURI != "" {
		q.Set("redirect_uri", cfg.RedirectURI)
	}

	if cfg.State != "" {
		q.Set("state", cfg.State)
	}

	// SHA256 the code verifier and base64 encode it.
	if cfg.CodeVerifier != "" {
		h := sha256.New()
		h.Write([]byte(cfg.CodeVerifier))
		codeChallenge := base64.RawURLEncoding.EncodeToString(h.Sum(nil))
		q.Set("code_challenge", codeChallenge)
		q.Set("code_challenge_method", "S256")
	}

	apiURL.RawQuery = q.Encode()

	log.Println("AuthorizeURL: ", apiURL.String())

	return apiURL.String(), nil
}
