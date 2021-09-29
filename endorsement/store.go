// Copyright 2021 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package endorsement

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-plugin"
	structpb "google.golang.org/protobuf/types/known/structpb"

	"github.com/veraison/common"
)

func ResponseFromError(err error) *Response {
	return &Response{ErrorValue: 1, ErrorDetail: err.Error()}
}

// Manager handles interadctions with the endorsement store. It is responsible
// for loading the appropriate plugin and maintaining the client connection to
// it.
type Store struct {
	UnimplementedStoreServer
	UnimplementedFetcherServer

	// BackendName is the name of the endorsement store plugin that has been
	// loaded, or "[none]" if there isn't one.
	BackendName string

	// Backend is the interface to the underlying store.
	Backend common.IEndorsementBackend

	// RpcClient is an instance of ClientProtocol responsible for handling
	// the underlying RPC channel to the plugin.
	RPCClient plugin.ClientProtocol

	// Client is the client interface to the plugin used for aquring the
	// handle for the interface implemented by the plugin and establishing
	// the RPC channel.
	Client *plugin.Client
}

func (s *Store) Open(ctx context.Context, args *OpenArgs) (*Response, error) {
	lp, err := common.LoadPlugin(args.PluginLocations, "endorsementstore", args.BackendName, args.Quiet)
	if err != nil {
		return ResponseFromError(err), nil
	}

	s.Backend = lp.Raw.(common.IEndorsementBackend)
	s.Client = lp.PluginClient
	s.RPCClient = lp.RPCClient

	if err = s.Backend.Init(args.BackendConfig.AsMap()); err != nil {
		s.Client.Kill()
		return ResponseFromError(err), nil
	}

	return &Response{}, nil
}

func (s Store) Close(ctx context.Context, args *CloseArgs) (*Response, error) {
	s.Client.Kill()
	s.RPCClient.Close()
	s.Backend.Close()

	return &Response{}, nil
}

func (s Store) GetEndorsements(
	ctx context.Context,
	args *GetEndorsementsArgs,
) (*GetEndorsementsResponse, error) {

	// TODO: this a a HACK to provide a minimal impleentation of the new interface.
	// Query Descriptors should no longer be required here. Additionally, the assempled
	// descriptors are for PSA only....
	if args.Id.Type != common.TokenFormat_PSA {
		return nil, fmt.Errorf("format %q not supported", args.Id.Type)
	}

	parts := args.Id.Parts.AsMap()

	qds := []common.QueryDescriptor{
		common.QueryDescriptor{
			Name: "hardware_id",
			Args: common.QueryArgs{
				"platform_id": parts["platform_id"],
			},
		},
		common.QueryDescriptor{
			Name: "software_components",
			Args: common.QueryArgs{
				"platform_id":  parts["platform_id"],
				"measurements": parts["measurements"],
			},
		},
	}
	// end HACK

	backendResponse, err := s.Backend.GetEndorsements(qds...)
	if err != nil {
		return nil, err
	}

	endorsements := make(map[string]interface{})
	for k, v := range backendResponse {
		if len(v) != 1 {
			return nil, fmt.Errorf("no unique match")
		}
		endorsements[k] = v[0]
	}

	endStruct, err := structpb.NewStruct(endorsements)
	if err != nil {
		return nil, err
	}

	response := &GetEndorsementsResponse{
		Endorsements: endStruct,
	}

	return response, nil
}
