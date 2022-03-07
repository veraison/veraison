// Copyright 2021-2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package trustedservices

import (
	"context"
	"reflect"

	"github.com/veraison/common"
)

func NewRPCServerParamStore() (*common.ParamStore, error) {
	store := common.NewParamStore("vts-http")
	err := PopulateRPCServerParams(store)
	return store, err
}

func PopulateRPCServerParams(store *common.ParamStore) error {
	return store.AddParamDefinitions(map[string]*common.ParamDescription{
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

func (c *RPCServer) Init(ctx context.Context, params *common.ParamStore) (*common.InitResponse, error) {
	return nil, nil
}
func (c *RPCServer) GetAttestation(
	ctx context.Context,
	token *common.AttestationToken,
) (*common.Attestation, error) {
	return c.Client.GetAttestation(token)
}

func (c *RPCServer) Close(ctx context.Context, args *common.CloseArgs) (*common.CloseResponse, error) {
	err := c.Client.Close()
	return nil, err
}
