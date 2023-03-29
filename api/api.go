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

package api

import (
	"context"
	"nuts-foundation/nuts-monitor/client"
	"nuts-foundation/nuts-monitor/config"
)

var _ StrictServerInterface = (*Wrapper)(nil)

const (
	DOWN = "DOWN"
	UP   = "UP"
)

type Wrapper struct {
	Config config.Config
	Client client.HTTPClient
}

func (w Wrapper) Diagnostics(ctx context.Context, _ DiagnosticsRequestObject) (DiagnosticsResponseObject, error) {
	diagnostics, err := w.Client.Diagnostics(ctx)
	if err != nil {
		return nil, err
	}

	return Diagnostics200JSONResponse(*diagnostics), nil
}

func (w Wrapper) CheckHealth(ctx context.Context, _ CheckHealthRequestObject) (CheckHealthResponseObject, error) {
	if w.Config.NutsNodeAddr != "" {
		_, err := w.Client.CheckHealth(ctx)
		if err != nil {
			var errString interface{} = err.Error()
			return CheckHealth200JSONResponse{
				Status: "DOWN",
				Details: map[string]client.HealthCheckResult{
					"node": {
						Details: &errString,
						Status:  "DOWN",
					},
				},
			}, nil
		}

		return CheckHealth200JSONResponse{
			Status: "UP",
			Details: map[string]client.HealthCheckResult{
				"node": {
					Status: "UP",
				},
			},
		}, nil
	}
	return CheckHealth200JSONResponse{
		Status: "UP",
	}, nil
}
