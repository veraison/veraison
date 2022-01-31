// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/veraison/corim/comid"
)

func TestDecoder_GetName(t *testing.T) {
	d := &Decoder{}

	expected := PluginName

	actual := d.GetName()

	assert.Equal(t, expected, actual)
}

func TestDecoder_GetSupportedMediaTypes(t *testing.T) {
	d := &Decoder{}

	expected := []string{
		SupportedMediaType,
	}

	actual := d.GetSupportedMediaTypes()

	assert.Equal(t, expected, actual)
}

func TestDecoder_Init(t *testing.T) {
	d := &Decoder{}

	assert.Nil(t, d.Init(nil))
}

func TestDecoder_Close(t *testing.T) {
	d := &Decoder{}

	assert.Nil(t, d.Close())
}

func TestDecoder_Decode_empty_data(t *testing.T) {
	d := &Decoder{}

	emptyData := []byte{}

	expectedErr := `empty data`

	_, err := d.Decode(emptyData)

	assert.EqualError(t, err, expectedErr)
}

func TestDecoder_Decode_invalid_data(t *testing.T) {
	d := &Decoder{}

	invalidCbor := []byte("invalid CBOR")

	expectedErr := `CBOR decoding failed: cbor: cannot unmarshal UTF-8 text string into Go value of type corim.UnsignedCorim`

	_, err := d.Decode(invalidCbor)

	assert.EqualError(t, err, expectedErr)
}

func TestDecoder_Decode_OK(t *testing.T) {
	tvs := []string{
		unsignedCorimComidPsaIakPubOne,
		unsignedCorimComidPsaIakPubTwo,
		unsignedCorimComidPsaRefValOne,
		unsignedCorimComidPsaRefValThree,
		unsignedCorimComidPsaRefValOnlyMandIDAttr,
	}

	d := &Decoder{}

	for _, tv := range tvs {
		data := comid.MustHexDecode(t, tv)
		_, err := d.Decode(data)
		assert.NoError(t, err)
	}
}

func TestDecoder_Decode_negative_tests(t *testing.T) {
	tvs := []struct {
		desc        string
		input       string
		expectedErr string
	}{
		{
			desc:        "multiple verification keys for an instance",
			input:       unsignedCorimComidPsaMultIak,
			expectedErr: `bad key in CoMID at index 0: expecting exactly one IAK public key`,
		},
		{
			desc:        "multiple digests in the same measurement",
			input:       unsignedCorimComidPsaRefValMultDigest,
			expectedErr: `bad software component in CoMID at index 0: extracting measurement at index 0: expecting exactly one digest`,
		},
		{
			desc:        "missing measurement identifier",
			input:       unsignedCorimComidPsaRefValNoMkey,
			expectedErr: `bad software component in CoMID at index 0: extracting measurement at index 0: measurement key is not present`,
		},
		{
			desc:        "no implementation id specified in the measurement",
			input:       unsignedCorimComidPsaRefValNoImplID,
			expectedErr: `bad software component in CoMID at index 0: could not extract PSA class attributes: could not extract implementation-id from class-id: class-id type is: comid.TaggedUUID`,
		},
		{
			desc:        "no instance id specified in the verification key triple",
			input:       unsignedCorimComidPsaIakPubNoUeID,
			expectedErr: `bad key in CoMID at index 0: could not extract PSA instance-id: expecting instance in environment`,
		},
		{
			desc:        "no implementation id specified in the verification key triple",
			input:       unsignedCorimComidPsaIakPubNoImplID,
			expectedErr: `bad key in CoMID at index 0: could not extract PSA class attributes: could not extract implementation-id from class-id: class-id type is: comid.TaggedUUID`,
		}}

	for _, tv := range tvs {
		data := comid.MustHexDecode(t, tv.input)
		d := &Decoder{}
		_, err := d.Decode(data)
		assert.EqualError(t, err, tv.expectedErr)
	}
}
