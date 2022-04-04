// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package kvstore

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKVStore_New_nil_config(t *testing.T) {
	m, err := New(nil)

	expectedErr := `nil configuration`

	assert.EqualError(t, err, expectedErr)
	assert.Nil(t, m)
}

func TestKVStore_New_missing_backend_directive(t *testing.T) {
	cfg := Config{}

	m, err := New(cfg)

	expectedErr := `backend directive not found`

	assert.EqualError(t, err, expectedErr)
	assert.Nil(t, m)
}

func TestKVStore_New_unsupported_backend(t *testing.T) {
	cfg := Config{
		"backend": "xyz",
	}

	m, err := New(cfg)

	expectedErr := `backend "xyz" is not supported`

	assert.EqualError(t, err, expectedErr)
	assert.Nil(t, m)
}

func TestKVStore_New_memory_backend_ok(t *testing.T) {
	cfg := Config{
		"backend": "memory",
	}

	m, err := New(cfg)

	assert.NoError(t, err)
	assert.IsType(t, &Memory{}, m)
}

func TestKVStore_New_SQL_backend_failed_init(t *testing.T) {
	cfg := Config{
		"backend":        "sql",
		"sql_tablename":  "endorsement",
		"sql_datasource": "db.sql",
		// no sql_driver
	}

	m, err := New(cfg)

	expectedErr := `missing directive: "sql_driver"`

	assert.EqualError(t, err, expectedErr)
	assert.Nil(t, m)
}
