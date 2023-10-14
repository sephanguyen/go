package mapper

import (
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/spike/modules/email/constants"
	"github.com/manabie-com/backend/internal/spike/modules/email/domain/dto"
	"github.com/manabie-com/backend/internal/spike/modules/email/domain/model"

	"go.uber.org/multierr"
)

func ToEmailEntity(email *dto.Email) (*model.Email, error) {
	emailEnt := &model.Email{}
	database.AllNullEntity(emailEnt)

	emailContentEnt := &model.EmailContent{}
	emailContentEnt.HTMLContent = email.Content.HTMLContent
	emailContentEnt.PlainTextContent = email.Content.PlainTextContent

	err := multierr.Combine(
		emailEnt.EmailID.Set(email.EmailID),
		emailEnt.EmailFrom.Set(email.EmailFrom),
		emailEnt.SendGridMessageID.Set(email.SendGridMessageID),
		emailEnt.Subject.Set(email.Subject),
		emailEnt.Content.Set(emailContentEnt),
		emailEnt.EmailRecipients.Set(email.EmailRecipients),
		emailEnt.Status.Set(email.Status),
	)

	if err != nil {
		return nil, err
	}

	return emailEnt, nil
}

func ToEmailRecipientEntities(email *dto.Email) (model.EmailRecipients, error) {
	emailRecipientEnts := make([]*model.EmailRecipient, 0)

	for _, recipient := range email.EmailRecipients {
		emailRecipientEnt := &model.EmailRecipient{}
		database.AllNullEntity(emailRecipientEnt)

		err := multierr.Combine(
			emailRecipientEnt.EmailID.Set(email.EmailID),
			emailRecipientEnt.EmailRecipientID.Set(idutil.ULIDNow()),
			emailRecipientEnt.RecipientAddress.Set(recipient),
		)
		if err != nil {
			return nil, err
		}

		emailRecipientEnts = append(emailRecipientEnts, emailRecipientEnt)
	}

	return emailRecipientEnts, nil
}

func ToEmailRecipientEventEntities(events []dto.SGEmailEvent) (model.EmailRecipientEvents, error) {
	emailRecipientEventEnts := make([]*model.EmailRecipientEvent, 0)

	for _, ev := range events {
		emailRecipientEventEnt := &model.EmailRecipientEvent{}
		database.AllNullEntity(emailRecipientEventEnt)

		err := multierr.Combine(
			emailRecipientEventEnt.Event.Set(ev.Event),
			emailRecipientEventEnt.EmailRecipientEventID.Set(ev.SGEventID),
			emailRecipientEventEnt.Type.Set(constants.EventTypeDelivery),
		)

		if err != nil {
			return nil, err
		}

		emailRecipientEventEnts = append(emailRecipientEventEnts, emailRecipientEventEnt)
	}

	return emailRecipientEventEnts, nil
}
