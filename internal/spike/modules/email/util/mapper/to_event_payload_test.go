package mapper

import (
	"testing"

	"github.com/manabie-com/backend/internal/spike/modules/email/domain/dto"

	"github.com/bxcodec/faker/v3"
	"github.com/stretchr/testify/assert"
)

func TestToSendEmailEventPayload(t *testing.T) {
	t.Parallel()

	var email = &dto.Email{}
	assert.NoError(t, faker.FakeData(email))

	payload := ToSendEmailEventPayload(email)
	assert.Equal(t, email.EmailID, payload.EmailID)
	assert.Equal(t, email.SendGridMessageID, payload.SendGridMessageID)
	assert.Equal(t, email.EmailFrom, payload.EmailFrom)
	assert.Equal(t, email.Content.HTMLContent, payload.Content.HTMLContent)
	assert.Equal(t, email.Content.PlainTextContent, payload.Content.PlainTextContent)
	assert.Equal(t, email.EmailRecipients, payload.EmailRecipients)
	assert.Equal(t, email.Status, payload.Status)
	assert.Equal(t, email.Subject, payload.Subject)
}
