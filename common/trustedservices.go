// Copyright 2021 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package common

// ITrustedServicesConnector provides a means of establishing connection to the trusted services component.
type ITrustedServicesConnector interface {
	Connect(host string, port int, params map[string]string) (ITrustedServicesClient, error)
}

// ITrustedServicesClient specifies the client interface for the trusted services component.
type ITrustedServicesClient interface {
	Init(params *ParamStore) error

	// GetAttestation returns attestation information -- evidences,
	// endorsed claims, trust vector, etc -- for the provided attestation
	// token data.
	GetAttestation(token *AttestationToken) (*Attestation, error)

	Close() error
}
