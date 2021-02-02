// Copyright 2021 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"encoding/json"
	"log"
	"net/rpc"

	"github.com/hashicorp/go-plugin"
)

type PolicyStorePlugin struct {
	Impl IPolicyStore
}

func (p *PolicyStorePlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &PolicyStoreServer{Impl: p.Impl}, nil
}

func (p *PolicyStorePlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &PolicyStoreRPC{client: c}, nil
}

type GetPolicyArgs struct {
	TenantId    int
	TokenFormat TokenFormat
}

type PutPolicyArgs struct {
	TenantId int
	Policy   *Policy
}

type PolicyStoreServer struct {
	Impl IPolicyStore
}

func (s *PolicyStoreServer) GetName(args interface{}, resp *string) error {
	*resp = s.Impl.GetName()
	return nil
}

func (s *PolicyStoreServer) Init(args PolicyStoreParams, resp *string) error {
	return s.Impl.Init(args)
}

func (s *PolicyStoreServer) GetPolicy(args GetPolicyArgs, resp *Policy) error {
	policy, err := s.Impl.GetPolicy(args.TenantId, args.TokenFormat)
	*resp = *policy
	return err
}

func (s *PolicyStoreServer) PutPolicy(args PutPolicyArgs, resp *interface{}) error {
	return s.Impl.PutPolicy(args.TenantId, args.Policy)
}

func (s *PolicyStoreServer) Close(args interface{}, resp *interface{}) error {
	return s.Impl.Close()
}

type PolicyStoreRPC struct {
	client *rpc.Client
}

func (e *PolicyStoreRPC) GetName() string {
	var resp string
	err := e.client.Call("Plugin.GetName", new(interface{}), &resp)
	if err != nil {
		log.Printf("ERROR during GetName RPC: %v\n", err)
		return ""
	}
	return resp
}

func (e *PolicyStoreRPC) Init(args PolicyStoreParams) error {
	return e.client.Call("Plugin.Init", args, nil)
}

func (e *PolicyStoreRPC) GetPolicy(tenantId int, tokenFormat TokenFormat) (*Policy, error) {
	var resp Policy

	args := GetPolicyArgs{TenantId: tenantId, TokenFormat: tokenFormat}
	err := e.client.Call("Plugin.GetPolicy", args, &resp)

	return &resp, err
}

func (e *PolicyStoreRPC) PutPolicy(tenantId int, policy *Policy) error {
	args := PutPolicyArgs{TenantId: tenantId, Policy: policy}
	// TODO: figure out why the last argument here must be non-nil
	return e.client.Call("Plugin.PutPolicy", args, new(interface{}))
}

func (e *PolicyStoreRPC) Close() error {
	return e.client.Call("Plugin.Close", new(interface{}), nil)
}

type PolicyEnginePlugin struct {
	Impl IPolicyEngine
}

func (p *PolicyEnginePlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &PolicyEngineServer{Impl: p.Impl}, nil
}

func (p *PolicyEnginePlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &PolicyEngineRPC{client: c}, nil
}

type PolicyEngineArgs struct {
	Evidence     map[string]interface{}
	Endorsements map[string]interface{}
}

type GetAttetationResultArgs struct {
	PolicyEngineArgs
	Simple bool
}

type PolicyEngineServer struct {
	Impl IPolicyEngine
}

func (s *PolicyEngineServer) GetName(args interface{}, resp *string) error {
	*resp = s.Impl.GetName()
	return nil
}

func (s *PolicyEngineServer) Init(args PolicyEngineParams, resp *interface{}) error {
	return s.Impl.Init(args)
}

func (s *PolicyEngineServer) LoadPolicy(policy []byte, resp *interface{}) error {
	return s.Impl.LoadPolicy(policy)
}

func (s *PolicyEngineServer) GetClaims(args PolicyEngineArgs, resp *map[string]interface{}) error {
	var err error
	*resp, err = s.Impl.GetClaims(args.Evidence, args.Endorsements)
	return err
}

func (s *PolicyEngineServer) CheckValid(args PolicyEngineArgs, resp *bool) error {
	var err error
	*resp, err = s.Impl.CheckValid(args.Evidence, args.Endorsements)
	return err
}

func (s *PolicyEngineServer) GetAttetationResult(argBlob []byte, resp *[]byte) error {
	var args GetAttetationResultArgs
	// NOTE: encoding/gob used to serialize objects by net/rpc cannot handle []interface{}, which
	//       necessitates pre-serialing any objects that may contain arbitrary JSON decodings.
	if err := json.Unmarshal(argBlob, &args); err != nil {
		return err
	}

	var result AttestationResult
	if err := s.Impl.GetAttetationResult(args.Evidence, args.Endorsements, args.Simple, &result); err != nil {
		return err
	}

	var err error
	*resp, err = json.Marshal(result)
	return err
}

func (s *PolicyEngineServer) Stop(args interface{}, resp *interface{}) error {
	return s.Impl.Stop()
}

type PolicyEngineRPC struct {
	client *rpc.Client
}

func (e *PolicyEngineRPC) GetName() string {
	var resp string
	err := e.client.Call("Plugin.GetName", new(interface{}), &resp)
	if err != nil {
		log.Printf("ERROR during GetName RPC: %v\n", err)
		return ""
	}
	return resp
}

func (e *PolicyEngineRPC) Init(args PolicyEngineParams) error {
	return e.client.Call("Plugin.Init", args, new(interface{}))
}

func (e *PolicyEngineRPC) LoadPolicy(policy []byte) error {
	return e.client.Call("Plugin.LoadPolicy", policy, new(interface{}))
}

func (e *PolicyEngineRPC) CheckValid(
	evidence map[string]interface{},
	endorsements map[string]interface{},
) (bool, error) {
	var resp bool

	args := PolicyEngineArgs{Evidence: evidence, Endorsements: endorsements}
	err := e.client.Call("Plugin.CheckValid", args, &resp)

	return resp, err
}

func (e *PolicyEngineRPC) GetClaims(
	evidence map[string]interface{},
	endorsements map[string]interface{},
) (map[string]interface{}, error) {
	resp := make(map[string]interface{})

	args := PolicyEngineArgs{Evidence: evidence, Endorsements: endorsements}
	err := e.client.Call("Plugin.GetClaims", args, &resp)

	return resp, err
}

func (e *PolicyEngineRPC) GetAttetationResult(
	evidence map[string]interface{},
	endorsements map[string]interface{},
	simple bool,
	result *AttestationResult,
) error {

	args := GetAttetationResultArgs{
		PolicyEngineArgs: PolicyEngineArgs{
			Evidence:     evidence,
			Endorsements: endorsements,
		},
		Simple: simple,
	}

	// NOTE: encoding/gob used to serialize objects by net/rpc cannot handle []interface{}, which
	//       necessitates pre-serialing any objects that may contain arbitrary JSON decodings.
	argsBlob, err := json.Marshal(args)
	if err != nil {
		return err
	}

	var resultBlob []byte
	err = e.client.Call("Plugin.GetAttetationResult", argsBlob, &resultBlob)
	if err != nil {
		return err
	}

	return json.Unmarshal(resultBlob, result)
}

func (e *PolicyEngineRPC) Stop() error {
	return e.client.Call("Plugin.Stop", new(interface{}), new(interface{}))
}
