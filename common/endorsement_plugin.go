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

// EndorsementStorePlugin defines the plugin interface for an IEndorsementStore implementation.
type EndorsementStorePlugin struct {

	// Impl is the underlying IEndorsementStore implementation.
	Impl IEndorsementStore
}

// Server Returns the EndorsementStoreServer instance used by the plugin
// process to provide access to IEndorsementStore methods.
func (p *EndorsementStorePlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &EndorsementStoreServer{Impl: p.Impl}, nil
}

// Client returns the RPC client instance used to interact with a plugin
// running an EndorsementStoreServer via an RPC channel.
func (p *EndorsementStorePlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &EndorsementStoreRPC{client: c}, nil
}

// EndorsementStoreServer provides plugin-side IEndorsementStore implementation.
type EndorsementStoreServer struct {
	Impl IEndorsementStore
}

func (s *EndorsementStoreServer) GetName(ags interface{}, resp *string) error {
	*resp = s.Impl.GetName()
	return nil
}

func (s *EndorsementStoreServer) Init(args EndorsementStoreParams, resp *string) error {
	return s.Impl.Init(args)
}

func (s *EndorsementStoreServer) GetEndorsements(qds []QueryDescriptor, resp *EndorsementMatches) error {
	var err error
	*resp, err = s.Impl.GetEndorsements(qds...)
	return err
}

func (s *EndorsementStoreServer) RunQuery(qdBlob []byte, resp *[]byte) error {
	var qd QueryDescriptor
	// NOTE: encoding/gob used to serialize objects by net/rpc cannot handle []interface{}, which
	//       necessitates pre-serialing any objects that may contain arbitrary JSON decodings.
	err := json.Unmarshal(qdBlob, &qd)
	if err != nil {
		return err
	}

	qresult, err := s.Impl.RunQuery(qd.Name, qd.Args)
	if err != nil {
		return err
	}

	*resp, err = json.Marshal(qresult)
	return err
}
func (s *EndorsementStoreServer) GetSupportedQueries(args interface{}, resp *[]string) error {
	*resp = s.Impl.GetSupportedQueries()
	return nil
}

func (s *EndorsementStoreServer) SupportsQuery(name string, resp *bool) error {
	*resp = s.Impl.SupportsQuery(name)
	return nil
}

func (s *EndorsementStoreServer) Close(args interface{}, resp *interface{}) error {
	return s.Impl.Close()
}

type EndorsementStoreRPC struct {
	client *rpc.Client
}

func (e *EndorsementStoreRPC) GetName() string {
	var resp string
	err := e.client.Call("Plugin.GetName", new(interface{}), &resp)
	if err != nil {
		log.Printf("ERROR during GetName RPC: %v\n", err)
		return ""
	}
	return resp
}

func (e *EndorsementStoreRPC) Init(args EndorsementStoreParams) error {
	return e.client.Call("Plugin.Init", args, nil)
}

func (e *EndorsementStoreRPC) GetSupportedQueries() []string {
	var resp []string

	err := e.client.Call("Plugin.GetSupportedQueries", new(interface{}), &resp)
	if err != nil {
		log.Printf("ERROR during GetSupportedQueries RPC: %v\n", err)
		return nil
	}

	return resp
}

func (e *EndorsementStoreRPC) SupportsQuery(query string) bool {
	var resp bool

	err := e.client.Call("Plugin.SupportsQuery", query, &resp)
	if err != nil {
		log.Printf("ERROR during SupportsQuery RPC: %v\n", err)
	}

	return resp
}

func (e *EndorsementStoreRPC) Close() error {
	return e.client.Call("Plugin.Close", new(interface{}), nil)
}

func (e *EndorsementStoreRPC) RunQuery(name string, args QueryArgs) (QueryResult, error) {
	var resp QueryResult
	qd := QueryDescriptor{Name: name, Args: args}

	// NOTE: encoding/gob used to serialize objects by net/rpc cannot handle []interface{}, which
	//       necessitates pre-serialing any objects that may contain arbitrary JSON decodings.
	qdBlob, err := json.Marshal(qd)
	if err != nil {
		return nil, err
	}

	var resBlob []byte

	err = e.client.Call("Plugin.RunQuery", qdBlob, &resBlob)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(resBlob, &resp)
	return resp, err
}

func (e *EndorsementStoreRPC) GetEndorsements(qds ...QueryDescriptor) (EndorsementMatches, error) {
	matches := make(EndorsementMatches)

	for _, qd := range qds {
		qr, err := e.RunQuery(qd.Name, qd.Args)
		if err != nil {
			return nil, err
		}

		if !checkConstraintHolds(qr, qd.Constraint) {
			return nil, fmt.Errorf("Result for query '%v' failed to satisfy constraint", qd.Name)
		}

		matches[qd.Name] = qr
	}

	return matches, nil
}
