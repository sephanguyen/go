package consumers

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/sendgrid"
	sendgrid_builders "github.com/manabie-com/backend/internal/golibs/sendgrid/builders"
	"github.com/manabie-com/backend/internal/spike/modules/email/application/consumers/payloads"
	"github.com/manabie-com/backend/internal/spike/modules/email/constants"
	"github.com/manabie-com/backend/internal/spike/modules/email/domain/model"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_sendgrid "github.com/manabie-com/backend/mock/golibs/sendgrid"
	mock_repo "github.com/manabie-com/backend/mock/spike/modules/email/infrastructure/repositories"
	spb "github.com/manabie-com/backend/pkg/manabuf/spike/v1"

	"github.com/stretchr/testify/assert"
	"go.uber.org/multierr"
)

func TestSendEmailHandler_Handle(t *testing.T) {
	t.Parallel()

	mockSendGridClient := mock_sendgrid.NewSendGridClient(t)
	mockDB := new(mock_database.Ext)
	mockEmailRepo := mock_repo.MockEmailRepo{}
	mockEmailRecipientRepo := mock_repo.MockEmailRecipientRepo{}

	handler := SendEmailHandler{
		DB:                 mockDB,
		SendGridClient:     mockSendGridClient,
		EmailRepo:          &mockEmailRepo,
		EmailRecipientRepo: &mockEmailRecipientRepo,
	}

	sgMessageID := idutil.ULIDNow()
	emailID := idutil.ULIDNow()
	emailRecipientID := idutil.ULIDNow()
	dummySGError := fmt.Errorf("call SendEmail with SendGrid error")
	dummyUpdateEmailError := fmt.Errorf("call emailRepo.UpdateEmail error")

	emailEventExample := payloads.SendEmailEvent{
		EmailID:           emailID,
		SendGridMessageID: "",
		Subject:           "Unit test email",
		Content: payloads.EmailContent{
			PlainTextContent: "Unit test email content plain text.",
			HTMLContent:      "Unit test email content HTML.",
		},
		EmailFrom: constants.ManabieDomainEmail,
		Status:    "EMAIL_STATUS_QUEUED",
		EmailRecipients: []string{
			"example@manabie.com",
		},
	}

	emailRecipientsExample := model.EmailRecipients{
		{
			EmailID:          database.Text(emailID),
			EmailRecipientID: database.Text(emailRecipientID),
			RecipientAddress: database.Text("example@manabie.com"),
		},
	}

	resourcePath := "-2147483648"
	correspondents := makeSendGridCorrespondents(emailRecipientsExample, resourcePath)
	sgEmail := sendgrid_builders.NewEmailLegacyBuilder().
		WithSubject(emailEventExample.Subject).
		WithSender(sendgrid.Correspondent{
			Name:    constants.ManabieSenderName,
			Address: emailEventExample.EmailFrom,
		}).
		WithRecipients(correspondents...).
		WithContents(constants.EmailContentPlainTextType, emailEventExample.Content.PlainTextContent).
		WithContents(constants.EmailContentHTMLTextType, emailEventExample.Content.HTMLContent).
		BuildEmail()

	ctx := interceptors.ContextWithJWTClaims(context.Background(), &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			ResourcePath: resourcePath,
		},
	})

	t.Run("happy case", func(t *testing.T) {
		emailEventExampleByte, _ := json.Marshal(emailEventExample)
		mockEmailRecipientRepo.On("GetEmailRecipientsByEmailID", ctx, mockDB, emailID).Once().Return(emailRecipientsExample, nil)

		mockSendGridClient.On("SendWithContext", ctx, sgEmail).Once().Return(sgMessageID, nil)
		mockEmailRepo.On("UpdateEmail", ctx, mockDB, emailID, map[string]interface{}{
			"status":        spb.EmailStatus_EMAIL_STATUS_PROCESSED.String(),
			"sg_message_id": sgMessageID,
		}).Once().Return(nil)

		isRetry, err := handler.Handle(ctx, emailEventExampleByte)

		assert.NoError(t, err)
		assert.Equal(t, false, isRetry)
	})

	t.Run("sendgrid failed, email update successfully", func(t *testing.T) {
		emailEventExampleByte, _ := json.Marshal(emailEventExample)
		mockEmailRecipientRepo.On("GetEmailRecipientsByEmailID", ctx, mockDB, emailID).Once().Return(emailRecipientsExample, nil)
		mockSendGridClient.On("SendWithContext", ctx, sgEmail).Once().Return("", dummySGError)
		mockEmailRepo.On("UpdateEmail", ctx, mockDB, emailID, map[string]interface{}{
			"status": spb.EmailStatus_EMAIL_STATUS_PROCESSED_FAILED.String(),
		}).Once().Return(nil)

		isRetry, err := handler.Handle(ctx, emailEventExampleByte)
		assert.EqualError(t, err, fmt.Sprintf("error on call SendGrid SendEmail: [%v]", dummySGError))
		assert.Equal(t, true, isRetry)
	})

	t.Run("sendgrid failed, email update failed", func(t *testing.T) {
		emailEventExampleByte, _ := json.Marshal(emailEventExample)
		mockEmailRecipientRepo.On("GetEmailRecipientsByEmailID", ctx, mockDB, emailID).Once().Return(emailRecipientsExample, nil)
		mockSendGridClient.On("SendWithContext", ctx, sgEmail).Once().Return("", dummySGError)
		mockEmailRepo.On("UpdateEmail", ctx, mockDB, emailID, map[string]interface{}{
			"status": spb.EmailStatus_EMAIL_STATUS_PROCESSED_FAILED.String(),
		}).Once().Return(dummyUpdateEmailError)

		combinedErr := multierr.Combine(
			dummySGError,
			fmt.Errorf("error on call EmailRepo.UpdateEmail: [%v]", dummyUpdateEmailError),
		)

		isRetry, err := handler.Handle(ctx, emailEventExampleByte)

		assert.EqualError(t, err, fmt.Sprintf("occurred some errors: [%v]", combinedErr))
		assert.Equal(t, true, isRetry)
	})

	t.Run("sendgrid successfully, email update failed", func(t *testing.T) {
		emailEventExampleByte, _ := json.Marshal(emailEventExample)
		mockEmailRecipientRepo.On("GetEmailRecipientsByEmailID", ctx, mockDB, emailID).Once().Return(emailRecipientsExample, nil)
		mockSendGridClient.On("SendWithContext", ctx, sgEmail).Once().Return(sgMessageID, nil)
		mockEmailRepo.On("UpdateEmail", ctx, mockDB, emailID, map[string]interface{}{
			"status":        spb.EmailStatus_EMAIL_STATUS_PROCESSED.String(),
			"sg_message_id": sgMessageID,
		}).Once().Return(dummyUpdateEmailError)

		isRetry, err := handler.Handle(ctx, emailEventExampleByte)

		assert.EqualError(t, err, fmt.Sprintf("error on update status and sg_message_id: [%v], email payload: [%v]", dummyUpdateEmailError, emailEventExample))
		assert.Equal(t, false, isRetry)
	})
}
