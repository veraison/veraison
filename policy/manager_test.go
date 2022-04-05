// Copyright 2021-2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package policy

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/veraison/common"

	hclog "github.com/hashicorp/go-hclog"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestPutPolicyBytesAndGetPolicy(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("%v", err)
	}

	pluginDir := filepath.Join(wd, "..", "plugins", "bin")

	pluginManager, err := common.NewPluginManager([]string{pluginDir}, hclog.Info)
	require.Nil(err)
	defer pluginManager.Close()

	pm := NewManager(pluginManager)

	schemaFile := filepath.Join(wd, "test", "iat-policy.sqlite")
	dbPath, err := initDb(schemaFile)
	if err != nil {
		t.Errorf("%v", err)
	}
	defer finiDb(dbPath)

	managerParams, err := NewManagerParamStore()
	require.Nil(err)

	err = managerParams.PopulateFromMap(map[string]interface{}{
		"PolicyStoreName":   "sqlite",
		"PolicyStoreParams": map[string]interface{}{"dbpath": dbPath},
	})
	if err != nil {
		t.Errorf("%v", err)
	}

	err = pm.Init(managerParams)
	if err != nil {
		t.Fatalf("%v", err)
	}

	assert.NotNil(pm.Store)

	policyBundle := filepath.Join(wd, "test", "policy.zip")
	policyBytes, err := ioutil.ReadFile(policyBundle)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if err = pm.PutPolicyBytes(1, policyBytes); err != nil {
		t.Fatalf("%v", err)
	}

	policy, err := pm.GetPolicy(1, common.AttestationFormat_PSA_IOT)
	if err != nil {
		t.Fatalf("%v", err)
	}
	assert.Equal(common.AttestationFormat_PSA_IOT, policy.AttestationFormat)
	assert.Equal("$.implementation_id", policy.QueryMap["hardware_id"]["platform_id"])
	assert.Equal("$.sw_components[*].measurement_value",
		policy.QueryMap["software_components"]["measurements"])
}
