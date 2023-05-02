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
	"crypto/tls"
	"encoding/json"
	"github.com/sirupsen/logrus"
	"strings"
	"sync"

	"github.com/nuts-foundation/go-did/did"
	"nuts-foundation/nuts-monitor/client/vdr"
)

// NetworkTopology holds vertices and edges, also used in webAPI
type NetworkTopology struct {
	// Edges map of PeerID -> PeerID
	Edges []Tuple `json:"edges"`

	// PeerID own node's network ID
	PeerID string `json:"peerID"`

	// Peers array of Peer information
	Peers []Peer `json:"peers"`

	TxCount int `json:"tx_count"`
}

// Peer contains info from PeerDiagnostics and the DID Document (if available)
type Peer struct {
	PeerID           string  `json:"peer_id"`
	NodeDID          *string `json:"node_did,omitempty"`
	Address          string  `json:"address"`
	Authenticated    bool    `json:"authenticated"`
	CN               string  `json:"cn"`
	TransactionCount int     `json:"tx_count"`
	ContactName      string  `json:"contact_name"`
	ContactPhone     string  `json:"contact_phone"`
	ContactWeb       string  `json:"contact_web"`
	ContactEmail     string  `json:"contact_email"`
	SoftwareVersion  string  `json:"software_version"`
	SoftwareID       string  `json:"software_id"`
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

func existingPeer(peers []Peer, peerID string) (*Peer, bool) {
	for i, v := range peers {
		if v.PeerID == peerID {
			return &peers[i], true
		}
	}
	return &Peer{PeerID: peerID}, false
}

func containsEdge(edges []Tuple, tuple Tuple) bool {
	for _, v := range edges {
		if v.equals(tuple) {
			return true
		}
	}
	return false
}

// TopologyService is a helper service to combine data from multiple API calls into a single data presentation for the frontend
type TopologyService struct {
	HTTPClient HTTPClient
}

// NetworkTopology returns data about connected peers and "friends of friends".
func (ts TopologyService) NetworkTopology(ctx context.Context) (NetworkTopology, error) {
	// first get diagnostics for our node's information
	networkTopology := NetworkTopology{
		Edges: []Tuple{},
	}
	diagnostics, err := ts.HTTPClient.Diagnostics(ctx)
	if err != nil {
		return networkTopology, err
	}
	networkTopology.PeerID = diagnostics.Network.Connections.PeerId
	networkTopology.TxCount = diagnostics.Network.State.TransactionCount

	// Add own as vertice, needed for empty node
	networkTopology.Peers = append(networkTopology.Peers, Peer{PeerID: diagnostics.Network.Connections.PeerId, TransactionCount: diagnostics.Network.State.TransactionCount})

	// Peer diagnostics, this will include connections from our node to others as well.
	peerDiagnostics, err := ts.HTTPClient.PeerDiagnostics(ctx)
	if err != nil {
		return networkTopology, err
	}

	for k, v := range peerDiagnostics {
		peerID := realPeerID(k)
		peer, ok := existingPeer(networkTopology.Peers, peerID)
		if v.TransactionNum != nil {
			peer.TransactionCount = int(*v.TransactionNum)
		}
		if v.SoftwareVersion != nil {
			peer.SoftwareVersion = *v.SoftwareVersion
		}
		if v.SoftwareID != nil {
			peer.SoftwareID = *v.SoftwareID
		}
		if !ok {
			networkTopology.Peers = append(networkTopology.Peers, *peer)
		}
		if v.Peers != nil {
			for _, connectedPeer := range *v.Peers {
				otherPeerID := realPeerID(connectedPeer)
				t := Tuple([2]string{otherPeerID, peerID})
				if !containsEdge(networkTopology.Edges, t) {
					networkTopology.Edges = append(networkTopology.Edges, t)
				}
				if p, ok := existingPeer(networkTopology.Peers, otherPeerID); !ok {
					networkTopology.Peers = append(networkTopology.Peers, *p)
				}
			}
		}
	}

	// add data to peers we are connected to
	for _, cp := range diagnostics.Network.Connections.ConnectedPeers {
		peer, _ := existingPeer(networkTopology.Peers, cp.Id)
		peer.NodeDID = cp.Nodedid
		peer.Address = cp.Address
		peer.Authenticated = cp.Authenticated
	}

	// this is a blocking call that opens TLS connections to all peers
	ts.addInfoToPeers(ctx, networkTopology.Peers)

	return networkTopology, nil
}

// addInfoToPeers adds information from DID Documents and the Certificate exposed at the NutsComm address.
func (ts TopologyService) addInfoToPeers(ctx context.Context, peers []Peer) {
	wgDocument := sync.WaitGroup{}
	wgTLS := sync.WaitGroup{}
	for i, p := range peers {
		if p.Authenticated {
			// connect to the node and certificate info
			wgDocument.Add(1)
			go func(peer *Peer) {
				document, err := ts.HTTPClient.DIDDocument(ctx, *peer.NodeDID)
				if err != nil {
					logrus.Errorf("failed to retrieve DID Document: %v", err)
				} else {
					nodeContactInfo := extractContactInfo(document.Document)
					peer.ContactName = nodeContactInfo.Name
					peer.ContactEmail = nodeContactInfo.Email
					peer.ContactWeb = nodeContactInfo.Web
					peer.ContactPhone = nodeContactInfo.Phone

					nutsComm, err := extractNutsCommAddress(document.Document)
					if err != nil {
						logrus.Errorf("failed to extract NutsComm address from DID Document: %v", err)
					} else {
						wgTLS.Add(1)
						go func(peer *Peer) {
							conn, err := tls.Dial("tcp", nutsComm, &tls.Config{InsecureSkipVerify: true})
							if err != nil {
								logrus.Errorf("failed to connect over TCP to %s: %v", nutsComm, err)
							} else {
								peerCerts := conn.ConnectionState().PeerCertificates
								if len(peerCerts) > 0 {
									peer.CN = conn.ConnectionState().PeerCertificates[0].Subject.String()
								}
								_ = conn.Close()
							}
							wgTLS.Done()
						}(peer)
					}
				}
				wgDocument.Done()
			}(&peers[i])
		}
	}
	wgDocument.Wait()
	wgTLS.Wait()
}

func extractNutsCommAddress(document vdr.DIDDocument) (string, error) {
	asJSON, _ := json.Marshal(document)
	asGoDID := did.Document{}
	_ = json.Unmarshal(asJSON, &asGoDID)

	_, url, err := asGoDID.ResolveEndpointURL("NutsComm")
	// remove grpc://
	stripped := url[7:]

	return stripped, err
}

// NodeContactInfo is a helper structure
type NodeContactInfo struct {
	Name  string
	Phone string
	Web   string
	Email string
}

func extractContactInfo(document vdr.DIDDocument) NodeContactInfo {
	asJSON, _ := json.Marshal(document)
	asGoDID := did.Document{}
	_ = json.Unmarshal(asJSON, &asGoDID)

	nci := NodeContactInfo{}
	for _, s := range asGoDID.Service {
		if s.Type == "node-contact-info" {
			_ = s.UnmarshalServiceEndpoint(&nci)
			break
		}
	}

	return nci
}
