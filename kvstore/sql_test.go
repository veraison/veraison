// Copyright 2021-2022 Contributors to the Veraison project.
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

	cfg := Config{
		"not_type_directive": "whatever",
		"sql_driver":         "sqlite3",
		"sql_datasource":     "db=veraison.sql",
	}

	expectedErr := `missing "type" directive`

	err := s.Init(cfg)
	assert.EqualError(t, err, expectedErr)
}

func TestSQL_Init_missing_driver_name(t *testing.T) {
	s := SQL{}

	cfg := Config{
		"type":           "trustanchor",
		"sql_datasource": "db=veraison-trustanchor.sql",
	}

	expectedErr := `missing "sql_driver" directive`

	err := s.Init(cfg)
	assert.EqualError(t, err, expectedErr)
}

func TestSQL_Init_missing_datasource_name(t *testing.T) {
	s := SQL{}

	cfg := Config{
		"type":       "trustanchor",
		"sql_driver": "postgres",
	}

	expectedErr := `missing "sql_datasource" directive`

	err := s.Init(cfg)
	assert.EqualError(t, err, expectedErr)
}

// SQL drivers need to be anonymously imported by the calling code
func TestSQL_Init_db_open_unknown_driver_postgres(t *testing.T) {
	s := SQL{}

	cfg := Config{
		"type":           "trustanchor",
		"sql_driver":     "postgres",
		"sql_datasource": "db=veraison-trustanchor.sql",
	}

	expectedErr := `sql: unknown driver "postgres" (forgotten import?)`

	err := s.Init(cfg)
	assert.EqualError(t, err, expectedErr)
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

	e := mock.ExpectQuery(regexp.QuoteMeta("SELECT vals FROM endorsement WHERE key = ?"))
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

	rows := sqlmock.NewRows([]string{"vals"})
	rows.AddRow(nil)

	e := mock.ExpectQuery(regexp.QuoteMeta("SELECT vals FROM endorsement WHERE key = ?"))
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

	rows := sqlmock.NewRows([]string{"vals"})
	rows.AddRow("[1, 2]")

	e := mock.ExpectQuery(regexp.QuoteMeta("SELECT vals FROM endorsement WHERE key = ?"))
	e.WithArgs("key")
	e.WillReturnRows(rows)

	s := SQL{Type: TypeEndorsement, DB: db}

	vals, err := s.Get("key")
	assert.NoError(t, err)
	assert.Equal(t, []string{"[1, 2]"}, vals)

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

	e := mock.ExpectExec(regexp.QuoteMeta("INSERT INTO endorsement(key, vals) VALUES(?, ?)"))
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

	e := mock.ExpectExec(regexp.QuoteMeta("INSERT INTO endorsement(key, vals) VALUES(?, ?)"))
	e.WithArgs(testKey, testVal)
	e.WillReturnResult(sqlmock.NewResult(1, 1))

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

	e := mock.ExpectExec(regexp.QuoteMeta("DELETE FROM endorsement WHERE key = ?"))
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

	e := mock.ExpectExec(regexp.QuoteMeta("DELETE FROM endorsement WHERE key = ?"))
	e.WithArgs(testKey)
	e.WillReturnResult(sqlmock.NewResult(1, 1))

	err = s.Del(testKey)
	assert.NoError(t, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}
