// Copyright 2021 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

// TokenFormat indicates the format of the attestation token.
type TokenFormat int

// String returns the textual representation of the token format name.
func (tf TokenFormat) String() string {
	switch tf {
	case PsaIatToken:
		return "psa"
	case DiceToken:
		return "dice"
	default:
		return fmt.Sprintf("TokenFormat(%d)", tf)
	}
}

const (
	// PsaIatToken is the PSA initial attestation token (based on:
	// https://datatracker.ietf.org/doc/draft-tschofenig-rats-psa-token/)
	PsaIatToken = TokenFormat(iota)

	// DiceToken is a token based on the TCG DICE specification
	// https://trustedcomputinggroup.org/wp-content/uploads/TCG_DICE_Attestation_Architecture_r22_02dec2020.pdf
	DiceToken

	// UnknownToken is used to indicate that the format of the token could
	// not be established.
	UnknownToken = TokenFormat(math.MaxInt64)
)

var tokenRegex = regexp.MustCompile(`token\((\d+)\)`)

// TokenFormatFromString converts a string name into the corresponding TokenFormat.
func TokenFormatFromString(value string) (TokenFormat, error) {
	value = strings.ToLower(value)

	if value == "psa_iat" || value == "psa-iat" || value == "psa" {
		return PsaIatToken, nil
	} else if matched := tokenRegex.FindSubmatch([]byte(value)); len(matched) != 0 {
		i, err := strconv.Atoi(string(matched[1]))
		return TokenFormat(i), err
	}

	return 0, fmt.Errorf("not a valid token format: %v", value)
}
