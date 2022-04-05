// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package frontend

import (
	"github.com/veraison/common"
)

type ResponseEvidence struct {
	Type  string `json:"type" binding:"required"`
	Value []byte `json:"value" binding:"required"`
}

type ResponseBody struct {
	SessionInfo
	Evidence ResponseEvidence          `json:"evidence" binding:"required"`
	Result   *common.AttestationResult `json:"result" binding:"required"`
}
