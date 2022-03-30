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

type EndorsementAddArg struct {
	Name   string
	Args   QueryArgs
	Update bool
}

// EndorsementBackendPlugin defines the plugin interface for an IEndorsementBackend implementation.
type EndorsementBackendPlugin struct {

	// Impl is the underlying IEndorsementBackend implementation.
	Impl IEndorsementBackend
}

// Server Returns the EndorsementBackendServer instance used by the plugin
// process to provide access to IEndorsementBackend methods.
func (p *EndorsementBackendPlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &EndorsementBackendServer{Impl: p.Impl}, nil
}

// Client returns the RPC client instance used to interact with a plugin
// running an EndorsementBackendServer via an RPC channel.
func (p *EndorsementBackendPlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &EndorsementBackendRPC{client: c}, nil
}

// EndorsementBackendServer provides plugin-side IEndorsementBackend implementation.
type EndorsementBackendServer struct {
	Impl IEndorsementBackend
}

func (s *EndorsementBackendServer) GetName(ags interface{}, resp *string) error {
	*resp = s.Impl.GetName()
	return nil
}

func (s *EndorsementBackendServer) Init(args EndorsementBackendParams, resp *string) error {
	return s.Impl.Init(args)
}

func (s *EndorsementBackendServer) GetEndorsements(qds []QueryDescriptor, resp *EndorsementMatches) error {
	var err error
	*resp, err = s.Impl.GetEndorsements(qds...)
	return err
}

func (s *EndorsementBackendServer) RunQuery(qdBlob []byte, resp *[]byte) error {
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

func (s *EndorsementBackendServer) AddEndorsement(argBlob []byte, resp *[]byte) error {
	var arg EndorsementAddArg
	// NOTE: encoding/gob used to serialize objects by net/rpc cannot handle []interface{}, which
	//       necessitates pre-serialing any objects that may contain arbitrary JSON decodings.
	err := json.Unmarshal(argBlob, &arg)
	if err != nil {
		return err
	}

	return s.Impl.AddEndorsement(arg.Name, arg.Args, arg.Update)
}

func (s *EndorsementBackendServer) GetSupportedQueries(args interface{}, resp *[]string) error {
	*resp = s.Impl.GetSupportedQueries()
	return nil
}

func (s *EndorsementBackendServer) SupportsQuery(name string, resp *bool) error {
	*resp = s.Impl.SupportsQuery(name)
	return nil
}

func (s *EndorsementBackendServer) Close(args interface{}, resp *interface{}) error {
	return s.Impl.Close()
}

type EndorsementBackendRPC struct {
	client *rpc.Client
}

func (e *EndorsementBackendRPC) GetName() string {
	var resp string
	err := e.client.Call("Plugin.GetName", new(interface{}), &resp)
	if err != nil {
		log.Printf("ERROR during GetName RPC: %v\n", err)
		return ""
	}
	return resp
}

func (e *EndorsementBackendRPC) Init(args EndorsementBackendParams) error {
	return e.client.Call("Plugin.Init", args, nil)
}

func (e *EndorsementBackendRPC) GetSupportedQueries() []string {
	var resp []string

	err := e.client.Call("Plugin.GetSupportedQueries", new(interface{}), &resp)
	if err != nil {
		log.Printf("ERROR during GetSupportedQueries RPC: %v\n", err)
		return nil
	}

	return resp
}

func (e *EndorsementBackendRPC) SupportsQuery(query string) bool {
	var resp bool

	err := e.client.Call("Plugin.SupportsQuery", query, &resp)
	if err != nil {
		log.Printf("ERROR during SupportsQuery RPC: %v\n", err)
	}

	return resp
}

func (e *EndorsementBackendRPC) Close() error {
	return e.client.Call("Plugin.Close", new(interface{}), nil)
}

func (e *EndorsementBackendRPC) RunQuery(name string, args QueryArgs) (QueryResult, error) {
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

func (e *EndorsementBackendRPC) AddEndorsement(name string, args QueryArgs, update bool) error {
	arg := EndorsementAddArg{Name: name, Args: args, Update: update}

	// NOTE: encoding/gob used to serialize objects by net/rpc cannot handle []interface{}, which
	//       necessitates pre-serialing any objects that may contain arbitrary JSON decodings.
	argBlob, err := json.Marshal(arg)
	if err != nil {
		return err
	}

	var resBlob []byte

	return e.client.Call("Plugin.AddEndorsement", argBlob, &resBlob)
}

func (e *EndorsementBackendRPC) GetEndorsements(qds ...QueryDescriptor) (EndorsementMatches, error) {
	matches := make(EndorsementMatches)

	for _, qd := range qds {
		qr, err := e.RunQuery(qd.Name, qd.Args)
		if err != nil {
			return nil, err
		}

		if !checkConstraintHolds(qr, qd.Constraint) {
			return nil, fmt.Errorf("result for query '%v' failed to satisfy constraint", qd.Name)
		}

		matches[qd.Name] = qr
	}

	return matches, nil
}
