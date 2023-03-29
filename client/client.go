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
	"net/http"
	"nuts-foundation/nuts-monitor/config"
)

// HTTPClient holds the server address and other basic settings for the http client
type HTTPClient struct {
	Config config.Config
}

func (hb HTTPClient) client() ClientInterface {
	response, err := NewClientWithResponses(hb.Config.NutsNodeAddress, WithHTTPClient(MustCreateHTTPClient(hb.Config)))
	if err != nil {
		panic(err)
	}
	return response
}

func (hb HTTPClient) CheckHealth(ctx context.Context, reqEditors ...RequestEditorFn) (*Health, error) {
	response, err := hb.client().CheckHealth(ctx, reqEditors...)
	if err != nil {
		return nil, err
	}
	if err := TestResponseCode(http.StatusOK, response); err != nil {
		return nil, err
	}
	result, err := ParseCheckHealthResponse(response)
	if err != nil {
		return nil, err
	}
	return result.JSON200, nil
}
