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

func TestDecoder_Decode_OK(t *testing.T) {
	tvs := []string{
		unsignedCorimComidTpmEnactTrustAKOne,
		unsignedCorimComidTpmEnactTrustGoldenOne,
	}

	d := &Decoder{}

	for _, tv := range tvs {
		data := comid.MustHexDecode(t, tv)
		_, err := d.Decode(data)
		assert.NoError(t, err)
	}
}
