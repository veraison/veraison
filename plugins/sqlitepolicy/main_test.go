// Copyright 2021-2022 Contributors to the Veraison project.
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
	"github.com/stretchr/testify/require"

	"github.com/veraison/common"
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
	require := require.New(t)
	assert := assert.New(t)

	wd, err := os.Getwd()
	require.Nil(err)

	schemaFile := filepath.Join(wd, "test", "iat-policy.sqlite")
	dbPath, err := initDb(schemaFile)
	require.Nil(err)
	defer finiDb(dbPath)

	var pm PolicyStore

	params, err := CreatePolicyStoreParams()
	require.Nil(err)
	require.Nil(params.SetString("dbpath", dbPath))

	err = pm.Init(params)
	require.Nil(err)

	policy, err := pm.GetPolicy(1, common.AttestationFormat_PSA_IOT)
	require.Nil(err)

	assert.Equal(common.AttestationFormat_PSA_IOT, policy.AttestationFormat)
	assert.Equal("$.implementation_id", policy.QueryMap["hardware_id"]["platform_id"])
	assert.Equal("$.sw_components[*].measurement_value",
		policy.QueryMap["software_components"]["measurements"])

	policy, err = pm.GetPolicy(1, common.AttestationFormat(123))
	require.NotNil(err)
	assert.Contains(err.Error(), "no rows")
	assert.Nil(policy)
}

func TestSqliteDeletePolicy(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	wd, err := os.Getwd()
	require.Nil(err)

	schemaFile := filepath.Join(wd, "test", "iat-policy.sqlite")
	dbPath, err := initDb(schemaFile)
	require.Nil(err)
	defer finiDb(dbPath)

	var pm PolicyStore

	params, err := CreatePolicyStoreParams()
	require.Nil(err)
	require.Nil(params.SetString("dbpath", dbPath))

	err = pm.Init(params)
	require.Nil(err)

	err = pm.DeletePolicy(1, common.AttestationFormat_PSA_IOT)
	assert.Nil(err)

	err = pm.DeletePolicy(1, common.AttestationFormat_PSA_IOT)
	require.NotNil(err)
	assert.Contains(err.Error(), "no rows")
}
