// Copyright 2021-2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/veraison/common"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetTrustAnchorID(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	tokenBytes, err := ioutil.ReadFile("test/psa-token.cbor")
	require.Nil(err)

	expectedTaID := "psa://1/BwYFBAMCAQAPDg0MCwoJCBcWFRQTEhEQHx4dHBsaGRg=/AQcGBQQDAgEADw4NDAsKCQgXFhUUExIREB8eHRwbGhkY"

	scheme := new(Scheme)

	token := common.AttestationToken{
		TenantId: 1,
		Format:   common.AttestationFormat_PSA_IOT,
		Data:     tokenBytes,
	}

	taID, err := scheme.GetTrustAnchorID(&token)
	require.Nil(err)
	assert.Equal(expectedTaID, taID)
}

func TestExtractEvidence(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	tokenBytes, err := ioutil.ReadFile("test/psa-token.cbor")
	require.Nil(err)

	trustAnchor := `
-----BEGIN PUBLIC KEY-----
MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAETl4iCZ47zrRbRG0TVf0dw7VFlHtv
18HInYhnmMNybo+A1wuECyVqrDSmLt4QQzZPBECV8ANHS5HgGCCSr7E/Lg==
-----END PUBLIC KEY-----`

	scheme := new(Scheme)

	token := common.AttestationToken{
		TenantId: 1,
		Format:   common.AttestationFormat_PSA_IOT,
		Data:     tokenBytes,
	}

	extracted, err := scheme.ExtractEvidence(&token, trustAnchor)

	require.Nil(err)
	assert.Equal("PSA_IOT_PROFILE_1", extracted.Evidence["profile"].(string))

	swComponents := extracted.Evidence["software-components"].([]interface{})
	assert.Len(swComponents, 4)
	assert.Equal("BL", swComponents[0].(map[string]interface{})["measurement-type"].(string))

}

func TestGetAttestation(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	extractedBytes, err := ioutil.ReadFile("test/extracted.json")
	require.Nil(err)

	var ec common.EvidenceContext
	err = json.Unmarshal(extractedBytes, &ec)
	require.Nil(err)

	endorsementsBytes, err := ioutil.ReadFile("test/endorsements.json")
	require.Nil(err)

	scheme := new(Scheme)

	attestation, err := scheme.GetAttestation(&ec, string(endorsementsBytes))
	require.Nil(err)

	assert.Equal(common.Status_SUCCESS, attestation.Result.Status)
}
