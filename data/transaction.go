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
	"errors"
	"github.com/lestrrat-go/jwx/jws"
	"strings"
	"time"
)

// ErrInvalidSigner is returned when the signer is not a valid DID signer
var ErrInvalidSigner = errors.New("invalid signer")

// ErrNoSigTime is returned when the transaction does not contain a sigt field
var ErrNoSigTime = errors.New("no sigt field")

// Transaction represents a Nuts transaction.
// It is parsed from a JWS token. It does not check the signature.
type Transaction struct {
	// Signer is extracted from the key used to sign the transaction
	Signer string
	// SigTime is the signature time in seconds since the Unix epoch
	SigTime time.Time
}

func FromJWS(transaction string) (*Transaction, error) {
	// we use the lestrrat-go/jwx library to parse the JWS
	jwsToken, err := jws.ParseString(transaction)
	if err != nil {
		return nil, err
	}

	// first extract the signature time from the "sigt" field
	// the "sigt" field is a Unix Timestamp in seconds
	// we convert it to an int64
	sigt, ok := jwsToken.Signatures()[0].ProtectedHeaders().Get("sigt")
	if !ok {
		return nil, ErrNoSigTime
	}
	// parse the sigt string value to time field
	sigTime := time.Unix(int64(sigt.(float64)), 0)

	// the signer can either be extracted from the "kid" header or from the embedded key
	// we first try to extract it from the "kid" header
	signer, ok := jwsToken.Signatures()[0].ProtectedHeaders().Get("kid")
	if ok {
		// the kid is a combination of DID and key ID, we only want the DID part
		// the DID is the part before the first #
		// example: did:nuts:0x1234567890abcdef#key-1 -> did:nuts:0x1234567890abcdef
		// check if # is contained in the string, return an error if not
		index := strings.Index(signer.(string), "#")
		if index == -1 {
			return nil, ErrInvalidSigner
		}
		return &Transaction{Signer: signer.(string)[:index], SigTime: sigTime}, nil
	}

	// if the "kid" header is not present, we try to extract the signer from the embedded key
	// the embedded key is a JWK, we extract the "kid" header from it
	// the "kid" header is a combination of DID and key ID, we only want the DID part
	// the DID is the part before the first #
	// example: did:nuts:0x1234567890abcdef#key-1 -> did:nuts:0x1234567890abcdef
	jwk := jwsToken.Signatures()[0].ProtectedHeaders().JWK()
	if jwk != nil {
		kid := jwk.KeyID()
		index := strings.Index(kid, "#")
		if index == -1 {
			return nil, ErrInvalidSigner
		}
		return &Transaction{Signer: kid[:index], SigTime: sigTime}, nil
	}

	return &Transaction{}, nil
}
