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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

// ExampleJWS contains a correct transaction in condensed JWS format with a jwk field
const ExampleJWS = "eyJhbGciOiJFUzI1NiIsImNyaXQiOlsic2lndCIsInZlciIsInByZXZzIiwiandrIl0sImN0eSI6ImFwcGxpY2F0aW9uL2RpZCtqc29uIiwiandrIjp7ImNydiI6IlAtMjU2Iiwia2lkIjoiZGlkOm51dHM6Q29yMzI4SjUxaE54U3V5RXVCZ2FWdVZuUXBFZ0tzOTFzTUpHYVB1M0I2SnIjcjNDM25kWHFMT0YzWkpCTkh5SVM4SFEzSjRVQmlKRGplQTRGREFRSk51OCIsImt0eSI6IkVDIiwieCI6IlpvMTRYR0pwRzIwSXdYUmFINGhjZ2p0bXUzTHF6dnNoUUlBTTZIWXZJN1UiLCJ5IjoibVJrOTZkRjVSd05Zd0tPUGxncTVxeUtoQUhkQ0UyeHM2bHFJaWtndGJJTSJ9LCJsYyI6MCwicHJldnMiOltdLCJzaWd0IjoxNjUzOTg2MTMwLCJ2ZXIiOjF9.Y2UxOTI3ZTQ1NTdjNDNmMmM1YWVkYzg1OWI4OTg3ZmY2NmI3ZDk3YjhmZmVhZDJkNjEyZDE1ZjNkNTIwMmJlOQ.PEZyffKoWPliezsUlfAm7cdcHTDCImwa5w6inVxC8QQg9swJM3ozjZEV2b3_DzOVDpN7jecvb1WeIf7PDMHTKQ"

// ExampleJWS2 contains a correct transaction in condensed JWS format with a kid field
const ExampleJWS2 = "eyJhbGciOiJFUzI1NiIsImNyaXQiOlsic2lndCIsInZlciIsInByZXZzIiwia2lkIl0sImN0eSI6ImFwcGxpY2F0aW9uL2RpZCtqc29uIiwia2lkIjoiZGlkOm51dHM6OWJXOFZoVmsyazRXNXAxNVpmVER0cFFEYVNrNm9MNnlhR0JaenU3ZzI5UFojZjdRdDcyVlRIZzIybC0tVmZ2VWZQTkR0V0ZfWWxENmFDTUotQmdvS3l4USIsImxjIjoxMCwicHJldnMiOlsiYjA2OGRkMjc0NDFjMDBkNzU1MTdjYjUwZmJhMjIyYjczZjU5NDFlZGY4ZDNmNGQxMzk2NjBjNDkyZTZkNmZkNyJdLCJzaWd0IjoxNjU0MjQ4OTU4LCJ2ZXIiOjF9.NDBmMmY0NGYxNDUwYjBiYzE2ZGYxMzYyMzZmM2I0MDkzZDc3NTEyM2IyZGM5YWFjMzg3YzJmZGMxYzIyYTliZA.1Ak130yrjlqEGgg3HcVm1JB0iEOlnxOhIGojo9icthyg50h72ByRzKg4Pa7oWnuKH4JnIzlZXkPH0vvC6BL2Bw"

// ExampleJWS3 contains a transaction without a sigtime field
const ExampleJWS3 = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"

// ExampleJWS4 contains a transaction with an incorrect kid field
const ExampleJWS4 = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6ImRpZDpudXRzOmFiY2RlZmcxMjM0NSJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.p3sW4cvesjhCCS05Uwhow12WFjH2fGKAQH5wmy5MdCQ"

// ExampleJWS5 contains a transaction without key fields
const ExampleJWS5 = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCIsInNpZ3QiOjEwfQ.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.NPkBysCGI47ey7RiDi_zPl9IxN34t3Kdk1Lpw3Dzzok"

func TestFromJWS(t *testing.T) {
	t.Run("extract transaction from a valid JWS with a jwk field", func(t *testing.T) {
		transaction, err := FromJWS(ExampleJWS)

		require.NoError(t, err)
		assert.NotNil(t, transaction)
		assert.Equal(t, time.Unix(1653986130, 0), transaction.SigTime)
	})

	t.Run("extract transaction from a valid JWS without a jwk field", func(t *testing.T) {
		transaction, err := FromJWS(ExampleJWS2)

		require.NoError(t, err)
		assert.NotNil(t, transaction)
	})
	t.Run("extract transaction from a valid JWS without a jwk field and without a kid field", func(t *testing.T) {
		transaction, err := FromJWS(ExampleJWS5)

		require.NoError(t, err)
		assert.NotNil(t, transaction)
	})
	t.Run("extract transaction from a valid JWS without a sigt field", func(t *testing.T) {
		transaction, err := FromJWS(ExampleJWS3)

		assert.Error(t, ErrNoSigTime, err)
		assert.Nil(t, transaction)
	})
	t.Run("extract transaction from a JWS with an incorrect kid field", func(t *testing.T) {
		transaction, err := FromJWS(ExampleJWS4)

		assert.Error(t, err)
		assert.Nil(t, transaction)
	})
}
