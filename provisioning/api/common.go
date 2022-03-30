// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package api

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/moogar0880/problems"
)

func ReportProblem(c *gin.Context, status int, details ...string) {
	prob := problems.NewStatusProblem(status)

	if len(details) > 0 {
		prob.Detail = strings.Join(details, ", ")
	}

	c.Header("Content-Type", "application/problem+json")
	c.AbortWithStatusJSON(status, prob)
}
