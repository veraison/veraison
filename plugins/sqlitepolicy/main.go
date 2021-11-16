// Copyright 2021 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/hashicorp/go-plugin"
	_ "github.com/mattn/go-sqlite3"

	"github.com/veraison/common"
)

var paramDescriptions = map[string]*common.ParamDescription{
	"dbpath": {
		Kind:     uint32(reflect.String),
		Path:     "policy.store_params.dbPath",
		Required: common.ParamNecessity_REQUIRED,
	},
}

func CreatePolicyStoreParams() (*common.ParamStore, error) {
	params := common.NewParamStore("policy_store")
	if err := params.AddParamDefinitions(paramDescriptions); err != nil {
		return nil, err
	}
	params.Freeze()
	return params, nil
}

type PolicyStore struct {
	db   *sql.DB
	path string
}

func (ps *PolicyStore) GetName() string {
	return "SQLITE"
}

func (ps PolicyStore) GetParamDescriptions() (map[string]*common.ParamDescription, error) {
	return paramDescriptions, nil
}

// Init opens the database connection.
// Expected parameters:
//    dbPath -- the path to the database file.
func (ps *PolicyStore) Init(params *common.ParamStore) error {
	dbPath, err := params.TryGetString("dbpath")
	if err != nil {
		return err
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

// ListPolicies returns a list of entries for policies within the store. If
// porvided tenantID is greater than zero, only etries for that tenant will be
// returned. Otherwise, all entries will be returned.
func (ps *PolicyStore) ListPolicies(tenantID int) ([]common.PolicyListEntry, error) {
	var result []common.PolicyListEntry
	var rows *sql.Rows
	var err error

	if tenantID > 0 {
		rows, err = ps.db.Query(
			"select tenant_id, token_format from policy where tenant_id = ?",
			tenantID,
		)
	} else {
		rows, err = ps.db.Query("select tenant_id, token_format from policy")
	}

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var entry common.PolicyListEntry
		if err := rows.Scan(&entry.TenantID, &entry.AttestationFormatName); err != nil {
			return nil, err
		}
		result = append(result, entry)
	}

	return result, nil
}

// GetPolicy returns a policy matching a tenant and the Evidence format
func (ps *PolicyStore) GetPolicy(tenantID int, tokenFormat common.AttestationFormat) (*common.Policy, error) {
	policy := common.NewPolicy()

	policy.AttestationFormat = tokenFormat

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
		tenantID, policy.AttestationFormat.String(), QueryMapBytes, policy.Rules,
	)

	return err
}

// DeletePolicy removes the policy identified by the tenantID and AttestationFormat
func (ps *PolicyStore) DeletePolicy(tenantID int, tokenFormat common.AttestationFormat) error {
	// Make sure the policy is present before deleting
	row := ps.db.QueryRow(
		"select query_map, rules from policy where tenant_id = ? and token_format = ?",
		tenantID, tokenFormat.String(),
	)

	// NOTE: row.Err() does not seem to return ErrNoRows when no matches
	// where found, and just claims the query was run successfully, so a
	// Scan() is needed to check for that.
	var queryMap []byte
	var rules []byte

	err := row.Scan(&queryMap, &rules)
	if err != nil {
		return err
	}

	_, err = ps.db.Exec(
		"delete from policy where tenant_id = ? and token_format = ?",
		tenantID, tokenFormat.String(),
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
