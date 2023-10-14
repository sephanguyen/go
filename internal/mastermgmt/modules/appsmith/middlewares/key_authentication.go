package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	MasterHeaderKey = "Mastermgmt-signature-v1"
	MasterAuthValue = "M@nabie-mastermgmt"
)

func Authenticate(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		signature := c.GetHeader(MasterHeaderKey)
		logger.Info("master authenticate", zap.String("signature", signature))
		if MasterAuthValue != signature {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "signature is not match",
			})
			return
		}

		c.Next()
	}
}
