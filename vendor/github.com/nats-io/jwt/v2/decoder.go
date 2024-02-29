/*
 * Copyright 2020-2022 The NATS Authors
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package jwt

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/nats-io/nkeys"
)

const libVersion = 2

type identifier struct {
	Type          ClaimType `json:"type,omitempty"`
	GenericFields `json:"nats,omitempty"`
}

func (i *identifier) Kind() ClaimType {
	if i.Type != "" {
		return i.Type
	}
	return i.GenericFields.Type
}

func (i *identifier) Version() int {
	if i.Type != "" {
		return 1
	}
	return i.GenericFields.Version
}

type v1ClaimsDataDeletedFields struct {
	Tags          TagList   `json:"tags,omitempty"`
	Type          ClaimType `json:"type,omitempty"`
	IssuerAccount string    `json:"issuer_account,omitempty"`
}

// Decode takes a JWT string decodes it and validates it
// and return the embedded Claims. If the token header
// doesn't match the expected algorithm, or the claim is
// not valid or verification fails an error is returned.
func Decode(token string) (Claims, error) {
	// must have 3 chunks
	chunks := strings.Split(token, ".")
	if len(chunks) != 3 {
		return nil, errors.New("expected 3 chunks")
	}

	// header
	if _, err := parseHeaders(chunks[0]); err != nil {
		return nil, err
	}
	// claim
	data, err := decodeString(chunks[1])
	if err != nil {
		return nil, err
	}
	ver, claim, err := loadClaims(data)
	if err != nil {
		return nil, err
	}

	// sig
	sig, err := decodeString(chunks[2])
	if err != nil {
		return nil, err
	}

	if ver <= 1 {
		if !claim.verify(chunks[1], sig) {
			return nil, errors.New("claim failed V1 signature verification")
		}
	} else {
		if !claim.verify(token[:len(chunks[0])+len(chunks[1])+1], sig) {
			return nil, errors.New("claim failed V2 signature verification")
		}
	}

	prefixes := claim.ExpectedPrefixes()
	if prefixes != nil {
		ok := false
		issuer := claim.Claims().Issuer
		for _, p := range prefixes {
			switch p {
			case nkeys.PrefixByteAccount:
				if nkeys.IsValidPublicAccountKey(issuer) {
					ok = true
				}
			case nkeys.PrefixByteOperator:
				if nkeys.IsValidPublicOperatorKey(issuer) {
					ok = true
				}
			case nkeys.PrefixByteUser:
				if nkeys.IsValidPublicUserKey(issuer) {
					ok = true
				}
			case nkeys.PrefixByteServer:
				if nkeys.IsValidPublicServerKey(issuer) {
					ok = true
				}
			}
		}
		if !ok {
			return nil, fmt.Errorf("unable to validate expected prefixes - %v", prefixes)
		}
	}
	return claim, nil
}

func loadClaims(data []byte) (int, Claims, error) {
	var id identifier
	if err := json.Unmarshal(data, &id); err != nil {
		return -1, nil, err
	}

	if id.Version() > libVersion {
		return -1, nil, errors.New("JWT was generated by a newer version ")
	}

	var claim Claims
	var err error
	switch id.Kind() {
	case OperatorClaim:
		claim, err = loadOperator(data, id.Version())
	case AccountClaim:
		claim, err = loadAccount(data, id.Version())
	case UserClaim:
		claim, err = loadUser(data, id.Version())
	case ActivationClaim:
		claim, err = loadActivation(data, id.Version())
	case AuthorizationRequestClaim:
		claim, err = loadAuthorizationRequest(data, id.Version())
	case AuthorizationResponseClaim:
		claim, err = loadAuthorizationResponse(data, id.Version())
	case "cluster":
		return -1, nil, errors.New("ClusterClaims are not supported")
	case "server":
		return -1, nil, errors.New("ServerClaims are not supported")
	default:
		var gc GenericClaims
		if err := json.Unmarshal(data, &gc); err != nil {
			return -1, nil, err
		}
		return -1, &gc, nil
	}

	return id.Version(), claim, err
}
