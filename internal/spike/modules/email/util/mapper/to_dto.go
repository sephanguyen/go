package mapper

import (
	"encoding/json"

	"github.com/manabie-com/backend/internal/spike/modules/email/domain/dto"
	spb "github.com/manabie-com/backend/pkg/manabuf/spike/v1"
)

func ToEmailDTO(email *spb.SendEmailRequest, emailFrom string) *dto.Email {
	emailDTO := &dto.Email{
		Subject: email.Subject,
		Content: dto.EmailContent{
			HTMLContent:      email.Content.HTML,
			PlainTextContent: email.Content.PlainText,
		},
		EmailRecipients: email.Recipients,
		EmailFrom:       emailFrom,
	}

	return emailDTO
}

func ToEmailEventDTO(data []byte) ([]dto.SGEmailEvent, error) {
	emailEvents := &[]dto.SGEmailEvent{}
	err := json.Unmarshal(data, emailEvents)
	if err != nil {
		return nil, err
	}
	return *emailEvents, nil
}
