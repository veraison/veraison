// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package storeclient

import (
	"context"
	"log"

	"github.com/veraison/endorsement"
	"google.golang.org/grpc"
)

type Dummy struct{}

func NewDummy() IStoreClient {
	return &Dummy{}
}

func (o *Dummy) AddSwComponents(
	ctx context.Context,
	in *endorsement.AddSwComponentsRequest,
	opts ...grpc.CallOption,
) (*endorsement.AddSwComponentsResponse, error) {
	log.Println("><> AddSwComponents()")
	// always return a successful response
	return &endorsement.AddSwComponentsResponse{
		Status: &endorsement.Status{
			Result: true,
		},
	}, nil
}

func (o *Dummy) AddTrustAnchor(
	ctx context.Context,
	in *endorsement.AddTrustAnchorRequest,
	opts ...grpc.CallOption,
) (*endorsement.AddTrustAnchorResponse, error) {
	log.Println("><> AddTrustAnchor()")
	// always return a successful response
	return &endorsement.AddTrustAnchorResponse{
		Status: &endorsement.Status{
			Result: true,
		},
	}, nil
}

func (o *Dummy) GetTrustAnchor(
	ctx context.Context,
	in *endorsement.GetTrustAnchorRequest,
	opts ...grpc.CallOption,
) (*endorsement.GetTrustAnchorResponse, error) {
	return &endorsement.GetTrustAnchorResponse{
		Status: &endorsement.Status{
			Result: true,
		},
	}, nil
}

func (o *Dummy) GetSwComponent(
	ctx context.Context,
	in *endorsement.GetSwComponentRequest,
	opts ...grpc.CallOption,
) (*endorsement.GetSwComponentResponse, error) {
	// always return a successful response
	return &endorsement.GetSwComponentResponse{
		Status: &endorsement.Status{
			Result: true,
		},
	}, nil
}

func (o *Dummy) Open(
	ctx context.Context,
	in *endorsement.OpenRequest,
	opts ...grpc.CallOption,
) (*endorsement.OpenResponse, error) {
	// always return a successful response
	return &endorsement.OpenResponse{
		Status: &endorsement.Status{
			Result: true,
		},
	}, nil
}
func (o *Dummy) Close(
	ctx context.Context,
	in *endorsement.CloseRequest,
	opts ...grpc.CallOption,
) (*endorsement.CloseResponse, error) {
	// always return a successful response
	return &endorsement.CloseResponse{
		Status: &endorsement.Status{
			Result: true,
		},
	}, nil
}
