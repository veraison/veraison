// Copyright 2021-2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package verifier

import (
	"reflect"

	"go.uber.org/zap"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"

	"github.com/veraison/common"
)

func NewVerifierParams() (*common.ParamStore, error) {
	store := common.NewParamStore("verifier")
	err := store.AddParamDefinitions(map[string]*common.ParamDescription{
		"pluginLocations": {
			Kind:     uint32(reflect.String),
			Path:     "plugin.locations",
			Required: common.ParamNecessity_REQUIRED,
		},
		"policyEngineName": {
			Kind:     uint32(reflect.String),
			Path:     "policy.engine_name",
			Required: common.ParamNecessity_REQUIRED,
		},
		"policyStoreName": {
			Kind:     uint32(reflect.String),
			Path:     "policy.store_name",
			Required: common.ParamNecessity_REQUIRED,
		},
		"policyStoreParams": {
			Kind:     uint32(reflect.Map),
			Path:     "policy.store_name",
			Required: common.ParamNecessity_OPTIONAL,
		},
		"vtsHost": {
			Kind:     uint32(reflect.String),
			Path:     "vts.host",
			Required: common.ParamNecessity_OPTIONAL,
		},
		"vtsPort": {
			Kind:     uint32(reflect.Int),
			Path:     "vts.port",
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
	conn common.ITrustedServicesConnector,
	pm common.IPolicyManager,
	pe common.IPolicyEngine,
) error {
	v.config = config
	v.pm = pm
	v.pe = pe

	var err error
	v.vts, err = conn.Connect(
		v.config.GetString("VtsHost"),
		v.config.GetInt("VtsPort"),
		v.config.GetStringMapString("VtsParams"),
	)
	if err != nil {
		return err
	}

	return nil
}

func (v *Verifier) Close() error {
	return v.vts.Close()
}

func (v *Verifier) Verify(
	token *common.AttestationToken,
) (*common.AttestationResult, error) {
	policy, err := v.pm.GetPolicy(int(token.TenantId), token.Format)
	if err != nil {
		return nil, err
	}

	attestation, err := v.vts.GetAttestation(token)
	if err != nil {
		return nil, err
	}

	attestation.Result.RawEvidence = token.Data
	attestation.Result.Timestamp = timestamppb.Now()

	err = v.pe.Appraise(attestation, policy)
	if err != nil {
		return nil, err
	}

	return attestation.Result, nil
}
