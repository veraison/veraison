// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package verifier

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/veraison/common"

	"go.uber.org/zap"
)

type StubVtsClient struct {
	dbPath      string //nolint
	evidenceDir string
}

func (c *StubVtsClient) Init(params *common.ParamStore) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	c.evidenceDir = filepath.Join(wd, "test", "samples")

	return nil
}

func (c StubVtsClient) Close() error {
	return nil
}

func (c StubVtsClient) GetAttestation(
	token *common.AttestationToken,
) (*common.Attestation, error) {
	certDetails, err := structpb.NewStruct(map[string]interface{}{
		"certification_authority": "Acme",
	})
	if err != nil {
		return nil, fmt.Errorf(
			"could not cert details for token %q: %s",
			token.Data,
			err.Error(),
		)
	}

	evidence, err := structpb.NewStruct(map[string]interface{}{})
	if err != nil {
		return nil, err
	}

	attestation := common.Attestation{
		Result: &common.AttestationResult{
			TrustVector: &common.TrustVector{
				HardwareAuthenticity: common.AR_Status_SUCCESS,
				SoftwareIntegrity:    common.AR_Status_SUCCESS,
				SoftwareUpToDateness: common.AR_Status_UNKNOWN,
				RuntimeIntegrity:     common.AR_Status_UNKNOWN,
				ConfigIntegrity:      common.AR_Status_UNKNOWN,
				CertificationStatus:  common.AR_Status_FAILURE,
			},
			Status:      common.AR_Status_FAILURE,
			RawEvidence: token.Data,
			Timestamp:   timestamppb.Now(),
			EndorsedClaims: &common.EndorsedClaims{
				CertificationDetails: certDetails,
			},
		},
		Evidence: &common.EvidenceContext{
			TenantId: 1,
			Format:   token.Format,
			Evidence: evidence,
		},
	}

	return &attestation, nil
}

type StubManager struct {
}

func (m *StubManager) Init(params *common.ParamStore) error {
	return nil
}

func (m *StubManager) ListPolicies(tenantID int) ([]common.PolicyListEntry, error) {
	return nil, nil
}

func (m *StubManager) GetPolicy(
	tenantID int,
	tokenFormat common.AttestationFormat,
) (*common.Policy, error) {
	return nil, nil
}

func (m *StubManager) PutPolicy(tenantID int, policy *common.Policy) error {
	return nil
}

func (m *StubManager) PutPolicyBytes(tenantID int, policyBytes []byte) error {
	return nil
}

func (m *StubManager) DeletePolicy(
	tenantID int,
	tokenFormat common.AttestationFormat,
) error {
	return nil
}

func (m *StubManager) Close() error {
	return nil
}

type StubEngine struct {
}

func (m StubEngine) GetName() string {
	return "stub-policy-engine"
}

func (m *StubEngine) Init(params *common.ParamStore) error {
	return nil
}

func (m *StubEngine) Appraise(
	attestation *common.Attestation,
	policy *common.Policy,
) error {

	attestation.Result.TrustVector.CertificationStatus = common.AR_Status_FAILURE
	attestation.Result.Status = common.AR_Status_FAILURE

	return nil
}

func (m StubEngine) Close() error {
	return nil
}

func TestVerifier(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("%v", err)
	}

	var vc *common.ParamStore
	vc, err = NewVerifierParams()
	require.Nil(err)
	require.Nil(vc.SetString("pluginLocations", filepath.Join(wd, "..", "plugins", "bin")))
	require.Nil(vc.SetStringMapString("VtsParams", map[string]string{"schemaPath": "/tmp/schema.sql"}))

	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("%v", err)
	}

	v, err := NewVerifier(logger)
	if err != nil {
		t.Fatalf("%v", err)
	}

	err = v.Init(vc, new(StubVtsClient), new(StubManager), new(StubEngine))
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer v.Close()

	token := common.AttestationToken{
		Format: common.AttestationFormat_PSA_IOT,
		Data:   []byte("valid-iat"),
	}

	result, err := v.Verify(&token)
	require.Nil(err)
	assert.Equal(result.Status, common.AR_Status_FAILURE)
	assert.Equal(result.TrustVector.CertificationStatus, common.AR_Status_FAILURE)
}
