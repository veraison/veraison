// Copyright 2021 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"encoding/json"
	"fmt"
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
	TenantID    int
	TokenFormat AttestationFormat
}

type PutPolicyArgs struct {
	TenantID int
	Policy   *Policy
}

type PolicyStoreServer struct {
	Impl IPolicyStore
}

func (s *PolicyStoreServer) GetName(args interface{}, resp *string) error {
	*resp = s.Impl.GetName()
	return nil
}

func (s *PolicyStoreServer) GetParamDescriptions(args interface{}, resp *[]byte) error {
	result, err := s.Impl.GetParamDescriptions()
	if err != nil {
		return err
	}

	*resp, err = json.Marshal(result)
	return err
}

func (s *PolicyStoreServer) Init(bytes []byte, resp *string) error {
	var params ParamStore
	if err := json.Unmarshal(bytes, &params); err != nil {
		return err
	}

	return s.Impl.Init(&params)
}

func (s *PolicyStoreServer) ListPolicies(tenantID int, resp *[]byte) error {
	result, err := s.Impl.ListPolicies(tenantID)
	if err != nil {
		return err
	}

	*resp, err = json.Marshal(result)
	return err
}

func (s *PolicyStoreServer) GetPolicy(args GetPolicyArgs, resp *Policy) error {
	policy, err := s.Impl.GetPolicy(args.TenantID, args.TokenFormat)

	if policy != nil {
		*resp = *policy
	}

	return err
}

func (s *PolicyStoreServer) PutPolicy(args PutPolicyArgs, resp *interface{}) error {
	return s.Impl.PutPolicy(args.TenantID, args.Policy)
}

