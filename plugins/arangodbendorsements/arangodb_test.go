// Copyright 2021 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/veraison/common"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// initDb does the script invocation to start the Docker for
// Arango DB and sets DB elements from json containers
func initDb(t *testing.T) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	dbpath := filepath.Join(wd, "test", "scripts")

	require.Nil(t, os.Chdir(dbpath))
	defer func() {
		require.Nil(t, os.Chdir(wd))
	}()

	CmdStartArango := &exec.Cmd{
		Path:   "./arango-start.sh",
		Args:   []string{""},
		Stdout: os.Stdout,
		Stderr: os.Stdout,
	}
	log.Printf("Starting Command = %s", CmdStartArango.String())
	if err := CmdStartArango.Run(); err != nil {
		return fmt.Errorf("failed starting arango: %w", err)
	}
	return nil
}

// finiDb does the clean up of DB and brings the docker down
func finiDb(t *testing.T) {
	wd, err := os.Getwd()
	require.Nil(t, err)

	dbpath := filepath.Join(wd, "test", "scripts")

	require.Nil(t, os.Chdir(dbpath))
	defer func() {
		require.Nil(t, os.Chdir(wd))
	}()

	CmdStopArango := &exec.Cmd{
		Path:   "./arango-stop.sh",
		Args:   []string{""},
		Stdout: os.Stdout,
		Stderr: os.Stdout,
	}

	fmt.Println("Starting Command", CmdStopArango.String())

	require.Nil(t, CmdStopArango.Run())
}

// testGetSupportedQueries checks the number of supported queries
func testGetSupportedQueries(t *testing.T, fetcher common.IEndorsementStore) {
	assert := assert.New(t)

	var expected = map[string]bool{
		"hardware_id":           true,
		"software_components":   true,
		"all_sw_components":     true,
		"linked_sw_comp_latest": true,
		"sw_component_latest":   true}

	fetcherSupported := fetcher.GetSupportedQueries()

	assert.Equal(len(expected), len(fetcherSupported),
		"Did not retrieve the right number of supported queries")

	for _, q := range fetcherSupported {
		_, ok := expected[q]
		assert.Truef(ok, "Retrieved unexpected query '%s'", q)
	}
}

// testSupportsQuery checks whether the required query is supported or not?
func testSupportsQuery(t *testing.T, fetcher common.IEndorsementStore) {
	assert := assert.New(t)

	assert.True(fetcher.SupportsQuery("hardware_id"),
		"fetcher does not support expected query 'hardware_id'")
	assert.False(fetcher.SupportsQuery("fake_query"),
		"fetcher claims to support a non-existing query")
}

// testQueryHardwareID sets the required query parameter for GetHardwareId query
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

// testQuerySoftwareComponents sets the required parameter for the SW Component Query
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

// testQueryAlternativeSwComponents checks the Alternative Software Algorithm for Query
// to fetch the components
func testQueryAlternativeSwComponents(t *testing.T, fetcher common.IEndorsementStore) {
	assert := assert.New(t)

	qd := common.QueryDescriptor{
		Name: "software_components",
		Args: map[string]interface{}{
			"platform_id": "76543210fedcba9817161514131211101f1e1d1c1b1a1918",
			"measurements": []string{
				"76543210fedcba9817161514131211101f1e1d1c1b1a1916",
				"76543210fedcba9817161514131211101f1e1d1c1b1a3117",
				"76543210fedcba9817161514131211101f1e1d1c1b1a1918",
				"76543210fedcba9817161514131211101f1e1d1c1b1a1919",
			},
		},
	}

	qr, err := fetcher.GetEndorsements(qd)
	assert.Nil(err)
	assert.NotEmpty(qr["software_components"], "Did not match software components")
}

// testQueryAllSoftwareRelations sets the query parameter to fetch all Software relations
// for a given platform
func testQueryAllSoftwareRelations(t *testing.T, fetcher common.IEndorsementStore) {
	assert := assert.New(t)

	qd := common.QueryDescriptor{
		Name: "all_sw_components",
		Args: map[string]interface{}{
			"platform_id": "76543210fedcba9817161514131211101f1e1d1c1b1a1918",
		},
		Constraint: common.QcOne,
	}

	qr, err := fetcher.GetEndorsements(qd)
	if err != nil {
		t.Fatalf("%v", err)
	}

	assert.Nil(err)
	assert.NotEmpty(qr["all_sw_components"], "Did not match software components")
}

