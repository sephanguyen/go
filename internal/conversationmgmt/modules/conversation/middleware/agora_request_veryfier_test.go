package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/manabie-com/backend/internal/conversationmgmt/modules/conversation/constants"
	"github.com/manabie-com/backend/internal/conversationmgmt/modules/conversation/controller/http/payload"
	"github.com/manabie-com/backend/internal/conversationmgmt/modules/conversation/utils"
	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/spike/modules/email/util"
	mock_agora "github.com/manabie-com/backend/mock/golibs/chatvendor/agora"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func Test_VerifyWebhookRequest(t *testing.T) {
	t.Parallel()
	logger := zaptest.NewLogger(t)
	dummySecret := "dummy-secret"
	verifier := mock_agora.NewWebhookVerifier(t)
	agoraCfg := configs.AgoraConfig{WebhookSecret: dummySecret}

	getHandler := func() gin.HandlerFunc {
		// using sendgrid provider
		return VerifyWebhookRequest(agoraCfg, logger, verifier)
	}

	t.Run("happy case", func(t *testing.T) {
		content := map[string]interface{}{
			"foo": "bar",
		}
		byteContent, _ := json.Marshal(content)
		verifier.On("Verify", dummySecret, byteContent).Once().Return(true, nil)

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		req, expectedBody := utils.NewMockRequest("POST", content, nil)
		ctx.Request = req

		verifyHandler := getHandler()
		verifyHandler(ctx)

		// assertion to make use Body content is still exist to passed to the handler chain
		assert.Equal(t, expectedBody, ctx.Request.Body)
	})

	t.Run("invalid http request", func(t *testing.T) {
		content := map[string]interface{}{
			"foo": "bar",
		}
		byteContent, _ := json.Marshal(content)
		verifier.On("Verify", dummySecret, byteContent).Once().Return(false, nil)

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request, _ = util.NewMockRequest("POST", content, nil)

		verifyHandler := getHandler()
		verifyHandler(ctx)

		errMsg := &payload.MessageEventResponse{}
		err := json.Unmarshal(w.Body.Bytes(), &errMsg)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Equal(t, constants.ResponseWebhookFAILED, errMsg.Code)
	})

	t.Run("server error", func(t *testing.T) {
		content := map[string]interface{}{
			"foo": "bar",
		}
		byteContent, _ := json.Marshal(content)
		verifier.On("Verify", dummySecret, byteContent).Once().Return(false, fmt.Errorf("dummy"))

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request, _ = util.NewMockRequest("POST", content, nil)

		verifyHandler := getHandler()
		verifyHandler(ctx)

		errMsg := &payload.MessageEventResponse{}
		err := json.Unmarshal(w.Body.Bytes(), &errMsg)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.NotEmpty(t, errMsg.Code)
	})
}
