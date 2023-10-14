package validation

import (
	"fmt"

	spb "github.com/manabie-com/backend/pkg/manabuf/spike/v1"
)

var (
	ErrMissingSubject    = "your email is missing subject field (must have)."
	ErrMissingContent    = "your email is missing content field: html and plain text."
	ErrMissingRecipients = "your email is missing repcipients field: at least one recipient here."
)

func ValidateSendEmailRequiredFields(emailReq *spb.SendEmailRequest) error {
	if emailReq.Subject == "" {
		return fmt.Errorf(ErrMissingSubject)
	}

	if emailReq.Content == nil {
		return fmt.Errorf(ErrMissingContent)
	}

	if emailReq.Content.HTML == "" && emailReq.Content.PlainText == "" {
		return fmt.Errorf(ErrMissingContent)
	}

	if emailReq.Recipients == nil || len(emailReq.Recipients) == 0 {
		return fmt.Errorf(ErrMissingRecipients)
	}
	return nil
}
