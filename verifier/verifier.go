// Copyright 2021-2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package verifier

import (
	"fmt"
	"reflect"

	"go.uber.org/zap"

	"github.com/veraison/common"
)

func NewVerifierParams() (*common.ParamStore, error) {
	store := common.NewParamStore("verifier")
	err := store.AddParamDefinitions(map[string]*common.ParamDescription{
		"VtsClientType": {
			Kind:     uint32(reflect.Map),
			Path:     "policy.store_name",
			Required: common.ParamNecessity_OPTIONAL,
		},
	})
	if err != nil {
		return nil, err
	}
	store.Freeze()

	return store, nil
}

func NewVerifier(logger *zap.Logger) (*Verifier, error) {
	v := new(Verifier)

	v.logger = logger

	return v, nil
}

type Verifier struct {
	config *common.ParamStore
	vts    common.ITrustedServicesClient
	pm     common.IPolicyManager
	pe     common.IPolicyEngine
	logger *zap.Logger
}

func (v *Verifier) Init(
	config *common.ParamStore,
	vts common.ITrustedServicesClient,
	pm common.IPolicyManager,
	pe common.IPolicyEngine,
) error {
	v.config = config
	v.pm = pm
	v.pe = pe
	v.vts = vts

	return nil
}

func (v *Verifier) Close() error {
	return v.vts.Close()
}

func (v *Verifier) Verify(
	token *common.AttestationToken,
) (*common.AttestationResult, error) {
	policy, err := v.pm.GetPolicy(int(token.TenantId), token.Format)
	if err != nil && err != common.ErrPolicyNotFound {
		return nil, fmt.Errorf("could not obtain policy: %v", err)
	}

	attestation, err := v.vts.GetAttestation(token)
	if err != nil {
		return nil, err
	}

	attestation.Result.RawEvidence = token.Data

	if policy != nil {
		err = v.pe.Appraise(attestation, policy)
		if err != nil {
			return nil, err
		}
	}

	return attestation.Result, nil
}