// testQueryLatestLinkSwComp sets the query parameter to fetch the latest
// and greates SW component (update+patch) for a given software component
// linked to the platform
func testQueryLatestLinkSwComp(t *testing.T, fetcher common.IEndorsementStore) {
	assert := assert.New(t)

	qd := common.QueryDescriptor{
		Name: "linked_sw_comp_latest",
		Args: map[string]interface{}{
			"platform_id": "76543210fedcba9817161514131211101f1e1d1c1b1a1918",
			"measurements": []string{
				"76543210fedcba9817161514131211101f1e1d1c1b1a1917",
			},
		},
		Constraint: common.QcOne,
	}

	qr, err := fetcher.GetEndorsements(qd)
	if err != nil {
		t.Fatalf("%v", err)
	}
	assert.Nil(err)
	assert.NotEmpty(qr["linked_sw_comp_latest"], "Did not match software components")
}

// testQueryLatestSwComp sets the query parameter to fetch the latest
// and greates SW component (update+patch) for a given software component
// which could either be a patch or an update
func testQueryLatestSwComp(t *testing.T, fetcher common.IEndorsementStore) {
	assert := assert.New(t)

	qd := common.QueryDescriptor{
		Name: "sw_component_latest",
		Args: map[string]interface{}{
			"platform_id": "76543210fedcba9817161514131211101f1e1d1c1b1a1918",
			"measurements": []string{
				"76543210fedcba9817161514131211101f1e1d1c1b1a2017",
			},
		},
		Constraint: common.QcOne,
	}

	qr, err := fetcher.GetEndorsements(qd)
	if err != nil {
		t.Fatalf("%v", err)
	}
	assert.Nil(err)
	assert.NotEmpty(qr["sw_component_latest"], "Did not match software components")
}

// TestArangoDBStoreParams_OK checks for valid ArangoStore Params
func TestArangoDBStoreParams_OK(t *testing.T) {
	dbParam := ArangoDBparams{
		ConEndPoint:    "http://psaverifier.org:2829",
		StoreName:      "ArangoGraphDB",
		GraphName:      "psa-endorsements",
		Login:          "root",
		Password:       "rootpassword",
		HwCollection:   "hwid_collection",
		SwCollection:   "swid_collection",
		EdgeCollection: "edge_verif_scheme",
		RelCollection:  "edge_rel_scheme",
	}
	_, err := NewArangoStore(dbParam)
	assert.Nil(t, err)
}

// TestArangoDBStoreParams_NOK checks for invalid ArangoStore Params
func TestArangoDBStoreParams_NOK(t *testing.T) {
	dbParam := ArangoDBparams{
		ConEndPoint:    "psaverifier.org",
		StoreName:      "ArangoGraphDB",
		GraphName:      "psa-endorsements",
		Login:          "root",
		Password:       "rootpassword",
		HwCollection:   "hwid_collection",
		SwCollection:   "swid_collection",
		EdgeCollection: "edge_verif_scheme",
		RelCollection:  "edge_rel_scheme",
	}
	_, err := NewArangoStore(dbParam)
	wantErr := "init failed, no valid connection endpoint: supplied URL psaverifier.org is not absolute"
	assert.EqualError(t, err, wantErr)
}

// TestArangoDbEndorsementStore is the main test case, which checks for
// the required steps of ArangoDB Endoresements
func TestArangoDbEndorsementStore(t *testing.T) {
	dbParam := ArangoDBparams{
		StoreName:      "ArangoGraphDB",
		GraphName:      "psa-endorsements",
		Login:          "root",
		Password:       "rootpassword",
		HwCollection:   "hwid_collection",
		SwCollection:   "swid_collection",
		EdgeCollection: "edge_verif_scheme",
		RelCollection:  "edge_rel_scheme",
	}
	as, err := NewArangoStore(dbParam)
	if err != nil {
		t.Fatalf("%v", err)
	}
	argList := common.EndorsementStoreParams{
		"storeInstance": as,
	}
	require.Nil(t, initDb(t))

	defer finiDb(t)
	fetcher := new(EndorsementStore)

	err = fetcher.Init(argList)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer fetcher.Close()

	testGetSupportedQueries(t, fetcher)
	testSupportsQuery(t, fetcher)

	testQueryHardwareID(t, fetcher)

	testQuerySoftwareComponents(t, fetcher)
}

