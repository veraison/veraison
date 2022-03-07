// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package decoder

type Params map[string]interface{}

type IDecoder interface {
	Init(params Params) error
	Close() error
	GetName() string
	GetSupportedMediaTypes() []string
	Decode([]byte) (*EndorsementDecoderResponse, error)
}
