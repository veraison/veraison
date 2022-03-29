// Copyright 2021-2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package trustedservices

import (
	"context"
	"errors"
	"reflect"

	"github.com/veraison/common"
)

func NewRPCServerParamStore() (*common.ParamStore, error) {
	store := common.NewParamStore("vts-http")
	if store == nil {
		return nil, errors.New("param store initialization failed")
	}
	err := PopulateRPCServerParams(store)
	return store, err
}

func PopulateRPCServerParams(store *common.ParamStore) error {
	return store.AddParamDefinitions(map[string]*common.ParamDescription{
		// NOTE(tho) should we use Path as map key?  (It'd make debug messages clearer)
		"Port": {
			Kind:     uint32(reflect.Int),
			Path:     "vts.port",
			Required: common.ParamNecessity_REQUIRED,
		},
	})
}

type RPCServer struct {
	common.UnimplementedVTSServer
	Client *LocalClient
}

func (o *RPCServer) Init(params *common.ParamStore) error {
	o.Client = &LocalClient{}
	return o.Client.Init(params)
}

func (o *RPCServer) GetAttestation(
	unusedCtx context.Context, token *common.AttestationToken,
) (*common.Attestation, error) {
	return o.Client.GetAttestation(token)
}

func (o *RPCServer) AddSwComponents(
	unusedCtx context.Context, req *common.AddSwComponentsRequest,
) (*common.AddSwComponentsResponse, error) {
	return o.Client.AddSwComponents(req)
}

func (o *RPCServer) AddTrustAnchor(
	unusedCtx context.Context, req *common.AddTrustAnchorRequest,
) (*common.AddTrustAnchorResponse, error) {
	return o.Client.AddTrustAnchor(req)
}
