package queries

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/notification/modules/system_notification/domain/dto"
	"github.com/manabie-com/backend/internal/notification/modules/system_notification/domain/model"
	"github.com/manabie-com/backend/internal/notification/modules/system_notification/infrastructure/repo"
	mock_repositories "github.com/manabie-com/backend/mock/notification/modules/system_notification/infrastructure/repo"
	"github.com/manabie-com/backend/mock/testutil"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_RetrieveSystemNotifications(t *testing.T) {
	t.Parallel()

	mockDB := testutil.NewMockDB()

	systemNotificationRepo := &mock_repositories.MockSystemNotificationRepo{}
	systemNotificationContentRepo := &mock_repositories.MockSystemNotificationContentRepo{}

	handler := &SystemNotificationQueryHandler{
		DB:                            mockDB.DB,
		SystemNotificationRepo:        systemNotificationRepo,
		SystemNotificationContentRepo: systemNotificationContentRepo,
	}

	ctx := context.Background()
	snID1 := "system-notification-id-1"
	snID2 := "system-notification-id-2"

	eventInDB1 := &model.SystemNotification{
		SystemNotificationID: database.Text(snID1),
	}
	eventInDB2 := &model.SystemNotification{
		SystemNotificationID: database.Text(snID2),
	}

	contentList1 := &model.SystemNotificationContent{
		SystemNotificationID: database.Text(snID1),
		Language:             database.Text("en"),
		Text:                 database.Text("text_en"),
	}
	contentList2 := &model.SystemNotificationContent{
		SystemNotificationID: database.Text(snID2),
		Language:             database.Text("vi"),
		Text:                 database.Text("text_vi"),
	}

	contents := model.SystemNotificationContents{
		contentList1,
		contentList2,
	}
	userID := "user-id"

	payload := &RetrieveSystemNotificationPayload{
		Limit:    10,
		Offset:   0,
		UserID:   userID,
		Language: "en",
		Status:   npb.SystemNotificationStatus_SYSTEM_NOTIFICATION_STATUS_NEW.String(),
		Keyword:  "keyword",
	}

	filter := repo.NewFindSystemNotificationFilter()
	_ = filter.Limit.Set(payload.Limit)
	_ = filter.Offset.Set(payload.Offset)
	_ = filter.ValidFrom.Set(time.Now())
	_ = filter.Language.Set(payload.Language)
	_ = filter.Status.Set(payload.Status)
	_ = filter.Keyword.Set(payload.Keyword)
	_ = filter.UserID.Set(payload.UserID)

	totalForStatus := map[string]uint32{
		npb.SystemNotificationStatus_SYSTEM_NOTIFICATION_STATUS_NEW.String():  1,
		npb.SystemNotificationStatus_SYSTEM_NOTIFICATION_STATUS_DONE.String(): 2,
		npb.SystemNotificationStatus_SYSTEM_NOTIFICATION_STATUS_NONE.String(): 3,
	}

	t.Run("happy case", func(t *testing.T) {
		systemNotificationRepo.
			On("FindSystemNotifications", ctx, mockDB.DB, mock.MatchedBy(func(f *repo.FindSystemNotificationFilter) bool {
				if f.Limit != filter.Limit || f.Offset != filter.Offset {
					return false
				}
				if f.ValidFrom.Status == pgtype.Null {
					return false
				}
				if f.UserID.String != userID {
					return false
				}
				return true
			})).
			Once().
			Return(model.SystemNotifications{eventInDB1, eventInDB2}, nil)

		systemNotificationContentRepo.
			On("FindBySystemNotificationIDs", ctx, mockDB.DB, mock.Anything).
			Once().
			Return(contents, nil)

		systemNotificationRepo.
			On("CountSystemNotifications", ctx, mockDB.DB, mock.MatchedBy(func(f *repo.FindSystemNotificationFilter) bool {
				if f.Limit != filter.Limit || f.Offset != filter.Offset {
					return false
				}
				if f.ValidFrom.Status == pgtype.Null {
					return false
				}
				if f.UserID.String != userID {
					return false
				}
				return true
			})).
			Once().
			Return(totalForStatus, nil)

		resp := handler.RetrieveSystemNotifications(ctx, payload)

		assert.Nil(t, resp.Error)
		for _, e := range resp.SystemNotifications {
			assert.Equal(t, 1, len(e.Content), "Length of content must be 1")
			assert.IsType(t, &dto.SystemNotification{}, e, "Object is not type dto.SystemNotification")
		}
		assert.Equal(t, totalForStatus[npb.SystemNotificationStatus_SYSTEM_NOTIFICATION_STATUS_NONE.String()], resp.TotalCount)
	})

	t.Run("find system notifications error", func(t *testing.T) {
		systemNotificationRepo.
			On("FindSystemNotifications", ctx, mockDB.DB, mock.MatchedBy(func(f *repo.FindSystemNotificationFilter) bool {
				if f.Limit != filter.Limit || f.Offset != filter.Offset {
					return false
				}
				if f.ValidFrom.Status == pgtype.Null {
					return false
				}
				if f.UserID.String != userID {
					return false
				}
				return true
			})).
			Once().
			Return(model.SystemNotifications{}, pgx.ErrNoRows)

		systemNotificationContentRepo.
			On("FindBySystemNotificationIDs", ctx, mockDB.DB, mock.Anything).
			Once().
			Return(contents, nil)

		systemNotificationRepo.
			On("CountSystemNotifications", ctx, mockDB.DB, mock.MatchedBy(func(f *repo.FindSystemNotificationFilter) bool {
				if f.Limit != filter.Limit || f.Offset != filter.Offset {
					return false
				}
				if f.ValidFrom.Status == pgtype.Null {
					return false
				}
				if f.UserID.String != userID {
					return false
				}
				return true
			})).
			Once().
			Return(totalForStatus, nil)

		resp := handler.RetrieveSystemNotifications(ctx, payload)

		assert.Nil(t, resp.SystemNotifications)
		assert.Equal(t, fmt.Errorf("query.SystemNotificationRepo.FindSystemNotifications: %v", pgx.ErrNoRows), resp.Error)
		assert.Equal(t, uint32(0), resp.TotalCount)
	})
}
