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

func NewSlidingWindow(resolution, length, evictionInterval time.Duration) *slidingWindow {
	s := &slidingWindow{
		resolution:       resolution,
		length:           length,
		evictionInterval: evictionInterval,
		dataPoints:       []DataPoint{},
	}

	s.consolidate()

	return s
}

func (s *slidingWindow) Start(ctx context.Context) {
	done := ctx.Done()

	go func() {
		ticker := time.NewTicker(s.evictionInterval)

		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				s.mutex.Lock()
				s.slide()
				s.consolidate()
				s.mutex.Unlock()
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

	s.dataPoints = s.dataPoints[cutoff:]
}

// AddCount adds +1 to the DataPoint at the correct moment
func (s *slidingWindow) AddCount(at time.Time) {
	// first we truncate the at time to the correct resolution
	at = at.Truncate(s.resolution)

	// then we check if the at time is in the sliding window
	if at.Before(time.Now().Add(-s.length).Truncate(s.resolution)) {
		// the at time is before the sliding window, we ignore it
		return
	}

	// add the Count to the correct DataPoint
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for i, dataPoint := range s.dataPoints {
		if dataPoint.Timestamp.Equal(at) {
			s.dataPoints[i].Count++
			return
		}
	}

	// no DataPoint found, create a new one
	s.dataPoints = append(s.dataPoints, DataPoint{
		Timestamp: at,
		Count:     1,
	})

	// consolidate the window
	s.consolidate()

	return
}

// consolidate the window, this will fill any gaps in the dataPoints slice
// it'll first truncate the slice and remove all old dataPoints
// then it'll fill the slice with new DataPoints
// and sort it afterwards
func (s *slidingWindow) consolidate() {
	s.slide()

	now := time.Now().Truncate(s.resolution)

	// first create a new slice with maxLen
	newDataPoints := make([]DataPoint, s.maxLength())

	// prefill the new slice with DataPoints
	for i := len(newDataPoints) - 1; i >= 0; i-- {
		newDataPoints[i] = DataPoint{
			Timestamp: now.Add(time.Duration(i-len(newDataPoints)+1) * s.resolution),
			Count:     0,
		}
	}

	// then loop over the current dataPoints and add the count from those points to the correct new DataPoint
	for _, dataPoint := range s.dataPoints {
		for i, newDataPoint := range newDataPoints {
			if newDataPoint.Timestamp.Equal(dataPoint.Timestamp) {
				newDataPoints[i].Count += dataPoint.Count
			}
		}
	}

	// sort the dataPoints
	sort.Slice(newDataPoints, func(i, j int) bool {
		return newDataPoints[i].Timestamp.Before(newDataPoints[j].Timestamp)
	})

	// set the new dataPoints
	s.dataPoints = newDataPoints
}
