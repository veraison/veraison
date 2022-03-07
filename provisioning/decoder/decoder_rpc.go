// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package decoder

import (
	"fmt"
	"net/rpc"

	"google.golang.org/protobuf/encoding/protojson"
)

/*
  Server-side RPC adapter around the Decoder plugin implementation
  (plugin-side)
*/

type RPCServer struct {
	Impl IDecoder
}

func (s *RPCServer) Init(params Params, unused *interface{}) error {
	return s.Impl.Init(params)
}

func (s RPCServer) Close(unused0 interface{}, unused1 *interface{}) error {
	return s.Impl.Close()
}

func (s *RPCServer) GetName(args interface{}, resp *string) error {
	*resp = s.Impl.GetName()
	return nil
}

func (s *RPCServer) GetSupportedMediaTypes(args interface{}, resp *[]string) error {
	*resp = s.Impl.GetSupportedMediaTypes()
	return nil
}

func (s RPCServer) Decode(data []byte, resp *[]byte) error {
	j, err := s.Impl.Decode(data)
	if err != nil {
		return fmt.Errorf("plugin returned error: %w", err)
	}

	*resp, err = protojson.Marshal(j)
	if err != nil {
		return fmt.Errorf("failed to marshal plugin response: %w", err)
	}

	return nil
}

/*
  RPC client
  (plugin caller side)
*/

type RPCClient struct {
	client *rpc.Client
}

func (c RPCClient) Init(params Params) error {
	var unused interface{}

	return c.client.Call("Plugin.Init", params, &unused)
}

func (c RPCClient) Close() error {
	var (
		unused0 interface{}
		unused1 interface{}
	)

	return c.client.Call("Plugin.Close", unused0, unused1)
}

func (c RPCClient) GetName() string {
	var (
		err    error
		resp   string
		unused interface{}
	)

	err = c.client.Call("Plugin.GetName", &unused, &resp)
	if err != nil {
		return ""
	}

	return resp
}

func (c RPCClient) GetSupportedMediaTypes() []string {
	var (
		err    error
		resp   []string
		unused interface{}
	)

	err = c.client.Call("Plugin.GetSupportedMediaTypes", &unused, &resp)
	if err != nil {
		return nil
	}

	return resp
}

func (c RPCClient) Decode(data []byte) (*EndorsementDecoderResponse, error) {
	var (
		err  error
		resp EndorsementDecoderResponse
		j    []byte
	)

	err = c.client.Call("Plugin.Decode", data, &j)
	if err != nil {
		return nil, fmt.Errorf("RPC server returned error: %w", err)
	}

	err = protojson.Unmarshal(j, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed unmarshaling response from RPC server: %w", err)
	}

	return &resp, nil
}
