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
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/veraison/common"
	"github.com/veraison/policy"
)

func initDb(schemaFile string) (string, error) {
	schema, err := ioutil.ReadFile(schemaFile)
	if err != nil {
		return "", err
	}

	dbf, err := ioutil.TempFile(os.TempDir(), "policy-db-")
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

func Test_PolicyTool_FullCycle(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	wd, err := os.Getwd()
	require.Nil(err)

	schemaFile := filepath.Join(wd, "test", "policy-db.sqlite")
	dbPath, err := initDb(schemaFile)
	require.Nil(err)
	defer finiDb(dbPath)

	pluginDir := filepath.Join(wd, "..", "..", "plugins", "bin")

	config := &common.Config{
		PluginLocations:      []string{pluginDir},
		PolicyEngineName:     "opa",
		PolicyStoreName:      "sqlite",
		EndorsementStoreName: "sqlite", // thos won't be used
		PolicyStoreParams: common.PolicyStoreParams{
			"dbpath": dbPath,
		},
		EndorsementStoreParams: common.EndorsementStoreParams{
			"dbpath": "", // this won't be used.
		},
	}

	pm := policy.NewManager()
	err = pm.InitializeStore(
		config.PluginLocations, config.PolicyStoreName, config.PolicyStoreParams,
	)
	require.Nil(err)
	defer pm.Close()

	policyZipFile := filepath.Join(wd, "test", "policy.zip")
	logger := zap.NewNop()

	err = runSetCommand(config, []string{policyZipFile}, pm, logger)
	require.Nil(err)

	err = runSetCommand(config, []string{policyZipFile}, pm, logger)
	require.NotNil(err)
	assert.Contains(err.Error(), "already exists")

	err = runSetCommand(config, []string{"-f", policyZipFile}, pm, logger)
	assert.Nil(err)

	err = runDeleteCommand(config, []string{"-t", "1", "psa"}, pm, logger)
	require.Nil(err)

	err = runDeleteCommand(config, []string{"psa"}, pm, logger)
	require.NotNil(err)
	assert.Contains(err.Error(), "no rows")
}
