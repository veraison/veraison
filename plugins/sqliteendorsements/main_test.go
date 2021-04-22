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

	"github.com/veraison/common"
)

func initDb(schemaFile string) (string, error) {
	schema, err := ioutil.ReadFile(schemaFile)
	if err != nil {
		return "", err
	}

	dbf, err := ioutil.TempFile(os.TempDir(), "store-db-")
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

	store := new(SqliteEndorsementStore)

	err = store.Init(common.EndorsementStoreParams{"dbpath": dbPath})
	if err != nil {
		t.Fatalf("%v", err)
	}

	testGetSupportedQueries(t, store)
	testSupportsQuery(t, store)
	testQueryHardwareID(t, store)
	testQuerySoftwareComponents(t, store)
	testAddHardwareID(t, store)
	testAddSoftwareComponents(t, store)
}

func testGetSupportedQueries(t *testing.T, store common.IEndorsementStore) {
	assert := assert.New(t)

	var expected = map[string]bool{"hardware_id": true, "software_components": true}

	storeSupported := store.GetSupportedQueries()

	assert.Equal(len(expected), len(storeSupported),
		"Did not retrieve the right number of supported queries")

	for _, q := range storeSupported {
		_, ok := expected[q]
		assert.Truef(ok, "Retrieved unexpected query '%s'", q)
	}
}

func testSupportsQuery(t *testing.T, store common.IEndorsementStore) {
	assert := assert.New(t)

	assert.True(store.SupportsQuery("hardware_id"),
		"store does not support expected query 'hardware_id'")
	assert.False(store.SupportsQuery("fake_query"),
		"store claims to support a non-existing query")
}

func testQueryHardwareID(t *testing.T, store common.IEndorsementStore) {
	runHwIDQuery(t, store, "76543210fedcba9817161514131211101f1e1d1c1b1a1918", "acme-rr-trap")
}

func testQuerySoftwareComponents(t *testing.T, store common.IEndorsementStore) {
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

	qr, err := store.GetEndorsements(qd)
	assert.Nil(err)
	assert.NotEmpty(qr["software_components"], "Did not match software components")
}

func testAddHardwareID(t *testing.T, store common.IEndorsementStore) {
	assert := assert.New(t)

	err := store.AddEndorsement(
		"hardware_id",
		common.QueryArgs{
			"platform_id": "123456789abcdef123456789abcdef123456789abcdef123",
			"hardware_id": "test-hardware",
		},
		false,
	)
	assert.Nil(err)

	err = store.AddEndorsement(
		"hardware_id",
		common.QueryArgs{
			"platform_id": "123456789abcdef123456789abcdef123456789abcdef123",
			"hardware_id": "test-hardware",
		},
		false,
	)
	assert.Equal("UNIQUE constraint failed: hardware.platform_id", err.Error())

	runHwIDQuery(t, store, "123456789abcdef123456789abcdef123456789abcdef123", "test-hardware")

	// test update
	err = store.AddEndorsement(
		"hardware_id",
		common.QueryArgs{
			"platform_id": "123456789abcdef123456789abcdef123456789abcdef123",
			"hardware_id": "production-hardware",
		},
		true,
	)
	assert.Nil(err)

	runHwIDQuery(t, store, "123456789abcdef123456789abcdef123456789abcdef123", "production-hardware")
}

func runHwIDQuery(t *testing.T, store common.IEndorsementStore, platID string, expected string) {
	assert := assert.New(t)

	qd := common.QueryDescriptor{
		Name: "hardware_id",
		Args: map[string]interface{}{
			"platform_id": platID,
		},
		Constraint: common.QcOne,
	}

	qr, err := store.GetEndorsements(qd)
	if err != nil {
		t.Fatalf("%v", err)
	}

	assert.Equal(1, len(qr["hardware_id"]),
		"hardware_id constraint of exactly 1 match was not met")
	assert.Equal(expected, qr["hardware_id"][0],
		"hardware_id failed to match")
}

func testAddSoftwareComponents(t *testing.T, store common.IEndorsementStore) {
	assert := assert.New(t)
	require := require.New(t)

	scs := []common.SoftwareEndorsement{
		common.SoftwareEndorsement{
			Type:        "M4",
			SignerID:    "76543210fedcba9817161514131211101f1e1d1c1b1a1918",
			Version:     "1.0.1",
			Measurement: "76543210fedcba9817161514131211101f1e1d1c1b1a1920",
		},
		common.SoftwareEndorsement{
			Type:        "M5",
			SignerID:    "76543210fedcba9817161514131211101f1e1d1c1b1a1918",
			Version:     "1.0.2",
			Measurement: "76543210fedcba9817161514131211101f1e1d1c1b1a1921",
		},
	}
	scsArg, err := json.Marshal(scs)
	require.Nil(err)

	args := common.QueryArgs{
		"platform_id":         "76543210fedcba9817161514131211101f1e1d1c1b1a1918",
		"software_components": scsArg,
	}

	err = store.AddEndorsement("software_components", args, false)
	assert.Nil(err)

	// Adding an exact duplicate should succeed
	scs = []common.SoftwareEndorsement{
		common.SoftwareEndorsement{
			Type:        "M5",
			SignerID:    "76543210fedcba9817161514131211101f1e1d1c1b1a1918",
			Version:     "1.0.2",
			Measurement: "76543210fedcba9817161514131211101f1e1d1c1b1a1921",
		},
	}
	scsArg, err = json.Marshal(scs)
	require.Nil(err)

	args = common.QueryArgs{
		"platform_id":         "76543210fedcba9817161514131211101f1e1d1c1b1a1918",
		"software_components": scsArg,
	}

	err = store.AddEndorsement("software_components", args, false)
	assert.Nil(err)

	// Adding a different component for the same measurement should fail...
	scs = []common.SoftwareEndorsement{
		common.SoftwareEndorsement{
			Type:        "M5",
			SignerID:    "76543210fedcba9817161514131211101f1e1d1c1b1a1918",
			Version:     "2.0.0", // version updated
			Measurement: "76543210fedcba9817161514131211101f1e1d1c1b1a1921",
		},
	}
	scsArg, err = json.Marshal(scs)
	require.Nil(err)

	args = common.QueryArgs{
		"platform_id":         "76543210fedcba9817161514131211101f1e1d1c1b1a1918",
		"software_components": scsArg,
	}

	err = store.AddEndorsement("software_components", args, false)
	assert.Equal(
		"component with measurement \"76543210fedcba9817161514131211101f1e1d1c1b1a1921\" is already registered",
		err.Error(),
	)

	// ...unless it is being updated.
	err = store.AddEndorsement("software_components", args, true)
	assert.Nil(err)
}
