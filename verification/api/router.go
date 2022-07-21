// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package api

import (
	"github.com/gin-gonic/gin"
)

func NewRouter(handler IHandler) *gin.Engine {
	router := gin.New()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	router.Group("/challenge-response/v1").
		POST("/newSession", handler.NewChallengeResponse).
		POST("/session/:id", handler.SubmitEvidence).
		GET("/session/:id", handler.GetSession)

	return router
}
