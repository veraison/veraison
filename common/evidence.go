// Copyright 2021-2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package common

// IEvidenceExtractor defined the interface that must be implemented by
// plugins used to extract evidence from attestation tokens.
type IEvidenceExtractor interface {

	// Init initializes the token extractor, performing any one-time setup.
	// This must be invoked before attempting to GetTrustAnchorID or
	// ExtractEvidence.
	Init(params *ParamStore) error

	// GetTrustAnchorID returns the TrustAnchorID associated with this
	// token. This is used to retrieve a trust anchor form a store that may
	// be used to decrypt and/or verify the signature on the token.
	GetTrustAnchorID(token []byte) (TrustAnchID, error)

	// ExtractEvidence verifies the token structure and signature using the
	// provided trust anchor and extracts evidence claims. The claims are
	// collected in a map that may be serialized as a JSON structure. The
	// contents of the map are specific to the token format being
	// processed.
	ExtractEvidence(token []byte, ta []byte) (map[string]interface{}, error)

	// Close ensures the evidence extractor is cleanly terminated.
	Close() error
}
