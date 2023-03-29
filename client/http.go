/*
 * Copyright (C) 2023 Nuts community
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 *
 */

package client

import (
	"fmt"
	"io"
	"net/http"
	"nuts-foundation/nuts-monitor/config"
	"time"
)

// CreateHTTPClient creates a new HTTP client with the given client configuration.
// The result HTTPRequestDoer can be supplied to OpenAPI generated clients for executing requests.
// This does not use the generated client options for e.g. authentication,
// because each generated OpenAPI client reimplements the client options using structs,
// which makes them incompatible with each other, making it impossible to use write generic client code for common traits like authorization.
// If the given authorization token builder is non-nil, it calls it and passes the resulting token as bearer token with requests.
func CreateHTTPClient(cfg config.Config) (HTTPRequestDoer, error) {
	var result *httpRequestDoerAdapter
	client := &http.Client{}
	client.Timeout = 10 * time.Second
	result = &httpRequestDoerAdapter{
		fn: client.Do,
	}

	generator := func() (string, error) {
		return "", nil
	}

	if cfg.NutsNodeAPIKeyFile != "" {
		generator = createTokenGenerator(cfg)
	}

	fn := result.fn
	result = &httpRequestDoerAdapter{fn: func(req *http.Request) (*http.Response, error) {
		token, err := generator()
		if err != nil {
			return nil, fmt.Errorf("failed to generate authorization token: %w", err)
		}
		if len(token) > 0 {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		return fn(req)
	}}

	return result, nil
}

// MustCreateHTTPClient is like CreateHTTPClient but panics if it returns an error.
func MustCreateHTTPClient(cfg config.Config) HTTPRequestDoer {
	client, err := CreateHTTPClient(cfg)
	if err != nil {
		panic(err)
	}
	return client
}

// AuthorizationTokenGenerator is a function type definition for creating authorization tokens
type AuthorizationTokenGenerator func() (string, error)

// HTTPRequestDoer defines the Do method of the http.Client interface.
type HTTPRequestDoer interface {
	Do(*http.Request) (*http.Response, error)
}

// httpRequestDoerAdapter wraps a HTTPRequestFn in a struct, so it can be used where HTTPRequestDoer is required.
type httpRequestDoerAdapter struct {
	fn func(req *http.Request) (*http.Response, error)
}

// Do calls the wrapped HTTPRequestFn.
func (w httpRequestDoerAdapter) Do(req *http.Request) (*http.Response, error) {
	return w.fn(req)
}

// TestResponseCode checks whether the returned HTTP status response code matches the expected code.
// If it doesn't match it returns an error, containing the received and expected status code, and the response body.
func TestResponseCode(expectedStatusCode int, response *http.Response) error {
	if response.StatusCode != expectedStatusCode {
		responseData, _ := io.ReadAll(response.Body)
		return fmt.Errorf("server returned HTTP %d (expected: %d), body: %s", response.StatusCode, expectedStatusCode, responseData)
	}
	return nil
}
