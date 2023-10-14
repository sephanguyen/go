package mapper

import (
	"github.com/manabie-com/backend/internal/spike/modules/email/application/consumers/payloads"
	"github.com/manabie-com/backend/internal/spike/modules/email/domain/dto"
)

func ToSendEmailEventPayload(email *dto.Email) *payloads.SendEmailEvent {
	sendEmailEvent := &payloads.SendEmailEvent{}

	emailContent := payloads.EmailContent{}
	emailContent.HTMLContent = email.Content.HTMLContent
	emailContent.PlainTextContent = email.Content.PlainTextContent

	sendEmailEvent.EmailID = email.EmailID
	sendEmailEvent.SendGridMessageID = email.SendGridMessageID
	sendEmailEvent.EmailFrom = email.EmailFrom
	sendEmailEvent.Subject = email.Subject
	sendEmailEvent.Content = emailContent
	sendEmailEvent.EmailRecipients = email.EmailRecipients
	sendEmailEvent.Status = email.Status

	return sendEmailEvent
}
