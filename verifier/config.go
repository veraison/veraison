// Copyright 2021 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package verifier

import (
	"veraison/common"
)

type VerifierConfig struct {
	PluginLocations        []string
	PolicyStoreName        string
	PolicyEngineName       string
	EndorsementStoreName   string
	PolicyStoreParams      common.PolicyStoreParams
	PolicyEngineParams     common.PolicyEngineParams
	EndorsementStoreParams common.EndorsementStoreParams
}
