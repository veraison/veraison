// Copyright 2021 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package endorsement

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/veraison/common"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

func initDb(schemaFile string) (string, error) {
	schema, err := ioutil.ReadFile(schemaFile)
	if err != nil {
		return "", err
	}

	dbf, err := ioutil.TempFile(os.TempDir(), "endorsement-db-")
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

func TestLoadStore(t *testing.T) {
	assert := assert.New(t)

	pm := NewManager()

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("%v", err)
	}

	schemaFile := filepath.Join(wd, "test", "iat-endorsement.sqlite")
	dbPath, err := initDb(schemaFile)
	if err != nil {
		t.Errorf("%v", err)
	}
	defer finiDb(dbPath)

	pluginDir := filepath.Join(wd, "..", "plugins", "bin")

	params := common.EndorsementStoreParams{"dbPath": dbPath}
	err = pm.InitializeStore([]string{pluginDir}, "sqlite", params)
	if err != nil {
		t.Fatalf("%v", err)
	}

	assert.NotNil(pm.Store)

	supported := pm.Store.GetSupportedQueries()
	sort.Strings(supported)
	assert.Equal([]string{"hardware_id", "software_components"}, supported)
}
