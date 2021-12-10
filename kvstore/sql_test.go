// Copyright 2021 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package kvstore

import (
	"errors"
	"regexp"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSQL_Init_missing_store_type(t *testing.T) {
	s := SQL{}

	cfg := map[string]interface{}{
		"no-type-directive":         "whatever",
		"store-sql-driver-name":     "sqlite",
		"store-sql-datasource-name": "db=veraison",
	}

	expectedErr := `missing "store-type" directive`

	err := s.Init(cfg)
	assert.EqualError(t, err, expectedErr)
}

func TestSQL_Init_missing_driver_name(t *testing.T) {
	s := SQL{}

	cfg := map[string]interface{}{
		"store-type":                "trustanchor",
		"store-sql-datasource-name": "db=veraison",
	}

	expectedErr := `missing "store-sql-driver-name" directive`

	err := s.Init(cfg)
	assert.EqualError(t, err, expectedErr)
}

func TestSQL_Init_missing_datasource_name(t *testing.T) {
	s := SQL{}

	cfg := map[string]interface{}{
		"store-type":            "trustanchor",
		"store-sql-driver-name": "postgres",
	}

	expectedErr := `missing "store-sql-datasource-name" directive`

	err := s.Init(cfg)
	assert.EqualError(t, err, expectedErr)
}

// SQL drivers need to be anonymously imported by the calling code
func TestSQL_Init_db_open_unknown_driver_postgres(t *testing.T) {
	s := SQL{}

	cfg := map[string]interface{}{
		"store-type":                "trustanchor",
		"store-sql-driver-name":     "postgres",
		"store-sql-datasource-name": "db=veraison",
	}

	expectedErr := `sql: unknown driver "postgres" (forgotten import?)`

	err := s.Init(cfg)
	assert.EqualError(t, err, expectedErr)
}

func TestSQL_Init_Close_cycle_ok(t *testing.T) {
	s := Memory{}

	for _, typ := range []string{"trustanchor", "endorsement"} {
		cfg := map[string]interface{}{"store-type": typ}

		err := s.Init(cfg)
		assert.NoError(t, err)
		assert.Equal(t, s.Type.String(), typ)
		assert.Len(t, s.Data, 0)

		err = s.Close()
		assert.NoError(t, err)
	}
}

func TestSQL_Set_Get_Del_with_uninitialised_store(t *testing.T) {
	s := SQL{}

	expectedErr := `SQL store uninitialized`

	err := s.Set(testKey, testVal)
	assert.EqualError(t, err, expectedErr)

	err = s.Del(testKey)
	assert.EqualError(t, err, expectedErr)

	_, err = s.Get(testKey)
	assert.EqualError(t, err, expectedErr)
}

func TestSQL_Get_empty_key(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	s := SQL{Type: TypeEndorsement, DB: db}

	emptyKey := ""

	expectedErr := `the supplied key is empty`

	_, err = s.Get(emptyKey)
	assert.EqualError(t, err, expectedErr)
}

func TestSQL_Get_db_layer_failure(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	s := SQL{Type: TypeEndorsement, DB: db}

	dbErrorString := "a DB error"

	e := mock.ExpectQuery(regexp.QuoteMeta("SELECT val FROM endorsement WHERE key = ?"))
	e.WithArgs("key")
	e.WillReturnError(errors.New(dbErrorString))

	expectedErr := dbErrorString

	_, err = s.Get("key")
	assert.EqualError(t, err, expectedErr)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestSQL_Get_broken_invariant_null_val_panic(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	s := SQL{Type: TypeEndorsement, DB: db}

	rows := sqlmock.NewRows([]string{"val"})
	rows.AddRow(nil)

	e := mock.ExpectQuery(regexp.QuoteMeta("SELECT val FROM endorsement WHERE key = ?"))
	e.WithArgs("key")
	e.WillReturnRows(rows)

	assert.Panics(t, func() { _, _ = s.Get("key") })

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestSQL_Get_ok(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"val"})
	rows.AddRow("[1, 2]")

	e := mock.ExpectQuery(regexp.QuoteMeta("SELECT val FROM endorsement WHERE key = ?"))
	e.WithArgs("key")
	e.WillReturnRows(rows)

	s := SQL{Type: TypeEndorsement, DB: db}

	val, err := s.Get("key")
	assert.NoError(t, err)
	assert.Equal(t, "[1, 2]", val)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestSQL_Set_empty_key(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	s := SQL{Type: TypeEndorsement, DB: db}

	emptyKey := ""

	expectedErr := `the supplied key is empty`

	err = s.Set(emptyKey, testVal)
	assert.EqualError(t, err, expectedErr)
}

func TestSQL_Set_bad_val(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	s := SQL{Type: TypeEndorsement, DB: db}

	invalidJSON := ""

	expectedErr := `the supplied val contains invalid JSON: unexpected end of JSON input`

	err = s.Set(testKey, invalidJSON)
	assert.EqualError(t, err, expectedErr)
}

func TestSQL_Set_db_layer_failure(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	s := SQL{Type: TypeEndorsement, DB: db}

	dbErrorString := "a DB error"

	e := mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO endorsement(key, val) VALUES(?, ?)"))
	e.WithArgs(testKey, testVal)
	e.WillReturnError(errors.New(dbErrorString))

	expectedErr := dbErrorString

	err = s.Set(testKey, testVal)
	assert.EqualError(t, err, expectedErr)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestSQL_Set_ok(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	s := SQL{Type: TypeEndorsement, DB: db}

	e := mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO endorsement(key, val) VALUES(?, ?)"))
	e.WithArgs(testKey, testVal)
	e.WillReturnRows()

	err = s.Set(testKey, testVal)
	assert.NoError(t, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestSQL_Del_empty_key(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	s := SQL{Type: TypeEndorsement, DB: db}

	emptyKey := ""

	expectedErr := `the supplied key is empty`

	err = s.Set(emptyKey, testVal)
	assert.EqualError(t, err, expectedErr)
}

func TestSQL_Del_db_layer_failure(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	s := SQL{Type: TypeEndorsement, DB: db}

	dbErrorString := "a DB error"

	e := mock.ExpectQuery(regexp.QuoteMeta("DELETE FROM endorsement WHERE key = ?"))
	e.WithArgs(testKey)
	e.WillReturnError(errors.New(dbErrorString))

	expectedErr := dbErrorString

	err = s.Del(testKey)
	assert.EqualError(t, err, expectedErr)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestSQL_Del_ok(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	s := SQL{Type: TypeEndorsement, DB: db}

	e := mock.ExpectQuery(regexp.QuoteMeta("DELETE FROM endorsement WHERE key = ?"))
	e.WithArgs(testKey)
	e.WillReturnRows()

	err = s.Del(testKey)
	assert.NoError(t, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}
