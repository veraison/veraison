// Copyright 2021 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"encoding/json"
	"log"
	"net/rpc"

	"github.com/hashicorp/go-plugin"
)

type EvidenceExtractorPlugin struct {
	Impl IEvidenceExtractor
}

func (p EvidenceExtractorPlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &EvidenceExtractorServer{Impl: p.Impl}, nil
}

func (p EvidenceExtractorPlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &EvidenceExtractorRPC{client: c}, nil
}

type EvidenceExtractorServer struct {
	Impl IEvidenceExtractor
}

func (s EvidenceExtractorServer) GetName(args interface{}, resp *string) error {
	*resp = s.Impl.GetName()
	return nil
}

func (s EvidenceExtractorServer) Init(params EvidenceExtractorParams, resp *interface{}) error {
	return s.Impl.Init(params)
}

func (s EvidenceExtractorServer) Close(args interface{}, resp *interface{}) error {
	return s.Impl.Close()
}

func (s EvidenceExtractorServer) GetTrustAnchorID(token []byte, resp *[]byte) error {
	result, err := s.Impl.GetTrustAnchorID(token)
	if err != nil {
		return err
	}

	*resp, err = json.Marshal(result)
	return err
}

type ExtractEvidenceArgs struct {
	Token       []byte
	TrustAnchor []byte
}

func (s EvidenceExtractorServer) ExtractEvidence(args ExtractEvidenceArgs, resp *[]byte) error {
	result, err := s.Impl.ExtractEvidence(args.Token, args.TrustAnchor)
	if err != nil {
		return err
	}

	*resp, err = json.Marshal(result)
	return err
}

type EvidenceExtractorRPC struct {
	client *rpc.Client
}

func (e EvidenceExtractorRPC) GetName() string {
	var resp string
	err := e.client.Call("Plugin.GetName", new(interface{}), &resp)
	if err != nil {
		log.Printf("ERROR during GetName RPC: %v\n", err)
		return ""
	}
	return resp
}

func (e EvidenceExtractorRPC) Init(params EvidenceExtractorParams) error {
	return e.client.Call("Plugin.Init", params, new(interface{}))
}

func (e EvidenceExtractorRPC) Close() error {
	return e.client.Call("Plugin.Close", new(interface{}), new(interface{}))
}

func (e EvidenceExtractorRPC) GetTrustAnchorID(token []byte) (TrustAnchorID, error) {
	var resp []byte

	err := e.client.Call("Plugin.GetTrustAnchorID", token, &resp)
	if err != nil {
		return TrustAnchorID{}, err
	}

	var taID TrustAnchorID

	err = json.Unmarshal(resp, &taID)
	return taID, err
}

func (e EvidenceExtractorRPC) ExtractEvidence(token []byte, ta []byte) (map[string]interface{}, error) {
	args := &ExtractEvidenceArgs{
		Token:       token,
		TrustAnchor: ta,
	}

	var resp []byte

	err := e.client.Call("Plugin.ExtractEvidence", args, &resp)
	if err != nil {
		return nil, err
	}

	var evidence map[string]interface{}

	err = json.Unmarshal(resp, &evidence)
	return evidence, err
}
