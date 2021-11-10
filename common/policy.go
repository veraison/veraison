// Copyright 2021 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/go-plugin"
)

// Policy encapsulates the information that indicates how a particular token
// should be validated.
type Policy struct {

	// AttestationFormat indicates the tokens to which this policy applies.
	AttestationFormat AttestationFormat

	// QueryMap specifies the queries that need to be run in order to
	// obtains the endorsements needed to validate the token.
	QueryMap QueryMap

	// Rules is a binary blob containing the rules needed to validate the
	// token. The format of these rules is determined by the verifier used.
	Rules []byte
}

// NewPolicy allocates a new policy and returns a pointer to it.
func NewPolicy() *Policy {
	policy := new(Policy)
	policy.QueryMap = make(QueryMap)
	return policy
}

// Write serializes the policy to the specified writer as a zip archive
// containing a directory named after the token format with two files --
// "rules" containing the Rules binary blob, and "query-map" containing the
// JSON serialization of the QueryMap.
func (p *Policy) Write(w io.Writer) error {
	zw := zip.NewWriter(w)

	formatName := p.AttestationFormat.String()

	rulesName := filepath.Join(formatName, "rules")
	rulesWriter, err := zw.Create(rulesName)
	if err != nil {
		return err
	}

	if _, err = rulesWriter.Write(p.Rules); err != nil {
		return err
	}

	qmapName := filepath.Join(formatName, "query-map")
	qmapWriter, err := zw.Create(qmapName)
	if err != nil {
		return err
	}

	qmapBytes, err := json.Marshal(p.QueryMap)
	if err != nil {
		return err
	}

	if _, err = qmapWriter.Write(qmapBytes); err != nil {
		return err
	}

	return zw.Close()
}

// WriteToPath functions the same way as Write above, except it takes a string
// file path instead of an io.Writer as input.
func (p *Policy) WriteToPath(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	return p.Write(f)
}

// ReadPolicies deserializes policies from the specified io.ReaderAt which is
// assumed to point ot a zip archive. The archive must contain one or more
// directories named after the token format. Each directory should contain
// exactly two files -- "rules", containing the rules for the Verifier, and
// "query-map" containing a JSON serialization of a QueryMap (see QueryMap
// documentation).
func ReadPolicies(r io.ReaderAt, size int64) ([]*Policy, error) {
	reader, err := zip.NewReader(r, size)
	if err != nil {
		return nil, err
	}
	return doReadPolicies(reader)
}

// ReadPoliciesFromBytes reads policies from a byte array buffer assumed to
// contain a zip archive (see ReadPolicies above).
func ReadPoliciesFromBytes(buffer []byte) ([]*Policy, error) {
	reader := bytes.NewReader(buffer)
	return ReadPolicies(reader, int64(len(buffer)))
}

// ReadPoliciesFromPath reads Policy zip archive (see ReadPolicies) from the
// specified path.
func ReadPoliciesFromPath(path string) ([]*Policy, error) {
	rc, err := zip.OpenReader(path)
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	return doReadPolicies(&rc.Reader)
}

// WritePoliciesToPath creates a zip archive with the specified path and
// serializes the specfied policies into it. Each policy is serialized as a
// directory with the name matching the token format of the policy. The
// directory contains two files: query_map.json and rules, that contain the
// corresponding parts of the policy.
// NOTE: each policy must be for a different token format, and it is assumed
// that the policies are for the same tenant.
func WritePoliciesToPath(policies []*Policy, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("could not open/create %q: %v", path, err)
	}
	defer file.Close()

	writer := zip.NewWriter(file)

	if err = doWritePolicies(writer, policies); err != nil {
		return fmt.Errorf("could not write policies: %v", err)
	}

	if err = writer.Close(); err != nil {
		return fmt.Errorf("could not finalize zip archive: %v", err)
	}

	return nil
}

