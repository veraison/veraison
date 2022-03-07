// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"github.com/veraison/veraison/provisioning/decoder"
	plugin_common "github.com/veraison/veraison/provisioning/plugins/common"
)

const (
	SupportedMediaType = "application/corim-unsigned+cbor; profile=http://arm.com/psa/iot/1"
	PluginName         = "unsigned-corim (PSA_IOT profile)"
)

type Decoder struct{}

func (o Decoder) Init(params decoder.Params) error {
	return nil // no-op
}

func (o Decoder) Close() error {
	return nil // no-op
}

func (o Decoder) GetName() string {
	return PluginName
}

func (o Decoder) GetSupportedMediaTypes() []string {
	return []string{
		SupportedMediaType,
	}
}

func (o Decoder) Decode(data []byte) (*decoder.EndorsementDecoderResponse, error) {
	return plugin_common.UnsignedCorimDecoder(data, Extractor{})
}
