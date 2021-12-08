// Copyright 2021 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package kvstore

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	testKey    = `psa://tenant-1/deadbeef/beefdead`
	testVal    = `{"some": "json"}`
	altTestKey = `psa://tenant-2/cafecafe/cafecafe`
	altTestVal = `[1, 2, 3]`
)

func TestMemory_Init_missing_store_type(t *testing.T) {
	s := Memory{}

	cfg := map[string]interface{}{
		"some-random-directive": "blabla",
	}

	expectedErr := `missing "store-type" directive`

	err := s.Init(cfg)
	assert.EqualError(t, err, expectedErr)
}

func TestMemoryStore_Init_empty_store_type(t *testing.T) {
	s := Memory{}

	cfg := map[string]interface{}{
		"store-type": "",
	}

	expectedErr := `invalid "store-type": unknown type ""`

	err := s.Init(cfg)
	assert.EqualError(t, err, expectedErr)
}

func TestMemoryStore_Init_unknown_store_type(t *testing.T) {
	s := Memory{}

	cfg := map[string]interface{}{
		"store-type": "some-random-type",
	}

	expectedErr := `invalid "store-type": unknown type "some-random-type"`

	err := s.Init(cfg)
	assert.EqualError(t, err, expectedErr)
}

func TestMemoryStore_Init_bad_store_type(t *testing.T) {
	s := Memory{}

	cfg := map[string]interface{}{
		"store-type": []string{"invalid array type"},
	}

	expectedErr := `"store-type" wants string values`

	err := s.Init(cfg)
	assert.EqualError(t, err, expectedErr)
}

func TestMemoryStore_Init_Term_cycle_ok(t *testing.T) {
	s := Memory{}

	for _, typ := range []string{"trustanchor", "endorsement"} {
		cfg := map[string]interface{}{"store-type": typ}

		err := s.Init(cfg)
		assert.NoError(t, err)
		assert.Equal(t, s.Type.String(), typ)
		assert.Len(t, s.Data, 0)

		err = s.Term()
		assert.NoError(t, err)
	}
}

func TestMemoryStore_Set_Get_Del_with_uninitialised_store(t *testing.T) {
	s := Memory{}

	expectedErr := `memory store uninitialized`

	err := s.Set(testKey, testVal)
	assert.EqualError(t, err, expectedErr)

	err = s.Del(testKey)
	assert.EqualError(t, err, expectedErr)

	_, err = s.Get(testKey)
	assert.EqualError(t, err, expectedErr)
}

func TestMemoryStore_Set_Get_ok(t *testing.T) {
	s := Memory{}
	cfg := map[string]interface{}{"store-type": "endorsement"}

	err := s.Init(cfg)
	require.NoError(t, err)

	err = s.Set(testKey, testVal)
	assert.NoError(t, err)

	val, err := s.Get(testKey)
	assert.NoError(t, err)
	assert.Equal(t, testVal, val)
}

func TestMemoryStore_Get_empty_key(t *testing.T) {
	s := Memory{}
	cfg := map[string]interface{}{"store-type": "endorsement"}

	err := s.Init(cfg)
	require.NoError(t, err)

	emptyKey := ""
	expectedErr := `the supplied key is empty`

	_, err = s.Get(emptyKey)
	assert.EqualError(t, err, expectedErr)
}

func TestMemoryStore_Del_empty_key(t *testing.T) {
	s := Memory{}
	cfg := map[string]interface{}{"store-type": "endorsement"}

	err := s.Init(cfg)
	require.NoError(t, err)

	emptyKey := ""
	expectedErr := `the supplied key is empty`

	err = s.Del(emptyKey)
	assert.EqualError(t, err, expectedErr)
}

func TestMemoryStore_Set_empty_key(t *testing.T) {
	s := Memory{}
	cfg := map[string]interface{}{"store-type": "endorsement"}

	err := s.Init(cfg)
	require.NoError(t, err)

	emptyKey := ""
	expectedErr := `the supplied key is empty`

	err = s.Set(emptyKey, testVal)
	assert.EqualError(t, err, expectedErr)
}

func TestMemoryStore_Set_bad_json(t *testing.T) {
	s := Memory{}
	cfg := map[string]interface{}{"store-type": "endorsement"}

	err := s.Init(cfg)
	require.NoError(t, err)

	badJSON := "[1, 2"
	expectedErr := `the supplied val contains invalid JSON: unexpected end of JSON input`

	err = s.Set(testKey, badJSON)
	assert.EqualError(t, err, expectedErr)
}

func TestMemoryStore_Set_using_same_key_fails(t *testing.T) {
	s := Memory{}
	cfg := map[string]interface{}{"store-type": "endorsement"}

	err := s.Init(cfg)
	require.NoError(t, err)

	err = s.Set(testKey, testVal)
	require.NoError(t, err)

	expectedErr := `key "psa://tenant-1/deadbeef/beefdead" already exist, delete it first`

	// try to set with the same key fails
	err = s.Set(testKey, altTestVal)
	assert.EqualError(t, err, expectedErr)

	// delete key and retry
	err = s.Del(testKey)
	assert.NoError(t, err)

	err = s.Set(testKey, altTestVal)
	assert.NoError(t, err)

	val, err := s.Get(testKey)
	assert.NoError(t, err)
	assert.Equal(t, altTestVal, val)
}

func TestMemoryStore_Get_no_such_key(t *testing.T) {
	s := Memory{}
	cfg := map[string]interface{}{"store-type": "endorsement"}

	err := s.Init(cfg)
	require.NoError(t, err)

	expectedErr := `key "psa://tenant-1/deadbeef/beefdead" not found`

	_, err = s.Get(testKey)
	assert.EqualError(t, err, expectedErr)
}

func TestMemoryStore_dump_ok(t *testing.T) {
	s := Memory{}
	cfg := map[string]interface{}{"store-type": "endorsement"}

	err := s.Init(cfg)
	require.NoError(t, err)

	err = s.Set(testKey, testVal)
	require.NoError(t, err)
	err = s.Set(altTestKey, altTestVal)
	require.NoError(t, err)

	expectedTbl := `Key                              Val
---                              ---
psa://tenant-1/deadbeef/beefdead {"some": "json"}
psa://tenant-2/cafecafe/cafecafe [1, 2, 3]
`
	tbl := s.dump()
	assert.Equal(t, expectedTbl, tbl)

}
