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
	var instanceAttrs InstanceAttributes

	if err := instanceAttrs.FromEnvironment(rv.Environment); err != nil {
		return nil, fmt.Errorf("could not extract instance attributes: %w", err)
	}

	if len(rv.Measurements) != 1 {
		return nil, fmt.Errorf("expecting one measurement only")
	}

	var (
		swComponents []*common.SwComponent
		swCompAttrs  SwCompAttributes
		measurement  comid.Measurement = rv.Measurements[0]
	)

	if err := swCompAttrs.FromMeasurement(measurement); err != nil {
		return nil, fmt.Errorf("extracting measurement: %w", err)
	}

	swID, err := makeSwID(instanceAttrs)
	if err != nil {
		return nil, fmt.Errorf("failed to create software component id: %w", err)
	}

	swAttrs, err := makeSwAttrs(swCompAttrs)
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

	if len(swComponents) == 0 {
		return nil, fmt.Errorf("no software components found")
	}

	return swComponents, nil
}

func makeSwID(i InstanceAttributes) (*structpb.Struct, error) {
	return structpb.NewStruct(
		map[string]interface{}{
			"enacttrust-tpm.node-id": i.NodeID,
		},
	)
}

func makeSwAttrs(s SwCompAttributes) (*structpb.Struct, error) {
	return structpb.NewStruct(
		map[string]interface{}{
			"enacttrust-tpm.digest": s.Digest,
			"enacttrust-tpm.alg-id": s.AlgID,
		},
	)
}

func (o Extractor) TaExtractor(avk comid.AttestVerifKey) (*common.TrustAnchor, error) {
	var instanceAttrs InstanceAttributes

	if err := instanceAttrs.FromEnvironment(avk.Environment); err != nil {
		return nil, fmt.Errorf("could not extract node id: %w", err)
	}

	taID, err := makeTaID(instanceAttrs)
	if err != nil {
		return nil, fmt.Errorf("failed to create trust anchor id: %w", err)
	}

	// extract AK pub
	if len(avk.VerifKeys) != 1 {
		return nil, errors.New("expecting exactly one AK public key")
	}

	akPub := avk.VerifKeys[0].Key

	taKey, err := makeTaRawPublicKey(akPub)
	if err != nil {
		return nil, fmt.Errorf("failed to create trust anchor raw public key: %w", err)
	}

	ta := &common.TrustAnchor{
		Id: &common.TrustAnchID{
			Type:  common.AttestationFormat_PSA_IOT,
			Parts: taID,
		},
		Value: &common.TrustAnchorValue{
			Type:  common.TaType_TA_RAWPUBLICKEY,
			Value: taKey,
		},
	}

	return ta, nil
}

func makeTaRawPublicKey(key string) (*structpb.Struct, error) {
	iakPub := map[string]interface{}{
		"enacttrust.ak-pub": key,
	}

	return structpb.NewStruct(iakPub)
}

func makeTaID(i InstanceAttributes) (*structpb.Struct, error) {
	return structpb.NewStruct(
		map[string]interface{}{
			"enacttrust-tpm.node-id": i.NodeID,
		},
	)
}
