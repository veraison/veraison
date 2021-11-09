// Copyright 2021 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package endorsement

import (
	"github.com/veraison/common"
)

type IEndorsementConfig interface {
	GetPluginLocations() []string
	GetEndorsementBackendName() string
	GetEndorsementBackendParams() common.EndorsementBackendParams
	GetQuiet() bool
}
