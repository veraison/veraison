// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package decoder

type IDecoderManager interface {
	Init(dir string) error
	Close() error
	Dispatch(mediaType string, data []byte) (*EndorsementDecoderResponse, error)
	IsSupportedMediaType(mediaType string) bool
	SupportedMediaTypes() string
}
