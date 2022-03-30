// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package common

import timestamppb "google.golang.org/protobuf/types/known/timestamppb"

func NewAttestation(ec *EvidenceContext) *Attestation {
	return &Attestation{
		Evidence: ec,
		Result: &AttestationResult{
			Status: AR_Status_UNKNOWN,
			TrustVector: &TrustVector{
				SoftwareIntegrity:    AR_Status_UNKNOWN,
				HardwareAuthenticity: AR_Status_UNKNOWN,
				SoftwareUpToDateness: AR_Status_UNKNOWN,
				ConfigIntegrity:      AR_Status_UNKNOWN,
				RuntimeIntegrity:     AR_Status_UNKNOWN,
				CertificationStatus:  AR_Status_UNKNOWN,
			},
			Timestamp:         timestamppb.Now(),
			ProcessedEvidence: ec.Evidence,
		},
	}
}

func (a *Attestation) GetEndorsements() map[string]interface{} {
	return map[string]interface{}{
		"HardwareDetails":      a.Result.EndorsedClaims.HardwareDetails.AsMap(),
		"SoftwareDetails":      a.Result.EndorsedClaims.SoftwareDetails.AsMap(),
		"CertificationDetails": a.Result.EndorsedClaims.CertificationDetails.AsMap(),
		"ConfigDetails":        a.Result.EndorsedClaims.ConfigDetails.AsMap(),
	}
}
