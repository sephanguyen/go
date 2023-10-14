package mappers

import (
	"context"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/notification/repositories"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
)

func Test_UnreadUserNotificationFilter(t *testing.T) {
	filter := UnreadUserNotificationFilter("noti1", 100)
	assert.Equal(t, []string{"noti1"}, database.FromTextArray(filter.NotiIDs))
	assert.Equal(t, int64(100), filter.Limit.Int)
	assert.Equal(t, []string{cpb.UserNotificationStatus_USER_NOTIFICATION_STATUS_NEW.String()}, database.FromTextArray(filter.UserStatus))
}

func Test_NotificationDetailToUserNotificationFilter(t *testing.T) {
	fakeUser := "user1"
	ctx := interceptors.ContextWithUserID(context.Background(), fakeUser)
	basicAssert := func(t *testing.T, filter repositories.FindUserNotificationFilter) {
		assert.Equal(t, int64(1), filter.Limit.Int)
		assert.Equal(t, pgtype.Null, filter.UserStatus.Status)
		assert.Equal(t, pgtype.Null, filter.OffsetTime.Status)
		assert.Equal(t, pgtype.Null, filter.OffsetText.Status)
		assert.Equal(t, []string{fakeUser}, database.FromTextArray(filter.UserIDs))
		assert.Equal(t, pgtype.Null, filter.IsImportant.Status)
	}
	t.Run("Has target id", func(t *testing.T) {
		req := &npb.RetrieveNotificationDetailRequest{
			NotificationId: "noti1",
			TargetId:       "student 1",
		}
		filter := NotificationDetailToUserNotificationFilter(ctx, req)
		assert.Equal(t, req.TargetId, filter.StudentID.String)
		assert.Equal(t, req.TargetId, filter.ParentID.String)
		assert.Equal(t, []string{req.NotificationId}, database.FromTextArray(filter.NotiIDs))
		basicAssert(t, filter)
	})
	t.Run("Has no target id", func(t *testing.T) {
		req := &npb.RetrieveNotificationDetailRequest{
			NotificationId: "noti1",
		}
		filter := NotificationDetailToUserNotificationFilter(ctx, req)
		basicAssert(t, filter)
		assert.Equal(t, []string{req.NotificationId}, database.FromTextArray(filter.NotiIDs))
		assert.Equal(t, pgtype.Null, filter.StudentID.Status)
		assert.Equal(t, pgtype.Null, filter.ParentID.Status)
	})
	t.Run("Has no target id, has user notification id", func(t *testing.T) {
		req := &npb.RetrieveNotificationDetailRequest{
			UserNotificationId: "user-noti-id-1",
		}
		filter := NotificationDetailToUserNotificationFilter(ctx, req)
		basicAssert(t, filter)
		assert.Equal(t, []string{req.UserNotificationId}, database.FromTextArray(filter.UserNotificationIDs))
		assert.Equal(t, pgtype.Null, filter.StudentID.Status)
		assert.Equal(t, pgtype.Null, filter.ParentID.Status)
	})
}
