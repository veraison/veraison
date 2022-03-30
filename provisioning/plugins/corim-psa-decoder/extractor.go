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

func (o Extractor) SwCompExtractor(rv comid.ReferenceValue) ([]*common.SwComponent, error) {
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
	var swComponents []*common.SwComponent

	for i, m := range rv.Measurements {
		var psaSwCompAttrs PSASwCompAttributes

		if err := psaSwCompAttrs.FromMeasurement(m); err != nil {
			return nil, fmt.Errorf("extracting measurement at index %d: %w", i, err)
		}

		swID, err := makeSwID(psaClassAttrs, psaSwCompAttrs)
		if err != nil {
			return nil, fmt.Errorf("failed to create software component id: %w", err)
		}

		swAttrs, err := makeSwAttrs(psaSwCompAttrs)
		if err != nil {
			return nil, fmt.Errorf("failed to create software component attributes: %w", err)
		}

		swComponent := common.SwComponent{
			Id: &common.SwComponentID{
				Type:  common.AttestationFormat_PSA_IOT,
				Parts: swID,
			},
			Attributes: swAttrs,
		}

		swComponents = append(swComponents, &swComponent)
	}

	if len(swComponents) == 0 {
		return nil, fmt.Errorf("no software components found")
	}

	return swComponents, nil
}

func makeSwID(c PSAClassAttributes, s PSASwCompAttributes) (*structpb.Struct, error) {
	swID := map[string]interface{}{
		"psa.impl-id":   c.ImplID,
		"psa.signer-id": s.SignerID,
	}

	if c.Vendor != "" {
		swID["psa.hw-vendor"] = c.Vendor
	}

	if c.Model != "" {
		swID["psa.hw-model"] = c.Model
	}

	if s.MeasurementType != "" {
		swID["psa.measurement-type"] = s.MeasurementType
	}

	if s.Version != "" {
		swID["psa.version"] = s.Version
	}

	return structpb.NewStruct(swID)
}

func makeSwAttrs(s PSASwCompAttributes) (*structpb.Struct, error) {
	return structpb.NewStruct(
		map[string]interface{}{
			"psa.measurement-value": s.MeasurementValue,
			"psa.measurement-desc":  s.AlgID,
		},
	)
}

func (o Extractor) TaExtractor(avk comid.AttestVerifKey) (*common.TrustAnchor, error) {
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

	taID, err := makeTaID(psaInstanceAttrs, psaClassAttrs)
	if err != nil {
		return nil, fmt.Errorf("failed to create trust anchor id: %w", err)
	}

	// extract IAK pub
	if len(avk.VerifKeys) != 1 {
		return nil, errors.New("expecting exactly one IAK public key")
	}

	iakPub := avk.VerifKeys[0].Key

	taKey, err := makeTaRawPublicKey(iakPub)
	if err != nil {
		return nil, fmt.Errorf("failed to create trust anchor raw public key: %w", err)
	}

	ta := &common.TrustAnchor{
		Id: &common.TrustAnchorID{
			Type:  common.AttestationFormat_PSA_IOT,
			Parts: taID,
		},
		Value: &common.TrustAnchorValue{
			Type:  common.TAType_TA_RAWPUBLICKEY,
			Value: taKey,
		},
	}

	return ta, nil
}

func makeTaRawPublicKey(key string) (*structpb.Struct, error) {
	iakPub := map[string]interface{}{
		"psa.iak-pub": key,
	}

	return structpb.NewStruct(iakPub)
}

func makeTaID(i PSAInstanceAttributes, c PSAClassAttributes) (*structpb.Struct, error) {
	taID := map[string]interface{}{
		"psa.impl-id": c.ImplID,
		"psa.inst-id": []byte(i.InstID),
	}

	if c.Vendor != "" {
		taID["psa.hw-vendor"] = c.Vendor
	}

	if c.Model != "" {
		taID["psa.hw-model"] = c.Model
	}

	return structpb.NewStruct(taID)
}
