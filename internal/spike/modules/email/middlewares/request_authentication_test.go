package middlewares

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/manabie-com/backend/internal/golibs/sendgrid"
	"github.com/manabie-com/backend/internal/spike/modules/email/util"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

type MiddlewareError struct {
	Error string `json:"error"`
}

func Test_Authenticate(t *testing.T) {
	t.Parallel()
	logger := zaptest.NewLogger(t)
	getHandler := func() gin.HandlerFunc {
		// using sendgrid provider
		auth := &sendgrid.Mock{}

		return AuthenticateWebhookRequest(logger, auth)
	}

	t.Run("happy case", func(t *testing.T) {
		content := map[string]interface{}{
			"foo": "bar",
		}

		headers := map[string][]string{
			"code": {"200"}, // tell the authenticator to return ok
		}

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)

		req, expectedBody := util.NewMockRequest("POST", content, headers)

		ctx.Request = req

		authHandler := getHandler()
		authHandler(ctx)

		// assertion to make use Body content is still exist to passed to the handler chain
		assert.Equal(t, expectedBody, ctx.Request.Body)
	})

	t.Run("invalid http request", func(t *testing.T) {
		content := map[string]interface{}{
			"foo": "bar",
		}

		headers := map[string][]string{
			"code": {"400"}, // tell the authenticator to return unauthorized code
		}

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)

		req, _ := util.NewMockRequest("POST", content, headers)

		ctx.Request = req

		authHandler := getHandler()
		authHandler(ctx)

		errMsg := &MiddlewareError{}
		err := json.Unmarshal(w.Body.Bytes(), &errMsg)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Equal(t, errorNotAValidRequestFromProvider, errMsg.Error)
	})

	t.Run("server error", func(t *testing.T) {
		content := map[string]interface{}{
			"foo": "bar",
		}

		headers := map[string][]string{
			"code": {"500"}, // tell the authenticator to return unauthorized code
		}

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)

		req, _ := util.NewMockRequest("POST", content, headers)

		ctx.Request = req

		authHandler := getHandler()
		authHandler(ctx)

		errMsg := &MiddlewareError{}
		err := json.Unmarshal(w.Body.Bytes(), &errMsg)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.NotEmpty(t, errMsg.Error)
	})
}
