// Copyright 2021 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"fmt"
)

// EndorsementStoreParams encapsulates parameters used to initialize the
// endorsement store. What parameters are required and values supported depends
// on the specific IEndorsementStore implementation.
type EndorsementStoreParams map[string]interface{}

// SoftwareEndoresement encapsulates an endorsed measurement value associated
// with a particular version of a software component. It represents a know good
// build of the component. How this measurement value is calculated is
// endorsement scheme dependant, but it is typically a hash of the code and/or
// read only memory layout of the process associated with the software
// component.
type SoftwareEndorsement struct {
	Measurement string `json:"measurement_value"`
	Type        string `json:"sw_component_type"`
	Version     string `json:"sw_component_version"`
	SignerID    string `json:"signer_id"`
}

// EndorsementMatches maps query name onto the corresponding QueryResult. This
// is the result of matching endorsements from a sore suing GetEndorsements().
type EndorsementMatches map[string]QueryResult

// IEndorsementStore defines the interface that must be provide by an
// endorsement store implementation. An endorsement store is used to store and
// query endorsement values against which an attested device may be evaluated.
type IEndorsementStore interface {

	// GetName returns the name of the implemented endorsement store. This
	// name is used to configure which store will be used by a deployment.
	// This is the only method that  may be safely called prior to Init.
	GetName() string

	// Init initializes the endorsement store based on the specified
	// parameters. What initialization is performed depends on the specific
	// implementation, but would typically involve things such as
	// establishing connection to a database server. This must be called
	// before running any queries against the store.
	Init(args EndorsementStoreParams) error

	// GetEndorsements queries the underlying store for endorsements based
	// on the specified QueryDescriptor's. Each query descriptor specified
	// the name of the query to run and the corresponding parameters, and
	// will produce a corresponding QueryResult entry in the returned
	// EndorsementMatches.
	GetEndorsements(qds ...QueryDescriptor) (EndorsementMatches, error)

	// AddEndorsement adds endorsment data of the type specified by name.
	// The endorsement data is interpreted from args based on the name. If
	// update flag is set to true, and existing endrsement (identified from
	// the args) is updated instead.
	AddEndorsement(name string, args QueryArgs, update bool) error

	// RunQuery executes the query with the specified name against the
	// parameter values specified by the provided QueryArgs and return the
	// corresponding QueryResult. The name must be of a query supported by
	// the particular IEndorsementStore implementation (see
	// GetSupportedQueries below) otherwise an error will be returned.
	RunQuery(name string, args QueryArgs) (QueryResult, error)

	// GetSupportedQueries returns a list of the names of the queries
	// supported by the IEndorsementStore implementation. Only the queries
	// whos names are turned may be run against this store.
	GetSupportedQueries() []string

	// SupportsQuery returns a boolean value indicating whether the query
	// with the specified name is supported by the IEndorsementStore
	// implementation.
	SupportsQuery(query string) bool

	// Close cleanly terminates everything that was initialized during Init
	// (e.g. closing database connections). No additional queries may be
	// run after this has been called with first re-initializing with Init.
	Close() error
}

// BaseEndorsementStore implements generic functionality that may be shared
// cross all IEndorsementStore implementations.
type BaseEndorsementStore struct {

	// Queries is a map of names to Query functions populated by a
	// particular IEndorsementStore implementation. Each implementation
	// defines a Query function for every type of query it supports and
	// "registers" them under a name here.
	Queries map[string]Query

	// Adders is a map of names of endorsement types to the Adder function
	// used to add/update those endorsements. Each IEndorsementStore
	// implementation defines its own adders and registers them here.
	Adders map[string]QueryAdder
}

func (e *BaseEndorsementStore) AddEndorsement(name string, args QueryArgs, update bool) error {
	adderFunc, ok := e.Adders[name]
	if !ok {
		return fmt.Errorf("no adder specified for %q", name)
	}

	return adderFunc(args, update)
}

// GetEndorsements invokes RunQuery for each of the specified QueryDescriptor's
// and collects returned QueryResult's inside EndorsementMatches.
func (e *BaseEndorsementStore) GetEndorsements(qds ...QueryDescriptor) (EndorsementMatches, error) {
	matches := make(EndorsementMatches)

	for _, qd := range qds {
		qr, err := e.RunQuery(qd.Name, qd.Args)
		if err != nil {
			return nil, err
		}

		if !checkConstraintHolds(qr, qd.Constraint) {
			return nil, fmt.Errorf("result for query %q failed to satisfy constraint", qd.Name)
		}

		matches[qd.Name] = qr
	}

	return matches, nil
}

// RunQuery retrieves the Query function associated with the specified name and
// executes it with the provided QueryArgs, returning the QueryResult.
func (e *BaseEndorsementStore) RunQuery(name string, args QueryArgs) (QueryResult, error) {
	queryFunc, ok := e.Queries[name]
	if !ok {
		return nil, fmt.Errorf("query %q not implemented", name)
	}

	result, err := queryFunc(args)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// GetSupportedQueries returns the list of name of queries supported by the
// IEndorsementStore implementation.
func (e *BaseEndorsementStore) GetSupportedQueries() []string {
	keys := make([]string, 0, len(e.Queries))
	for k := range e.Queries {
		keys = append(keys, k)
	}
	return keys
}

// SupportsQuery returns a boolean value indicating whether the query
// with the specified name is supported by the IEndorsementStore
// implementation.
func (e *BaseEndorsementStore) SupportsQuery(query string) bool {
	if _, ok := e.Queries[query]; ok {
		return true
	}
	return false
}

func checkConstraintHolds(r QueryResult, c QueryConstraint) bool {
	switch c {
	case QcNone:
		return true
	case QcZero:
		return len(r) == 0
	case QcOne:
		return len(r) == 1
	case QcOneOrMore:
		return len(r) != 0
	case QcMultiple:
		return len(r) > 1
	default:
		return false
	}
}
