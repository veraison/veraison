// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package main

import "context"

// Store represents an interface towards any suitable Database store
type Store interface {
	// IsInitialised returns true or false based on store been initialized or not?
	IsInitialised() bool

	// Connect makes a connection to the store prior to running any queries
	Connect(ctx context.Context) error

	// GetQueryParam returns the desired query parameter
	GetQueryParam(qParamID string) (string, error)

	// RunQuery executes a query on the store and returns a list of documents
	RunQuery(ctx context.Context, query string, queryArgs map[string]interface{}, document interface{}) ([]interface{}, error)

	// Close closes the data base
	Close(ctx context.Context) error
}
