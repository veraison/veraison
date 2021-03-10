// Copyright 2021 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"

	"github.com/hashicorp/go-plugin"
	"github.com/veraison/psatoken"

	"github.com/veraison/common"
)

type EvidenceExtractor struct {
}

func (ee EvidenceExtractor) GetName() string {
	return "psa"
}

func (ee EvidenceExtractor) Init(params common.EvidenceExtractorParams) error {
	return nil
}

func (ee EvidenceExtractor) Close() error {
	return nil
}

func (ee EvidenceExtractor) GetTrustAnchorID(token []byte) (common.TrustAnchorID, error) {
	var psaToken psatoken.PSAToken

	err := psaToken.FromCOSE(token)
	if err != nil {
		return common.TrustAnchorID{}, err
	}

	taID := make(map[string]interface{})
	taID["key-id"] = psaToken.InstID

	return common.TrustAnchorID{Type: common.TaTypeKey, Value: taID}, nil
}

func (ee EvidenceExtractor) ExtractEvidence(token []byte, trustAnchor []byte) (map[string]interface{}, error) {
	block, rest := pem.Decode(trustAnchor)
	if block == nil {
		return nil, errors.New("could not extract trust anchor PEM block")
	}

	if len(rest) != 0 {
		return nil, errors.New("trailing data found after PEM block")
	}

	if block.Type != "PUBLIC KEY" {
		return nil, fmt.Errorf("unsupported key type %q", block.Type)
	}

	pk, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	var psaToken psatoken.PSAToken

	if err = psaToken.FromCOSE(token); err != nil {
		return nil, err
	}

	if err = psaToken.Verify(pk); err != nil {
		return nil, err
	}

	return processClaims(psaToken.PSATokenClaims), nil
}

func processClaims(token psatoken.PSATokenClaims) map[string]interface{} {
	claims := make(map[string]interface{})

	claims["Profile"] = token.Profile
	claims["PartitionID"] = token.PartitionID
	claims["SecurityLifeCycle"] = token.SecurityLifeCycle
	claims["ImplID"] = token.ImplID
	claims["BootSeed"] = token.BootSeed
	claims["HwVersion"] = token.HwVersion
	claims["Nonce"] = token.Nonce
	claims["InstanceID"] = token.InstID
	claims["VSI"] = token.VSI
	claims["NoSwMeasurements"] = token.NoSwMeasurements

	if token.NoSwMeasurements == 0 {
		swComponents := make([]map[string]interface{}, len(token.SwComponents))

		for i, swComp := range token.SwComponents {
			swCompClaims := make(map[string]interface{})

			swCompClaims["MeasurementType"] = swComp.MeasurementType
			swCompClaims["MeasurementValue"] = swComp.MeasurementValue
			swCompClaims["Version"] = swComp.Version
			swCompClaims["SignerID"] = swComp.SignerID
			swCompClaims["MeasurementDesc"] = swComp.MeasurementDesc

			swComponents[i] = swCompClaims
		}

		claims["SwComponents"] = swComponents
	}

	return claims
}

func main() {
	var handshakeConfig = plugin.HandshakeConfig{
		ProtocolVersion:  1,
		MagicCookieKey:   "VERAISON_PLUGIN",
		MagicCookieValue: "VERAISON",
	}

	var pluginMap = map[string]plugin.Plugin{
		"evidenceextractor": &common.EvidenceExtractorPlugin{
			Impl: &EvidenceExtractor{},
		},
	}

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: handshakeConfig,
		Plugins:         pluginMap,
	})
}
