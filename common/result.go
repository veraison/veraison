// Copyright 2021 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"encoding/json"
	"time"
)

// AttestationResult encapsulates the result of validating an attestation
// token.
type AttestationResult struct {
	AttestationResultExtension

	Status            Status         `cbor:"0,keyasint" json:"status" binding:"required"`
	TrustVector       TrustVector    `cbor:"1,keyasint,omitempty" json:"trust-vector"`
	RawEvidence       []byte         `cbor:"2,keyasint,omitempty" json:"raw-evidence"`
	Timestamp         time.Time      `cbor:"3,keyasint" json:"timestamp" binding:"required"`
	EndorsedClaims    EndorsedClaims `cbor:"4,keyasint,omitempty" json:"endorsed-claims"`
	AppraisalPolicyID string         `cbor:"5,keyasint" json:"appraisal-policy-id"`
}

func NewAttestationResult() *AttestationResult {
	ar := new(AttestationResult)
	ar.ProcessedEvidence = make(ClaimsMap)

	return ar
}

// ToJSON returns a text representation of a JSON structure representation of the AttestationResult.
func (r AttestationResult) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}

// FromJSON populates the AttestationResult from the provided data which is
// interpreted as a JSON serialization of an AttestationResult.
func (r *AttestationResult) FromJSON(data []byte) error {
	return json.Unmarshal(data, r)
}

// ToCBOR returns a byte slice containing the CBOR encoded AttestationResult
func (r AttestationResult) ToCBOR() ([]byte, error) {
	return em.Marshal(r)
}

// FromCBOR populates the AttestationResult from the provided data, interpreted
// as a CBOR encoded AttestationResult.
func (r *AttestationResult) FromCBOR(data []byte) error {
	return dm.Unmarshal(data, r)
}

type AttestationResultExtension struct {
	ProcessedEvidence ClaimsMap `cbor:"100,keyasint" json:"veraison-processed-evidence"`
}

// Status enumerates possible states of the elements of the
// attestation result trust vector, as well as the attestation result as a
// whole.
type Status int8

const (
	StatusFailure = Status(iota)
	StatusSuccess
	StatusUnknown

	StatusInvalid // Must be last
)

func (s Status) String() string {
	if s < 0 {
		return "unknown"
	} else if s >= StatusInvalid {
		return "INVALID"
	} else {
		return []string{
			"failure",
			"success",
		}[s]
	}
}

// TrustVector is a collection of statuses for different
// aspects of the attester extracted from provided evidence.
type TrustVector struct {
	TrustVectorExtension

	HardwareAuthenticity Status `cbor:"0,keyasint" json:"hw-authenticity" binding:"required"`
	SoftwareIntegrity    Status `cbor:"1,keyasint" json:"sw-integrity" binding:"required"`
	SoftwareUpToDateness Status `cbor:"2,keyasint" json:"sw-up-to-dateness" binding:"required"`
	ConfigIntegrity      Status `cbor:"3,keyasint" json:"config-integrity" binding:"required"`
	RuntimeIntegrity     Status `cbor:"4,keyasint" json:"runtime-integrity" binding:"required"`
	CertificationStatus  Status `cbor:"5,keyasint" json:"certification-status" binding:"required"`
}

// Space for $$ar-trust-vector-extension
type TrustVectorExtension struct {
}

// ClaimsMap contains claims extracted or derived from evidence
type ClaimsMap map[Label]interface{}

func (c ClaimsMap) MarshalJSON() ([]byte, error) {
	outMap := make(map[string]interface{})
	for key, val := range c {
		outMap[key.String()] = val
	}

	return json.Marshal(outMap)
}

func (c *ClaimsMap) UnmarshalJSON(data []byte) error {
	inMap := make(map[string]interface{})

	if err := json.Unmarshal(data, &inMap); err != nil {
		return err
	}

	for key, value := range inMap {
		var label Label

		label.FromString(key)
		(*c)[label] = value
	}

	return nil
}

// EndorsedClaims contains claims about the attester's hardware, software, configuration, and certification that have either been extracted form the evidence, or derived from evidence and/or endorsements.
type EndorsedClaims struct {
	HardwareDetails      ClaimsMap `cbor:"0,keyasint,omitempty" json:"hw-details"`
	SoftwareDetails      ClaimsMap `cbor:"1,keyasint,omitempty" json:"sw-details"`
	CertificationDetails ClaimsMap `cbor:"2,keyasint,omitempty" json:"certification-details"`
	ConfigDetails        ClaimsMap `cbor:"3,keyasint,omitempty" json:"config-details"`
}
