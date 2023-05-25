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
	"sort"
	"sync"
	"time"
)

type DataPoint struct {
	Timestamp time.Time
	Count     uint32
}

type slidingWindow struct {
	resolution       time.Duration
	length           time.Duration
	evictionInterval time.Duration
	mutex            sync.Mutex
	dataPoints       []DataPoint
}

// init fill the dataPoints slice with DataPoints
func (s *slidingWindow) init() {
	s.dataPoints = make([]DataPoint, s.maxLength())

	now := time.Now().Truncate(s.resolution)

	for i := len(s.dataPoints) - 1; i >= 0; i-- {
		s.dataPoints[i] = DataPoint{
			Timestamp: now.Truncate(s.resolution).Add(time.Duration(i-len(s.dataPoints)+1) * s.resolution),
			Count:     0,
		}
	}
}

func (s *slidingWindow) start(ctx context.Context) {
	done := ctx.Done()

	go func() {
		ticker := time.NewTicker(s.evictionInterval)

		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				s.slide()
			}
		}
	}()
}

// maxLength calculates the maximum number of dataPoints in the sliding window
func (s *slidingWindow) maxLength() int {
	return int(s.length / s.resolution)
}

// slide removes dataPoints older than length
func (s *slidingWindow) slide() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	now := time.Now().Truncate(s.resolution)

	cutoff := -1
	for j := len(s.dataPoints) - 1; j >= 0; j-- {
		if s.dataPoints[j].Timestamp.Before(now.Add(-s.length)) {
			cutoff = j + 1
			break
		}
	}

	if cutoff == -1 {
		return
	}

	// create a new slice with length equal to maxLength and fill it with the dataPoints that are not older than length
	newSet := make([]DataPoint, s.maxLength())

	copy(newSet, s.dataPoints[cutoff:])

	// fill the remainder of the slice with new DataPoints, the last dataPoint should be time.Now().Truncate(s.resolution)
	// each dataPoint before that should be time.Now().Truncate(s.resolution).Add(-s.resolution)
	for i := len(newSet) - 1; i > cutoff; i-- {
		newSet[i] = DataPoint{
			Timestamp: now.Truncate(s.resolution).Add(time.Duration(i-len(newSet)+1) * s.resolution),
			Count:     0,
		}
	}
	s.dataPoints = newSet
}

// addCount adds +1 to the DataPoint at the correct moment
func (s *slidingWindow) addCount(at time.Time) {
	// first we truncate the at time to the correct resolution
	at = at.Truncate(s.resolution)

	// then we check if the at time is in the sliding window
	if at.Before(time.Now().Add(-s.length)) {
		// the at time is before the sliding window, we ignore it
		return
	}

	// add the Count to the correct DataPoint
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for i, dataPoint := range s.dataPoints {
		if dataPoint.Timestamp == at {
			s.dataPoints[i].Count++
			return
		}
	}

	// no DataPoint found, create a new one
	s.dataPoints = append(s.dataPoints, DataPoint{
		Timestamp: at,
		Count:     1,
	})

	// sort the dataPoints
	sort.Slice(s.dataPoints, func(i, j int) bool {
		return s.dataPoints[i].Timestamp.Before(s.dataPoints[j].Timestamp)
	})

	return
}
