package grpc

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/constants"
	spike_consts "github.com/manabie-com/backend/internal/spike/constants"
	"github.com/manabie-com/backend/internal/spike/modules/email/application/commands"
	email_consts "github.com/manabie-com/backend/internal/spike/modules/email/constants"
	"github.com/manabie-com/backend/internal/spike/modules/email/metrics"
	"github.com/manabie-com/backend/internal/spike/modules/email/util/mapper"
	"github.com/manabie-com/backend/internal/spike/modules/email/util/validation"
	spb "github.com/manabie-com/backend/pkg/manabuf/spike/v1"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (svc *EmailModifierService) SendEmail(ctx context.Context, req *spb.SendEmailRequest) (*spb.SendEmailResponse, error) {
	if err := validation.ValidateSendEmailRequiredFields(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	emailFrom := email_consts.ManabieDomainEmail
	if svc.Env == spike_consts.ProdEnv {
		emailFrom = email_consts.ManabieDomainEmailPROD
	}
	createEmailPayload := &commands.CreateEmailPayload{
		Email: mapper.ToEmailDTO(req, emailFrom),
	}
	email, err := svc.EmailCommandHandler.CreateEmail(ctx, createEmailPayload)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("svc.EmailCommandHandler.CreateEmail: %v", err))
	}

	// Publish send email event to Kafka, only enable on [local - stag]
	if svc.Env == spike_consts.LocalEnv || svc.Env == spike_consts.StagEnv {
		logger := ctxzap.Extract(ctx)
		sendEmailEventPayload := mapper.ToSendEmailEventPayload(email)
		var sendEmailEventBytesPayload []byte
		sendEmailEventBytesPayload, err = json.Marshal(sendEmailEventPayload)
		if err != nil {
			logger.Sugar().Errorf("error when call json.Marshal (email event): %v", err)
		}

		spanName := "PUBLISHER." + constants.EmailSendingTopic
		err = svc.KafkaMgmt.TracedPublishContext(ctx, spanName, constants.EmailSendingTopic, []byte(sendEmailEventPayload.EmailID), sendEmailEventBytesPayload)
		if err != nil {
			logger.Sugar().Errorf("error when call svc.KafkaMgmt.PublishContext: %v", err)
		}
	}

	// Error when publishing email sending payload to Kafka
	if err != nil {
		// Update processed failed status
		errStatusUpdated := svc.EmailRepo.UpdateEmail(ctx, svc.DB, email.EmailID, map[string]interface{}{
			"status": spb.EmailStatus_EMAIL_STATUS_QUEUED_FAILED.String(),
		})
		if errStatusUpdated != nil {
			err = multierr.Combine(err, fmt.Errorf("error when call EmailRepo.UpdateEmail: [%v]", errStatusUpdated))
			return &spb.SendEmailResponse{
				EmailId: email.EmailID,
			}, status.Error(codes.Internal, fmt.Sprintf("occurred some error when publishing email event and update email status: [%v]", err))
		}

		return &spb.SendEmailResponse{
			EmailId: email.EmailID,
		}, status.Error(codes.Internal, fmt.Sprintf("error when publishing email event: [%v]", err))
	}

	// Record [queued] email event
	svc.EmailMetrics.RecordEmailEvents(metrics.EmailQueued, float64(len(email.EmailRecipients)))

	return &spb.SendEmailResponse{
		EmailId: email.EmailID,
	}, nil
}
