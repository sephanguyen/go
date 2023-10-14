package http

import (
	"net/http/httptest"
	"testing"

	"github.com/manabie-com/backend/internal/spike/configurations"
	"github.com/manabie-com/backend/internal/spike/modules/email/application/commands"
	"github.com/manabie-com/backend/internal/spike/modules/email/domain/dto"
	"github.com/manabie-com/backend/internal/spike/modules/email/util"
	mock_commands "github.com/manabie-com/backend/mock/spike/modules/email/application/commands"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestEmailHTTPService_EmailStatusReceiver(t *testing.T) {
	t.Parallel()
	core, log := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	emailEventHandler := &mock_commands.MockUpsertEmailEventHandler{}
	httpService := &EmailHTTPService{
		Logger:            logger,
		EmailEventHandler: emailEventHandler,
	}
	t.Run("should fail to handle request", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)

		// dummy body data, expected to fail to marshalling this
		content := map[string]interface{}{
			"foo": "bar",
		}
		headers := map[string][]string{}

		req, _ := util.NewMockRequest("POST", content, headers)

		ctx.Request = req

		emailEventHandler.On("UpsertEmailEvent", ctx, mock.Anything).Once().Return(nil)

		httpService.EmailStatusReceiver(ctx)

		entry := log.All()[0]
		assert.Equal(t,
			"failed ToEmailEventDTO:json: cannot unmarshal object into Go value of type []dto.SGEmailEvent",
			entry.Message)
	})

	t.Run("success with webhook config", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		allowedOrgIDs := []string{"org-id-1", "org-id-2"}
		httpService.EmailWebhookCfg = configurations.EmailWebhook{
			ReceiveOnlyFronTenant: allowedOrgIDs,
		}

		content := &[]dto.SGEmailEvent{
			{
				OrganizationID: "org-id-1",
				EmailID:        "email-id",
			},
		}
		headers := map[string][]string{}

		req, _ := util.NewMockRequest("POST", content, headers)

		ctx.Request = req

		emailEventHandler.On("UpsertEmailEvent", ctx, mock.MatchedBy(func(in commands.UpsertEmailEventPayload) bool {
			if len(in.AllowedOrgIDs) != len(allowedOrgIDs) {
				return false
			}
			return true
		})).Once().Return(nil)

		httpService.EmailStatusReceiver(ctx)
	})
}
