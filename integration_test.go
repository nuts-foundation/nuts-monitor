/*
 * Nuts node
 * Copyright (C) 2022 Nuts community
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

package main

import (
	"fmt"
	"net"
	"net/http"
	"nuts-foundation/nuts-monitor/test"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestStatusCodes tests if the returned errors from the API implementations are correctly translated to status codes
func TestStatusCodes(t *testing.T) {
	httpPort := startServer(t)

	baseUrl := fmt.Sprintf("http://localhost:%d", httpPort)

	type operation struct {
		path string
		body interface{}
	}
	t.Run("200s", func(t *testing.T) {
		testCases := []operation{
			{path: "/status"},
			{path: "/health"},
		}

		for _, testCase := range testCases {
			resp, err := http.Get(fmt.Sprintf("%s%s", baseUrl, testCase.path))

			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		}
	})
}

func startServer(t *testing.T) int {
	e := newEchoServer()

	httpPort := test.FreeTCPPort()

	go func() {
		err := e.Start(fmt.Sprintf(":%d", httpPort))
		if err != nil {
			if err.Error() != "http: Server closed" {
				t.Fatal(err)
			}
		}
	}()

	if !test.WaitFor(t, func() (bool, error) {
		resp, err := http.Get(fmt.Sprintf("http://localhost:%d/status", httpPort))
		return err == nil && resp.StatusCode == http.StatusOK, nil
	}, time.Second*5, "Timeout while waiting for node to become available") {
		t.Fatal("time-out")
	}

	t.Cleanup(func() {
		err := e.Close()
		if err != nil {
			t.Fatal(err)
		}

		// wait for port to become free again
		test.WaitFor(t, func() (bool, error) {
			if a, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("localhost:%d", httpPort)); err == nil {
				if l, err := net.ListenTCP("tcp", a); err == nil {
					l.Close()
					return true, nil
				}
			}

			return false, nil
		}, 5*time.Second, "Timeout while waiting for node to shutdown")
	})

	return httpPort
}