// TestArangoDbAlternativeSwComponent tests the alternative algorithm for
// fetching Software components based on given measurements.
func TestArangoDbAlternativeSwComponent(t *testing.T) {
	dbParam := ArangoDBparams{
		StoreName:      "ArangoGraphDB",
		GraphName:      "psa-endorsements",
		Login:          "root",
		Password:       "rootpassword",
		HwCollection:   "hwid_collection",
		SwCollection:   "swid_collection",
		EdgeCollection: "edge_verif_scheme",
		RelCollection:  "edge_rel_scheme",
	}
	as, err := NewArangoStore(dbParam)
	if err != nil {
		t.Fatalf("%v", err)
	}

	argList := common.EndorsementStoreParams{
		"AltAlgorithm":  "Normal",
		"storeInstance": as,
	}
	require.Nil(t, initDb(t))

	defer finiDb(t)
	fetcher := new(EndorsementStore)
	err = fetcher.Init(argList)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer fetcher.Close()

	testQueryAlternativeSwComponents(t, fetcher)
}

// TestArangoDbEndorsementAllSwComponents fetches all SW Components that
// exist for a given platform
func TestArangoDbEndorsementAllSwComponents(t *testing.T) {
	dbParam := ArangoDBparams{
		StoreName:      "ArangoGraphDB",
		GraphName:      "psa-endorsements",
		Login:          "root",
		Password:       "rootpassword",
		HwCollection:   "hwid_collection",
		SwCollection:   "swid_collection",
		EdgeCollection: "edge_verif_scheme",
		RelCollection:  "edge_rel_scheme",
	}
	as, err := NewArangoStore(dbParam)
	if err != nil {
		t.Fatalf("%v", err)
	}

	argList := common.EndorsementStoreParams{
		"storeInstance": as,
	}

	require.Nil(t, initDb(t))
	defer finiDb(t)
	fetcher := new(EndorsementStore)

	err = fetcher.Init(argList)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer fetcher.Close()

	testQueryAllSoftwareRelations(t, fetcher)
}

// TestArangoDbEndorsementLatestLinkedSwComponent fetches the latest and greatest upto date
// measurement for a given software linked with a platform and a specific measurement
func TestArangoDbEndorsementLatestLinkedSwComponent(t *testing.T) {
	dbParam := ArangoDBparams{
		StoreName:      "ArangoGraphDB",
		GraphName:      "psa-endorsements",
		Login:          "root",
		Password:       "rootpassword",
		HwCollection:   "hwid_collection",
		SwCollection:   "swid_collection",
		EdgeCollection: "edge_verif_scheme",
		RelCollection:  "edge_rel_scheme",
	}
	as, err := NewArangoStore(dbParam)
	if err != nil {
		t.Fatalf("%v", err)
	}

	argList := common.EndorsementStoreParams{
		"storeInstance": as,
	}
	require.Nil(t, initDb(t))
	defer finiDb(t)
	fetcher := new(EndorsementStore)

	err = fetcher.Init(argList)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer fetcher.Close()

	testQueryLatestLinkSwComp(t, fetcher)
}

// TestArangoDbEndorsementLatestSwComponent fetches the latest and greatest upto date
// measurement for a given software specific measurement, which could itself be a
// patch or an update
func TestArangoDbEndorsementLatestSwComponent(t *testing.T) {
	dbParam := ArangoDBparams{
		StoreName:      "ArangoGraphDB",
		GraphName:      "psa-endorsements",
		Login:          "root",
		Password:       "rootpassword",
		HwCollection:   "hwid_collection",
		SwCollection:   "swid_collection",
		EdgeCollection: "edge_verif_scheme",
		RelCollection:  "edge_rel_scheme",
	}
	as, err := NewArangoStore(dbParam)
	if err != nil {
		t.Fatalf("%v", err)
	}

	argList := common.EndorsementStoreParams{
		"storeInstance": as,
	}

	require.Nil(t, initDb(t))
	defer finiDb(t)
	fetcher := new(EndorsementStore)

	err = fetcher.Init(argList)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer fetcher.Close()

	testQueryLatestSwComp(t, fetcher)
}
