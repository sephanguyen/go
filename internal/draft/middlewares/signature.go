package middlewares

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	HeaderKey = "Manabie-Signature"
)

func VerifySignature(headerKey, signingKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		buf, err := io.ReadAll(c.Request.Body)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		signature := c.GetHeader(headerKey)
		mac := hmac.New(sha256.New, []byte(signingKey))
		_, _ = mac.Write(buf)
		sign := hex.EncodeToString(mac.Sum(nil))

		if sign != signature {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "signature is not match",
			})
			return
		}

		c.Request.Body = io.NopCloser(bytes.NewBuffer(buf))
		c.Next()
	}
}