func (s *PolicyStoreServer) DeletePolicy(args GetPolicyArgs, resp *interface{}) error {
	return s.Impl.DeletePolicy(args.TenantID, args.TokenFormat)
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

func (e *PolicyStoreRPC) GetParamDescriptions() (map[string]*ParamDescription, error) {
	result := make(map[string]*ParamDescription)

	var resp []byte
	if err := e.client.Call("Plugin.GetParamDescriptions", new(interface{}), &resp); err != nil {
		return nil, err
	}

	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	return result, nil
}

func (e *PolicyStoreRPC) Init(params *ParamStore) error {
	bytes, err := json.Marshal(params)
	if err != nil {
		return err
	}

	return e.client.Call("Plugin.Init", bytes, nil)
}

func (e *PolicyStoreRPC) ListPolicies(tenantID int) ([]PolicyListEntry, error) {
	var result []PolicyListEntry

	var resp []byte

	if err := e.client.Call("Plugin.ListPolicies", tenantID, &resp); err != nil {
		return nil, err
	}

	// NOTE: encoding/gob used to serialize objects by net/rpc cannot handle []interface{}, which
	//       necessitates pre-serialing any objects that may contain arbitrary JSON decodings.
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	return result, nil
}

func (e *PolicyStoreRPC) GetPolicy(tenantID int, tokenFormat AttestationFormat) (*Policy, error) {
	var resp Policy

	args := GetPolicyArgs{TenantID: tenantID, TokenFormat: tokenFormat}
	err := e.client.Call("Plugin.GetPolicy", args, &resp)

	return &resp, err
}

func (e *PolicyStoreRPC) PutPolicy(tenantID int, policy *Policy) error {
	args := PutPolicyArgs{TenantID: tenantID, Policy: policy}
	// TODO: figure out why the last argument here must be non-nil
	return e.client.Call("Plugin.PutPolicy", args, new(interface{}))
}

func (e *PolicyStoreRPC) DeletePolicy(tenantID int, tokenFormat AttestationFormat) error {
	args := GetPolicyArgs{TenantID: tenantID, TokenFormat: tokenFormat}
	// TODO: figure out why the last argument here must be non-nil
	return e.client.Call("Plugin.DeletePolicy", args, new(interface{}))
}

func (e *PolicyStoreRPC) Close() error {
	return e.client.Call("Plugin.Close", new(interface{}), nil)
}

type PolicyEnginePlugin struct {
	Impl         IPolicyEngine
	PluginClient *plugin.Client
	RPCClient    plugin.ClientProtocol
}

func (p *PolicyEnginePlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &PolicyEngineServer{Impl: p.Impl}, nil
}

func (p *PolicyEnginePlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &PolicyEngineRPC{client: c}, nil
}

func (p *PolicyEnginePlugin) Load(lp *LoadedPlugin) error {
	p.Impl = lp.Raw.(IPolicyEngine)
	p.RPCClient = lp.RPCClient
	p.PluginClient = lp.PluginClient

	return nil
}

func (p *PolicyEnginePlugin) Init(params *ParamStore) error {
	return p.Impl.Init(params)
}

func (p *PolicyEnginePlugin) Close() error {
	p.PluginClient.Kill()
	return p.RPCClient.Close()
}

func (p *PolicyEnginePlugin) GetName() string {
	return p.Impl.GetName()
}

func (p *PolicyEnginePlugin) Appraise(attestation *Attestation, policy *Policy) error {
	return p.Impl.Appraise(attestation, policy)
}

func LoadPolicyEnginePlugin(locations []string, name string) (*PolicyEnginePlugin, error) {
	lp, err := LoadPlugin(locations, "policyenginge", name, false)
	if err != nil {
		return nil, err
	}

	var policyEnginePlugin PolicyEnginePlugin
	err = policyEnginePlugin.Load(lp)
	return &policyEnginePlugin, err
}

type AppraiseArgs struct {
	Attestation []byte
	Policy      []byte
}

type PolicyEngineServer struct {
	Impl IPolicyEngine
}

func (s *PolicyEngineServer) GetName(args interface{}, resp *string) error {
	*resp = s.Impl.GetName()
	return nil
}

func (s *PolicyEngineServer) Init(params *ParamStore, resp *interface{}) error {
	return s.Impl.Init(params)
}

func (s *PolicyEngineServer) Appraise(args AppraiseArgs, resp *[]byte) error {
	var attestation Attestation
	var policy Policy

	err := json.Unmarshal(args.Attestation, &attestation)
	if err != nil {
		return err
	}

	err = json.Unmarshal(args.Policy, &policy)
	if err != nil {
		return err
	}

	err = s.Impl.Appraise(&attestation, &policy)
	if err != nil {
		return err
	}

	*resp, err = json.Marshal(&attestation)

	return err
}

func (s *PolicyEngineServer) Close(args interface{}, resp *interface{}) error {
	return s.Impl.Close()
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

func (e *PolicyEngineRPC) Init(params *ParamStore) error {
	return e.client.Call("Plugin.Init", params, new(interface{}))
}

func (e *PolicyEngineRPC) Appraise(attestation *Attestation, policy *Policy) error {
	// NOTE: encoding/gob used to serialize objects by net/rpc cannot handle []interface{}, which
	//       necessitates pre-serialing any objects that may contain arbitrary JSON decodings.

	var err error
	args := AppraiseArgs{}

	args.Attestation, err = json.Marshal(attestation)
	if err != nil {
		return err
	}

	args.Policy, err = json.Marshal(policy)
	if err != nil {
		return err
	}

	var resultBlob []byte
	err = e.client.Call("Plugin.Appraise", args, &resultBlob)
	if err != nil {
		return err
	}

	return json.Unmarshal(resultBlob, attestation)
}

func (e *PolicyEngineRPC) Close() error {
	return e.client.Call("Plugin.Fini", new(interface{}), new(interface{}))
}

type PolicyEnginePluginContainer struct {
	Engine   IPolicyEngine
	Client   *plugin.Client
	Protocol plugin.ClientProtocol
}

func (c PolicyEnginePluginContainer) GetName() string {
	return c.Engine.GetName()
}

func (c PolicyEnginePluginContainer) Init(params *ParamStore) error {
	return c.Engine.Init(params)
}

func (c PolicyEnginePluginContainer) Appraise(
	attestation *Attestation,
	policy *Policy,
) error {
	return c.Engine.Appraise(attestation, policy)
}

func (c PolicyEnginePluginContainer) Close() error {
	err := c.Engine.Close()
	c.Client.Kill()
	c.Protocol.Close()
	return err
}

func LoadAndInitializePolicyEnginePlugin(
	params *ParamStore,
) (IPolicyEngine, error) {
	engineName := Canonize(params.GetString("PolicyEngineName"))
	pluginLocations := params.GetStringSlice("PluginLocations")
	quiet := params.GetBool("Quiet")

	pc := new(PolicyEnginePluginContainer)

	lp, err := LoadPlugin(pluginLocations, "policyengine", engineName, quiet)
	if err != nil {
		return nil, err
	}

	pc.Engine = lp.Raw.(IPolicyEngine)
	pc.Client = lp.PluginClient
	pc.Protocol = lp.RPCClient

	if pc.Client == nil {
		return nil, fmt.Errorf("failed to find policy engine with name '%v'", engineName)
	}

	err = pc.Init(params)
	if err != nil {
		pc.Client.Kill()
		return nil, err
	}

	return pc, nil
}
