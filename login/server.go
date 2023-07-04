// Copyright 2019 Adobe. All rights reserved.
// This file is licensed to you under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License. You may obtain a copy
// of the License at http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software distributed under
// the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR REPRESENTATIONS
// OF ANY KIND, either express or implied. See the License for the specific language
// governing permissions and limitations under the License.

package login

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net"
	"net/http"

	"github.com/adobe/ims-go/ims"
)

// Server is the local server used to perform a user login in IMS.
//
// The usage pattern for the login server (minus error handling, for simplicity)
// is the following.
//
//	// Start the service
//	srv := login.NewServer(cfg)
//
//	// Listen on a separate goroutine
//	go srv.Serve(listener)
//
//	// Wait for a response.
//	select {
//	case res := <-srv.Response():
//	    // Process the login response.
//	case err := <-srv.Error():
//	    // An error occurred.
//	case <- time.After(5 * time.Minute):
//	    // Bail out after some time.
//	}
//
//	// Close the server.
//	srv.Shutdown()
type Server struct {
	server *http.Server
	resCh  chan *ims.TokenResponse
	errCh  chan error
}

// ServerConfig is the configuration for the login server.
type ServerConfig struct {
	// The IMS client.
	Client *ims.Client
	// The client ID.
	ClientID string
	// The client secret.
	ClientSecret string
	// List of scopes to request.
	Scope []string
	// The URL to be redirected after authentication
	RedirectURI string
	// A custom handler for sending an error response to the client. If not
	// provided, a default response is sent.
	OnError http.Handler
	// A custom handler for sending a success response to the client. If not
	// provided, a default response is sent.
	OnSuccess http.Handler
	// Use PKCE in the authorization code flow.
	UsePKCE bool
}

// NewServer creates a new Server for the provided ServerConfig.
func NewServer(cfg *ServerConfig) (*Server, error) {
	state, err := randomState()
	if err != nil {
		return nil, fmt.Errorf("generate random state: %v", err)
	}

	codeVerifier := ""
	if cfg.UsePKCE {
		codeVerifier, err = randomCodeVerifier()
		if err != nil {
			return nil, fmt.Errorf("generate random code verifier: %v", err)
		}
	}

	var (
		resCh = make(chan *ims.TokenResponse)
		errCh = make(chan error)
	)

	result := &resultHandler{
		successHandler: cfg.OnSuccess,
		failureHandler: cfg.OnError,
		resCh:          resCh,
		errCh:          errCh,
	}

	route := &routeMiddleware{
		redirect: &redirectMiddleware{
			client:       cfg.Client,
			clientID:     cfg.ClientID,
			scope:        cfg.Scope,
			state:        state,
			redirectURI:  cfg.RedirectURI,
			next:         result,
			codeVerifier: codeVerifier,
		},

		callback: &callbackMiddleware{
			client:       cfg.Client,
			clientID:     cfg.ClientID,
			clientSecret: cfg.ClientSecret,
			scope:        cfg.Scope,
			state:        state,
			next:         result,
			codeVerifier: codeVerifier,
		},
	}

	server := &http.Server{
		Handler: route,
	}

	return &Server{
		server: server,
		resCh:  resCh,
		errCh:  errCh,
	}, nil
}

// Serve make the server listen to the provided listener.
func (s *Server) Serve(lst net.Listener) error {
	return s.server.Serve(lst)
}

// Shutdown closes the server and the channels returned from Error() and
// Response(). When closing the server, this method has the same semantics of
// http.Server.Shutdown.
func (s *Server) Shutdown(ctx context.Context) error {
	defer close(s.errCh)
	defer close(s.resCh)

	return s.server.Shutdown(ctx)
}

// Error returns a channel that can be listened to for error conditions.
func (s *Server) Error() <-chan error {
	return s.errCh
}

// Response returns a channel that can be listened to for a successful login.
func (s *Server) Response() <-chan *ims.TokenResponse {
	return s.resCh
}

func randomState() (string, error) {
	binaryData := make([]byte, 128)

	if _, err := rand.Read(binaryData); err != nil {
		return "", fmt.Errorf("error generating state parameter: %v", err)
	}

	return base64.StdEncoding.EncodeToString(binaryData), nil
}

// Returns a random code verifier parameter for PKCE
// https://datatracker.ietf.org/doc/html/rfc7636#section-4.1
func randomCodeVerifier() (string, error) {
	binaryData := make([]byte, 32)

	if _, err := rand.Read(binaryData); err != nil {
		return "", fmt.Errorf("error generating code verifier parameter: %v", err)
	}

	return base64.RawURLEncoding.EncodeToString(binaryData), nil
}
