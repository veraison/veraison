// Copyright 2021 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/hashicorp/go-plugin"
	_ "github.com/mattn/go-sqlite3"

	"github.com/veraison/common"
)

type SqliteEndorsementStore struct {
	common.BaseEndorsementBackend
	db   *sql.DB
	path string
}

func (e *SqliteEndorsementStore) GetName() string {
	return "SQLITE"
}

func retrieveDbPath(args common.EndorsementBackendParams) string {
	i, found := args["dbpath"]
	if !found {
		return ""
	}
	switch v := i.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	default:
		return ""
	}
}

// Init opens the database connection.
// Expected parameters:
//    args -- the input parameters to Init.
func (e *SqliteEndorsementStore) Init(args common.EndorsementBackendParams) error {
	dbPath := retrieveDbPath(args)
	if dbPath == "" {
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

	e.Adders = map[string]common.QueryAdder{
		"hardware_id":         e.AddHardwareID,
		"software_components": e.AddSoftwareComponents,
	}

	return nil
}

// Close the database connection.
func (e *SqliteEndorsementStore) Close() error {
	return e.db.Close()
}

// GetHardwareID returns the HardwareID for the platform
func (e *SqliteEndorsementStore) GetHardwareID(args common.QueryArgs) (common.QueryResult, error) {
	var result []interface{}

	platformID, err := args.GetString("platform_id")
	if err != nil {
		return nil, fmt.Errorf("error getting mandatory argument \"platform_id\": %s", err.Error())
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

func (e *SqliteEndorsementStore) AddHardwareID(args common.QueryArgs, update bool) error {
	platformID, err := args.GetString("platform_id")
	if err != nil {
		return fmt.Errorf("error getting mandatory argument \"platform_id\": %s", err.Error())
	}

	hardwareID, err := args.GetString("hardware_id")
	if err != nil {
		return fmt.Errorf("error getting mandatory argument \"hardware_id\": %s", err.Error())
	}

	if update {
		_, err = e.db.Exec(
			"update hardware set hw_id = ? where platform_id = ?",
			hardwareID,
			platformID,
		)
	} else {
		_, err = e.db.Exec(
			"insert into hardware(platform_id, hw_id) values (?, ?)",
			platformID,
			hardwareID,
		)
	}

	return err
}

// GetSoftwareComponents returns the matching measurements
func (e *SqliteEndorsementStore) GetSoftwareComponents(args common.QueryArgs) (common.QueryResult, error) {
	platformID, err := args.GetString("platform_id")
	if err != nil {
		return nil, fmt.Errorf("error getting mandatory argument \"platform_id\": %s", err.Error())
	}

	measurements, err := args.GetStringSlice("measurements")
	if err != nil {
		return nil, fmt.Errorf("error getting mandatory argument \"measurements\": %s", err.Error())
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

func (e *SqliteEndorsementStore) AddSoftwareComponents(args common.QueryArgs, update bool) error {
	platformID, err := args.GetString("platform_id")
	if err != nil {
		return fmt.Errorf("error getting mandatory argument \"platform_id\": %s", err.Error())
	}

	var components []common.SoftwareEndorsement
	if err = args.UnmarshalJSONObject("software_components", &components); err != nil {
		return err
	}

	tx, err := e.db.BeginTx(context.Background(), &sql.TxOptions{})
	if err != nil {
		return err
	}

	for _, component := range components {
		if update {
			err = e.updatePlatformSoftwareComponent(tx, platformID, component)
		} else {
			err = e.addPlatformSoftwareComponent(tx, platformID, component)
		}

		if err != nil {
			_ = tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func (e *SqliteEndorsementStore) addPlatformSoftwareComponent(
	tx *sql.Tx,
	platformID string,
	component common.SoftwareEndorsement,
) error {
	row := tx.QueryRow(
		"SELECT type, signer_id, version FROM sw_components WHERE measurement = ?",
		component.Measurement,
	)

	var typ, signerID, version string
	err := row.Scan(&typ, &signerID, &version)
	if err == nil {
		if typ != component.Type || signerID != component.SignerID || version != component.Version {
			return fmt.Errorf(
				"component with measurement \"%s\" is already registered",
				component.Measurement,
			)
		}

		// This exact component entry already exists; short-circuit here, indicating success.
		return nil
	} else if err != sql.ErrNoRows {
		return err
	}

	return e.doAddNewSoftwareComponent(tx, platformID, component)
}

func (e *SqliteEndorsementStore) updatePlatformSoftwareComponent(
	tx *sql.Tx,
	platformID string,
	component common.SoftwareEndorsement,
) error {
	row := tx.QueryRow(
		"SELECT sw_id from sw_components WHERE measurement = ?",
		component.Measurement,
	)

	var swID int
	if err := row.Scan(&swID); err != nil {
		if err == sql.ErrNoRows {
			return e.doAddNewSoftwareComponent(tx, platformID, component)
		}

		return err
	}

	// Found existing component for measurement
	_, err := tx.Exec(
		"UPDATE sw_components SET signer_id = ? , version = ? WHERE sw_id = ?",
		component.SignerID,
		component.Version,
		swID,
	)

	return err
}

func (e *SqliteEndorsementStore) doAddNewSoftwareComponent(
	tx *sql.Tx,
	platformID string,
	component common.SoftwareEndorsement,
) error {
	var swID int
	row := tx.QueryRow("SELECT MAX(sw_id) FROM sw_components")
	if err := row.Scan(&swID); err != nil {
		return err
	}
	swID++

	_, err := tx.Exec(
		"INSERT INTO sw_components(sw_id, type, signer_id, version, measurement) VALUES (?, ?, ?, ?, ?)",
		swID,
		component.Type,
		component.SignerID,
		component.Version,
		component.Measurement,
	)

	if err != nil {
		return err
	}

	_, err = tx.Exec(
		"INSERT INTO verif_scheme(scheme_id, platform_id, sw_id) VALUES (?, ?, ?)",
		1, // TODO: verification scheme has not been defined in the current implementation
		platformID,
		swID,
	)

	return err
}

func (e *SqliteEndorsementStore) processSchemeMeasurements(
	measurements []string,
	platformID string,
) (map[int][]string, error) {
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

func (e *SqliteEndorsementStore) getSoftwareEndorsements(
	schemeMeasMap map[int][]string,
	numMeasurements int,
) ([]interface{}, error) {
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
			se := new(common.SoftwareEndorsement)
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
		"endorsementstore": &common.EndorsementBackendPlugin{
			Impl: &SqliteEndorsementStore{},
		},
	}

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: handshakeConfig,
		Plugins:         pluginMap,
	})
}
