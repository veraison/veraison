// Copyright 2021-2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package kvstore

import "fmt"

type Config map[string]interface{}

// Common directives -- MUST NOT be reused by specialisations
const (
	DirectiveType    = "type"
	DirectiveBackend = "backend"
)

func (cfg Config) ReadVarString(directive string) (string, error) {
	i, ok := cfg[directive]
	if !ok {
		return "", fmt.Errorf("missing %q directive", directive)
	}

	s, ok := i.(string)
	if !ok {
		return "", fmt.Errorf("%q wants string values", directive)
	}

	return s, nil
}

type Type uint8

const (
	TypeUnset Type = iota
	TypeTrustAnchor
	TypeEndorsement
)

var (
	typeToString = map[Type]string{
		TypeUnset:       "unset",
		TypeTrustAnchor: "trustanchor",
		TypeEndorsement: "endorsement",
	}

	stringToType = map[string]Type{
		"unset":       TypeUnset,
		"trustanchor": TypeTrustAnchor,
		"endorsement": TypeEndorsement,
	}
)

func (o Type) String() string {
	v, ok := typeToString[o]
	if !ok {
		return "unknown"
	}
	return v
}

func (o *Type) FromString(s string) error {
	v, ok := stringToType[s]
	if !ok {
		return fmt.Errorf("unknown type %q", s)
	}
	*o = v
	return nil
}

func (o *Type) SetFromConfig(cfg Config) error {
	s, err := cfg.ReadVarString(DirectiveType)
	if err != nil {
		return err
	}

	err = o.FromString(s)
	if err != nil {
		return fmt.Errorf("invalid value for %q: %w", DirectiveType, err)
	}

	return nil
}
