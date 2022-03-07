// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package common

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestLabel(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	intLabel := NewIntLabel(1)
	assert.True(intLabel.IsInt())
	assert.Equal("1", intLabel.String())

	bytes, err := json.Marshal(intLabel)
	require.Nil(err)
	assert.Equal("\"1\"", string(bytes))

	stringLabel := NewStringLabel("test")
	assert.False(stringLabel.IsInt())
	assert.Equal("1", intLabel.String())

	bytes, err = json.Marshal(stringLabel)
	require.Nil(err)
	assert.Equal("\"test\"", string(bytes))

}

func TestResult_RoundTrip(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	result := AttestationResult{
		Status: Status_SUCCESS,
		TrustVector: &TrustVector{
			HardwareAuthenticity: Status_SUCCESS,
			SoftwareIntegrity:    Status_SUCCESS,
			SoftwareUpToDateness: Status_SUCCESS,
			ConfigIntegrity:      Status_SUCCESS,
			RuntimeIntegrity:     Status_SUCCESS,
			CertificationStatus:  Status_UNKNOWN,
		},
		RawEvidence:       []byte{0xDE, 0xAD, 0xBE, 0xEF},
		Timestamp:         &timestamppb.Timestamp{},
		EndorsedClaims:    &EndorsedClaims{},
		AppraisalPolicyID: "test-policy",
	}

	bytes, err := protojson.Marshal(&result)
	require.Nil(err)
	assert.JSONEq(`{"status":"SUCCESS","trust-vector":{"hw-authenticity":"SUCCESS","sw-integrity":"SUCCESS","sw-up-to-dateness":"SUCCESS","config-integrity":"SUCCESS","runtime-integrity":"SUCCESS","certification-status":"UNKNOWN"},"raw-evidence":"3q2+7w==","timestamp":"1970-01-01T00:00:00Z","endorsed-claims":{},"appraisal-policy-id":"test-policy"}`, string(bytes))

	var extractedResult AttestationResult
	err = protojson.Unmarshal(bytes, &extractedResult)
	require.Nil(err)
	assert.Equal(&result, &extractedResult)
}
