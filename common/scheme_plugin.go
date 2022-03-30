// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package common

import (
	"encoding/json"
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
	p.Impl = lp.Raw.(IScheme)
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

func (p *SchemePlugin) ExtractEvidence(
	token *AttestationToken,
	trustAnchor string,
) (*ExtractedEvidence, error) {
	return p.Impl.ExtractEvidence(token, trustAnchor)
}

func (p *SchemePlugin) GetAttestation(ec *EvidenceContext, endorsements string) (*Attestation, error) {
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

func (s *SchemeServer) GetTrustAnchorID(data []byte, resp *string) error {
	var token AttestationToken
	if err := json.Unmarshal(data, &token); err != nil {
		return err
	}

	var err error
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
	Endorsements string
}

func (s *SchemeServer) GetAttestation(args GetAttestationArgs, resp *[]byte) error {
	var ec EvidenceContext
	var err error

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
		log.Printf("ERROR during GetName RPC: %v\n", err)
		return AttestationFormat_UNKNOWN_FORMAT
	}
	return resp
}

func (s *SchemeRPC) GetTrustAnchorID(token *AttestationToken) string {
	data, err := json.Marshal(token)
	if err != nil {
		log.Printf("ERROR during token marshing: %v\n", err)
		return ""
	}

	var resp string
	err = s.client.Call("Plugin.GetTrustAnchorID", data, &resp)
	if err != nil {
		log.Printf("ERROR getting trust anchor ID: %v\n", err)
		return ""
	}

	return resp
}

func (s *SchemeRPC) ExtractEvidence(token *AttestationToken, trustAnchor string) (*ExtractedEvidence, error) {
	var err error
	var args ExtractEvidenceArgs
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

func (s *SchemeRPC) GetAttestation(ec *EvidenceContext, endorsements string) (*Attestation, error) {
	var args GetAttestationArgs
	var err error

	args.Evidence, err = json.Marshal(ec)
	if err != nil {
		log.Printf("ERROR unmarshaling evidence: %v\n", err)
		return nil, err
	}
	args.Endorsements = endorsements

	var resp []byte
	err = s.client.Call("Plugin.GetAttestation", args, &resp)
	if err != nil {
		log.Printf("ERROR getting attestaton: %v\n", err)
		return nil, err
	}

	var attestation Attestation
	err = json.Unmarshal(resp, &attestation)

	return &attestation, err
}

func LoadSchemePlugin(locations []string, format AttestationFormat) (*SchemePlugin, error) {
	lp, err := LoadPlugin(locations, "scheme", format.String(), false)
	if err != nil {
		return nil, err
	}

	var schemePlugin SchemePlugin
	err = schemePlugin.Init(lp)
	return &schemePlugin, err
}
