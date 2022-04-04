// Copyright 2021-2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package kvstore

import (
	"errors"
)

type Config map[string]interface{}

// Common directives -- MUST NOT be reused by specialisations
const (
	DirectiveBackend = "backend"
)

var (
	ErrMissingDirective = errors.New("missing directive")
	ErrInvalidDirective = errors.New("invalidly specified directive")
)

func (cfg Config) ReadVarString(directive string) (string, error) {
	i, ok := cfg[directive]
	if !ok {
		return "", ErrMissingDirective
	}

	s, ok := i.(string)
	if !ok {
		return "", ErrInvalidDirective
	}

	return s, nil
}
