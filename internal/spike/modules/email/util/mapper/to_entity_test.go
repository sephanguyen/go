package mapper

import (
	"testing"

	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/spike/modules/email/domain/dto"
	"github.com/manabie-com/backend/internal/spike/modules/email/domain/model"

	"github.com/bxcodec/faker/v3"
	"github.com/stretchr/testify/assert"
)

func Test_DTOToEmailEntity(t *testing.T) {
	t.Parallel()

	var email = &dto.Email{}
	assert.NoError(t, faker.FakeData(email))
	// can't fake enum type for now

	ent, err := ToEmailEntity(email)
	assert.NoError(t, err)
	assert.Equal(t, email.EmailID, ent.EmailID.String)
	assert.Equal(t, email.EmailFrom, ent.EmailFrom.String)
	assert.Equal(t, email.SendGridMessageID, ent.SendGridMessageID.String)
	assert.Equal(t, email.Subject, ent.Subject.String)
	assert.Equal(t, email.Status, ent.Status.String)

	recipients := []string{}
	err = ent.EmailRecipients.AssignTo(&recipients)
	assert.NoError(t, err)
	assert.Equal(t, email.EmailRecipients, recipients)

	emailContent := model.EmailContent{}
	err = ent.Content.AssignTo(&emailContent)
	assert.Equal(t, email.Content.HTMLContent, emailContent.HTMLContent)
	assert.Equal(t, email.Content.PlainTextContent, emailContent.PlainTextContent)
}

func Test_DTOToEmailRecipientEntities(t *testing.T) {
	t.Parallel()

	var email = &dto.Email{
		EmailID: idutil.ULIDNow(),
		EmailRecipients: []string{
			"example-1@manabie.com",
			"example-2@manabie.com",
			"example-3@manabie.com",
		},
	}

	emailRecipientEnts, err := ToEmailRecipientEntities(email)
	assert.NoError(t, err)
	assert.Equal(t, len(emailRecipientEnts), len(email.EmailRecipients))

	for i, emailRecipientEnt := range emailRecipientEnts {
		assert.Equal(t, email.EmailID, emailRecipientEnt.EmailID.String)
		assert.Equal(t, email.EmailRecipients[i], emailRecipientEnt.RecipientAddress.String)
	}
}
