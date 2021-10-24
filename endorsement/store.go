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

func ResponseFromError(err error) *OpenResponse {
	status := &Status{Result: false, ErrorDetail: err.Error()}
	return &OpenResponse{Status: status}
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

	config IEndorsementConfig
}

func (s *Store) Init(config IEndorsementConfig) error {
	lp, err := common.LoadPlugin(
		config.GetPluginLocations(),
		"endorsementstore",
		config.GetEndorsementBackendName(),
		config.GetQuiet(),
	)
	if err != nil {
		return err
	}

	s.Backend = lp.Raw.(common.IEndorsementBackend)
	s.Client = lp.PluginClient
	s.RPCClient = lp.RPCClient

	if err = s.Backend.Init(config.GetEndorsementBackendParams()); err != nil {
		s.Client.Kill()
		return err
	}

	return nil
}

func (s Store) Fini() {
	s.Client.Kill()
	s.RPCClient.Close()
	s.Backend.Close()
}

func (s Store) GetEndorsements(
	ctx context.Context,
	args *GetEndorsementsRequest,
) (*GetEndorsementsResponse, error) {

	// TODO: this a a HACK to provide a minimal impleentation of the new interface.
	// Query Descriptors should no longer be required here. Additionally, the assempled
	// descriptors are for PSA only....
	if args.Id.Type != common.AttestationFormat_PSA_IOT {
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
