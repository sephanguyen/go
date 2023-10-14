package middlewares

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var (
	errorNotAValidRequestFromProvider = "Not a valid request from provider"
)

type Authenticator interface {
	AuthenticateHTTPRequest(header http.Header, payload []byte) (bool, error)
}

func AuthenticateWebhookRequest(logger *zap.Logger, auth Authenticator) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		jsonData, err := io.ReadAll(ctx.Request.Body)
		if err != nil {
			logger.Error(fmt.Sprintf("io.ReadAll failed: %+v", err))
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		isAuthenticated, err := auth.AuthenticateHTTPRequest(ctx.Request.Header, jsonData)
		if err != nil {
			logger.Error(fmt.Sprintf("Authenticate failed: %s", err.Error()))
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		if !isAuthenticated {
			logger.Error(fmt.Sprintf("Authenticate failed: %s", errorNotAValidRequestFromProvider))
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": errorNotAValidRequestFromProvider,
			})
			return
		}
		// pass the request body to the next handler
		ctx.Request.Body = io.NopCloser(bytes.NewBuffer(jsonData))
		ctx.Next()
	}
}
