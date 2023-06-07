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
	"fmt"
	"net/http"
	"nuts-foundation/nuts-monitor/client/diagnostics"
	"nuts-foundation/nuts-monitor/client/network"
	"nuts-foundation/nuts-monitor/client/vdr"
	"nuts-foundation/nuts-monitor/config"
)

// HTTPClient holds the server address and other basic settings for the http client
type HTTPClient struct {
	Config config.Config
}

func (hb HTTPClient) networkClient() network.ClientInterface {
	addr := hb.Config.NutsNodeInternalAddr
	if addr == "" {
		addr = hb.Config.NutsNodeAddr
	}
	response, err := network.NewClientWithResponses(addr, network.WithHTTPClient(MustCreateHTTPClient(hb.Config)))
	if err != nil {
		panic(err)
	}
	return response
}

func (hb HTTPClient) vdrClient() vdr.ClientInterface {
	addr := hb.Config.NutsNodeInternalAddr
	if addr == "" {
		addr = hb.Config.NutsNodeAddr
	}
	response, err := vdr.NewClientWithResponses(addr, vdr.WithHTTPClient(MustCreateHTTPClient(hb.Config)))
	if err != nil {
		panic(err)
	}
	return response
}

func (hb HTTPClient) diagnosticsClient() diagnostics.ClientInterface {
	response, err := diagnostics.NewClientWithResponses(hb.Config.NutsNodeAddr, diagnostics.WithHTTPClient(MustCreateHTTPClient(hb.Config)))
	if err != nil {
		panic(err)
	}
	return response
}

func (hb HTTPClient) CheckHealth(ctx context.Context, reqEditors ...diagnostics.RequestEditorFn) (*diagnostics.Health, error) {
	response, err := hb.diagnosticsClient().CheckHealth(ctx, reqEditors...)
	if err != nil {
		return nil, err
	}
	if err := TestResponseCode(http.StatusOK, response); err != nil {
		return nil, err
	}
	result, err := diagnostics.ParseCheckHealthResponse(response)
	if err != nil {
		return nil, err
	}
	if result.JSON503 != nil {
		return result.JSON503, nil
	}
	if result.JSON200 != nil {
		return result.JSON200, nil
	}
	return nil, fmt.Errorf("received incorrect response from node: %s", string(result.Body))
}

func (hb HTTPClient) Diagnostics(ctx context.Context) (*diagnostics.Diagnostics, error) {
	reqEditors := []diagnostics.RequestEditorFn{
		func(ctx context.Context, req *http.Request) error {
			req.Header.Add("Accept", "application/json")
			return nil
		},
	}

	response, err := hb.diagnosticsClient().Diagnostics(ctx, reqEditors...)
	if err != nil {
		return nil, err
	}
	if err := TestResponseCode(http.StatusOK, response); err != nil {
		return nil, err
	}
	result, err := diagnostics.ParseDiagnosticsResponse(response)
	if err != nil {
		return nil, err
	}
	if result.JSON200 != nil {
		return result.JSON200, nil
	}
	return nil, fmt.Errorf("received incorrect response from node: %s", string(result.Body))
}

// PeerDiagnostics returns the PeerDiagnostics per PeerID.
func (hb HTTPClient) PeerDiagnostics(ctx context.Context) (map[string]network.PeerDiagnostics, error) {
	// first get diagnostics for our node's information
	peerDiagnostics := make(map[string]network.PeerDiagnostics)

	response, err := hb.networkClient().GetPeerDiagnostics(ctx)
	if err != nil {
		return peerDiagnostics, err
	}
	if err := TestResponseCode(http.StatusOK, response); err != nil {
		return peerDiagnostics, err
	}
	parsedResponse, err := network.ParseGetPeerDiagnosticsResponse(response)
	if err != nil {
		return peerDiagnostics, err
	}
	if parsedResponse.JSON200 != nil {
		return *parsedResponse.JSON200, nil
	}
	return peerDiagnostics, fmt.Errorf("received incorrect response from node: %s", string(parsedResponse.Body))
}

func (hb HTTPClient) DIDDocument(ctx context.Context, did string) (*vdr.DIDResolutionResult, error) {
	response, err := hb.vdrClient().GetDID(ctx, did, &vdr.GetDIDParams{})
	if err != nil {
		return nil, err
	}
	if err := TestResponseCode(http.StatusOK, response); err != nil {
		return nil, err
	}
	result, err := vdr.ParseGetDIDResponse(response)
	if err != nil {
		return nil, err
	}
	if result.JSON200 != nil {
		return result.JSON200, nil
	}
	return nil, fmt.Errorf("received incorrect response from node: %s", string(result.Body))
}

// ListTransactions returns transactions in a certain range according to LC value
func (hb HTTPClient) ListTransactions(ctx context.Context, start int, end int) ([]string, error) {
	var transactions []string

	response, err := hb.networkClient().ListTransactions(ctx, &network.ListTransactionsParams{
		Start: &start,
		End:   &end,
	})
	if err != nil {
		return transactions, err
	}
	if err := TestResponseCode(http.StatusOK, response); err != nil {
		return transactions, err
	}
	parsedResponse, err := network.ParseListTransactionsResponse(response)
	if err != nil {
		return transactions, err
	}
	if parsedResponse.JSON200 != nil {
		return *parsedResponse.JSON200, nil
	}
	return transactions, nil
}
