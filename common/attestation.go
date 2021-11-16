// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package common

func (a *Attestation) GetEndorsements() map[string]interface{} {
	return map[string]interface{}{
		"HardwareDetails":      a.Result.EndorsedClaims.HardwareDetails.AsMap(),
		"SoftwareDetails":      a.Result.EndorsedClaims.SoftwareDetails.AsMap(),
		"CertificationDetails": a.Result.EndorsedClaims.CertificationDetails.AsMap(),
		"ConfigDetails":        a.Result.EndorsedClaims.ConfigDetails.AsMap(),
	}
}
