package middleware

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"

	httppayload "github.com/manabie-com/backend/internal/conversationmgmt/modules/conversation/controller/http/payload"
	"github.com/manabie-com/backend/internal/golibs/chatvendor/agora"
	"github.com/manabie-com/backend/internal/golibs/configs"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var (
	errorNotAValidRequestFromProvider = "Not a valid request from provider"
)

func VerifyWebhookRequest(agoraCfg configs.AgoraConfig, logger *zap.Logger, verifier agora.WebhookVerifier) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		payload, err := io.ReadAll(ctx.Request.Body)
		if err != nil {
			logger.Error(fmt.Sprintf("io.ReadAll failed: %+v", err))
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, httppayload.NewMessageEventResponse(false, err))
			return
		}
		isVerified, err := verifier.Verify(agoraCfg.WebhookSecret, payload)
		if err != nil {
			logger.Error(fmt.Sprintf("Verify failed: %s", err.Error()))
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, httppayload.NewMessageEventResponse(false, err))
			return
		}
		if !isVerified {
			logger.Error(fmt.Sprintf("Verify failed: %s", errorNotAValidRequestFromProvider))
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, httppayload.NewMessageEventResponse(false, errors.New(errorNotAValidRequestFromProvider)))
			return
		}
		// pass the request body to the next handler
		ctx.Request.Body = io.NopCloser(bytes.NewBuffer(payload))
		ctx.Next()
	}
}
