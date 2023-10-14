package mapper

import (
	"testing"

	"github.com/manabie-com/backend/internal/spike/modules/email/constants"
	"github.com/manabie-com/backend/internal/spike/modules/email/domain/dto"
	spb "github.com/manabie-com/backend/pkg/manabuf/spike/v1"

	"github.com/bxcodec/faker/v3"
	"github.com/stretchr/testify/assert"
)

func Test_PbToEmailDTO(t *testing.T) {
	t.Parallel()

	var email = &spb.SendEmailRequest{}
	assert.NoError(t, faker.FakeData(email))

	dto := ToEmailDTO(email, constants.ManabieDomainEmail)
	assert.Equal(t, constants.ManabieDomainEmail, dto.EmailFrom)
	assert.Equal(t, email.Content.HTML, dto.Content.HTMLContent)
	assert.Equal(t, email.Content.PlainText, dto.Content.PlainTextContent)
	assert.Equal(t, email.Recipients, dto.EmailRecipients)
}

func Test_ToEmailEventDTO(t *testing.T) {
	t.Parallel()

	t.Run("both case full properties and lack of some properties", func(t *testing.T) {
		payload := `[
			{
				"email": "email1@gmail.com",
				"event": "delivered",
				"sg_event_id": "sg_event_id1",
				"sg_message_id": "sg_message_id1",
				"reason": "some reason",
				"status": "2.0.0",
				"attempt": "5",
				"type": "bounce",
				"bounce_classification": "invalid",
				"timestamp": 1513299569,
				"organization_id": "-2147483648"
			},
			{
				"email":"email1@gmail.com",
				"event":"processed",
				"organization_id": "-2147483648"
			}
		]`

		expectEmailEvents := []dto.SGEmailEvent{
			{
				Email:                "email1@gmail.com",
				Event:                "delivered",
				SGEventID:            "sg_event_id1",
				SGMessageID:          "sg_message_id1",
				Reason:               "some reason",
				Status:               "2.0.0",
				Attempt:              "5",
				Type:                 "bounce",
				BounceClassification: "invalid",
				Timestamp:            1513299569,
				OrganizationID:       "-2147483648",
			},
			{
				Email:          "email1@gmail.com",
				Event:          "processed",
				OrganizationID: "-2147483648",
			},
		}

		emailEvents, err := ToEmailEventDTO([]byte(payload))
		assert.Nil(t, err)
		assert.Equal(t, expectEmailEvents, emailEvents)
	})
}
