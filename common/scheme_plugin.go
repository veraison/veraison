// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package common

import (
	"encoding/json"
	"fmt"
	"log"
	"net/rpc"

	"github.com/hashicorp/go-plugin"
)

type SchemePlugin struct {
	Impl IScheme
}

func (p SchemePlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &SchemeServer{Impl: p.Impl}, nil
}

func (p SchemePlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &SchemeRPC{client: c}, nil
}

func (p *SchemePlugin) GetName() string {
	return p.Impl.GetName()
}

func (p *SchemePlugin) GetFormat() AttestationFormat {
	return p.Impl.GetFormat()
}

func (p *SchemePlugin) GetTrustAnchorID(token *AttestationToken) (string, error) {
	return p.Impl.GetTrustAnchorID(token)
}

func (p *SchemePlugin) SynthKeysFromSwComponent(tenantID string, swComp *Endorsement) ([]string, error) {
	return p.Impl.SynthKeysFromSwComponent(tenantID, swComp)
}

func (p *SchemePlugin) SynthKeysFromTrustAnchor(tenantID string, ta *Endorsement) ([]string, error) {
	return p.Impl.SynthKeysFromTrustAnchor(tenantID, ta)
}

func (p *SchemePlugin) ExtractEvidence(
	token *AttestationToken,
	trustAnchor string,
) (*ExtractedEvidence, error) {
	return p.Impl.ExtractEvidence(token, trustAnchor)
}

func (p *SchemePlugin) GetAttestation(ec *EvidenceContext, endorsements []string) (*Attestation, error) {
	return p.Impl.GetAttestation(ec, endorsements)
}

type SchemeServer struct {
	Impl IScheme
}

func (s *SchemeServer) GetName(args interface{}, resp *string) error {
	*resp = s.Impl.GetName()
	return nil
}

func (s *SchemeServer) GetFormat(args interface{}, resp *AttestationFormat) error {
	*resp = s.Impl.GetFormat()
	return nil
}

type SynthKeysArgs struct {
	TenantID        string
	EndorsementJSON []byte
}

func (s *SchemeServer) SynthKeysFromSwComponent(args SynthKeysArgs, resp *[]string) error {
	var (
		err    error
		swComp Endorsement
	)

	err = json.Unmarshal(args.EndorsementJSON, &swComp)
	if err != nil {
		return fmt.Errorf("unmarshaling software component: %w", err)
	}

	*resp, err = s.Impl.SynthKeysFromSwComponent(args.TenantID, &swComp)

	return err
}

func (s *SchemeServer) SynthKeysFromTrustAnchor(args SynthKeysArgs, resp *[]string) error {
	var (
		err error
		ta  Endorsement
	)

	err = json.Unmarshal(args.EndorsementJSON, &ta)
	if err != nil {
		return fmt.Errorf("unmarshaling trust anchor: %w", err)
	}

	*resp, err = s.Impl.SynthKeysFromTrustAnchor(args.TenantID, &ta)

	return err
}

func (s *SchemeServer) GetTrustAnchorID(data []byte, resp *string) error {
	var (
		err   error
		token AttestationToken
	)

	err = json.Unmarshal(data, &token)
	if err != nil {
		return fmt.Errorf("unmarshaling attestation token: %w", err)
	}

	*resp, err = s.Impl.GetTrustAnchorID(&token)

	return err
}

type ExtractEvidenceArgs struct {
	Token       []byte
	TrustAnchor string
}

func (s *SchemeServer) ExtractEvidence(args ExtractEvidenceArgs, resp *[]byte) error {
	var token AttestationToken
	if err := json.Unmarshal(args.Token, &token); err != nil {
		return fmt.Errorf("unmarshaling token: %w", err)
	}

	extracted, err := s.Impl.ExtractEvidence(&token, args.TrustAnchor)
	if err != nil {
		return err
	}

	*resp, err = json.Marshal(extracted)
	return err
}

type GetAttestationArgs struct {
	Evidence     []byte
	Endorsements []string
}

func (s *SchemeServer) GetAttestation(args GetAttestationArgs, resp *[]byte) error {
	var (
		ec  EvidenceContext
		err error
	)

	err = json.Unmarshal(args.Evidence, &ec)
	if err != nil {
		return fmt.Errorf("unmarshaling evidence: %w", err)
	}

	attestation, err := s.Impl.GetAttestation(&ec, args.Endorsements)
	if err != nil {
		return err
	}

	*resp, err = json.Marshal(attestation)

	return err
}

