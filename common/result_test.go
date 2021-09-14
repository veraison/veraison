package common

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		Status: StatusSuccess,
		TrustVector: TrustVector{
			HardwareAuthenticity: StatusSuccess,
			SoftwareIntegrity:    StatusSuccess,
			SoftwareUpToDateness: StatusSuccess,
			ConfigIntegrity:      StatusSuccess,
			RuntimeIntegrity:     StatusSuccess,
			CertificationStatus:  StatusUnknown,
		},
		RawEvidence:       []byte{0xDE, 0xAD, 0xBE, 0xEF},
		Timestamp:         time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),
		EndorsedClaims:    EndorsedClaims{},
		AppraisalPolicyID: "test-policy",
	}

	bytes, err := result.ToJSON()
	require.Nil(err)
	assert.JSONEq(`{"veraison-processed-evidence":{},"status":1,"trust-vector":{"hw-authenticity":1,"sw-integrity":1,"sw-up-to-dateness":1,"config-integrity":1,"runtime-integrity":1,"certification-status":2},"raw-evidence":"3q2+7w==","timestamp":"1970-01-01T00:00:00Z","endorsed-claims":{"hw-details":{},"sw-details":{},"certification-details":{},"config-details":{}},"appraisal-policy-id":"test-policy"}`, string(bytes))

	var extractedResult AttestationResult
	err = extractedResult.FromJSON(bytes)
	require.Nil(err)
	assert.Equal(result, extractedResult)

	bytes, err = result.ToCBOR()
	require.Nil(err)
	/*
		a6                           # map(6)
		   00                        #   unsigned(0)
		   01                        #   unsigned(1)
		   01                        #   unsigned(1)
		   a6                        #   map(6)
		      00                     #     unsigned(0)
		      01                     #     unsigned(1)
		      01                     #     unsigned(1)
		      01                     #     unsigned(1)
		      02                     #     unsigned(2)
		      01                     #     unsigned(1)
		      03                     #     unsigned(3)
		      01                     #     unsigned(1)
		      04                     #     unsigned(4)
		      01                     #     unsigned(1)
		      05                     #     unsigned(5)
		      02                     #     unsigned(2)
		   02                        #   unsigned(2)
		   44                        #   bytes(4)
		      deadbeef               #     "\xde\xad\xbe\xef"
		   03                        #   unsigned(3)
		   c1                        #   epoch datetime value, tag(1)
		      00                     #     unsigned(0)
					     #     datetime(1970-01-01T00:00:00Z)
		   05                        #   unsigned(5)
		   6b                        #   text(11)
		      746573742d706f6c696379 #     "test-policy"
		   18 64                     #   unsigned(100)
		   f6                        #   null, simple(22)
	*/
	assert.Equal([]byte{0xa6, 0x0, 0x1, 0x1, 0xa6, 0x0, 0x1, 0x1, 0x1, 0x2, 0x1, 0x3, 0x1, 0x4, 0x1, 0x5, 0x2, 0x2, 0x44, 0xde, 0xad, 0xbe, 0xef, 0x3, 0xc1, 0x0, 0x5, 0x6b, 0x74, 0x65, 0x73, 0x74, 0x2d, 0x70, 0x6f, 0x6c, 0x69, 0x63, 0x79, 0x18, 0x64, 0xf6}, bytes)
	var extractedResult2 AttestationResult
	err = extractedResult2.FromCBOR(bytes)
	require.Nil(err)
	assert.Equal(result, extractedResult)
}
