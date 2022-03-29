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
	Impl         IScheme
	PluginClient *plugin.Client
	RPCClient    plugin.ClientProtocol
}

func (p SchemePlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &SchemeServer{Impl: p.Impl}, nil
}

func (p SchemePlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &SchemeRPC{client: c}, nil
}

func (p *SchemePlugin) Init(lp *LoadedPlugin) error {
	var ok bool

	p.Impl, ok = lp.Raw.(IScheme)
	if !ok {
		return fmt.Errorf("the loaded plugin (%T) does not implement common.IScheme", lp.Raw)
	}

	p.RPCClient = lp.RPCClient
	p.PluginClient = lp.PluginClient

	return nil
}

func (p *SchemePlugin) Close() error {
	p.PluginClient.Kill()
	return p.RPCClient.Close()
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

func (p *SchemePlugin) SynthKeysFromSwComponent(tenantID string, swComp *SwComponent) ([]string, error) {
	return p.Impl.SynthKeysFromSwComponent(tenantID, swComp)
}

func (p *SchemePlugin) SynthKeysFromTrustAnchor(tenantID string, ta *TrustAnchor) ([]string, error) {
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
		swComp SwComponent
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
		ta  TrustAnchor
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
		return err
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
		return err
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
		log.Printf("ERROR during GetName RPC: %v\n", err)
		return ""
	}
	return resp
}

func (s *SchemeRPC) GetFormat() AttestationFormat {
	var resp AttestationFormat
	err := s.client.Call("Plugin.GetFormat", new(interface{}), &resp)
	if err != nil {
		log.Printf("ERROR during GetFormat RPC: %v\n", err)
		return AttestationFormat_UNKNOWN_FORMAT
	}
	return resp
}

func (s *SchemeRPC) SynthKeysFromSwComponent(tenantID string, swComp *SwComponent) ([]string, error) {
	var (
		err  error
		resp []string
		args SynthKeysArgs
	)

	args.TenantID = tenantID

	args.EndorsementJSON, err = json.Marshal(swComp)
	if err != nil {
		return nil, fmt.Errorf("failed unmarshaling response from RPC server: %w", err)
	}

	err = s.client.Call("Plugin.SynthKeysFromSwComponent", args, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed call to SynthKeysFromSwComponent: %w", err)
	}

	return resp, nil
}

func (s *SchemeRPC) SynthKeysFromTrustAnchor(tenantID string, ta *TrustAnchor) ([]string, error) {
	var (
		err  error
		resp []string
		args SynthKeysArgs
	)

	args.TenantID = tenantID

	args.EndorsementJSON, err = json.Marshal(ta)
	if err != nil {
		return nil, fmt.Errorf("failed unmarshaling response from RPC server: %w", err)
	}

	err = s.client.Call("Plugin.SynthKeysFromTrustAnchor", args, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed call to SynthKeysFromTrustAnchor: %w", err)
	}

	return resp, nil
}

func (s *SchemeRPC) GetTrustAnchorID(token *AttestationToken) (string, error) {
	data, err := json.Marshal(token)
	if err != nil {
		return "", err
	}

	var resp string
	err = s.client.Call("Plugin.GetTrustAnchorID", data, &resp)
	if err != nil {
		return "", err
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
		log.Printf("ERROR during token marshing: %v\n", err)
		return nil, err
	}
	args.TrustAnchor = trustAnchor

	var resp []byte
	err = s.client.Call("Plugin.ExtractedEvidence", args, &resp)
	if err != nil {
		log.Printf("ERROR extracting evidence: %v\n", err)
		return nil, err
	}

	var extracted ExtractedEvidence
	err = json.Unmarshal(resp, &extracted)
	if err != nil {
		log.Printf("ERROR unmarshaling extracted evidence: %v\n", err)
		return nil, err
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
		log.Printf("ERROR unmarshaling evidence: %v\n", err)
		return nil, err
	}
	args.Endorsements = endorsements

	err = s.client.Call("Plugin.GetAttestation", args, &resp)
	if err != nil {
		log.Printf("ERROR getting attestaton: %v\n", err)
		return nil, err
	}

	err = json.Unmarshal(resp, &attestation)

	return &attestation, err
}

func LoadSchemePlugin(locations []string, format AttestationFormat) (*SchemePlugin, error) {
	log.Printf("loading plugin for scheme: %s", format.String())

	lp, err := LoadPlugin(locations, "scheme", format.String(), false)
	if err != nil {
		return nil, err
	}

	var schemePlugin SchemePlugin
	err = schemePlugin.Init(lp)
	return &schemePlugin, err
}
