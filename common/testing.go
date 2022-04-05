// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package common

// This file contains some utilities to help with plugin testing.

import (
	"log"
	"net/rpc"

	"github.com/hashicorp/go-plugin"
)

var TestPluginSet = plugin.PluginSet{
	"test": &TestPlugin{},
}

type ITestPluginInterface interface {
	GetName() string
}

type TestPlugin struct {
	Impl ITestPluginInterface
}

func (p *TestPlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &TestPluginServer{Impl: p.Impl}, nil
}

func (p *TestPlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &TestPluginClient{RPCClient: c}, nil
}

type TestPluginServer struct {
	Impl ITestPluginInterface
}

func (s *TestPluginServer) GetName(args interface{}, resp *string) error {
	*resp = s.Impl.GetName()
	return nil
}

type TestPluginClient struct {
	RPCClient *rpc.Client
}

func (c TestPluginClient) GetName() string {
	var resp string

	if err := c.RPCClient.Call("Plugin.GetName", new(interface{}), &resp); err != nil {
		log.Printf("ERROR during GetName RPC: %v\n", err)
		return ""
	}

	return resp
}
