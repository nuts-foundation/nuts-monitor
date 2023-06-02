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
	"nuts-foundation/nuts-monitor/data"
	"time"
)

var _ StrictServerInterface = (*Wrapper)(nil)

const (
	DOWN = "DOWN"
	UP   = "UP"
)

type Wrapper struct {
	Config    config.Config
	Client    client.HTTPClient
	DataStore *data.Store
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
	ts := client.TopologyService{
		HTTPClient: w.Client,
	}

	networkTopology, err := ts.NetworkTopology(ctx)
	if err != nil {
		return nil, err
	}

	return NetworkTopology200JSONResponse(networkTopology), nil
}

func (w Wrapper) GetWebTransactionsAggregated(ctx context.Context, _ GetWebTransactionsAggregatedRequestObject) (GetWebTransactionsAggregatedResponseObject, error) {
	// get data from the store
	dataPoints := w.DataStore.GetTransactions()

	// convert the data points to the response object
	response := AggregatedTransactions{}
	// loop over the 3 categories of data points
	// for each category, loop over the data points and add them to the correct category in the response object
	for _, dp := range dataPoints[0] {
		response.Hourly = append(response.Hourly, toDataPoint(dp))
	}
	for _, dp := range dataPoints[1] {
		response.Daily = append(response.Daily, toDataPoint(dp))
	}
	for _, dp := range dataPoints[2] {
		response.Monthly = append(response.Monthly, toDataPoint(dp))
	}

	return GetWebTransactionsAggregated200JSONResponse(response), nil
}

func toDataPoint(dp data.DataPoint) DataPoint {
	return DataPoint{
		Timestamp: int(dp.Timestamp.Unix()),
		Label:     dp.Timestamp.Format(time.RFC3339),
		Value:     int(dp.Count),
	}
}
