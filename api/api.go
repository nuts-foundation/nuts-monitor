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
	"sort"
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

func (w Wrapper) AggregatedTransactions(_ context.Context, _ AggregatedTransactionsRequestObject) (AggregatedTransactionsResponseObject, error) {
	// get data from the store
	dataPoints := w.DataStore.GetTransactions()

	// convert the data points to the response object
	response := AggregatedTransactions{
		Hourly:  make([]DataPoint, 0),
		Daily:   make([]DataPoint, 0),
		Monthly: make([]DataPoint, 0),
	}
	// loop over the 3 categories of data points
	// for each category, loop over the data points and add them to the correct category in the response object
	for cty, dp := range dataPoints[0] {
		for _, a := range dp {
			response.Hourly = append(response.Hourly, toDataPoint(cty, a))
		}
	}
	for cty, dp := range dataPoints[1] {
		for _, a := range dp {
			response.Daily = append(response.Daily, toDataPoint(cty, a))
		}
	}
	for cty, dp := range dataPoints[2] {
		for _, a := range dp {
			response.Monthly = append(response.Monthly, toDataPoint(cty, a))
		}
	}

	return AggregatedTransactions200JSONResponse(response), nil
}

func (w Wrapper) TransactionCounts(_ context.Context, _ TransactionCountsRequestObject) (TransactionCountsResponseObject, error) {
	// get counts from the store
	mapping, count := w.DataStore.GetTransactionCounts()

	// create the basic response object
	response := TransactionCounts200JSONResponse{
		RootCount: int(count),
	}

	// loop over the mapping and place each entry in a tuple, then sort the tuples on count descending and take the top 10
	// then add the top 10 to the response object
	type tuple struct {
		count int
		did   string
	}
	var tuples []tuple
	for k, v := range mapping {
		tuples = append(tuples, tuple{
			count: int(v),
			did:   k,
		})
	}
	sort.Slice(tuples, func(i, j int) bool {
		return tuples[i].count > tuples[j].count
	})

	for i := 0; i < 10 && i < len(tuples); i++ {
		response.TransactionsPerRoot = append(response.TransactionsPerRoot, TransactionsPerRoot{
			Did:   tuples[i].did,
			Count: tuples[i].count,
		})
	}
	return response, nil
}

func toDataPoint(cty string, dp data.DataPoint) DataPoint {
	return DataPoint{
		ContentType: cty,
		Timestamp:   int(dp.Timestamp.Unix()),
		Label:       dp.Timestamp.Format(time.RFC3339),
		Value:       int(dp.Count),
	}
}
