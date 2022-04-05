// Copyright 2021-2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"net/url"
	"strings"

	plugin "github.com/hashicorp/go-plugin"
	"github.com/veraison/common"
	psatoken "github.com/veraison/psatoken"
	structpb "google.golang.org/protobuf/types/known/structpb"
)

type Endorsements struct {
	ImplID       *[]byte                   `json:"implementation-id"`
	HwVersion    *string                   `json:"hardware-version,omitempty"`
	SwComponents []psatoken.PSASwComponent `json:"software-components"`
}

type Scheme struct {
}

func (s Scheme) GetName() string {
	return common.AttestationFormat_PSA_IOT.String()
}

func (s Scheme) GetFormat() common.AttestationFormat {
	return common.AttestationFormat_PSA_IOT
}

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

func (s Scheme) SynthKeysFromSwComponent(tenantID string, swComp *common.Endorsement) ([]string, error) {
	var (
		absPath []string
		fields  map[string]*structpb.Value
		err     error
	)

	fields, err = getFieldsFromParts(swComp.GetAttributes())
	if err != nil {
		return nil, fmt.Errorf("unable to synthesize software component abs-path: %w", err)
	}

	absPath, err = appendMandatoryPathSegment(absPath, "psa.impl-id", fields)
	if err != nil {
		return nil, fmt.Errorf("unable to synthesize software component abs-path: %w", err)
	}

	verificationLookupKey := url.URL{
		Scheme: "psa-iot",
		Host:   tenantID,
		Path:   strings.Join(absPath, "/"),
	}

	return []string{verificationLookupKey.String()}, nil
}

func (s Scheme) SynthKeysFromTrustAnchor(tenantID string, ta *common.Endorsement) ([]string, error) {
	var (
		absPath []string
		fields  map[string]*structpb.Value
		err     error
	)

	fields, err = getFieldsFromParts(ta.GetAttributes())
	if err != nil {
		return nil, fmt.Errorf("unable to synthesize trust anchor abs-path: %w", err)
	}

	for _, k := range []string{"psa.impl-id", "psa.inst-id"} {
		absPath, err = appendMandatoryPathSegment(absPath, k, fields)
		if err != nil {
			return nil, fmt.Errorf("unable to synthesize trust anchor abs-path: %w", err)
		}
	}

	taLookupKey := url.URL{
		Scheme: "psa-iot",
		Host:   tenantID,
		Path:   strings.Join(absPath, "/"),
	}

	return []string{taLookupKey.String()}, nil
}

func (s Scheme) GetTrustAnchorID(token *common.AttestationToken) (string, error) {
	var psaToken psatoken.PSAToken

	err := psaToken.FromCOSE(token.Data)
	if err != nil {
		return "", fmt.Errorf("could not decode COSE: %v", err)
	}

	implIDString := base64.StdEncoding.EncodeToString(*psaToken.ImplID)
	instIDString := base64.StdEncoding.EncodeToString(*psaToken.InstID)

	return fmt.Sprintf("psa://%d/%s/%s", token.TenantId, implIDString, instIDString), nil
}

func (s Scheme) ExtractEvidence(
	token *common.AttestationToken,
	trustAnchor string,
) (*common.ExtractedEvidence, error) {
	block, rest := pem.Decode([]byte(trustAnchor))
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

	if err = psaToken.FromCOSE(token.Data); err != nil {
		return nil, err
	}

	if err = psaToken.Verify(pk); err != nil {
		return nil, err
	}

	var extracted common.ExtractedEvidence

	claimsMap, err := claimsToMap(&psaToken.PSATokenClaims)
	if err != nil {
		return nil, err
	}

	extracted.Evidence = claimsMap

	implIDString := base64.StdEncoding.EncodeToString(*psaToken.ImplID)
	extracted.SoftwareID = fmt.Sprintf("psa://%d/%s/", token.TenantId, implIDString)

	return &extracted, nil
}

