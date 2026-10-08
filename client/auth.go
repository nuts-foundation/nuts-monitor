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
	"crypto"
	"crypto/ecdsa"
	"crypto/rsa"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/v3/jwa"
	"github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/lestrrat-go/jwx/v3/jwt"
	"nuts-foundation/nuts-monitor/config"
	"time"
)

// createTokenGenerator generates valid API tokens for the Nuts node and signs them with the private key
func createTokenGenerator(config config.Config) AuthorizationTokenGenerator {
	return func() (string, error) {
		key, err := jwkKey(config.ApiKey)
		if err != nil {
			return "", err
		}

		issuedAt := time.Now()
		notBefore := issuedAt
		expires := notBefore.Add(time.Second * time.Duration(5))
		token, err := jwt.NewBuilder().
			Issuer(config.NutsNodeAPIUser).
			Subject(config.NutsNodeAPIUser).
			Audience([]string{config.NutsNodeAPIAudience}).
			IssuedAt(issuedAt).
			NotBefore(notBefore).
			Expiration(expires).
			JwtID(uuid.New().String()).
			Build()

		alg, ok := key.Algorithm()
		if !ok {
			return "", errors.New("signing key has no algorithm")
		}
		bytes, err := jwt.Sign(token, jwt.WithKey(alg, key))
		if err != nil {
			return "", err
		}
		return string(bytes), nil
	}
}

func jwkKey(signer crypto.Signer) (key jwk.Key, err error) {
	// ssh key format
	key, err = jwk.Import(signer)
	if err != nil {
		return nil, err
	}

	switch k := signer.(type) {
	case *rsa.PrivateKey:
		key.Set(jwk.AlgorithmKey, jwa.PS512())
	case *ecdsa.PrivateKey:
		var alg jwa.SignatureAlgorithm
		alg, err = ecAlg(k)
		key.Set(jwk.AlgorithmKey, alg)
	default:
		err = fmt.Errorf("unsupported signing private key: %T", k)
		return
	}

	err = jwk.AssignKeyID(key)

	return
}

func ecAlg(key *ecdsa.PrivateKey) (alg jwa.SignatureAlgorithm, err error) {
	alg, err = ecAlgUsingPublicKey(key.PublicKey)
	return
}

func ecAlgUsingPublicKey(key ecdsa.PublicKey) (alg jwa.SignatureAlgorithm, err error) {
	switch key.Params().BitSize {
	case 256:
		alg = jwa.ES256()
	case 384:
		alg = jwa.ES384()
	case 521:
		alg = jwa.ES512()
	default:
		err = errors.New("unsupported key")
	}
	return
}
