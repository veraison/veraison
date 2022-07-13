// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package verifier

type IVerifier interface {
	IsSupportedMediaType(mt string) bool
	SupportedMediaTypes() []string
	ProcessEvidence(data []byte, mt string) ([]byte, error)
}