func (s Scheme) GetAttestation(
	ec *common.EvidenceContext,
	endorsementsStrings []string,
) (*common.Attestation, error) {

	attestation := common.Attestation{
		Evidence: ec,
		Result:   new(common.AttestationResult),
	}

	var endorsements []Endorsements
	for i, e := range endorsementsStrings {
		var endorsement Endorsements
		if err := json.Unmarshal([]byte(e), &endorsement); err != nil {
			return nil, fmt.Errorf("could not decode endorsement at index %d: %w", i, err)
		}
		endorsements = append(endorsements, endorsement)
	}

	err := populateAttestationResult(&attestation, endorsements)

	return &attestation, err
}

func claimsToMap(claims *psatoken.PSATokenClaims) (map[string]interface{}, error) {
	data, err := json.Marshal(claims)
	if err != nil {
		return nil, err
	}

	var out map[string]interface{}
	err = json.Unmarshal(data, &out)

	return out, err
}

func mapToClaims(in map[string]interface{}) (*psatoken.PSATokenClaims, error) {
	data, err := json.Marshal(in)
	if err != nil {
		return nil, err
	}

	var claims psatoken.PSATokenClaims
	err = json.Unmarshal(data, &claims)

	return &claims, err
}

func populateAttestationResult(attestation *common.Attestation, endorsements []Endorsements) error {
	tv := common.TrustVector{
		SoftwareUpToDateness: common.AR_Status_UNKNOWN,
		ConfigIntegrity:      common.AR_Status_UNKNOWN,
		RuntimeIntegrity:     common.AR_Status_UNKNOWN,
		CertificationStatus:  common.AR_Status_UNKNOWN,
	}

	claims, err := mapToClaims(attestation.Evidence.Evidence.AsMap())
	if err != nil {
		return err
	}

	match := matchSoftware(claims, endorsements)
	if match == nil {
		tv.SoftwareIntegrity = common.AR_Status_FAILURE
		tv.HardwareAuthenticity = common.AR_Status_UNKNOWN
	} else {
		tv.SoftwareIntegrity = common.AR_Status_SUCCESS

		if claims.HwVersion == nil {
			tv.HardwareAuthenticity = common.AR_Status_UNKNOWN
		} else {
			if *claims.HwVersion == "" || *claims.HwVersion == *match.HwVersion {
				tv.HardwareAuthenticity = common.AR_Status_SUCCESS
			} else {
				tv.HardwareAuthenticity = common.AR_Status_FAILURE
			}
		}
	}

	attestation.Result.TrustVector = &tv

	if tv.SoftwareIntegrity != common.AR_Status_FAILURE && tv.HardwareAuthenticity != common.AR_Status_FAILURE {
		attestation.Result.Status = common.AR_Status_SUCCESS
	} else {
		attestation.Result.Status = common.AR_Status_FAILURE
	}

	attestation.Result.ProcessedEvidence = attestation.Evidence.Evidence
	return nil
}

func matchSoftware(evidence *psatoken.PSATokenClaims, endorsements []Endorsements) *Endorsements {

	evidenceComponents := make(map[string]psatoken.PSASwComponent)
	for _, c := range evidence.SwComponents {
		key := base64.StdEncoding.EncodeToString(*c.MeasurementValue)
		evidenceComponents[key] = c
	}

	for _, endorsement := range endorsements {
		matched := true
		for _, comp := range endorsement.SwComponents {
			key := base64.StdEncoding.EncodeToString(*comp.MeasurementValue)
			evComp, ok := evidenceComponents[key]
			if !ok {
				matched = false
				break
			}

			typeMatched := comp.MeasurementType == "" || comp.MeasurementType == evComp.MeasurementType
			sigMatched := comp.SignerID == nil || bytes.Compare(*comp.SignerID, *evComp.SignerID) == 0
			versionMatched := comp.Version == "" || comp.Version == evComp.Version

			if !(typeMatched && sigMatched && versionMatched) {
				matched = false
				break
			}
		}

		if matched {
			return &endorsement
		}
	}

	return nil
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
