package services

import (
	"context"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/notification/entities"
	"github.com/manabie-com/backend/internal/notification/repositories"
	mock_bob_repositories "github.com/manabie-com/backend/mock/bob/repositories"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_infra "github.com/manabie-com/backend/mock/notification/infra"
	mock_repositories "github.com/manabie-com/backend/mock/notification/repositories"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNotificationModifierService_NotifyUnreadUser(t *testing.T) {
	db := &mock_database.Ext{}
	infoNotificationRepo := &mock_repositories.MockInfoNotificationRepo{}
	infoNotificationMsgRepo := &mock_repositories.MockInfoNotificationMsgRepo{}
	userNotificationRepo := &mock_repositories.MockUsersInfoNotificationRepo{}
	activityLogRepo := &mock_bob_repositories.MockActivityLogRepo{}
	pushNotificationService := &mock_infra.PushNotificationService{}

	userDeviceTokenRepo := &mock_repositories.MockUserDeviceTokenRepo{}
	svc := &NotificationModifierService{
		DB:                      db,
		InfoNotificationRepo:    infoNotificationRepo,
		InfoNotificationMsgRepo: infoNotificationMsgRepo,
		UserNotificationRepo:    userNotificationRepo,
		ActivityLogRepo:         activityLogRepo,
		PushNotificationService: pushNotificationService,
		UserDeviceTokenRepo:     userDeviceTokenRepo,
	}

	testCases := []struct {
		Name  string
		Req   *npb.NotifyUnreadUserRequest
		Err   error
		Setup func(ctx context.Context)
	}{
		{
			Name: "happy case",
			Err:  nil,
			Req: &npb.NotifyUnreadUserRequest{
				NotificationId: "notification-id-1",
			},
			Setup: func(ctx context.Context) {
				filter := repositories.NewFindNotificationFilter()
				filter.NotiIDs.Set([]string{"notification-id-1"})
				filter.Status.Set([]string{cpb.NotificationStatus_NOTIFICATION_STATUS_SENT.String()})
				notis := entities.InfoNotifications([]*entities.InfoNotification{
					{
						NotificationID:    database.Text("notification-id-1"),
						NotificationMsgID: database.Text("notification-msg-id-1"),
					},
				})

				notiMsgs := entities.InfoNotificationMsgs([]*entities.InfoNotificationMsg{
					{
						NotificationMsgID: database.Text("notification-msg-id-1"),
					},
				})
				infoNotificationRepo.On("Find", ctx, db, filter).Once().Return(notis, nil)
				infoNotificationMsgRepo.On("GetByIDs", ctx, db, database.TextArray([]string{"notification-msg-id-1"})).Once().Return(notiMsgs, nil)

				userNotificationsMap := make(map[string]entities.UserInfoNotifications)
				userNotificationsMap["notification-id-1"] = entities.UserInfoNotifications{
					{
						NotificationID: database.Text("notification-id-1"),
						UserID:         database.Text("unread-user-id-1"),
					},
					{
						NotificationID: database.Text("notification-id-1"),
						UserID:         database.Text("unread-user-id-2"),
					},
					{
						NotificationID: database.Text("notification-id-1"),
						UserID:         database.Text("unread-user-id-3"),
					},
					{
						NotificationID: database.Text("notification-id-1"),
						UserID:         database.Text("unread-user-id-4"),
					},
				}

				findUserNotiFilter := repositories.NewFindUserNotificationFilter()
				findUserNotiFilter.UserIDs = pgtype.TextArray{Status: pgtype.Null}
				findUserNotiFilter.NotiIDs = database.TextArray([]string{"notification-id-1"})
				findUserNotiFilter.UserStatus = database.TextArray([]string{cpb.UserNotificationStatus_USER_NOTIFICATION_STATUS_NEW.String()})
				findUserNotiFilter.OffsetText = pgtype.Text{Status: pgtype.Null}
				findUserNotiFilter.Limit = database.Int8(1000)
				userNotificationRepo.On("FindUserIDs", ctx, db, findUserNotiFilter).Once().Return(userNotificationsMap, nil)

				userDeviceTokens := entities.UserDeviceTokens{
					{
						UserID:            database.Text("unread-user-id-1"),
						DeviceToken:       database.Text("device-token-1"),
						AllowNotification: database.Bool(true),
					},
					{
						UserID:            database.Text("unread-user-id-2"),
						DeviceToken:       database.Text("device-token-2"),
						AllowNotification: database.Bool(true),
					},
					{
						UserID:            database.Text("unread-user-id-3"),
						DeviceToken:       database.Text("device-token-3"),
						AllowNotification: database.Bool(true),
					},
					{
						UserID:            database.Text("unread-user-id-4"),
						DeviceToken:       database.Text("device-token-4"),
						AllowNotification: database.Bool(true),
					},
				}

				pushNotificationService.On("PushNotificationForUser", ctx, userDeviceTokens, notis[0], notiMsgs[0]).Once().Return(0, 0, nil)

				userIDs := []string{"unread-user-id-1", "unread-user-id-2", "unread-user-id-3", "unread-user-id-4"}

				userNotificationRepo.On("UpdateUnreadUser", ctx, db, database.Text("notification-id-1"), database.TextArray(userIDs)).Once().Return(nil)

				findUserNotiFilterNext := repositories.NewFindUserNotificationFilter()
				findUserNotiFilterNext.UserIDs = pgtype.TextArray{Status: pgtype.Null}
				findUserNotiFilterNext.NotiIDs = database.TextArray([]string{"notification-id-1"})
				findUserNotiFilterNext.UserStatus = database.TextArray([]string{cpb.UserNotificationStatus_USER_NOTIFICATION_STATUS_NEW.String()})
				findUserNotiFilterNext.OffsetText = database.Text(userIDs[len(userIDs)-1])
				findUserNotiFilterNext.Limit = database.Int8(1000)
				userDeviceTokenRepo.On("FindByUserIDs", ctx, db, database.TextArray(userIDs)).Return(userDeviceTokens, nil)
				userNotificationRepo.On("FindUserIDs", ctx, db, findUserNotiFilterNext).Once().Return(make(map[string]entities.UserInfoNotifications), nil)
				activityLogRepo.On("Create", ctx, db, mock.Anything).Once().Return(nil)
			},
		},
	}

	userID := "user_id_1"
	ctx := context.Background()
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			ctx = interceptors.ContextWithUserID(ctx, userID)
			testCase.Setup(ctx)
			_, err := svc.NotifyUnreadUser(ctx, testCase.Req)
			assert.Nil(t, testCase.Err)
			if testCase.Err != nil {
				assert.Equal(t, testCase.Err, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
