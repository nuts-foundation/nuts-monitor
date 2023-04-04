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
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"net/http"
	"nuts-foundation/nuts-monitor/config"
	"nuts-foundation/nuts-monitor/test"
	"testing"
)

func TestClient_CheckHealth(t *testing.T) {
	ts := test.BasicTestNode(t)

	client := HTTPClient{Config: config.Config{NutsNodeAddress: ts.URL()}}
	resp, err := client.CheckHealth(context.Background())
	if err != nil {
		t.Fatalf("Failed to get correct response: %v", err)
	}

	assert.Equal(t, "UP", resp.Status)
}

func TestClient_Diagnostics(t *testing.T) {
	ts := test.BasicTestNode(t)
	d := Diagnostics{
		Status: Status{
			SoftwareVersion: "v1.0.0",
		},
	}
	dBytes, _ := json.Marshal(d)
	ts.HandleFunc("/status/diagnostics", func(w http.ResponseWriter, request *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(dBytes)
	})

	client := HTTPClient{Config: config.Config{NutsNodeAddress: ts.URL()}}
	resp, err := client.Diagnostics(context.Background())
	if err != nil {
		t.Fatalf("Failed to get correct response: %v", err)
	}

	assert.Equal(t, "v1.0.0", resp.Status.SoftwareVersion)
}
