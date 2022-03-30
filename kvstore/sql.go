// Copyright 2021 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package kvstore

import (
	"database/sql"
	"errors"
	"fmt"
)

type SQL struct {
	Type Type
	DB   *sql.DB
}

func (o *SQL) Init(cfg Config) error {
	if err := o.Type.SetFromConfig(cfg); err != nil {
		return err
	}

	driverName, err := cfg.ReadVarString(SQLDriverNameDirective)
	if err != nil {
		return err
	}

	dataSourceName, err := cfg.ReadVarString(SQLDataSourceNameDirective)
	if err != nil {
		return err
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

func (o SQL) Get(key string) (string, error) {
	if o.DB == nil {
		return "", errors.New("SQL store uninitialized")
	}

	if err := sanitizeK(key); err != nil {
		return "", err
	}

	var s sql.NullString

	// the lint warning here is spurious because o.Type.String() is not really
	// under the control of the caller -- there is a very limited set of fixed
	// outputs that it can produce
	q := fmt.Sprintf("SELECT val FROM %s WHERE key = ?", o.Type.String())

	err := o.DB.QueryRow(q, key).Scan(&s)
	if err != nil {
		return "", err
	}

	if !s.Valid {
		panic("broken invariant: found val with null string")
	}

	return s.String, nil
}

func (o SQL) Set(key string, val string) error {
	if o.DB == nil {
		return errors.New("SQL store uninitialized")
	}

	if err := sanitizeKV(key, val); err != nil {
		return err
	}

	q := fmt.Sprintf("INSERT INTO %s(key, val) VALUES(?, ?)", o.Type.String()) // nolint: gosec

	_, err := o.DB.Query(q, key, val)
	if err != nil {
		return err
	}

	return nil
}

func (o SQL) Del(key string) error {
	if o.DB == nil {
		return errors.New("SQL store uninitialized")
	}

	if err := sanitizeK(key); err != nil {
		return err
	}

	// the lint warning here is spurious because o.Type.String() is not really
	// under the control of the caller -- there is a very limited set of fixed
	// outputs that it can produce
	q := fmt.Sprintf("DELETE FROM %s WHERE key = ?", o.Type.String()) // nolint: gosec

	_, err := o.DB.Query(q, key)
	if err != nil {
		return err
	}

	return nil
}
