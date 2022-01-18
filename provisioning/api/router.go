package api

import (
	"github.com/gin-gonic/gin"
)

func NewRouter(handler IHandler) *gin.Engine {
	router := gin.New()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	v1 := router.Group("/endorsement-provisioning/v1")
	{
		v1.POST("/submit", handler.Submit)
	}

	return router
}