// Copyright 2021-2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package kvstore

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"
)

const (
	DefaultTableName = "kvstore"
)

var (
	safeTblNameRe = regexp.MustCompile(`[a-zA-Z0-9_]+`)
)

type SQL struct {
	TableName string
	DB        *sql.DB
}

func isSafeTblName(s string) bool {
	return safeTblNameRe.MatchString(s)
}

func (o *SQL) Init(cfg Config) error {
	tableName, err := cfg.ReadVarString(DirectiveSQLTableName)
	if err != nil {
		switch err {
		case ErrMissingDirective:
			o.TableName = DefaultTableName
		default:
			return fmt.Errorf("%w: %q", err, DirectiveSQLTableName)
		}
	} else {
		o.TableName = tableName
	}

	if !isSafeTblName(o.TableName) {
		return fmt.Errorf("unsafe table name: %q (MUST match %s)", o.TableName, safeTblNameRe)
	}

	driverName, err := cfg.ReadVarString(DirectiveSQLDriverName)
	if err != nil {
		return fmt.Errorf("%w: %q", err, DirectiveSQLDriverName)
	}

	dataSourceName, err := cfg.ReadVarString(DirectiveSQLDataSourceName)
	if err != nil {
		return fmt.Errorf("%w: %q", err, DirectiveSQLDataSourceName)
	}

	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return err
	}

	o.DB = db

	return nil
}

func (o *SQL) Close() error {
	return o.DB.Close()
}

func (o SQL) Get(key string) ([]string, error) {
	if o.DB == nil {
		return nil, errors.New("SQL store uninitialized")
	}

	if err := sanitizeK(key); err != nil {
		return nil, err
	}

	// nolint: gosec
	// o.TableName has been checked by isSafeTblName on init
	q := fmt.Sprintf("SELECT DISTINCT vals FROM %s WHERE key = ?", o.TableName)

	rows, err := o.DB.Query(q, key)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var vals []string

	for rows.Next() {
		var s sql.NullString

		if err := rows.Scan(&s); err != nil {
			return nil, err
		}

		if !s.Valid {
			panic("broken invariant: found val with null string")
		}

		vals = append(vals, s.String)
	}

	return vals, nil
}

func (o SQL) Add(key string, val string) error {
	if o.DB == nil {
		return errors.New("SQL store uninitialized")
	}

	if err := sanitizeKV(key, val); err != nil {
		return err
	}

	q := fmt.Sprintf("INSERT INTO %s(key, vals) VALUES(?, ?)", o.TableName)

	_, err := o.DB.Exec(q, key, val)
	if err != nil {
		return err
	}

	return nil
}

func (o SQL) Set(key string, val string) error {
	if o.DB == nil {
		return errors.New("SQL store uninitialized")
	}

	if err := sanitizeKV(key, val); err != nil {
		return err
	}

	txn, err := o.DB.Begin()
	if err != nil {
		return err
	}

	defer func() { _ = txn.Rollback() }()

	delQ := fmt.Sprintf("DELETE FROM %s WHERE key = ?", o.TableName)

	if _, err = o.DB.Exec(delQ); err != nil {
		return err
	}

	insQ := fmt.Sprintf("INSERT INTO %s(key, vals) VALUES(?, ?)", o.TableName)

	if _, err = o.DB.Exec(insQ, key, val); err != nil {
		return err
	}

	return txn.Commit()
}

func (o SQL) Del(key string) error {
	if o.DB == nil {
		return errors.New("SQL store uninitialized")
	}

	if err := sanitizeK(key); err != nil {
		return err
	}

	q := fmt.Sprintf("DELETE FROM %s WHERE key = ?", o.TableName)

	_, err := o.DB.Exec(q, key)
	if err != nil {
		return err
	}

	return nil
}
