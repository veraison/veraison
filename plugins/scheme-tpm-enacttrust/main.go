// Copyright 2021-2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"net/url"
	"strings"

	tpm2 "github.com/google/go-tpm/tpm2"
	uuid "github.com/google/uuid"
	plugin "github.com/hashicorp/go-plugin"
	common "github.com/veraison/common"
	structpb "google.golang.org/protobuf/types/known/structpb"
)

type Endorsements struct {
	Digest string
}

func (e *Endorsements) Populate(strings []string) error {
	if len(strings) != 1 {
		return fmt.Errorf("incorrect endorsements number: expected exactly 1, but found %v", strings)
	}

	e.Digest = strings[0]

	return nil
}

type Scheme struct{}

func (s Scheme) GetName() string {
	return common.AttestationFormat_TPM_ENACTTRUST.String()
}

func (s Scheme) GetFormat() common.AttestationFormat {
	return common.AttestationFormat_TPM_ENACTTRUST
}

func (s Scheme) SynthKeysFromSwComponent(tenantID string, swComp *common.Endorsement) ([]string, error) {
	return synthKeysFromParts("software component", tenantID, swComp.GetAttributes())
}

func (s Scheme) SynthKeysFromTrustAnchor(tenantID string, ta *common.Endorsement) ([]string, error) {
	return synthKeysFromParts("trust anchor", tenantID, ta.GetAttributes())
}

func (s Scheme) GetTrustAnchorID(token *common.AttestationToken) (string, error) {
	if token.Format != common.AttestationFormat_TPM_ENACTTRUST {
		return "", fmt.Errorf("wrong format: expect %q, but found %q",
			common.AttestationFormat_TPM_ENACTTRUST.String(),
			token.Format.String(),
		)
	}

	var decoded Token

	if err := decoded.Decode(token.Data); err != nil {
		return "", err
	}

	nodeID, err := uuid.FromBytes(decoded.AttestationData.ExtraData)
	if err != nil {
		return "", fmt.Errorf("could not decode node-id: %v", err)
	}

	return fmt.Sprintf("tpm-enacttrust://%d/%s", token.TenantId, nodeID.String()), nil
}

func (s Scheme) ExtractEvidence(
	token *common.AttestationToken,
	trustAnchor string,
) (*common.ExtractedEvidence, error) {
	if token.Format != common.AttestationFormat_TPM_ENACTTRUST {
		return nil, fmt.Errorf("wrong format: expect %q, but found %q",
			common.AttestationFormat_TPM_ENACTTRUST.String(),
			token.Format.String(),
		)
	}

	var decoded Token

	if err := decoded.Decode(token.Data); err != nil {
		return nil, fmt.Errorf("could not decode token: %v", err)
	}

	pubKey, err := parseKey(trustAnchor)
	if err != nil {
		return nil, fmt.Errorf("could not parse trust anchor: %v", err)
	}

	if err = decoded.VerifySignature(pubKey); err != nil {
		return nil, fmt.Errorf("could not verify token signature: %v", err)
	}

	if decoded.AttestationData.Type != tpm2.TagAttestQuote {
		return nil, fmt.Errorf("wrong TPMS_ATTEST type: expected %d, but got %d",
			tpm2.TagAttestQuote, decoded.AttestationData.Type)
	}

	var pcrs []int64
	for _, pcr := range decoded.AttestationData.AttestedQuoteInfo.PCRSelection.PCRs {
		pcrs = append(pcrs, int64(pcr))
	}

	evidence := common.NewExtractedEvidence()
	evidence.Evidence["pcr-selection"] = pcrs
	evidence.Evidence["hash-algorithm"] = int64(decoded.AttestationData.AttestedQuoteInfo.PCRSelection.Hash)
	evidence.Evidence["pcr-digest"] = []byte(decoded.AttestationData.AttestedQuoteInfo.PCRDigest)

	nodeID, err := uuid.FromBytes(decoded.AttestationData.ExtraData)
	if err != nil {
		return nil, fmt.Errorf("could not decode node-id: %v", err)
	}
	evidence.SoftwareID = fmt.Sprintf("tpm-enacttrust://%d/%s", token.TenantId, nodeID.String())

	return evidence, nil
}

func (s Scheme) GetAttestation(
	ec *common.EvidenceContext,
	endorsementStrings []string,
) (*common.Attestation, error) {

	attestation := common.NewAttestation(ec)

	attestation.Result.TrustVector.SoftwareIntegrity = common.AR_Status_FAILURE
	attestation.Result.Status = common.AR_Status_FAILURE

	digestValue, ok := ec.Evidence.AsMap()["pcr-digest"]
	if !ok {
		return attestation, fmt.Errorf("evidence does not contain \"pcr-digest\" entry")
	}

	evidenceDigest, ok := digestValue.(string)
	if !ok {
		return attestation, fmt.Errorf("wrong type value \"pcr-digest\" entry; expected string but found %T", digestValue)
	}

	var endorsements Endorsements
	if err := endorsements.Populate(endorsementStrings); err != nil {
		return attestation, err
	}

	if endorsements.Digest == evidenceDigest {
		attestation.Result.TrustVector.SoftwareIntegrity = common.AR_Status_SUCCESS
		attestation.Result.Status = common.AR_Status_SUCCESS
	}

	return attestation, nil
}

// TODO(tho) factor out (== scheme-tpm-enacttrust)
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

// TODO(tho) factor out (== scheme-tpm-enactrust)
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

func parseKey(keyString string) (*ecdsa.PublicKey, error) {
	buf, err := base64.StdEncoding.DecodeString(keyString)
	if err != nil {
		return nil, err
	}

	key, err := x509.ParsePKIXPublicKey(buf)
	if err != nil {
		return nil, fmt.Errorf("could not parse public key: %v", err)
	}

	ret, ok := key.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("could not extract EC public key; got [%T]: %v", key, err)
	}

	return ret, nil
}

// Token is the container for the decoded EnactTrust token
type Token struct {
	// TPMS_ATTEST decoded from the token
	AttestationData *tpm2.AttestationData
	// Raw token bytes
	Raw []byte
	// TPMT_SIGNATURE decoded from the token
	Signature *tpm2.Signature
}

func (t *Token) Decode(data []byte) error {
	// The first two bytes are the SIZE of the following TPMS_ATTEST
	// structure. The following SIZE bytes are the TPMS_ATTEST structure,
	// the remaining bytes are the signature.
	if len(data) < 3 {
		return fmt.Errorf("could not get data size; token too small (%d)", len(data))
	}

	size := binary.BigEndian.Uint16(data[:2])
	if len(data) < int(2+size) {
		return fmt.Errorf("TPMS_ATTEST appears truncated; expected %d, but got %d bytes",
			size, len(data)-2)
	}

	var err error

	t.Raw = data[2 : 2+size]
	t.AttestationData, err = tpm2.DecodeAttestationData(t.Raw)
	if err != nil {
		return fmt.Errorf("could not decode TPMS_ATTEST: %v", err)
	}

	t.Signature, err = tpm2.DecodeSignature(bytes.NewBuffer(data[2+size:]))
	if err != nil {
		return fmt.Errorf("could not decode TPMT_SIGNATURE: %v", err)
	}

	return nil
}

func (t Token) VerifySignature(key *ecdsa.PublicKey) error {
	digest := sha256.Sum256(t.Raw)

	if !ecdsa.Verify(key, digest[:], t.Signature.ECC.R, t.Signature.ECC.S) {
		return fmt.Errorf("failed to verify signature")
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
