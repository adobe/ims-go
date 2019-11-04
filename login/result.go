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
	"fmt"
	"net/http"

	"github.com/adobe/ims-go/ims"
)

type resultHandler struct {
	successHandler http.Handler
	failureHandler http.Handler
	resCh          chan<- *ims.TokenResponse
	errCh          chan<- error
}

func (h *resultHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if result, ok := r.Context().Value(contextKeyResult).(*ims.TokenResponse); ok {
		if h.successHandler != nil {
			h.successHandler.ServeHTTP(w, r)
		} else {
			fmt.Fprintf(w, "Success!")
		}

		select {
		case h.resCh <- result:
			// Result sent.
		case <-r.Context().Done():
			// Request cancelled.
		}

		return
	}

	var serr error

	if err, ok := r.Context().Value(contextKeyError).(error); ok {
		serr = err
	} else {
		serr = fmt.Errorf("neither error nor result returned")
	}

	if h.failureHandler != nil {
		h.failureHandler.ServeHTTP(w, r)
	} else {
		fmt.Fprintf(w, "Error: %v", serr)
	}

	select {
	case h.errCh <- serr:
		// Error sent.
	case <-r.Context().Done():
		// Request cancelled.
	}
}
