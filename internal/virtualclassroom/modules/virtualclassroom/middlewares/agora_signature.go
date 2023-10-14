package middlewares

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"time"

	"github.com/buger/jsonparser"
	"github.com/gin-gonic/gin"
)

const (
	AgoraHeaderKey = "Agora-Signature-V2"
)

const payloadKey = "payload"

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

		c.Set(payloadKey, buf)
		c.Next()
	}
}

func PayloadFromContext(c *gin.Context) []byte {
	if payload, ok := c.Get(payloadKey); ok {
		return payload.([]byte)
	}
	return nil
}

func VerifyTimestamp(maxSecondPayloadExpired int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		payload := PayloadFromContext(c)
		if len(payload) == 0 {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "body is empty",
			})
			return
		}

		ts, err := jsonparser.GetInt(payload, "timestamp")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		now := time.Now().Unix()
		if now-ts > maxSecondPayloadExpired {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "time is hack",
			})
			return
		}

		c.Next()
	}
}
