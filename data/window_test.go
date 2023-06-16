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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestSlidingWindow_AddCount(t *testing.T) {
	t.Run("adds a new DataPoint", func(t *testing.T) {
		now := time.Now()

		window := slidingWindow{
			resolution: time.Second,
			length:     time.Second * 10,
			dataPoints: map[string][]DataPoint{},
		}

		window.AddCount("test", now)

		assert.Len(t, window.dataPoints["test"], 10)
		assert.Equal(t, now.Truncate(time.Second), window.dataPoints["test"][9].Timestamp)
	})

	t.Run("adds a new DataPoint with a clockdrift", func(t *testing.T) {
		now := time.Now().Add(time.Second * 2)

		window := slidingWindow{
			resolution: time.Second,
			length:     time.Second * 10,
			clockdrift: time.Second * 5,
			dataPoints: map[string][]DataPoint{},
		}

		window.AddCount("test", now)

		assert.Len(t, window.dataPoints["test"], 10)
		assert.Equal(t, uint32(1), window.dataPoints["test"][6].Count)
	})

	t.Run("increases Count of existing DataPoint", func(t *testing.T) {
		now := time.Now().Truncate(time.Second)
		window := slidingWindow{
			resolution: time.Second,
			length:     time.Second * 10,
			dataPoints: map[string][]DataPoint{
				"test": {
					{Timestamp: now, Count: 1},
				},
			},
		}

		window.AddCount("test", now)

		assert.Len(t, window.dataPoints, 1)
		assert.Equal(t, uint32(2), window.dataPoints["test"][0].Count)
	})
}

func TestSlidingWindow_slide(t *testing.T) {
	t.Run("removes dataPoints older than length", func(t *testing.T) {
		now := time.Now().Truncate(time.Second)
		slightlyOlder := now.Add(-time.Millisecond)
		window := slidingWindow{
			resolution: time.Second,
			length:     time.Second * 10,
			dataPoints: map[string][]DataPoint{
				"test": {
					{Timestamp: slightlyOlder.Add(time.Second * -10), Count: 1},
					{Timestamp: now.Add(time.Second * -9), Count: 1},
					{Timestamp: now.Add(time.Second * -8), Count: 1},
				},
			},
		}

		window.slide(now)

		assert.Len(t, window.dataPoints["test"], 2)
		assert.Equal(t, now.Add(time.Second*-9), window.dataPoints["test"][0].Timestamp)
		assert.Equal(t, uint32(1), window.dataPoints["test"][0].Count)
		assert.Equal(t, now.Add(time.Second*-8), window.dataPoints["test"][1].Timestamp)
		assert.Equal(t, uint32(1), window.dataPoints["test"][1].Count)
	})
}

func TestSlidingWindow_consolidate(t *testing.T) {
	t.Run("it fills up a window to the length", func(t *testing.T) {
		window := slidingWindow{
			resolution: time.Second,
			length:     time.Second * 10,
			dataPoints: map[string][]DataPoint{"test": {}},
		}

		window.consolidate()

		assert.Len(t, window.dataPoints["test"], 10)
		assert.Equal(t, uint32(0), window.dataPoints["test"][0].Count)
	})

	t.Run("it fills gaps in the window", func(t *testing.T) {
		now := time.Now().Truncate(time.Second)
		window := slidingWindow{
			resolution: time.Second,
			length:     time.Second * 5,
			dataPoints: map[string][]DataPoint{"test": {
				{Timestamp: now.Add(time.Second * -4), Count: 1},
				{Timestamp: now.Add(time.Second * -2), Count: 1},
				{Timestamp: now, Count: 2},
			},
			},
		}

		window.consolidate()

		require.Len(t, window.dataPoints["test"], 5)
		assert.Equal(t, uint32(1), window.dataPoints["test"][0].Count)
		assert.Equal(t, uint32(0), window.dataPoints["test"][1].Count)
		assert.Equal(t, uint32(1), window.dataPoints["test"][2].Count)
		assert.Equal(t, uint32(0), window.dataPoints["test"][3].Count)
		assert.Equal(t, uint32(2), window.dataPoints["test"][4].Count)
	})
}

func TestSlidingWindow_Start(t *testing.T) {
	t.Run("slides & consolidates periodically", func(t *testing.T) {
		window := slidingWindow{
			resolution:       time.Second,
			length:           time.Second * 5,
			evictionInterval: time.Millisecond,
			dataPoints: map[string][]DataPoint{"test": {
				{Timestamp: time.Now().Truncate(time.Second).Add(-5 * time.Second), Count: 2},
				{Timestamp: time.Now().Truncate(time.Second).Add(-4 * time.Second), Count: 1},
			},
			},
		}
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		window.Start(ctx)

		time.Sleep(time.Millisecond * 10)

		window.mutex.Lock()
		defer window.mutex.Unlock()

		require.Len(t, window.dataPoints["test"], 5)
		assert.Equal(t, uint32(1), window.dataPoints["test"][0].Count)
	})
}

func TestSlidingWindow_toIndex(t *testing.T) {
	// 5 second window, 1 second resolution
	// add a datapoint for each second starting at -11 * time.Now()
	// call toIndex with each datapoint
	// assert that the index is correct
	now := time.Now().Truncate(time.Second)
	window := slidingWindow{
		resolution: time.Second,
		length:     time.Second * 5,
		dataPoints: map[string][]DataPoint{"test": {
			{Timestamp: now.Add(time.Second * -4), Count: 1}, // -4 to -3 interval
			{Timestamp: now.Add(time.Second * -3), Count: 1}, // -3 to -2 interval
			{Timestamp: now.Add(time.Second * -2), Count: 1}, // -2 to -1 interval
			{Timestamp: now.Add(time.Second * -1), Count: 1}, // -1 to truncated interval
			{Timestamp: now, Count: 1},                       // truncated at current second
		},
		},
	}

	for i := 0; i < 5; i++ {
		assert.Equal(t, i, window.toIndex(window.dataPoints["test"][i], now))
	}
}
