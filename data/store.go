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
	"time"
)

// Store is an in-memory store that contains a mapping from transaction signer to its controller.
// It also contains three sliding windows with length and resolution of: (1 hour, 1 minute), (1 day, 1 hour), (30 days, 1 day).
// A transaction can be added, the store will resolve the signer and the controller of the signer.
type Store struct {
	mapping        map[string]string
	slidingWindows []*slidingWindow
}

func NewStore() *Store {
	s := &Store{
		mapping: make(map[string]string),
	}

	// initialize all windows with empty dataPoints using the init function
	s.slidingWindows = append(s.slidingWindows, NewSlidingWindow(time.Minute, time.Hour, time.Second))
	s.slidingWindows = append(s.slidingWindows, NewSlidingWindow(time.Hour, 24*time.Hour, time.Minute))
	s.slidingWindows = append(s.slidingWindows, NewSlidingWindow(24*time.Hour, 30*24*time.Hour, time.Minute))

	return s
}

// Start starts the sliding windows
func (s *Store) Start(ctx context.Context) {
	for i := range s.slidingWindows {
		s.slidingWindows[i].Start(ctx)
	}
}

// Add a transaction to the sliding windows and resolve the controller of the signer
func (s *Store) Add(transaction Transaction) {
	// first add the transaction to the sliding windows
	for i := range s.slidingWindows {
		s.slidingWindows[i].AddCount(transaction.SigTime)
	}

	// todo: resolve the controller of the signer
}

// GetTransactions returns the transactions of the sliding windows
// The smallest resolution is first, the largest resolution is last
func (s *Store) GetTransactions() [3][]DataPoint {
	var transactions [3][]DataPoint

	for i, window := range s.slidingWindows {
		transactions[i] = window.dataPoints
	}

	return transactions
}
