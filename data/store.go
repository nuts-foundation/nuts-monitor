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

package data

import (
	"context"
	"log"
	"nuts-foundation/nuts-monitor/client"
	"time"
)

// Store is an in-memory store that contains a mapping from transaction signer to its controller.
// It also contains three sliding windows with length and resolution of: (1 hour, 1 minute), (1 day, 1 hour), (30 days, 1 day).
// A transaction can be added, the store will resolve the signer and the controller of the signer.
type Store struct {
	client         client.HTTPClient
	mapping        map[string]string
	slidingWindows []*slidingWindow
	didCount       map[string]uint32
	// rootDIDCount is the number of unique root DIDs, we can't use the length of the mapping because
	// the mapping may contain multiple levels of mapping before getting to the root DID
	rootDIDCount uint32
}

func NewStore(client client.HTTPClient) *Store {
	s := &Store{
		client:   client,
		mapping:  make(map[string]string),
		didCount: make(map[string]uint32),
	}

	// initialize all windows with empty dataPoints using the init function
	s.slidingWindows = append(s.slidingWindows, NewSlidingWindow(time.Minute, time.Hour, time.Second))
	s.slidingWindows = append(s.slidingWindows, NewSlidingWindow(time.Hour, 24*time.Hour, time.Minute))
	s.slidingWindows = append(s.slidingWindows, NewSlidingWindow(24*time.Hour, 30*24*time.Hour, time.Minute))

	return s
}

// Start the sliding windows
func (s *Store) Start(ctx context.Context) {
	for i := range s.slidingWindows {
		s.slidingWindows[i].Start(ctx)
	}
}

// Add a transaction to the sliding windows and resolve the controller of the signer
func (s *Store) Add(transaction Transaction) {
	// first add the transaction to the sliding windows
	for i := range s.slidingWindows {
		s.slidingWindows[i].AddCount(transaction.ContentType, transaction.SigTime)
	}

	controller, newRoot := s.resolveController(transaction.Signer)
	if newRoot {
		// a new root so add it to the count
		s.rootDIDCount++
	}
	s.didCount[controller]++
}

// GetTransactions returns the transactions of the sliding windows
// The smallest resolution is first, the largest resolution is last
func (s *Store) GetTransactions() [3]map[string][]DataPoint {
	var transactions [3]map[string][]DataPoint

	for i, window := range s.slidingWindows {
		transactions[i] = map[string][]DataPoint{}
		for cty, a := range window.dataPoints {
			transactions[i][cty] = a
		}
	}

	return transactions
}

// GetTransactionCounts returns the transaction count per root DID and the total number of roots
func (s *Store) GetTransactionCounts() (map[string]uint32, uint32) {
	return s.didCount, s.rootDIDCount
}

func (s *Store) resolveController(txDID string) (string, bool) {
	// check if the did is already resolved
	if controller, ok := s.mapping[txDID]; ok {
		return controller, false
	}

	// resolve the did
	result, err := s.client.DIDDocument(context.Background(), txDID)
	if err != nil {
		// resolving failed, just return the original did
		log.Printf("error resolving did: %s\n", err.Error())
		return txDID, true
	}

	root := txDID
	newRoot := true

	// check if the DID document contains a controller that differs from the did
	for _, controller := range result.Document.Controller {
		if controller.String() != txDID {
			// call resolveController recursively to resolve the controller of the controller
			root, newRoot = s.resolveController(controller.String())
			break
		}
	}
	// add the mapping to the store
	s.mapping[txDID] = root

	return root, newRoot
}
