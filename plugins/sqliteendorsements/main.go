// Copyright 2021 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"database/sql"
	"fmt"

	"reflect"

	"github.com/hashicorp/go-plugin"
	_ "github.com/mattn/go-sqlite3"

	"veraison/common"
)

type SqliteEndorsementStore struct {
	common.BaseEndorsementStore
	db   *sql.DB
	path string
}

func (e *SqliteEndorsementStore) GetName() string {
	return "sqlite"
}

// Init opens the database connection.
// Expected parameters:
//    dbPath -- the path to the database file.
func (e *SqliteEndorsementStore) Init(args common.EndorsementStoreParams) error {
	dbPath, found := args["dbPath"]
	if !found {
		return fmt.Errorf("dbPath not specified inside FetcherParams")
	}

	dbConfig := fmt.Sprintf("file:%s?cache=shared", dbPath)
	db, err := sql.Open("sqlite3", dbConfig)
	if err != nil {
		return err
	}

	e.db = db
	e.path = dbPath
	e.Queries = map[string]common.Query{
		"hardware_id":         e.GetHardwareId,
		"software_components": e.GetSoftwareComponents,
	}

	return nil
}

// Close the database connection.
func (e *SqliteEndorsementStore) Close() error {
	return e.db.Close()
}

func (e *SqliteEndorsementStore) GetHardwareId(args common.QueryArgs) (common.QueryResult, error) {
	var platformId string
	var result []interface{}

	platformIdArg, ok := args["platform_id"]
	if !ok {
		return nil, fmt.Errorf("Missing mandatory query argument 'platform_id'")
	}

	switch platformIdArg.(type) {
	case string:
		platformId = platformIdArg.(string)
	case []interface{}:
		platformId, ok = platformIdArg.([]interface{})[0].(string)
		if !ok {
			return nil, fmt.Errorf("Unexpected type for 'platform_id'; must be a string")
		}
	default:
		t := reflect.TypeOf(platformIdArg)
		return nil, fmt.Errorf("Unexpected type for 'platform_id'; must be a string; found: %v", t)
	}

	rows, err := e.db.Query("select hw_id from hardware where platform_id = ?", platformId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var hwId string
		if err := rows.Scan(&hwId); err != nil {
			return nil, err
		}

		result = append(result, hwId)
	}

	return result, nil
}

func (e *SqliteEndorsementStore) GetSoftwareComponents(args common.QueryArgs) (common.QueryResult, error) {
	//TODO: this method is horrible; do not read this; it needs to be destroyed with fire
	var platformId string
	var measurements []string

	platformIdArg, ok := args["platform_id"]
	if !ok {
		return nil, fmt.Errorf("Missing mandatory query argument 'platform_id'")
	}
	switch platformIdArg.(type) {
	case string:
		platformId = platformIdArg.(string)
	case []interface{}:
		platformId, ok = platformIdArg.([]interface{})[0].(string)
		if !ok {
			return nil, fmt.Errorf("Unexpected type for 'platform_id'; must be a string")
		}
	default:
		t := reflect.TypeOf(platformIdArg)
		return nil, fmt.Errorf("Unexpected type for 'platform_id'; must be a string; found: %v", t)
	}

	measurementsArg, ok := args["measurements"]
	if !ok {
		return nil, fmt.Errorf("Missing mandatory query argument 'platform_id'")
	}

	if measurementsArg == nil {
		return common.QueryResult{[]interface{}{}}, nil
	}

	switch measurementsArg.(type) {
	case []interface{}:
		for _, elt := range measurementsArg.([]interface{}) {
			measure, ok := elt.(string)
			if !ok {
				return nil, fmt.Errorf("Unexpected element type for 'measurements' slice; must be a string")
			}
			measurements = append(measurements, measure)
		}
	case []string:
		measurements = measurementsArg.([]string)
	default:
		return nil, fmt.Errorf("Unexpected type for 'measurements'; must be a []string")
	}

	// If no measurements provided, we automatically "match" an empty set of components
	if len(measurements) == 0 {
		return common.QueryResult{[]interface{}{}}, nil
	}

	schemeMeasMap, err := e.processSchemeMeasurements(measurements, platformId)
	if err != nil {
		return nil, err
	}

	return e.getSoftwareEndorsements(schemeMeasMap, len(measurements))
}

func (e *SqliteEndorsementStore) processSchemeMeasurements(measurements []string, platformId string) (map[int][]string, error) {
	var schemeMeasMap = make(map[int][]string)

	for _, measure := range measurements {

		rows, err := e.db.Query(
			"select scheme_id from verif_scheme_sw where measurement = ? and platform_id = ?",
			measure,
			platformId,
		)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var schemeId int
			if err := rows.Scan(&schemeId); err != nil {
				return nil, err
			}

			if _, ok := schemeMeasMap[schemeId]; !ok {
				schemeMeasMap[schemeId] = make([]string, 0, len(measurements))
			}

			schemeMeasMap[schemeId] = append(schemeMeasMap[schemeId], measure)
		}
	}

	return schemeMeasMap, nil
}

func (e *SqliteEndorsementStore) getSoftwareEndorsements(schemeMeasMap map[int][]string, numMeasurements int) ([]interface{}, error) {
	var result []interface{}

	for schemeId, schemeMeasures := range schemeMeasMap {
		if len(schemeMeasures) != numMeasurements {
			continue
		}

		var schemeEndorsements []map[string]string

		rows, err := e.db.Query(
			"select measurement, type, version, signer_id "+
				"from verif_scheme_sw "+
				"where scheme_id = ?",
			schemeId,
		)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			se := new(common.SoftwareEndoresement)
			err := rows.Scan(&se.Measurement, &se.Type, &se.Version, &se.SignerId)
			if err != nil {
				return nil, err
			}

			swMap := map[string]string{
				"measurement":          se.Measurement,
				"sw_component_type":    se.Type,
				"sw_component_version": se.Version,
				"signer_id":            se.SignerId,
			}

			schemeEndorsements = append(schemeEndorsements, swMap)
		}

		result = append(result, schemeEndorsements)
	}

	return result, nil
}

func main() {
	var handshakeConfig = plugin.HandshakeConfig{
		ProtocolVersion:  1,
		MagicCookieKey:   "VERAISON_PLUGIN",
		MagicCookieValue: "VERAISON",
	}

	var pluginMap = map[string]plugin.Plugin{
		"endorsementstore": &common.EndorsementStorePlugin{
			Impl: &SqliteEndorsementStore{},
		},
	}

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: handshakeConfig,
		Plugins:         pluginMap,
	})
}
