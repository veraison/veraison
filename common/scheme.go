// Copyright 2021-2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package common

// ExtractedEvidence contains a map of claims extracted from an attestation
// token along with the corresponding SoftwareID that is used to fetch
// endorsements.
type ExtractedEvidence struct {
	Evidence   map[string]interface{} `json:"evidence"`
	SoftwareID string                 `json:"software_id"`
}

// IScheme defines the interface to attestation scheme specific functionality.
// An object implementing this interface encapsulates all functionality
// specific to a particular AttestationFormat, such as knowledge of token
// structure.
type IScheme interface {
	GetName() string
	GetFormat() AttestationFormat
	GetTrustAnchorID(token *AttestationToken) (string, error)
	ExtractEvidence(token *AttestationToken, trustAnchor string) (*ExtractedEvidence, error)
	GetAttestation(ec *EvidenceContext, endorsements string) (*Attestation, error)
}
