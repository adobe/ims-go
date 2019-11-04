// ADOBE CONFIDENTIAL
// ___________________
//
//  Copyright 2019 Adobe Systems Incorporated
//  All Rights Reserved.
//
// NOTICE:  All information contained herein is, and remains the property of
// Adobe Systems Incorporated and its suppliers, if any.  The intellectual and
// technical concepts contained herein are proprietary to Adobe Systems
// Incorporated and its suppliers and are protected by trade secret or copyright
// law. Dissemination of this information or reproduction of this material is
// strictly forbidden unless prior written permission is obtained from Adobe
// Systems Incorporated.

package login

import (
	"context"
	"net/http"

	"github.com/adobe/ims-go/ims"
)

type contextKey string

var (
	contextKeyError  = contextKey("error")
	contextKeyResult = contextKey("result")
)

func serveError(h http.Handler, w http.ResponseWriter, r *http.Request, err error) {
	h.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), contextKeyError, err)))
}

func serveResult(h http.Handler, w http.ResponseWriter, r *http.Request, res *ims.TokenResponse) {
	h.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), contextKeyResult, res)))
}
