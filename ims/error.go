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
	"errors"
	"fmt"
)

// Error is an error containing information returned by the IMS API.
type Error struct {
	Response
	// ErrorCode is an error code associated with the error response.
	ErrorCode string
	// ErrorMessage is a human-readable description of the error.
	ErrorMessage string
}

func (e *Error) Error() string {
	return fmt.Sprintf(
		"error response: statusCode=%d, errorCode='%s', errorMessage='%s', x-debug-id='%s'",
		e.StatusCode,
		e.ErrorCode,
		e.ErrorMessage,
		e.XDebugID,
	)
}

// IsError checks if the given error is an IMS error and, if it is, returns an
// instance of Error.
func IsError(err error) (*Error, bool) {
	var imsErr *Error
	ok := errors.As(err, &imsErr)
	return imsErr, ok
}

func errorResponse(res *Response) error {
	var payload struct {
		ErrorCode    string `json:"error"`
		ErrorMessage string `json:"error_description"`
	}

	// The error from json.Unmarshal() is voluntarily ignored. If the server
	// returns an empty or badly serialized response, we just go on. The library
	// did its best to extract meaningful information from the response. The
	// unparsed body is returned to the user anyway, who will take the final
	// decision about how to deal with this error.

	_ = json.Unmarshal(res.Body, &payload)

	return &Error{
		Response:     *res,
		ErrorCode:    payload.ErrorCode,
		ErrorMessage: payload.ErrorMessage,
	}
}
