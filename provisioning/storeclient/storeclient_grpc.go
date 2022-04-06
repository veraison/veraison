// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package storeclient

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/veraison/common"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	ErrNotImplemented = errors.New("not implemented")
	ErrNoClient       = errors.New("there is no active gRPC store client")
)

type GRPC struct {
	Config     GRPCConfig
	Connection *grpc.ClientConn
}

func (o *GRPC) getProvisionerClient() common.VTSClient {
	if o.Connection == nil {
		return nil
	}

	return common.NewVTSClient(o.Connection)
}

// Supported parameters:
// * store-server.addr: string w/ syntax specified in
//   https://github.com/grpc/grpc/blob/master/doc/naming.md
//
// * TODO(tho) load balancing config
//   See https://github.com/grpc/grpc/blob/master/doc/load-balancing.md
//
// * TODO(tho) auth'n credentials (e.g., TLS / JWT credentials)
type GRPCConfig map[string]interface{}

const (
	DefaultStoreServerAddr = "dns:localhost:12345"
)

func (o *GRPCConfig) GetString(k, dflt string) (string, error) {
	v, ok := (*o)[k]
	if !ok {
		return dflt, nil
	}

	s, ok := v.(string)
	if !ok {
		return "", fmt.Errorf("value for %s is not of type string (%T)", k, v)
	}

	return s, nil
}

// NewGRPC instantiate a new gRPC store client with the supplied configuration
func NewGRPC(c GRPCConfig) IStoreClient {
	return &GRPC{Config: c}
}

// ensureConnection makes sure the underlying gRPC connection is usable
func (o *GRPC) ensureConnection() error {
	if o.Connection != nil {
		return nil
	}

	// TODO(tho) check if it's worth addressing this lint warning:
	// "grpc.WithTimeout is deprecated: use DialContext instead of Dial and
	//  context.WithTimeout instead.  Will be supported throughout 1.x."
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithTimeout(time.Second), // nolint: staticcheck
	}

	storeServerAddr, err := o.Config.GetString("store-server.addr", DefaultStoreServerAddr)
	if err != nil {
		return fmt.Errorf("configuration error: %w", err)
	}

	conn, err := grpc.Dial(storeServerAddr, opts...)
	if err != nil {
		return fmt.Errorf("connection to gRPC store server %s failed: %w", storeServerAddr, err)
	}

	o.Connection = conn

	return nil
}

func (o *GRPC) AddSwComponents(ctx context.Context, in *common.AddSwComponentsRequest, opts ...grpc.CallOption,
) (*common.AddSwComponentsResponse, error) {
	if err := o.ensureConnection(); err != nil {
		return nil, fmt.Errorf("failed AddSwComponents: %w", err)
	}

	c := o.getProvisionerClient()
	if c == nil {
		return nil, ErrNoClient
	}

	return c.AddSwComponents(ctx, in, opts...)
}

func (o *GRPC) AddTrustAnchor(ctx context.Context, in *common.AddTrustAnchorRequest, opts ...grpc.CallOption,
) (*common.AddTrustAnchorResponse, error) {
	if err := o.ensureConnection(); err != nil {
		return nil, fmt.Errorf("failed AddTrustAnchor: %w", err)
	}

	c := o.getProvisionerClient()
	if c == nil {
		return nil, ErrNoClient
	}

	return c.AddTrustAnchor(ctx, in, opts...)
}

func (o *GRPC) GetAttestation(
	ctx context.Context, in *common.AttestationToken, opts ...grpc.CallOption,
) (*common.Attestation, error) {
	return nil, ErrNotImplemented
}
