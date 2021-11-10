// Copyright 2021 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"crypto/x509"
	"database/sql"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"

	"github.com/hashicorp/go-plugin"
	_ "github.com/mattn/go-sqlite3"

	"github.com/veraison/common"
)

var insertQueryText = "INSERT INTO trust_anchor(tenant_id, ta_type, ta_id, value) VALUES (?, ?, ?, ?)"
var byTypeQueryText = "SELECT value FROM trust_anchor  WHERE tenant_id = ? AND ta_type = ?"
var byIDQueryText = "SELECT value FROM trust_anchor  WHERE tenant_id = ? AND ta_id = ?"

type SqliteTrustAnchorStore struct {
	db   *sql.DB
	path string
}

func (s SqliteTrustAnchorStore) GetName() string {
	return "SQLITE"
}

// Init opens the database connection.
// Expected parameters:
//    dbPath -- the path to the database file.
func (s *SqliteTrustAnchorStore) Init(params common.TrustAnchorStoreParams) error {
	dbPath, found := params["dbpath"]
	if !found {
		return errors.New("\"dbpath\" trust anchor parameter not specified")
	}

	dbConfig := fmt.Sprintf("file:%s?cache=shared", dbPath)
	db, err := sql.Open("sqlite3", dbConfig)
	if err != nil {
		return err
	}

	s.db = db
	s.path = dbPath

	return nil
}

func (s SqliteTrustAnchorStore) AddCertsFromPEM(tenantID int, value []byte) error {
	rest := value
	var block *pem.Block

	for len(rest) != 0 {
		block, rest = pem.Decode(rest)
		if block == nil {
			return errors.New("problem extracting token cert PEM block")
		}

		// ensure the PEM block contains a well-structured certificate
		if _, err := x509.ParseCertificate(block.Bytes); err != nil {
			return err
		}

		if _, err := s.db.Exec(insertQueryText, tenantID, common.TaTypeCert, nil, block.Bytes); err != nil {
			return err
		}

	}

	return nil
}

func (s SqliteTrustAnchorStore) AddPublicKeyFromPEM(tenantID int, id interface{}, value []byte) error {
	block, rest := pem.Decode(value)

	if len(rest) != 0 {
		return errors.New("trailing data after PEM block")
	}

	if block == nil {
		return errors.New("problem extracting token cert PEM block")
	}

	// ensure the PEM block contains a well-structured key
	_, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return err
	}

	var kid string

	switch v := id.(type) {
	case string:
		kid = v
	case []byte:
		kid = hex.EncodeToString(v)
	case int32, int64, uint32, uint64:
		kid = fmt.Sprint(v)
	default:
		return fmt.Errorf("unsupported key value: %q", v)
	}

	if _, err := s.db.Exec(insertQueryText, tenantID, common.TaTypeKey, kid, block.Bytes); err != nil {
		return err
	}

	return nil
}

func (s SqliteTrustAnchorStore) GetTrustAnchor(tenantID int, taID common.TrustAnchorID) ([]byte, error) {
	switch taID.Type {
	case common.TaTypeCert:
		return s.getCerts(tenantID)
	case common.TaTypeKey:
		var kid string

		switch v := taID.Value["key-id"].(type) {
		case string:
			kid = v
		case []byte:
			kid = hex.EncodeToString(v)
		case int32, int64, uint32, uint64:
			kid = fmt.Sprint(v)
		default:
			return nil, fmt.Errorf("unsupported key value: %q", v)
		}

		return s.getKey(tenantID, kid)
	default:
		return nil, fmt.Errorf("trust anchor of type %s not supported", taID.Type.String())
	}
}

func (s SqliteTrustAnchorStore) getCerts(tenantID int) ([]byte, error) {
	var result []byte

	rows, err := s.db.Query(byTypeQueryText, tenantID, common.TaTypeCert)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var certData []byte
		if err := rows.Scan(&certData); err != nil {
			return nil, err
		}

		block := pem.Block{Type: "CERTIFICATE", Bytes: certData}
		result = append(result, pem.EncodeToMemory(&block)...)
	}

	return result, nil
}

func (s SqliteTrustAnchorStore) getKey(tenantID int, kid string) ([]byte, error) {
	var keyData []byte
	err := s.db.QueryRow(byIDQueryText, tenantID, kid).Scan(&keyData)
	if err != nil {
		return nil, err
	}

	block := pem.Block{Type: "PUBLIC KEY", Bytes: keyData}
	return pem.EncodeToMemory(&block), nil
}

// Close the database connection.
func (s SqliteTrustAnchorStore) Close() error {
	return s.db.Close()
}

func main() {
	var handshakeConfig = plugin.HandshakeConfig{
		ProtocolVersion:  1,
		MagicCookieKey:   "VERAISON_PLUGIN",
		MagicCookieValue: "VERAISON",
	}

	var pluginMap = map[string]plugin.Plugin{
		"trustanchorstore": &common.TrustAnchorStorePlugin{
			Impl: &SqliteTrustAnchorStore{},
		},
	}

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: handshakeConfig,
		Plugins:         pluginMap,
	})
}
