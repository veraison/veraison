// Copyright 2021 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"database/sql"
	"fmt"

	"github.com/hashicorp/go-plugin"
	_ "github.com/mattn/go-sqlite3"

	"github.com/veraison/common"
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
		"hardware_id":         e.GetHardwareID,
		"software_components": e.GetSoftwareComponents,
	}

	return nil
}

// Close the database connection.
func (e *SqliteEndorsementStore) Close() error {
	return e.db.Close()
}

// GetHardwareID returns the HardwareID for the platform
func (e *SqliteEndorsementStore) GetHardwareID(args common.QueryArgs) (common.QueryResult, error) {
	var platformID string
	var result []interface{}

	platformIDArg, ok := args["platform_id"]
	if !ok {
		return nil, fmt.Errorf("missing mandatory query argument 'platform_id'")
	}

	switch v := platformIDArg.(type) {
	case string:
		platformID = v
	case []interface{}:
		platformID, ok = v[0].(string)
		if !ok {
			return nil, fmt.Errorf("unexpected type for 'platform_id'; must be a string")
		}
	default:
		return nil, fmt.Errorf("unexpected type for 'platform_id'; must be a string; found: %T", v)
	}

	rows, err := e.db.Query("select hw_id from hardware where platform_id = ?", platformID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var hwID string
		if err := rows.Scan(&hwID); err != nil {
			return nil, err
		}

		result = append(result, hwID)
	}

	return result, nil
}

// GetSoftwareComponents returns the matching measurements
func (e *SqliteEndorsementStore) GetSoftwareComponents(args common.QueryArgs) (common.QueryResult, error) {
	var platformID string
	var measurements []string

	platformIDArg, ok := args["platform_id"]
	if !ok {
		return nil, fmt.Errorf("missing mandatory query argument 'platform_id'")
	}
	switch v := platformIDArg.(type) {
	case string:
		platformID = v
	case []interface{}:
		platformID, ok = v[0].(string)
		if !ok {
			return nil, fmt.Errorf("unexpected type for 'platform_id'; must be a string")
		}
	default:
		return nil, fmt.Errorf("unexpected type for 'platform_id'; must be a string; found: %T", v)
	}

	measurementsArg, ok := args["measurements"]
	if !ok {
		return nil, fmt.Errorf("missing mandatory query argument 'platform_id'")
	}

	if measurementsArg == nil {
		return common.QueryResult{[]interface{}{}}, nil
	}

	switch v := measurementsArg.(type) {
	case []interface{}:
		for _, elt := range v {
			measure, ok := elt.(string)
			if !ok {
				return nil, fmt.Errorf("unexpected element type for 'measurements' slice; must be a string")
			}
			measurements = append(measurements, measure)
		}
	case []string:
		measurements = v
	default:
		return nil, fmt.Errorf("unexpected type for 'measurements'; must be a []string; found %T", v)
	}

	// If no measurements provided, we automatically "match" an empty set of components
	if len(measurements) == 0 {
		return common.QueryResult{[]interface{}{}}, nil
	}

	schemeMeasMap, err := e.processSchemeMeasurements(measurements, platformID)
	if err != nil {
		return nil, err
	}

	return e.getSoftwareEndorsements(schemeMeasMap, len(measurements))
}

func (e *SqliteEndorsementStore) processSchemeMeasurements(measurements []string, platformID string) (map[int][]string, error) {
	var schemeMeasMap = make(map[int][]string)

	for _, measure := range measurements {

		rows, err := e.db.Query(
			"select scheme_id from verif_scheme_sw where measurement = ? and platform_id = ?",
			measure,
			platformID,
		)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var schemeID int
			if err := rows.Scan(&schemeID); err != nil {
				return nil, err
			}

			if _, ok := schemeMeasMap[schemeID]; !ok {
				schemeMeasMap[schemeID] = make([]string, 0, len(measurements))
			}

			schemeMeasMap[schemeID] = append(schemeMeasMap[schemeID], measure)
		}
	}

	return schemeMeasMap, nil
}

func (e *SqliteEndorsementStore) getSoftwareEndorsements(schemeMeasMap map[int][]string, numMeasurements int) ([]interface{}, error) {
	var result []interface{}

	for schemeID, schemeMeasures := range schemeMeasMap {
		if len(schemeMeasures) != numMeasurements {
			continue
		}

		var schemeEndorsements []map[string]string

		rows, err := e.db.Query(
			"select measurement, type, version, signer_id "+
				"from verif_scheme_sw "+
				"where scheme_id = ?",
			schemeID,
		)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			se := new(common.SoftwareEndoresement)
			err := rows.Scan(&se.Measurement, &se.Type, &se.Version, &se.SignerID)
			if err != nil {
				return nil, err
			}

			swMap := map[string]string{
				"measurement":          se.Measurement,
				"sw_component_type":    se.Type,
				"sw_component_version": se.Version,
				"signer_id":            se.SignerID,
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
