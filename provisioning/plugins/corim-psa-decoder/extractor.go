// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"errors"
	"fmt"

	"github.com/veraison/common"
	"github.com/veraison/corim/comid"
	structpb "google.golang.org/protobuf/types/known/structpb"
)

type Extractor struct{}

func (o Extractor) SwCompExtractor(rv comid.ReferenceValue) ([]*common.Endorsement, error) {
	var psaClassAttrs PSAClassAttributes

	if err := psaClassAttrs.FromEnvironment(rv.Environment); err != nil {
		return nil, fmt.Errorf("could not extract PSA class attributes: %w", err)
	}

	// Each measurement is encoded in a measurement-map of a CoMID
	// reference-triple-record.  Since a measurement-map can encode one or more
	// measurements, a single reference-triple-record can carry as many
	// measurements as needed, provided they belong to the same PSA RoT
	// identified in the subject of the "reference value" triple.  A single
	// reference-triple-record SHALL completely describe the updatable PSA RoT.
	var swComponents []*common.Endorsement

	for i, m := range rv.Measurements {
		var psaSwCompAttrs PSASwCompAttributes

		if err := psaSwCompAttrs.FromMeasurement(m); err != nil {
			return nil, fmt.Errorf("extracting measurement at index %d: %w", i, err)
		}

		swAttrs, err := makeSwAttrs(psaClassAttrs, psaSwCompAttrs)
		if err != nil {
			return nil, fmt.Errorf("failed to create software component attributes: %w", err)
		}

		swComponent := common.Endorsement{
			Scheme:     common.AttestationFormat_PSA_IOT,
			Type:       common.EndorsementType_REFERENCE_VALUE,
			Attributes: swAttrs,
		}

		swComponents = append(swComponents, &swComponent)
	}

	if len(swComponents) == 0 {
		return nil, fmt.Errorf("no software components found")
	}

	return swComponents, nil
}

func makeSwAttrs(c PSAClassAttributes, s PSASwCompAttributes) (*structpb.Struct, error) {
	swAttrs := map[string]interface{}{
		"psa.impl-id":           c.ImplID,
		"psa.signer-id":         s.SignerID,
		"psa.measurement-value": s.MeasurementValue,
		"psa.measurement-desc":  s.AlgID,
	}

	if c.Vendor != "" {
		swAttrs["psa.hw-vendor"] = c.Vendor
	}

	if c.Model != "" {
		swAttrs["psa.hw-model"] = c.Model
	}

	if s.MeasurementType != "" {
		swAttrs["psa.measurement-type"] = s.MeasurementType
	}

	if s.Version != "" {
		swAttrs["psa.version"] = s.Version
	}

	return structpb.NewStruct(swAttrs)
}

func (o Extractor) TaExtractor(avk comid.AttestVerifKey) (*common.Endorsement, error) {
	// extract instance ID
	var psaInstanceAttrs PSAInstanceAttributes

	if err := psaInstanceAttrs.FromEnvironment(avk.Environment); err != nil {
		return nil, fmt.Errorf("could not extract PSA instance-id: %w", err)
	}

	// extract implementation ID
	var psaClassAttrs PSAClassAttributes

	if err := psaClassAttrs.FromEnvironment(avk.Environment); err != nil {
		return nil, fmt.Errorf("could not extract PSA class attributes: %w", err)
	}

	// extract IAK pub
	if len(avk.VerifKeys) != 1 {
		return nil, errors.New("expecting exactly one IAK public key")
	}

	iakPub := avk.VerifKeys[0].Key

	// TODO(tho) check that format of IAK pub is as expected

	taAttrs, err := makeTaAttrs(psaInstanceAttrs, psaClassAttrs, iakPub)
	if err != nil {
		return nil, fmt.Errorf("failed to create trust anchor attributes: %w", err)
	}

	ta := &common.Endorsement{
		Scheme:     common.AttestationFormat_PSA_IOT,
		Type:       common.EndorsementType_VERIFICATION_KEY,
		Attributes: taAttrs,
	}

	return ta, nil
}

func makeTaAttrs(i PSAInstanceAttributes, c PSAClassAttributes, key string) (*structpb.Struct, error) {
	taID := map[string]interface{}{
		"psa.impl-id": c.ImplID,
		"psa.inst-id": []byte(i.InstID),
		"psa.iak-pub": key,
	}

	if c.Vendor != "" {
		taID["psa.hw-vendor"] = c.Vendor
	}

	if c.Model != "" {
		taID["psa.hw-model"] = c.Model
	}

	return structpb.NewStruct(taID)
}
