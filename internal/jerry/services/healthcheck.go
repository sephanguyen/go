package services

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterHealthCheckService(ge *gin.Engine) {
	ge.GET("/healthz", func(ctx *gin.Context) {
		ctx.Status(http.StatusOK)
	})
}
