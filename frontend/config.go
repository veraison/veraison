// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package frontend

import (
	"reflect"

	"github.com/veraison/common"
	"github.com/veraison/policy"
	"github.com/veraison/trustedservices"
	"github.com/veraison/verifier"
)

func NewFrontendParamStore() (*common.ParamStore, error) {
	store := common.NewParamStore("frontend")
	err := store.AddParamDefinitions(map[string]*common.ParamDescription{
		"Debug": {
			Kind:     uint32(reflect.Bool),
			Path:     "debug",
			Required: common.ParamNecessity_OPTIONAL,
		},
	})

	return store, err
}

func NewFrontendConfig() (*common.Config, error) {
	configPaths := common.NewConfigPaths()

	frontendParams, err := NewFrontendParamStore()
	if err != nil {
		return nil, err
	}

	verifierParams, err := verifier.NewVerifierParams()
	if err != nil {
		return nil, err
	}

	policyManagerParams, err := policy.NewManagerParamStore()
	if err != nil {
		return nil, err
	}

	vtsClientParams, err := trustedservices.NewLocalClientParamStore()
	if err != nil {
		return nil, err
	}

	return common.NewConfig(
		configPaths.Strings(),
		frontendParams,
		verifierParams,
		policyManagerParams,
		vtsClientParams,
	)
}
