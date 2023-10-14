package mappers

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/notification/repositories"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"github.com/jackc/pgtype"
)

func UnreadUserNotificationFilter(notiID string, limit int64) repositories.FindUserNotificationFilter {
	findUserNotiFilter := repositories.NewFindUserNotificationFilter()
	findUserNotiFilter.UserIDs = pgtype.TextArray{Status: pgtype.Null}
	findUserNotiFilter.NotiIDs = database.TextArray([]string{notiID})
	findUserNotiFilter.UserStatus = database.TextArray([]string{cpb.UserNotificationStatus_USER_NOTIFICATION_STATUS_NEW.String()})
	findUserNotiFilter.OffsetText = pgtype.Text{Status: pgtype.Null}
	findUserNotiFilter.Limit = database.Int8(limit)
	return findUserNotiFilter
}

func NotificationDetailToUserNotificationFilter(ctx context.Context, req *npb.RetrieveNotificationDetailRequest) repositories.FindUserNotificationFilter {
	userID := interceptors.UserIDFromContext(ctx)
	filter := repositories.FindUserNotificationFilter{
		UserNotificationIDs: pgtype.TextArray{Status: pgtype.Null},
		UserIDs:             database.TextArray([]string{userID}),
		NotiIDs:             pgtype.TextArray{Status: pgtype.Null},
		UserStatus:          pgtype.TextArray{Status: pgtype.Null},
		Limit:               database.Int8(1),
		OffsetTime:          pgtype.Timestamptz{Status: pgtype.Null},
		OffsetText:          pgtype.Text{Status: pgtype.Null},
		StudentID:           pgtype.Text{Status: pgtype.Null},
		ParentID:            pgtype.Text{Status: pgtype.Null},
		IsImportant:         pgtype.Bool{Status: pgtype.Null},
	}
	// filter using student_id is applicable for new app only
	// we keep backward compatibility with old app
	if req.GetTargetId() != "" {
		filter.StudentID = database.Text(req.GetTargetId())
		filter.ParentID = database.Text(req.GetTargetId())
	}
	if req.GetNotificationId() != "" {
		filter.NotiIDs = database.TextArray([]string{req.GetNotificationId()})
	}
	if req.GetUserNotificationId() != "" {
		filter.UserNotificationIDs = database.TextArray([]string{req.GetUserNotificationId()})
	}
	return filter
}
