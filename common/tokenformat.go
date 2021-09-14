// Copyright 2021 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"encoding/json"
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

// NOTE: lower case here, as input values are strings.ToLower'd before the regex is applied.
var tokenRegex = regexp.MustCompile(`tokenformat\((\d+)\)`)

// FromString sets the TokenFormat from a string by either converting a name or
// a string representation of a number.
func (tf *TokenFormat) FromString(value string) error {
	value = strings.ToLower(value)

	switch value {
	case "psa-iat", "psa_iat", "psa", fmt.Sprint(int(PsaIatToken)):
		*tf = PsaIatToken
	case "dice", fmt.Sprint(int(DiceToken)):
		*tf = DiceToken
	default:
		match := tokenRegex.FindStringSubmatch(value)
		if match == nil {
			return fmt.Errorf("invalid TokenFormat value: %q", value)
		}
		i, err := strconv.Atoi(match[1])
		if err != nil {
			return fmt.Errorf("invalid TokenFormat value: %q; got: %v", value, err)
		}

		*tf = TokenFormat(i)
	}

	return nil
}

// UnmarshalJSON deserializes the supplied JSON encoded token format into the receiver TokenFormat
func (tf *TokenFormat) UnmarshalJSON(data []byte) error {
	var val interface{}

	if err := json.Unmarshal(data, &val); err != nil {
		return err
	}

	switch t := val.(type) {
	case float64:
		if t != float64(int64(t)) {
			return fmt.Errorf("non-integer numeric value for TokenFormat: %v", val)
		}
		*tf = TokenFormat(int64(t))
		return nil
	case string:
		return tf.FromString(t)
	default:
		return fmt.Errorf("unexpected value for TokenFormat: %v (%T)", val, val)
	}
}

// MarshalJSON serializes the receiver TokenFormat into JSON encoded token format
func (tf TokenFormat) MarshalJSON() ([]byte, error) {
	return json.Marshal(tf.String())
}

// TokenFormatFromString converts a string name into the corresponding TokenFormat.
func TokenFormatFromString(value string) (TokenFormat, error) {
	var result TokenFormat

	err := result.FromString(value)
	return result, err
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
