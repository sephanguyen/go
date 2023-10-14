package infrastructure

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/spike/modules/email/domain/model"
)

type EmailRepo interface {
	UpsertEmail(ctx context.Context, db database.QueryExecer, email *model.Email) error
	UpdateEmail(ctx context.Context, db database.QueryExecer, emailID string, attributes map[string]interface{}) error
}

type EmailRecipientRepo interface {
	BulkUpsertEmailRecipients(ctx context.Context, db database.QueryExecer, emailRecipients model.EmailRecipients) error
	GetEmailRecipientsByEmailID(ctx context.Context, db database.QueryExecer, emailID string) (model.EmailRecipients, error)
}

type EmailRecipientEventRepo interface {
	BulkInsertEmailRecipientEventRepo(ctx context.Context, db database.QueryExecer, emailRecipientEvents model.EmailRecipientEvents) error
	GetMapEventsByEventsAndEmailRecipientIDs(ctx context.Context, db database.QueryExecer, events, emailRecipientIDs []string) (map[string]*model.EmailRecipientEvent, error)
}
