// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/veraison/common"
	"google.golang.org/protobuf/types/known/structpb"
)

func Test_DecodeAttestationData(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	data, err := ioutil.ReadFile("test/tokens/basic.token")
	require.NoError(err)

	var decoded Token

	err = decoded.Decode(data)
	require.NoError(err)

	assert.Equal(uint32(4283712327), decoded.AttestationData.Magic)
	assert.Equal(uint64(0x7), decoded.AttestationData.FirmwareVersion)
}

func Test_GetTrustAnchorID(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	data, err := ioutil.ReadFile("test/tokens/basic.token")
	require.NoError(err)

	ta := common.AttestationToken{
		TenantId: int64(0),
		Format:   common.AttestationFormat_TPM_ENACTTRUST,
		Data:     data,
	}

	var s Scheme

	taID, err := s.GetTrustAnchorID(&ta)
	require.NoError(err)
	assert.Equal("tpm-enacttrust://0/7df7714e-aa04-4638-bcbf-434b1dd720f1", taID)
}

func readPublicKeyBytes(path string) ([]byte, error) {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(buf)
	if block == nil || block.Type != "PUBLIC KEY" {
		return nil, fmt.Errorf("could not decode EC public key from PEM block: %q", block)
	}
	return block.Bytes, nil
}

func Test_ExtracteEvidence(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	data, err := ioutil.ReadFile("test/tokens/basic.token")
	require.NoError(err)

	ta := common.AttestationToken{
		TenantId: int64(0),
		Format:   common.AttestationFormat_TPM_ENACTTRUST,
		Data:     data,
	}

	var s Scheme

	trustAnchorBytes, err := readPublicKeyBytes("test/keys/basic.pem.pub")
	require.NoError(err)
	trustAnchor := base64.StdEncoding.EncodeToString(trustAnchorBytes)

	ev, err := s.ExtractEvidence(&ta, trustAnchor)
	require.Nil(err)

	assert.Equal("tpm-enacttrust://0/7df7714e-aa04-4638-bcbf-434b1dd720f1", ev.SoftwareID)
	assert.Equal([]int64{1, 2, 3, 4}, ev.Evidence["pcr-selection"])
	assert.Equal(int64(11), ev.Evidence["hash-algorithm"])
	assert.Equal([]byte{0x87, 0x42, 0x8f, 0xc5, 0x22, 0x80, 0x3d, 0x31, 0x6, 0x5e, 0x7b, 0xce, 0x3c, 0xf0, 0x3f, 0xe4, 0x75, 0x9, 0x66, 0x31, 0xe5, 0xe0, 0x7b, 0xbd, 0x7a, 0xf, 0xde, 0x60, 0xc4, 0xcf, 0x25, 0xc7},
		ev.Evidence["pcr-digest"])
}

func Test_GetAttestation(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	evStruct, err := structpb.NewStruct(map[string]interface{}{
		"pcr-selection":  []interface{}{1, 2, 3, 4},
		"hash-algorithm": int64(11),
		"pcr-digest":     []byte{0x87, 0x42, 0x8f, 0xc5, 0x22, 0x80, 0x3d, 0x31, 0x6, 0x5e, 0x7b, 0xce, 0x3c, 0xf0, 0x3f, 0xe4, 0x75, 0x9, 0x66, 0x31, 0xe5, 0xe0, 0x7b, 0xbd, 0x7a, 0xf, 0xde, 0x60, 0xc4, 0xcf, 0x25, 0xc7},
	})
	require.NoError(err)

	evidenceContext := &common.EvidenceContext{
		Format:        common.AttestationFormat_TPM_ENACTTRUST,
		TenantId:      int64(0),
		TrustAnchorId: "tpm-enacttrust://0/7df7714e-aa04-4638-bcbf-434b1dd720f1",
		SoftwareId:    "tpm-enacttrust://0/7df7714e-aa04-4638-bcbf-434b1dd720f1",
		Evidence:      evStruct,
	}
	endorsements := []string{"h0KPxSKAPTEGXnvOPPA/5HUJZjHl4Hu9eg/eYMTPJcc="}

	var scheme Scheme

	attestation, err := scheme.GetAttestation(evidenceContext, endorsements)
	require.NoError(err)

	assert.Equal(common.AR_Status_SUCCESS, attestation.Result.Status)
	assert.Equal(common.AR_Status_SUCCESS, attestation.Result.TrustVector.SoftwareIntegrity)
}
