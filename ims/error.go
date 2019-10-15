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
	"encoding/json"
	"fmt"
	"net/http"
)

// Error is an error containing information returned by the IMS API.
type Error struct {
	// StatusCode is the status code of the response returning the error.
	StatusCode int
	// ErrorCode is an error code associated with the error response.
	ErrorCode string
	// ErrorMessage is a human-readable description of the error.
	ErrorMessage string
}

func (e *Error) Error() string {
	return e.ErrorMessage
}

// IsError checks if the given error is an IMS error and, if it is, returns an
// instance of Error.
func IsError(err error) (*Error, bool) {
	imsErr, ok := err.(*Error)
	return imsErr, ok
}

func errorResponse(r *http.Response) error {
	var payload struct {
		ErrorCode    string `json:"error"`
		ErrorMessage string `json:"error_description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		return fmt.Errorf("decode error response: %v", err)
	}

	return &Error{
		StatusCode:   r.StatusCode,
		ErrorCode:    payload.ErrorCode,
		ErrorMessage: payload.ErrorMessage,
	}
}
