// Copyright 2021 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"veraison/common"
)

func TestGetTrustAnchorID(t *testing.T) {
	assert := assert.New(t)

	tokenBytes, err := ioutil.ReadFile("test/psa-token.cbor")
	if err != nil {
		t.Fatalf("Could not lead device certs file.")
	}

	expectedTaId := map[string]interface{}{
		"key-id": &[]byte{
			0x01, 0x07, 0x06, 0x05, 0x04, 0x03, 0x02, 0x01, 0x00, 0x0f, 0x0e, 0x0d,
			0x0c, 0x0b, 0x0a, 0x09, 0x08, 0x17, 0x16, 0x15, 0x14, 0x13, 0x12, 0x11,
			0x10, 0x1f, 0x1e, 0x1d, 0x1c, 0x1b, 0x1a, 0x19, 0x18,
		},
	}

	ee := new(EvidenceExtractor)

	taId, err := ee.GetTrustAnchorID(tokenBytes)
	assert.Nil(err)
	assert.Equal(common.TaTypeKey, taId.Type)
	assert.Equal(expectedTaId, taId.Value)
}

func TestExtractEvidence(t *testing.T) {
	assert := assert.New(t)

	tokenBytes, err := ioutil.ReadFile("test/psa-token.cbor")
	if err != nil {
		t.Fatalf("Could not lead device certs file.")
	}

	keyText := []byte(`
-----BEGIN PUBLIC KEY-----
MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAETl4iCZ47zrRbRG0TVf0dw7VFlHtv
18HInYhnmMNybo+A1wuECyVqrDSmLt4QQzZPBECV8ANHS5HgGCCSr7E/Lg==
-----END PUBLIC KEY-----`)

	ee := new(EvidenceExtractor)

	claims, err := ee.ExtractEvidence(tokenBytes, keyText)

	assert.Nil(err)
	assert.Equal("PSA_IOT_PROFILE_1", *claims["Profile"].(*string))

	swComponents := claims["SwComponents"].([]map[string]interface{})
	assert.Len(swComponents, 4)
	assert.Equal("BL", swComponents[0]["MeasurementType"].(string))

}
