// Copyright 2021 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package common

// AttestationResult encapsulates the result of validating an attestation
// token.
type AttestationResult struct {

	// IsValid is set to True iff the token is valid. A token is considered valid when all of the following are true:
	// - Token signature, if there is one, has been verified against a know trust anchor.
	// - The claims structure matches what is expected based on the TokenFormat.
	// - The claims validate against the policy associated with the TokenFormat.
	IsValid bool `json:"is_valid" binding:"required"`

	// Claims contains the claims extracted from the evidence, and/or
	// derived from the evidence and endorsements. This may or may not be
	// populated depending on whether "simple" validation was used.
	Claims map[string]interface{} `json:"claims" binding:"required"`
}
