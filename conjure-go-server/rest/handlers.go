// Copyright (c) 2018 Palantir Technologies. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package rest

import (
	"context"
	"net/http"

	"github.com/palantir/witchcraft-go-error"
	"github.com/palantir/witchcraft-go-logging/wlog/svclog/svc1log"

	"github.com/palantir/conjure-go-runtime/conjure-go-contract/errors"
)

// A HandlerFunc implements http.Handler. If the func returns an error, the corresponding status code and
// JSON-encoded response body are written to the ResponseWriter. It is assumed that, if the error is non-nil,
// nothing has been written to the ResponseWriter.
type HandlerFunc func(http.ResponseWriter, *http.Request) error

// ServeHTTP implements the http.Handler interface
func (h HandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := h(w, r); err != nil {
		var conjureErr errors.Error
		if cErr, ok := werror.RootCause(err).(errors.Error); ok {
			conjureErr = cErr
		} else {
			conjureErr = errors.NewInternal()
		}
		wErr := werror.Wrap(err, "error handling request", werror.SafeParams(map[string]interface{}{
			"errorCode":       conjureErr.Code(),
			"errorName":       conjureErr.Name(),
			"errorInstanceID": conjureErr.InstanceID(),
		}))

		logError(r.Context(), conjureErr.Code().StatusCode(), wErr)
		errors.WriteErrorResponse(w, conjureErr)
	}
}

// ErrHandler is an ErrorHandler that creates a log in the request context's svc1log logger when an error is received.
// This preserves request-scoped logging configuration added by wrouter.
func logError(ctx context.Context, statusCode int, err error) {
	logger := svc1log.FromContext(ctx)
	logFn := logger.Info
	if statusCode >= 500 {
		logFn = logger.Error
	}
	logFn(err.Error(), svc1log.Stacktrace(err))
}
