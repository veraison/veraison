// Copyright 2021 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"

	"veraison/common"
)

func initDb(schemaFile string) (string, error) {
	schema, err := ioutil.ReadFile(schemaFile)
	if err != nil {
		return "", err
	}

	dbf, err := ioutil.TempFile(os.TempDir(), "fetcher-db-")
	if err != nil {
		return "", err
	}
	dbPath := dbf.Name()
	dbf.Close()

	dbConfig := fmt.Sprintf("file:%s?cache=shared", dbPath)
	db, err := sql.Open("sqlite3", dbConfig)
	if err != nil {
		return dbPath, err
	}
	defer db.Close()

	commands := strings.Split(string(schema), ";")
	for _, command := range commands {
		_, err := db.Exec(command)
		if err != nil {
			return dbPath, err
		}
	}

	return dbPath, nil
}

func finiDb(path string) {
	if path != "" {
		os.RemoveAll(path)
	}
}

func TestSqliteGetPolicy(t *testing.T) {
	assert := assert.New(t)

	wd, err := os.Getwd()
	if err != nil {
		t.Errorf("%v", err)
	}

	schemaFile := filepath.Join(wd, "test", "iat-policy.sqlite")
	dbPath, err := initDb(schemaFile)
	if err != nil {
		t.Errorf("%v", err)
	}
	defer finiDb(dbPath)

	pm, err := NewPolicyStore(common.PolicyStoreParams{"dbPath": dbPath})
	if err != nil {
		t.Errorf("%v", err)
	}

	policy, err := pm.GetPolicy(1, common.PsaIatToken)
	if err != nil {
		t.Errorf("%v", err)
	}

	assert.Equal(common.PsaIatToken, policy.TokenFormat)
	assert.Equal("$.implementation_id", policy.QueryMap["hardware_id"]["platform_id"])
	assert.Equal("$.sw_components[*].measurement_value",
		policy.QueryMap["software_components"]["measurements"])
}
