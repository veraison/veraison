// Copyright 2021 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/go-plugin"
	_ "github.com/mattn/go-sqlite3"

	"veraison/common"
)

type PolicyStore struct {
	db   *sql.DB
	path string
}

func (ps *PolicyStore) GetName() string {
	return "sqlite"
}

// Init opens the database connection.
// Expected parameters:
//    dbPath -- the path to the database file.
func (ps *PolicyStore) Init(args common.PolicyStoreParams) error {
	dbPath, found := args["dbPath"]
	if !found {
		return fmt.Errorf("dbPath not specified inside FetcherParams")
	}

	dbConfig := fmt.Sprintf("file:%s?cache=shared", dbPath)
	db, err := sql.Open("sqlite3", dbConfig)
	if err != nil {
		return err
	}

	ps.db = db
	ps.path = dbPath

	return nil
}

// GetPolicy returns a policy matching a tenant and the Evidence format
func (ps *PolicyStore) GetPolicy(tenantID int, tokenFormat common.TokenFormat) (*common.Policy, error) {
	policy := common.NewPolicy()

	policy.TokenFormat = tokenFormat

	row := ps.db.QueryRow(
		"select query_map, rules from policy where tenant_id = ? and token_format = ?",
		tenantID, tokenFormat.String(),
	)
	if err := row.Err(); err != nil {
		return nil, err
	}
	var queryMapBytes []byte
	if err := row.Scan(&queryMapBytes, &policy.Rules); err != nil {
		return nil, err
	}

	err := json.Unmarshal(queryMapBytes, &policy.QueryMap)
	if err != nil {
		return nil, err
	}

	return policy, nil
}

// PutPolicy stores a policy under the given tenant
func (ps *PolicyStore) PutPolicy(tenantID int, policy *common.Policy) error {

	QueryMapBytes, err := json.Marshal(policy.QueryMap)
	if err != nil {
		return err
	}

	_, err = ps.db.Exec(
		"insert into policy (tenant_id, token_format, query_map, rules) values (?, ?, ?, ?)",
		tenantID, policy.TokenFormat.String(), QueryMapBytes, policy.Rules,
	)

	return err
}

// Close the database connection.
func (ps *PolicyStore) Close() error {
	return ps.db.Close()
}

func main() {
	var handshakeConfig = plugin.HandshakeConfig{
		ProtocolVersion:  1,
		MagicCookieKey:   "VERAISON_PLUGIN",
		MagicCookieValue: "VERAISON",
	}

	var pluginMap = map[string]plugin.Plugin{
		"policystore": &common.PolicyStorePlugin{
			Impl: &PolicyStore{},
		},
	}

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: handshakeConfig,
		Plugins:         pluginMap,
	})
}
