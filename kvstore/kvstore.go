// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package kvstore

import (
	"errors"
	"fmt"
)

func New(cfg Config) (IKVStore, error) {
	if cfg == nil {
		return nil, errors.New("nil configuration")
	}

	backend, err := cfg.ReadVarString(DirectiveBackend)
	if err != nil {
		return nil, errors.New(DirectiveBackend + " directive not found")
	}

	var s IKVStore

	switch backend {
	case "memory":
		s = &Memory{}
	case "sql":
		s = &SQL{}
	case "arango":
		s = &ArangoStore{}
	default:
		return nil, fmt.Errorf("backend %q is not supported", backend)
	}

	if err := s.Init(cfg); err != nil {
		return nil, err
	}

	return s, nil
}
