// Copyright 2021 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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

func readJSON(path string, v interface{}) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, v)
	return err
}

func TestPopulateQueryDescriptor(t *testing.T) {
	assert := assert.New(t)

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("%v", err)
	}

	var claims map[string]interface{}
	claimsFile := filepath.Join(wd, "test", "iat-input.json")
	require.Nil(t, readJSON(claimsFile, &claims))

	var querySpecs map[string]map[string]string
	specsFile := filepath.Join(wd, "test", "iat-queries.json")
	require.Nil(t, readJSON(specsFile, &querySpecs))

	var qd common.QueryDescriptor

	err = common.PopulateQueryDescriptor(claims, "hardware_id", querySpecs["hardware_id"], &qd)
	assert.Nil(err, "failed to populte query descriptor")

	assert.Equal(qd.Name, "hardware_id", "QueryDescriptor name not set properly")
	expectedArgs := common.QueryArgs{
		"platform_id": []interface{}{"76543210fedcba9817161514131211101f1e1d1c1b1a1918"},
	}
	assert.Equal(expectedArgs, qd.Args, "QueryDescriptor arguments not set properly")

	err = common.PopulateQueryDescriptor(claims, "software_components", querySpecs["software_components"], &qd)
	assert.Nil(err, "failed to populate query descriptor")

	expectedArgs = common.QueryArgs{
		"platform_id": []interface{}{"76543210fedcba9817161514131211101f1e1d1c1b1a1918"},
		"measurements": []interface{}{
			"76543210fedcba9817161514131211101f1e1d1c1b1a1916",
			"76543210fedcba9817161514131211101f1e1d1c1b1a1917",
			"76543210fedcba9817161514131211101f1e1d1c1b1a1918",
			"76543210fedcba9817161514131211101f1e1d1c1b1a1919",
		},
	}
	assert.Equal(expectedArgs, qd.Args, "QueryDescriptor arguments not set properly")
}

func TestParseQueryDescriptors(t *testing.T) {
	assert := assert.New(t)

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("%v", err)
	}

	var claims map[string]interface{}
	claimsFile := filepath.Join(wd, "test", "iat-input.json")
	require.Nil(t, readJSON(claimsFile, &claims))

	specsFile := filepath.Join(wd, "test", "iat-queries.json")
	data, err := ioutil.ReadFile(specsFile)
	if err != nil {
		t.Fatalf("%v", err)
	}

	expectedQds := []common.QueryDescriptor{
		{
			Name: "hardware_id",
			Args: common.QueryArgs{
				"platform_id": []interface{}{"76543210fedcba9817161514131211101f1e1d1c1b1a1918"},
			},
			Constraint: common.QcNone,
		},
		{
			Name: "software_components",
			Args: common.QueryArgs{
				"platform_id": []interface{}{"76543210fedcba9817161514131211101f1e1d1c1b1a1918"},
				"measurements": []interface{}{
					"76543210fedcba9817161514131211101f1e1d1c1b1a1916",
					"76543210fedcba9817161514131211101f1e1d1c1b1a1917",
					"76543210fedcba9817161514131211101f1e1d1c1b1a1918",
					"76543210fedcba9817161514131211101f1e1d1c1b1a1919",
				},
			},
			Constraint: common.QcNone,
		},
	}

	qds, err := common.ParseQueryDescriptors(claims, data)
	if err != nil {
		t.Fatalf("%v", err)
	}

	sort.Sort(common.QueryDescriptorsByName(qds))

	for i, qd := range qds {
		eqd := expectedQds[i]

		assert.Equal(eqd.Name, qd.Name, "descriptor query name matches")
	}

}

func TestSqliteEndorsementStore(t *testing.T) {

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("%v", err)
	}

	schemaFile := filepath.Join(wd, "test", "iat-data.sqlite")
	dbPath, err := initDb(schemaFile)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer finiDb(dbPath)

	fetcher := new(SqliteEndorsementStore)

	err = fetcher.Init(common.EndorsementStoreParams{"dbPath": dbPath})
	if err != nil {
		t.Fatalf("%v", err)
	}

	testGetSupportedQueries(t, fetcher)
	testSupportsQuery(t, fetcher)
	testQueryHardwareID(t, fetcher)
	testQuerySoftwareComponents(t, fetcher)
}

func testGetSupportedQueries(t *testing.T, fetcher common.IEndorsementStore) {
	assert := assert.New(t)

	var expected = map[string]bool{"hardware_id": true, "software_components": true}

	fetcherSupported := fetcher.GetSupportedQueries()

	assert.Equal(len(expected), len(fetcherSupported),
		"Did not retrieve the right number of supported queries")

	for _, q := range fetcherSupported {
		_, ok := expected[q]
		assert.Truef(ok, "Retrieved unexpected query '%s'", q)
	}
}

func testSupportsQuery(t *testing.T, fetcher common.IEndorsementStore) {
	assert := assert.New(t)

	assert.True(fetcher.SupportsQuery("hardware_id"),
		"fetcher does not support expected query 'hardware_id'")
	assert.False(fetcher.SupportsQuery("fake_query"),
		"fetcher claims to support a non-existing query")
}

func testQueryHardwareID(t *testing.T, fetcher common.IEndorsementStore) {
	assert := assert.New(t)

	qd := common.QueryDescriptor{
		Name: "hardware_id",
		Args: map[string]interface{}{
			"platform_id": "76543210fedcba9817161514131211101f1e1d1c1b1a1918",
		},
		Constraint: common.QcOne,
	}

	qr, err := fetcher.GetEndorsements(qd)
	if err != nil {
		t.Fatalf("%v", err)
	}

	assert.Equal(1, len(qr["hardware_id"]),
		"hardware_id constraint of exactly 1 match was not met")
	assert.Equal("acme-rr-trap", qr["hardware_id"][0],
		"hardware_id failed to match")
}

func testQuerySoftwareComponents(t *testing.T, fetcher common.IEndorsementStore) {
	assert := assert.New(t)

	qd := common.QueryDescriptor{
		Name: "software_components",
		Args: map[string]interface{}{
			"platform_id": "76543210fedcba9817161514131211101f1e1d1c1b1a1918",
			"measurements": []string{
				"76543210fedcba9817161514131211101f1e1d1c1b1a1916",
				"76543210fedcba9817161514131211101f1e1d1c1b1a1917",
				"76543210fedcba9817161514131211101f1e1d1c1b1a1918",
				"76543210fedcba9817161514131211101f1e1d1c1b1a1919",
			},
		},
	}

	qr, err := fetcher.GetEndorsements(qd)
	assert.Nil(err)
	assert.NotEmpty(qr["software_components"], "Did not match software components")
}
