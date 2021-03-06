// Copyright 2021 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package verifier

import (
	"fmt"

	"github.com/hashicorp/go-plugin"

	"github.com/veraison/common"
	"github.com/veraison/endorsement"
	"github.com/veraison/policy"

	"go.uber.org/zap"
)

type Verifier struct {
	pm        *policy.Manager
	em        *endorsement.Manager
	pe        common.IPolicyEngine
	rpcClient plugin.ClientProtocol
	client    *plugin.Client
	logger    *zap.Logger
}

func NewVerifier(logger *zap.Logger) (*Verifier, error) {
	v := new(Verifier)

	v.logger = logger
	v.pm = policy.NewManager()
	v.em = endorsement.NewManager()

	return v, nil
}

// Initialize bootstraps the verifier
func (v *Verifier) Initialize(vc Config) error {
	if err := v.em.InitializeStore(
		vc.PluginLocations,
		vc.EndorsementStoreName,
		vc.EndorsementStoreParams,
		false,
	); err != nil {
		return err
	}

	if err := v.pm.InitializeStore(
		vc.PluginLocations,
		vc.PolicyStoreName,
		vc.PolicyStoreParams,
		false,
	); err != nil {
		return err
	}

	pe, client, rpcClient, err := common.LoadAndInitializePolicyEngine(
		vc.PluginLocations,
		vc.PolicyEngineName,
		vc.PolicyEngineParams,
		false,
	)
	if err != nil {
		return err
	}

	v.pe = pe
	v.client = client
	v.rpcClient = rpcClient

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

	qds, err := policy.GetQueryDesriptors(ec.Evidence, common.QcNone)
	if err != nil {
		return nil, err
	}

	matches, err := v.em.GetEndorsements(qds...)
	if err != nil {
		return nil, err
	}

	endorsements := make(map[string]interface{})
	for name, qr := range matches {
		if len(qr) == 1 {
			endorsements[name] = qr[0]
		} else if len(qr) == 0 {
			return nil, fmt.Errorf("no matches for '%v'", name)
		} else {
			return nil, fmt.Errorf("too many matches for '%v'", name)
		}
	}

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
	v.em.Close()
	v.pm.Close()
	v.client.Kill()
	v.rpcClient.Close()
}
