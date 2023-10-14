package http

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/spike/configurations"
	"github.com/manabie-com/backend/internal/spike/modules/email/application/commands"
	"github.com/manabie-com/backend/internal/spike/modules/email/infrastructure/repositories"
	"github.com/manabie-com/backend/internal/spike/modules/email/metrics"
	"github.com/manabie-com/backend/internal/spike/modules/email/util/mapper"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type EmailHTTPService struct {
	Logger          *zap.Logger
	EmailWebhookCfg configurations.EmailWebhook

	EmailEventHandler interface {
		UpsertEmailEvent(ctx context.Context, payload commands.UpsertEmailEventPayload) error
	}
}

func NewEmailHTTPService(db database.Ext, logger *zap.Logger, metrics metrics.EmailMetrics, webhookCfg configurations.EmailWebhook) *EmailHTTPService {
	return &EmailHTTPService{
		Logger:          logger,
		EmailWebhookCfg: webhookCfg,
		EmailEventHandler: &commands.UpsertEmailEventHandler{
			DB:                      db,
			EmailMetrics:            metrics,
			EmailRecipientEventRepo: &repositories.EmailRecipientEventRepo{},
		},
	}
}

func (s *EmailHTTPService) EmailStatusReceiver(ctx *gin.Context) {
	data, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		s.Logger.Error(fmt.Sprintf("failed io.Read: %+v", err))
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	emailEvents, err := mapper.ToEmailEventDTO(data)
	if err != nil {
		s.Logger.Error(fmt.Sprintf("failed ToEmailEventDTO:%+v", err))
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	if len(emailEvents) == 0 {
		ctx.JSON(http.StatusOK, gin.H{"success": true})
		return
	}

	allowedTenantIDs := []string{}
	if !s.EmailWebhookCfg.ReceiveFromAllTenant {
		allowedTenantIDs = s.EmailWebhookCfg.ReceiveOnlyFronTenant
	}

	err = s.EmailEventHandler.UpsertEmailEvent(ctx, commands.UpsertEmailEventPayload{
		EmailEvents:   emailEvents,
		AllowedOrgIDs: allowedTenantIDs,
	})
	if err != nil {
		s.Logger.Error(fmt.Sprintf("failed UpsertEmailEvents:%+v", err))
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"success": true})
}
