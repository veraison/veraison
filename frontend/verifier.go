// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package frontend

import (
	"fmt"

	"github.com/veraison/common"
	"github.com/veraison/policy"
	"github.com/veraison/trustedservices"
	"github.com/veraison/verifier"

	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
)

func CreateVerifier(
	config *common.Config,
	pluginManager *common.PluginManager,
	logger *zap.Logger,
) (*verifier.Verifier, error) {
	pmParams, err := config.GetParamStore("policy_manager")
	if err != nil {
		return nil, err
	}

	pm := policy.NewManager(pluginManager)
	if err = pm.Init(pmParams); err != nil {
		return nil, err
	}

	loaded, err := pluginManager.Load("policyengine", "opa")
	if err != nil {
		return nil, err
	}

	pe, ok := loaded.(common.IPolicyEngine)
	if !ok {
		return nil, fmt.Errorf("could not get IPolicyEngine from plugin")
	}

	v, err := verifier.NewVerifier(logger)
	if err != nil {
		return nil, err
	}

	vtsParams, err := config.GetParamStore("vts-local")
	if err != nil {
		return nil, err
	}

	vts := trustedservices.NewLocalClient(pluginManager)
	if err = vts.Init(vtsParams); err != nil {
		return nil, err
	}

	verifierParams, err := config.GetParamStore("verifier")
	if err != nil {

		return nil, err
	}

	err = v.Init(verifierParams, vts, pm, pe)
	return v, err
}
