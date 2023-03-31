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
	"context"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"nuts-foundation/nuts-monitor/config"
	"testing"
)

func TestClient_CheckHealth(t *testing.T) {
	// Create a new test HTTP server that returns a 200 status code and
	// the Spring Boot Actuator return body when the /health endpoint is requested.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status": "UP"}`))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	client := HTTPClient{Config: config.Config{NutsNodeAddress: ts.URL}}
	resp, err := client.CheckHealth(context.Background())
	if err != nil {
		t.Fatalf("Failed to get correct response: %v", err)
	}

	assert.Equal(t, "UP", resp.Status)
}