func doWritePolicies(writer *zip.Writer, policies []*Policy) error {
	seenFormats := make(map[AttestationFormat]bool)

	for _, policy := range policies {

		if _, ok := seenFormats[policy.AttestationFormat]; ok {
			return fmt.Errorf("found multiple polcies with format %q", policy.AttestationFormat.String())
		}

		queryMapData, err := json.Marshal(policy.QueryMap)
		if err != nil {
			return fmt.Errorf(
				"could not serialize query map for %q: %v",
				policy.AttestationFormat.String(),
				err,
			)
		}

		queryMapPath := filepath.Join(policy.AttestationFormat.String(), "query_map.json")
		file, err := writer.Create(queryMapPath)
		if err != nil {
			return fmt.Errorf("could not create %q: %v", queryMapPath, err)
		}

		if _, err = file.Write(queryMapData); err != nil {
			return fmt.Errorf("could not write %q: %v", queryMapPath, err)
		}

		rulesPath := filepath.Join(policy.AttestationFormat.String(), "rules")
		file, err = writer.Create(rulesPath)
		if err != nil {
			return fmt.Errorf("could not create %q: %v", rulesPath, err)
		}

		if _, err := file.Write(policy.Rules); err != nil {
			return fmt.Errorf("could not write %q: %v", rulesPath, err)
		}

		seenFormats[policy.AttestationFormat] = true
	}

	return nil
}

// GetQueryDesriptors returns GetQueryDesriptor's for running queries against
// endorsement store in order to  obtain endorsements needed to validate a
// token. The query descriptors are populated from the Policy's QueryMap using
// the specified claims structure in order to resolve the parameter JSONpath's.
func (p *Policy) GetQueryDesriptors(
	claims map[string]interface{},
	qc QueryConstraint,
) ([]QueryDescriptor, error) {
	var qds []QueryDescriptor

	for queryName, argSpec := range p.QueryMap {
		qd := new(QueryDescriptor)
		qd.Constraint = qc

		if err := PopulateQueryDescriptor(claims, queryName, argSpec, qd); err != nil {
			return nil, err
		}

		qds = append(qds, *qd)
	}

	return qds, nil
}

func doReadPolicies(r *zip.Reader) ([]*Policy, error) {
	entries := make(map[string]*policyEntry)

	for _, f := range r.File {
		if strings.HasSuffix(f.Name, "/") {
			// do not process directory entries directly
			continue
		}

		segments := strings.Split(f.Name, string(os.PathSeparator))
		if len(segments) != 2 {
			return nil, fmt.Errorf("directory structure too deep/shallow: %v", f.Name)
		}

		formatName := Canonize(segments[0])

		entry, found := entries[formatName]
		if !found {
			entry = newPolicyEntry(formatName)
			entries[formatName] = entry
		}

		fileName := Canonize(segments[1])
		switch fileName {
		case "RULES":
			if entry.ReadRules {
				return nil, fmt.Errorf("multiple rules definitions found for %v", formatName)
			}

			rulesFile, err := f.Open()
			if err != nil {
				return nil, err
			}

			if entry.Policy.Rules, err = ioutil.ReadAll(rulesFile); err != nil {
				return nil, err
			}

			entry.ReadRules = true
		case "QUERY_MAP", "QUERY_MAP.JSON":
			if entry.ReadQMap {
				return nil, fmt.Errorf("multiple query map definitions found for %v", formatName)
			}

			qmapFile, err := f.Open()
			if err != nil {
				return nil, err
			}

			decoder := json.NewDecoder(qmapFile)
			if err = decoder.Decode(&entry.Policy.QueryMap); err != nil {
				return nil, err
			}
			entry.ReadQMap = true
		default:
			return nil, fmt.Errorf("unexpected entry: %v", f.Name)
		}

	}

	var policies []*Policy
	for _, entry := range entries {
		if !(entry.ReadQMap && entry.ReadRules) {
			return nil, fmt.Errorf("incomplete definition for '%v'", entry.AttestationFormatName)
		}
		policies = append(policies, entry.Policy)
	}

	return policies, nil
}

// QueryMap provides the mapping query_name --> param_name --> param_jsonpath,
// i.e. it maps the names of queries to run onto the maps of parameter names to
// JSONpath of the corresponding parameter in the evidence claims structure.
type QueryMap map[string]map[string]string

// PolicyStoreParams is a key-value map of configuration for initializing an
// IPolicyStore implementation. What parameters are supported is
// implementation-specific.
type PolicyStoreParams map[string]string

