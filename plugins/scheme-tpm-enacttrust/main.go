// Copyright 2021-2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	plugin "github.com/hashicorp/go-plugin"
	"github.com/veraison/common"
	structpb "google.golang.org/protobuf/types/known/structpb"
)

type Scheme struct{}

func (s Scheme) GetName() string {
	return common.AttestationFormat_TPM_ENACTTRUST.String()
}

func (s Scheme) GetFormat() common.AttestationFormat {
	return common.AttestationFormat_TPM_ENACTTRUST
}

func (s Scheme) SynthKeysFromSwComponent(tenantID string, swComp *common.SwComponent) ([]string, error) {
	return synthKeysFromParts("software component", tenantID, swComp.GetId().GetParts())
}

func (s Scheme) SynthKeysFromTrustAnchor(tenantID string, ta *common.TrustAnchor) ([]string, error) {
	return synthKeysFromParts("trust anchor", tenantID, ta.GetId().GetParts())
}

func (s Scheme) GetTrustAnchorID(token *common.AttestationToken) (string, error) {
	return "", errors.New("TODO")
}

func (s Scheme) ExtractEvidence(token *common.AttestationToken, trustAnchor string) (*common.ExtractedEvidence, error) {
	return nil, errors.New("TODO")
}

func (s Scheme) GetAttestation(ec *common.EvidenceContext, endorsements []string) (*common.Attestation, error) {
	return nil, errors.New("TODO")
}

// TODO(tho) factor out (== scheme-psa)
func appendMandatoryPathSegment(
	path []string, key string, fields map[string]*structpb.Value,
) ([]string, error) {
	v, ok := fields[key]
	if !ok {
		return path, fmt.Errorf("missing mandatory %s", key)
	}

	segment := v.GetStringValue()
	if segment == "" {
		return path, fmt.Errorf("empty mandatory %s", key)
	}

	return append(path, segment), nil
}

// TODO(tho) factor out (== scheme-psa)
func getFieldsFromParts(parts *structpb.Struct) (map[string]*structpb.Value, error) {
	if parts == nil {
		return nil, errors.New("no parts found")
	}

	fields := parts.GetFields()
	if fields == nil {
		return nil, errors.New("no fields found")
	}

	return fields, nil
}

func synthKeysFromParts(scope, tenantID string, parts *structpb.Struct) ([]string, error) {
	var (
		absPath []string
		fields  map[string]*structpb.Value
		err     error
	)

	fields, err = getFieldsFromParts(parts)
	if err != nil {
		return nil, fmt.Errorf("unable to synthesize %s abs-path: %w", scope, err)
	}

	absPath, err = appendMandatoryPathSegment(absPath, "enacttrust-tpm.node-id", fields)
	if err != nil {
		return nil, fmt.Errorf("unable to synthesize %s abs-path: %w", scope, err)
	}

	lookupKey := url.URL{
		Scheme: "tpm-enacttrust",
		Host:   tenantID,
		Path:   strings.Join(absPath, "/"),
	}

	return []string{lookupKey.String()}, nil
}

func main() {
	var handshakeConfig = plugin.HandshakeConfig{
		ProtocolVersion:  1,
		MagicCookieKey:   "VERAISON_PLUGIN",
		MagicCookieValue: "VERAISON",
	}

	var pluginMap = map[string]plugin.Plugin{
		"scheme": &common.SchemePlugin{
			Impl: &Scheme{},
		},
	}

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: handshakeConfig,
		Plugins:         pluginMap,
	})
}
