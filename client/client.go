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
	"nuts-foundation/nuts-monitor/config"
	"strings"
)

// HTTPClient holds the server address and other basic settings for the http client
type HTTPClient struct {
	Config config.Config
}

func (hb HTTPClient) networkClient() network.ClientInterface {
	response, err := network.NewClientWithResponses(hb.Config.NutsNodeAddr, network.WithHTTPClient(MustCreateHTTPClient(hb.Config)))
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

// NetworkTopology holds vertices and edges, also used in webAPI
type NetworkTopology struct {
	// Edges map of PeerID -> PeerID
	Edges []Tuple `json:"edges"`

	// PeerID own nodes network ID
	PeerID string `json:"peerID"`

	// Vertices array of PeerIDs
	Vertices []string `json:"vertices"`
}

// NetworkTopology returns the vertices and edges of the network.
func (hb HTTPClient) NetworkTopology(ctx context.Context) (NetworkTopology, error) {
	// first get diagnostics for our node's information
	nt := NetworkTopology{}
	diagnostics, err := hb.Diagnostics(ctx)
	if err != nil {
		return nt, err
	}
	nt.PeerID = diagnostics.Network.Connections.PeerId

	// Peer diagnostics, this will include connections from our node to others as well.
	response, err := hb.networkClient().GetPeerDiagnostics(ctx)
	if err != nil {
		return nt, err
	}
	if err := TestResponseCode(http.StatusOK, response); err != nil {
		return nt, err
	}
	parsedResponse, err := network.ParseGetPeerDiagnosticsResponse(response)
	if err != nil {
		return nt, err
	}

	for k, v := range *parsedResponse.JSON200 {
		peerID := realPeerID(k)
		if !containsVertice(nt.Vertices, peerID) {
			nt.Vertices = append(nt.Vertices, peerID)
		}
		if v.Peers != nil {
			for _, connectedPeer := range *v.Peers {
				otherPeerID := realPeerID(connectedPeer)
				t := Tuple([2]string{otherPeerID, peerID})
				if !containsEdge(nt.Edges, t) {
					nt.Edges = append(nt.Edges, t)
				}
				if !containsVertice(nt.Vertices, otherPeerID) {
					nt.Vertices = append(nt.Vertices, otherPeerID)
				}
			}
		}
	}

	return nt, nil
}

type Tuple [2]string

func (t Tuple) equals(other Tuple) bool {
	return (t[0] == other[0] && t[1] == other[1]) ||
		(t[0] == other[1] && t[1] == other[0])
}

// realPeerID removes the -bootstrap postfix
func realPeerID(peerID string) string {
	s := strings.Split(peerID, "-bootstrap")
	return s[0]
}

func containsVertice(vertices []string, vertice string) bool {
	for _, v := range vertices {
		if v == vertice {
			return true
		}
	}
	return false
}

func containsEdge(edges []Tuple, tuple Tuple) bool {
	for _, v := range edges {
		if v.equals(tuple) {
			return true
		}
	}
	return false
}
