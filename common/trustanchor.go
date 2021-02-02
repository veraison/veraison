// Copyright 2021 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package common

import "fmt"

const (
	// The trust anchor value is not of one of the supported types. The value is assumed to be
	// stored as a single blob unique identified by the specfied ID Value parts.
	TaTypeUnspecified = TrustAnchorType(iota)

	// The trust anchor is a public key. The returned value is the DER encoding of this key.
	// the ID Value is the unique KID identifying the key.
	TaTypeKey

	// The trust anchor is a cert pool. The returned value is the concatenation of PEM encodings
	// of all certs in the store that match the ID Value.
	TaTypeCert
)

// TrustAnchorType indicates the type of trust anchor to be stored or retrieved
// to/from the store.
type TrustAnchorType int8

// String provieds a textual representation of the TrustAnchorType.
func (t TrustAnchorType) String() string {
	switch t {
	case TaTypeUnspecified:
		return "unspecified"
	case TaTypeKey:
		return "public key"
	case TaTypeCert:
		return "certificate"
	default:
		return fmt.Sprintf("TrustAnchorType(%v)", t)
	}
}

// TrustAnchorID identifies the trust anchor to be retrieved from the store.
type TrustAnchorID struct {

	// Type indicates the type of the trust anchor, e.g. certificate.
	Type TrustAnchorType

	// The value is the actual ID of the trust anchor. It is a map with
	// key-value pairs of ID parts. What parts comprise an ID is defined by
	// the TrustAnchorType.
	Value map[string]interface{}
}

// TrustAnchorStoreParams is a map of key-value pairs of parameters used to
// initialize an ITrustAnchorStore instance. What parameters are valid depends
// on a particular implementation.
type TrustAnchorStoreParams map[string]string

// ITrustAnchorStore defines the interface that must be implemented by a trust
// anchor store. A trust anchor store is used to store and retrieve registered
// trust anchors that are used to authenticate attestation tokens.
type ITrustAnchorStore interface {

	// GetName returns the name of the ITrustAnchorStore implementation.
	// This identifies the store that should be initialized for a
	// deployment.
	GetName() string

	// Init initializes the ITrustAnchorStore instances using the provided
	// parameters, creating the necessary database connections, etc.
	Init(params TrustAnchorStoreParams) error

	// GetTrustAnchor retrieves a trust anchor based on the specified
	// tenant and TrustAnchorID.
	GetTrustAnchor(tenantId int, taId TrustAnchorID) ([]byte, error)

	// AddCertsFromPEM adds a certificate chain to the trust anchor store
	// for the specified tenant.
	AddCertsFromPEM(tenantId int, value []byte) error

	// AddPublicKeyFromPEM adds a public key to the store for the specified
	// tenant under the specified key ID.
	AddPublicKeyFromPEM(tenantId int, id interface{}, value []byte) error

	// Close cleanly terminates the ITrustAnchorStore instance.
	Close() error
}
