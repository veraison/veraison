// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package storeclient

import (
	"context"
	"errors"
	"log"

	"github.com/veraison/common"
	"google.golang.org/grpc"
)

type Dummy struct{}

func NewDummy() IStoreClient {
	return &Dummy{}
}

func (o *Dummy) AddSwComponents(
	ctx context.Context,
	in *common.AddSwComponentsRequest,
	opts ...grpc.CallOption,
) (*common.AddSwComponentsResponse, error) {
	log.Println("><> AddSwComponents()")
	// always return a successful response
	return &common.AddSwComponentsResponse{
		Status: &common.Status{
			Result: true,
		},
	}, nil
}

func (o *Dummy) AddTrustAnchor(
	ctx context.Context,
	in *common.AddTrustAnchorRequest,
	opts ...grpc.CallOption,
) (*common.AddTrustAnchorResponse, error) {
	log.Println("><> AddTrustAnchor()")
	// always return a successful response
	return &common.AddTrustAnchorResponse{
		Status: &common.Status{
			Result: true,
		},
	}, nil
}

func (o *Dummy) GetTrustAnchor(
	ctx context.Context,
	in *common.GetTrustAnchorRequest,
	opts ...grpc.CallOption,
) (*common.GetTrustAnchorResponse, error) {
	return &common.GetTrustAnchorResponse{
		Status: &common.Status{
			Result: true,
		},
	}, nil
}

func (o *Dummy) GetSwComponent(
	ctx context.Context,
	in *common.GetSwComponentRequest,
	opts ...grpc.CallOption,
) (*common.GetSwComponentResponse, error) {
	// always return a successful response
	return &common.GetSwComponentResponse{
		Status: &common.Status{
			Result: true,
		},
	}, nil
}

func (o *Dummy) GetAttestation(
	ctx context.Context,
	in *common.AttestationToken,
	opts ...grpc.CallOption,
) (*common.Attestation, error) {
	return nil, errors.New("not implemented")
}
