// Copyright 2021 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package common

// EvidenceContext is the wrapper for evidence extracted from a token.
type EvidenceContext struct {

	// TenantID identifies the tenant on multi-tenancy deployment for
	// which the evidence should be evaluated.
	TenantID int `json:"tenant_id"`

	// Format indicates the format of the token from which evidence was
	// extracted. This is used to specify how the Evidence structure should
	// be interpreted, and to identify which endorsements will be necessary
	// for verification.
	Format AttestationFormat `json:"attestation_format"`

	// Evidence contains the evidence claims extracted from the token.
	// Claims can be simple key-value pairs or more complicated nested
	// structures. This is specific to the AttestationFormat. The only constraint
	// is that the resulting structure must be serializable as JSON.
	Evidence map[string]interface{} `json:"evidence"`
}

// EvidenceExtractorParams is a map of key-value pairs of parameters used to
// initialize an IEvidenceExtractor implementation. Which parameters are
// supported is specific to each implementation.
type EvidenceExtractorParams map[string]string

// IEvidenceExtractor defined the interface that must be implemented by
// plugins used to extract evidence from attestation tokens.
type IEvidenceExtractor interface {

	// GetName returns the name of the IEvidenceExtractor implementation.
	GetName() string

	// Init initializes the token extractor, performing any one-time setup.
	// This must be invoked before attempting to GetTrustAnchorID or
	// ExtractEvidence.
	Init(params EvidenceExtractorParams) error

	// GetTrustAnchorID returns the TrustAnchorID associated with this
	// token. This is used to retrieve a trust anchor form a store that may
	// be used to decrypt and/or verify the signature on the token.
	GetTrustAnchorID(token []byte) (TrustAnchorID, error)

	// ExtractEvidence verifies the token structure and signature using the
	// provided trust anchor and extracts evidence claims. The claims are
	// collected in a map that may be serialized as a JSON structure. The
	// contents of the map are specific to the token format being
	// processed.
	ExtractEvidence(token []byte, ta []byte) (map[string]interface{}, error)

	// Close ensures the evidence extractor is cleanly terminated.
	Close() error
}