type SchemeRPC struct {
	client *rpc.Client
}

func (s *SchemeRPC) GetName() string {
	var resp string
	err := s.client.Call("Plugin.GetName", new(interface{}), &resp)
	if err != nil {
		log.Printf("Plugin.GetName RPC call failed: %v", err) // nolint
		return ""
	}
	return resp
}

func (s *SchemeRPC) GetFormat() AttestationFormat {
	var resp AttestationFormat
	err := s.client.Call("Plugin.GetFormat", new(interface{}), &resp)
	if err != nil {
		log.Printf("Plugin.GetFormat RPC call failed: %v", err)
		return AttestationFormat_UNKNOWN_FORMAT
	}
	return resp
}

func (s *SchemeRPC) SynthKeysFromSwComponent(tenantID string, swComp *Endorsement) ([]string, error) {
	var (
		err  error
		resp []string
		args SynthKeysArgs
	)

	args.TenantID = tenantID

	args.EndorsementJSON, err = json.Marshal(swComp)
	if err != nil {
		return nil, fmt.Errorf("marshaling software component: %w", err)
	}

	err = s.client.Call("Plugin.SynthKeysFromSwComponent", args, &resp)
	if err != nil {
		return nil, fmt.Errorf("Plugin.SynthKeysFromSwComponent RPC call failed: %w", err) // nolint
	}

	return resp, nil
}

func (s *SchemeRPC) SynthKeysFromTrustAnchor(tenantID string, ta *Endorsement) ([]string, error) {
	var (
		err  error
		resp []string
		args SynthKeysArgs
	)

	args.TenantID = tenantID

	args.EndorsementJSON, err = json.Marshal(ta)
	if err != nil {
		return nil, fmt.Errorf("marshaling trust anchor: %w", err)
	}

	err = s.client.Call("Plugin.SynthKeysFromTrustAnchor", args, &resp)
	if err != nil {
		return nil, fmt.Errorf("Plugin.SynthKeysFromTrustAnchor RPC call failed: %w", err) // nolint
	}

	return resp, nil
}

func (s *SchemeRPC) GetTrustAnchorID(token *AttestationToken) (string, error) {
	data, err := json.Marshal(token)
	if err != nil {
		return "", fmt.Errorf("marshaling token: %w", err)
	}

	var resp string
	err = s.client.Call("Plugin.GetTrustAnchorID", data, &resp)
	if err != nil {
		return "", fmt.Errorf("Plugin.GetTrustAnchorID RCP call failed: %w", err) // nolint
	}

	return resp, nil
}

func (s *SchemeRPC) ExtractEvidence(token *AttestationToken, trustAnchor string) (*ExtractedEvidence, error) {
	var (
		err  error
		args ExtractEvidenceArgs
	)
	args.Token, err = json.Marshal(token)
	if err != nil {
		return nil, fmt.Errorf("marshaling token: %w", err)
	}
	args.TrustAnchor = trustAnchor

	var resp []byte
	err = s.client.Call("Plugin.ExtractEvidence", args, &resp)
	if err != nil {
		return nil, fmt.Errorf("Plugin.ExtractEvidence RCP call failed: %w", err) // nolint
	}

	var extracted ExtractedEvidence
	err = json.Unmarshal(resp, &extracted)
	if err != nil {
		return nil, fmt.Errorf("unmarshaling extracted evidence: %w", err)
	}

	return &extracted, nil
}

func (s *SchemeRPC) GetAttestation(ec *EvidenceContext, endorsements []string) (*Attestation, error) {
	var (
		args        GetAttestationArgs
		attestation Attestation
		err         error
		resp        []byte
	)

	args.Evidence, err = json.Marshal(ec)
	if err != nil {
		return nil, fmt.Errorf("marshaling evidence: %w", err)
	}
	args.Endorsements = endorsements

	err = s.client.Call("Plugin.GetAttestation", args, &resp)
	if err != nil {
		return nil, fmt.Errorf("Plugin.GetAttestation RCP call failed: %w", err) // nolint
	}

	err = json.Unmarshal(resp, &attestation)

	return &attestation, err
}
