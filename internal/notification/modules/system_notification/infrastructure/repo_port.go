package infrastructure

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/notification/modules/system_notification/domain/model"
	"github.com/manabie-com/backend/internal/notification/modules/system_notification/infrastructure/repo"
)

type SystemNotificationRepo interface {
	UpsertSystemNotification(ctx context.Context, db database.QueryExecer, systemNotification *model.SystemNotification) error
	FindSystemNotifications(ctx context.Context, db database.QueryExecer, filter *repo.FindSystemNotificationFilter) (model.SystemNotifications, error)
	CountSystemNotifications(ctx context.Context, db database.QueryExecer, filter *repo.FindSystemNotificationFilter) (map[string]uint32, error)
	FindByReferenceID(ctx context.Context, db database.QueryExecer, referenceID string) (*model.SystemNotification, error)
	CheckUserBelongToSystemNotification(ctx context.Context, db database.QueryExecer, userID, systemNotificationID string) (bool, error)
	SetStatus(ctx context.Context, db database.QueryExecer, systemNotificationID, status string) error
}

type SystemNotificationRecipientRepo interface {
	BulkInsertSystemNotificationRecipients(ctx context.Context, db database.QueryExecer, systemNotificationRecipients model.SystemNotificationRecipients) error
	SoftDeleteBySystemNotificationID(ctx context.Context, db database.QueryExecer, systemNotificationID string) error
}

type SystemNotificationContentRepo interface {
	FindBySystemNotificationIDs(ctx context.Context, db database.QueryExecer, systemNotificationIDs []string) (model.SystemNotificationContents, error)
	SoftDeleteBySystemNotificationID(ctx context.Context, db database.QueryExecer, systemNotificationID string) error
	BulkInsertSystemNotificationContents(ctx context.Context, db database.QueryExecer, systemNotificationContents model.SystemNotificationContents) error
}
