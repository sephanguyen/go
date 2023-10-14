package consumers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/sendgrid"
	sendgrid_builders "github.com/manabie-com/backend/internal/golibs/sendgrid/builders"
	"github.com/manabie-com/backend/internal/spike/modules/email/application/consumers/payloads"
	"github.com/manabie-com/backend/internal/spike/modules/email/constants"
	"github.com/manabie-com/backend/internal/spike/modules/email/domain/model"
	"github.com/manabie-com/backend/internal/spike/modules/email/infrastructure"
	spb "github.com/manabie-com/backend/pkg/manabuf/spike/v1"
	"go.uber.org/multierr"
)

type SendEmailHandler struct {
	DB             database.Ext
	SendGridClient sendgrid.SendGridClient

	EmailRepo          infrastructure.EmailRepo
	EmailRecipientRepo infrastructure.EmailRecipientRepo
}

func (h *SendEmailHandler) Handle(ctx context.Context, value []byte) (bool, error) {
	emailEvent := payloads.SendEmailEvent{}
	err := json.Unmarshal(value, &emailEvent)
	if err != nil {
		return false, fmt.Errorf("cannot unmarshal message payload: %s", string(value))
	}

	org, err := interceptors.OrganizationFromContext(ctx)
	if err != nil {
		return false, fmt.Errorf("failed OrganizationFromContext: %+v", err)
	}

	emailRecipients, err := h.EmailRecipientRepo.GetEmailRecipientsByEmailID(ctx, h.DB, emailEvent.EmailID)
	if err != nil {
		return false, fmt.Errorf("failed EmailRecipientRepo.GetEmailRecipientsByEmailID: %+v", err)
	}

	correspondents := makeSendGridCorrespondents(emailRecipients, org.OrganizationID().String())

	sgEmailBuilder := sendgrid_builders.NewEmailLegacyBuilder().
		WithSubject(emailEvent.Subject).
		WithSender(sendgrid.Correspondent{
			Name:    constants.ManabieSenderName,
			Address: emailEvent.EmailFrom,
		}).
		WithRecipients(correspondents...)

	if emailEvent.Content.PlainTextContent != "" {
		sgEmailBuilder = sgEmailBuilder.WithContents(constants.EmailContentPlainTextType, emailEvent.Content.PlainTextContent)
	}
	if emailEvent.Content.HTMLContent != "" {
		sgEmailBuilder = sgEmailBuilder.WithContents(constants.EmailContentHTMLTextType, emailEvent.Content.HTMLContent)
	}

	sgEmail := sgEmailBuilder.BuildEmail()
	sgMessageID, err := h.SendGridClient.SendWithContext(ctx, sgEmail)
	if err != nil {
		// Update processed failed status
		errStatusUpdated := h.EmailRepo.UpdateEmail(ctx, h.DB, emailEvent.EmailID, map[string]interface{}{
			"status": spb.EmailStatus_EMAIL_STATUS_PROCESSED_FAILED.String(),
		})
		if errStatusUpdated != nil {
			err = multierr.Combine(err, fmt.Errorf("error on call EmailRepo.UpdateEmail: [%v]", errStatusUpdated))
			return true, fmt.Errorf("occurred some errors: [%v]", err)
		}

		// Because SendGrid return an error when SendEmail, so we can re-call it again (retry == true)
		return true, fmt.Errorf("error on call SendGrid SendEmail: [%v]", err)
	}

	// Update processed status and sg_message_id
	errUpdated := h.EmailRepo.UpdateEmail(ctx, h.DB, emailEvent.EmailID, map[string]interface{}{
		"status":        spb.EmailStatus_EMAIL_STATUS_PROCESSED.String(),
		"sg_message_id": sgMessageID,
	})
	if errUpdated != nil {
		// This error occurred when updating status and sg_message_id, meaning we called SendEmail of SendGrid successfully
		// so we don't retry this logic, maybe will send email twice
		return false, fmt.Errorf("error on update status and sg_message_id: [%v], email payload: [%v]", errUpdated, emailEvent)
	}

	return false, nil
}

func makeSendGridCorrespondents(emailRecipients model.EmailRecipients, orgID string) []sendgrid.Correspondent {
	correspondents := make([]sendgrid.Correspondent, 0)
	for _, emailRecipient := range emailRecipients {
		emailAddress := emailRecipient.RecipientAddress.String
		correspondent := sendgrid.Correspondent{
			Name:    emailAddress,
			Address: emailAddress,
			CustomArguments: []map[string]string{
				{
					"organization_id":    orgID,
					"email_id":           emailRecipient.EmailID.String,
					"email_recipient_id": emailRecipient.EmailRecipientID.String,
				},
			},
		}

		correspondents = append(correspondents, correspondent)
	}

	return correspondents
}
