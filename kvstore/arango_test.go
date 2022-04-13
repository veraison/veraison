// Copyright 2021-2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package kvstore

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestArango_Init_OK(t *testing.T) {
	s := ArangoStore{}
	cfg := Config{
		"conn_endpoint":   "http://psaverifier.org:2829",
		"store_name":      "postgres",
		"collection_name": "",
		"login":           "root",
		"password":        "rootpassword",
	}

	err := s.Init(cfg)
	require.NoError(t, err)
	err = s.Close()
	require.NoError(t, err)
}
