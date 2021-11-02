// Copyright 2021 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package verifier

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-plugin"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/veraison/common"
	"github.com/veraison/endorsement"
	"github.com/veraison/policy"
)

type Verifier struct {
	pm        *policy.Manager
	pe        common.IPolicyEngine
	rpcClient plugin.ClientProtocol
	client    *plugin.Client

	esConn *grpc.ClientConn
	es     endorsement.StoreClient
	ef     endorsement.FetcherClient

	logger *zap.Logger
}

func NewVerifier(logger *zap.Logger) (*Verifier, error) {
	v := new(Verifier)

	v.logger = logger
	v.pm = policy.NewManager()

	return v, nil
}

// Initialize bootstraps the verifier
func (v *Verifier) Initialize(vc Config) error {
	storeAddress := fmt.Sprintf("%s:%d", vc.EndorsementStoreHost, vc.EndorsementStorePort)
	esConn, err := grpc.Dial(storeAddress, grpc.WithInsecure())
	if err != nil {
		return err
	}

	endorsementStoreClient := endorsement.NewStoreClient(esConn)

	backendConfig, err := structpb.NewStruct(vc.EndorsementBackendParams)
	if err != nil {
		return err
	}

	openArgs := &endorsement.OpenRequest{}

	response, err := endorsementStoreClient.Open(context.Background(), openArgs)
	if err != nil {
		esConn.Close()
		return err
	}
	if !response.Status.Result {
		return fmt.Errorf(
			"could not connect to endorsement store; got: %q",
			response.Status.ErrorDetail,
		)
	}

	endorsementFetcherClient := endorsement.NewFetcherClient(esConn)

	if err = v.pm.InitializeStore(
		vc.PluginLocations,
		vc.PolicyStoreName,
		vc.PolicyStoreParams,
		false,
	); err != nil {
		esConn.Close()
		return err
	}

	pe, client, rpcClient, err := common.LoadAndInitializePolicyEngine(
		vc.PluginLocations,
		vc.PolicyEngineName,
		vc.PolicyEngineParams,
		false,
	)
	if err != nil {
		esConn.Close()
		return err
	}

	v.pe = pe
	v.client = client
	v.rpcClient = rpcClient

	v.esConn = esConn
	v.es = endorsementStoreClient
	v.ef = endorsementFetcherClient

	return nil
}

// Verify verifies the supplied Evidence
func (v *Verifier) Verify(ec *common.EvidenceContext, simple bool) (*common.AttestationResult, error) {
	v.logger.Debug("verify params", zap.Reflect("evidence context", ec), zap.Bool("simple", simple))
	policy, err := v.pm.GetPolicy(ec.TenantID, ec.Format)
	if err != nil {
		return nil, err
	}

	if err = v.pe.LoadPolicy(policy.Rules); err != nil {
		return nil, err
	}

	// TODO: this a a HACK to provide a minimal impleentation of the new interface.
	// Query Descriptors should no longer be required here
	qds, err := policy.GetQueryDesriptors(ec.Evidence, common.QcNone)
	if err != nil {
		return nil, err
	}

	parts := make(map[string]interface{})
	for _, qd := range qds {
		for key, val := range qd.Args {
			parts[key] = val
		}
	}

	partsStruct, err := structpb.NewStruct(parts)
	if err != nil {
		return nil, err
	}
	// end of HACK

	evStruct, err := structpb.NewStruct(ec.Evidence)
	if err != nil {
		return nil, err
	}

	args := &endorsement.GetEndorsementsRequest{
		Id: &endorsement.EndorsementID{
			Type:  ec.Format,
			Parts: partsStruct,
		},
		Evidence: &endorsement.Evidence{Value: evStruct},
	}

	// TODO: make the timeout configurable
	//ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Second))
	//defer cancel()

	response, err := v.ef.GetEndorsements(context.Background(), args)
	if err != nil {
		return nil, err
	}
	if !response.Status.Result {
		return nil, fmt.Errorf(
			"could not get endorsements; got: %q",
			response.Status.ErrorDetail,
		)
	}

	endorsements := response.Endorsements.AsMap()
	result := new(common.AttestationResult)

	v.logger.Debug("fetched endorsements", zap.Reflect("endorsements", endorsements))
	v.logger.Debug("extracted evidence", zap.Reflect("evidence", ec.Evidence))

	if err := v.pe.GetAttetationResult(ec.Evidence, endorsements, simple, result); err != nil {
		return nil, err
	}

	v.logger.Debug("attestation result", zap.Reflect("result", result))
	return result, nil
}

func (v *Verifier) Close() {
	v.esConn.Close()
	v.pm.Close()
	v.client.Kill()
	v.rpcClient.Close()
}
