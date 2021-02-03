// Copyright 2021 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"encoding/json"
	"log"
	"net/rpc"

	"github.com/hashicorp/go-plugin"
)

type TrustAnchorStorePlugin struct {
	Impl ITrustAnchorStore
}

func (p TrustAnchorStorePlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &TrustAnchorStoreServer{Impl: p.Impl}, nil
}

func (p TrustAnchorStorePlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &TrustAnchorStoreRPC{client: c}, nil
}

type TrustAnchorStoreServer struct {
	Impl ITrustAnchorStore
}

func (s TrustAnchorStoreServer) GetName(args interface{}, resp *string) error {
	*resp = s.Impl.GetName()
	return nil
}

func (s TrustAnchorStoreServer) Init(params TrustAnchorStoreParams, resp *string) error {
	return s.Impl.Init(params)
}

func (s TrustAnchorStoreServer) Close(args interface{}, resp *interface{}) error {
	return s.Impl.Close()
}

type GetTrustAnchorArgs struct {
	TenantID      int
	TrustAnchorID TrustAnchorID
}

func (s TrustAnchorStoreServer) GetTrustAnchor(argsData []byte, resp *[]byte) error {
	var args GetTrustAnchorArgs

	err := json.Unmarshal(argsData, &args)
	if err != nil {
		return err
	}

	*resp, err = s.Impl.GetTrustAnchor(args.TenantID, args.TrustAnchorID)
	return err
}

type AddCertsFromPEMArgs struct {
	TenantID int
	Value    []byte
}

type AddPublicKeyFromPEMArgs struct {
	TenantID int
	KeyID    interface{}
	Value    []byte
}

func (s TrustAnchorStoreServer) AddCertsFromPEM(argsData []byte, resp *[]byte) error {
	var args AddCertsFromPEMArgs

	err := json.Unmarshal(argsData, &args)
	if err != nil {
		return err
	}

	return s.Impl.AddCertsFromPEM(args.TenantID, args.Value)
}

func (s TrustAnchorStoreServer) AddPublicKeyFromPEM(argsData []byte, resp *[]byte) error {
	var args AddPublicKeyFromPEMArgs

	err := json.Unmarshal(argsData, &args)
	if err != nil {
		return err
	}

	return s.Impl.AddPublicKeyFromPEM(args.TenantID, args.KeyID, args.Value)
}

type TrustAnchorStoreRPC struct {
	client *rpc.Client
}

func (r TrustAnchorStoreRPC) GetName() string {
	var resp string
	err := r.client.Call("Plugin.GetName", new(interface{}), &resp)
	if err != nil {
		log.Printf("ERROR during GetName RPC: %v\n", err)
		return ""
	}
	return resp
}

func (r TrustAnchorStoreRPC) Init(params TrustAnchorStoreParams) error {
	return r.client.Call("Plugin.Init", params, nil)
}

func (r TrustAnchorStoreRPC) Close() error {
	return r.client.Call("Plugin.Close", new(interface{}), nil)
}

func (r TrustAnchorStoreRPC) GetTrustAnchor(tenantId int, taId TrustAnchorID) ([]byte, error) {
	args := GetTrustAnchorArgs{TenantID: tenantId, TrustAnchorID: taId}
	argsData, err := json.Marshal(args)
	if err != nil {
		return nil, err
	}

	var result []byte

	err = r.client.Call("Plugin.GetTrustAnchor", argsData, &result)
	return result, err
}

func (r TrustAnchorStoreRPC) AddCertsFromPEM(tenantId int, value []byte) error {
	taId := TrustAnchorID{Type: TaTypeCert}
	args := GetTrustAnchorArgs{TenantID: tenantId, TrustAnchorID: taId}
	argsData, err := json.Marshal(args)
	if err != nil {
		return err
	}

	return r.client.Call("Plugin.AddCertsFromPEM", argsData, nil)
}

func (r TrustAnchorStoreRPC) AddPublicKeyFromPEM(tenantId int, kid interface{}, value []byte) error {
	args := AddPublicKeyFromPEMArgs{TenantID: tenantId, KeyID: kid, Value: value}

	argsData, err := json.Marshal(args)
	if err != nil {
		return err
	}

	return r.client.Call("Plugin.AddPublicKeyFromPEM", argsData, nil)
}
