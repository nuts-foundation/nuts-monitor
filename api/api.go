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
	"nuts-foundation/nuts-monitor/client/diagnostics"
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
	upResponse := CheckHealth200JSONResponse{
		Status: UP,
	}
	downResponse := CheckHealth503JSONResponse{
		Status: DOWN,
	}

	if w.Config.NutsNodeAddr != "" {
		h, err := w.Client.CheckHealth(ctx)
		if err != nil {
			var errString interface{} = err.Error()
			downResponse.Details = map[string]diagnostics.HealthCheckResult{
				"node": {
					Details: &errString,
					Status:  "UNKNOWN",
				},
			}
			return downResponse, nil
		}
		if h.Status != UP {
			downResponse.Details = map[string]diagnostics.HealthCheckResult{
				"node": {
					Status: h.Status,
				},
			}
			return downResponse, nil
		}

		upResponse.Details = map[string]diagnostics.HealthCheckResult{
			"node": {
				Status: h.Status,
			},
		}

	}
	return upResponse, nil
}

func (w Wrapper) NetworkTopology(ctx context.Context, _ NetworkTopologyRequestObject) (NetworkTopologyResponseObject, error) {
	networkTopology, err := w.Client.NetworkTopology(ctx)
	if err != nil {
		return nil, err
	}

	// test
	//networkTopology.Vertices = append(networkTopology.Vertices, "1")
	//networkTopology.Vertices = append(networkTopology.Vertices, "2")
	//networkTopology.Edges = append(networkTopology.Edges, client.Tuple{"1", "2"})
	//networkTopology.Edges = append(networkTopology.Edges, client.Tuple{networkTopology.PeerID, "2"})
	//networkTopology.Edges = append(networkTopology.Edges, client.Tuple{networkTopology.PeerID, "1"})

	return NetworkTopology200JSONResponse(networkTopology), nil
}