// IPolicyStore defines the interface that must be implemented by a policy
// store. A policy store is used to store Policy's that specify how a
// particular toke should be validated.
type IPolicyStore interface {

	// GetName returns the name of this IPolicyStore implementation. It may
	// be called without first initializing the store.
	GetName() string

	// Init initializes the policy store, creating the necessary database connections, etc.
	Init(args PolicyStoreParams) error

	// ListPolicies returns a list of entries for policies within the
	// store. If porvided tenantID is greater than zero, only etries for
	// that tenant will be returned. Otherwise, all entries will be
	// returned.
	ListPolicies(tenantID int) ([]PolicyListEntry, error)

	// GetPolicy returns the Policy stored for the specified tenant and
	// token format. If such a policy is not found, an error will be
	// returned.
	GetPolicy(tenantID int, AttestationFormat AttestationFormat) (*Policy, error)

	// PutPolicy adds a Policy for the specified tenant.
	PutPolicy(tenantID int, policy *Policy) error

	// DeletePolicy removes the policy identified by the specfied TenantID
	// and AttestationFormat from the store. If such a policy does not exist or,
	// for whatever reason, could not be removed, an error is returned.
	DeletePolicy(tenantID int, AttestationFormat AttestationFormat) error

	// Close ensures a clean shut down of the policy store, closing the
	// underlying database connections, etc.
	Close() error
}

// PolicyEngineParams params is map of key-value parameter pairs used to
// initialize an IPolicyEngine implementation. What parameters are supported is
// implementation dependent.
type PolicyEngineParams map[string]string

// IPolicyEngine defines the interface that must be implemented by a policy
// engine. The policy engine is responsible for evaluating a Policy against a
// set of evidence claims from an attestation token in order to determine
// whether or not the token is valid, as well as optionally adding some derived
// claims.
type IPolicyEngine interface {
	GetName() string

	// Init initializes a policy engine, creating any underlying
	// connections, etc. This must be called before a Policy is loaded.
	Init(args PolicyEngineParams) error

	// LoadPolicy loads a policy into the engine. This policy will then be
	// used to evaluate a set of claims. A Policy must be loaded before
	// attempting to process evidence.
	LoadPolicy(policy []byte) error

	// CheckValid returns a boolean value indicating whether the specified
	// set of evidence claims are valid with respect to the specified
	// endorsements.
	CheckValid(
		evidence map[string]interface{},
		endorsements map[string]interface{},
	) (Status, error)

	// GetClaims generates a set of additional claims derived from the
	// provided evidence claims and endorsements.
	GetClaims(
		evidence map[string]interface{},
		endorsements map[string]interface{},
	) (map[string]interface{}, error)

	// GetAttetationResult  populates the specified attestation result,
	// based on the provided evidence and endorsements. The simple flag
	// indicates wither the derived claims will be generated or only the
	// valid/invalid flag will be set.
	GetAttetationResult(
		evidence map[string]interface{},
		endorsements map[string]interface{},
		simple bool,
		result *AttestationResult,
	) error

	// Stop cleanly terminates the policy engine.
	Stop() error
}

func LoadAndInitializePolicyEngine(
	locations []string,
	name string,
	params PolicyEngineParams,
	quiet bool,
) (IPolicyEngine, *plugin.Client, plugin.ClientProtocol, error) {
	engineName := Canonize(name)

	lp, err := LoadPlugin(locations, "policyengine", engineName, quiet)
	if err != nil {
		return nil, nil, nil, err
	}

	pe := lp.Raw.(IPolicyEngine)
	client := lp.PluginClient
	rpcClient := lp.RPCClient

	if client == nil {
		return nil, nil, nil, fmt.Errorf("failed to find policy engine with name '%v'", engineName)
	}

	err = pe.Init(params)
	if err != nil {
		client.Kill()
		return nil, nil, nil, err
	}

	return pe, client, rpcClient, nil
}

// PolicyListEntry contains the listing for a policy inside an IPolicyStore.
type PolicyListEntry struct {

	// TenantID is the ID of the tenant to whom the policy belongs
	TenantID int `json:"tenant_id"`

	//  AttestationFormatName the name of the token format to which the policy applies.
	AttestationFormatName string `json:"token_format_name"`
}

type policyEntry struct {
	AttestationFormatName string
	ReadRules             bool
	ReadQMap              bool
	Policy                *Policy
}

func newPolicyEntry(name string) *policyEntry {
	entry := new(policyEntry)
	entry.AttestationFormatName = name
	entry.Policy = NewPolicy()
	entry.Policy.AttestationFormat = getTokenFromat(name)
	return entry
}

func getTokenFromat(name string) AttestationFormat {
	value, ok := AttestationFormat_value[name]
	if !ok {
		return AttestationFormat_UNKNOWN_FORMAT
	}

	return AttestationFormat(value)
}
