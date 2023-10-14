package services

import (
	"context"
	"fmt"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/notification/entities"
	"github.com/manabie-com/backend/internal/notification/repositories"
	"github.com/manabie-com/backend/internal/notification/services/utils"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_repositories "github.com/manabie-com/backend/mock/notification/repositories"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestNotificationModifierService_RetrieveNotificationDetail(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}

	infoNotificationRepo := &mock_repositories.MockInfoNotificationRepo{}
	infoNotificationMsgRepo := &mock_repositories.MockInfoNotificationMsgRepo{}
	userInfoNotificationRepo := &mock_repositories.MockUsersInfoNotificationRepo{}

	svc := &NotificationReaderService{
		DB:                       db,
		InfoNotificationRepo:     infoNotificationRepo,
		InfoNotificationMsgRepo:  infoNotificationMsgRepo,
		UserInfoNotificationRepo: userInfoNotificationRepo,
	}

	testCases := []struct {
		UserID string
		Name   string
		Req    *npb.RetrieveNotificationDetailRequest
		Err    error
		Setup  func(ctx context.Context)
	}{
		{
			UserID: "user_id_1",
			Name:   "happy case",
			Req: &npb.RetrieveNotificationDetailRequest{
				NotificationId: "notification_id_1",
			},
			Setup: func(ctx context.Context) {
				noti, notiMsg := utils.GenSampleNotificationWithMsg()
				userNoti := utils.GenUserNotificationEntity()
				filter := repositories.NewFindNotificationFilter()
				filter.NotiIDs.Set([]string{"notification_id_1"})
				filter.Status.Set([]string{cpb.NotificationStatus_NOTIFICATION_STATUS_SENT.String()})

				findUserNotificationfilter := repositories.FindUserNotificationFilter{
					UserNotificationIDs: pgtype.TextArray{Status: pgtype.Null},
					UserIDs:             database.TextArray([]string{"user_id_1"}),
					NotiIDs:             database.TextArray([]string{"notification_id_1"}),
					UserStatus:          pgtype.TextArray{Status: pgtype.Null},
					Limit:               database.Int8(1),
					StudentID:           pgtype.Text{Status: pgtype.Null},
					ParentID:            pgtype.Text{Status: pgtype.Null},
					OffsetTime:          pgtype.Timestamptz{Status: pgtype.Null},
					OffsetText:          pgtype.Text{Status: pgtype.Null},
					IsImportant:         pgtype.Bool{Status: pgtype.Null},
				}

				userInfoNotificationRepo.On("Find", ctx, db, findUserNotificationfilter).Once().Return(entities.UserInfoNotifications{&userNoti}, nil)
				infoNotificationRepo.On("Find", ctx, db, filter).Once().Return(entities.InfoNotifications([]*entities.InfoNotification{noti}), nil)
				infoNotificationMsgRepo.On("GetByIDs", ctx, db, database.TextArray([]string{noti.NotificationMsgID.String})).Once().Return(entities.InfoNotificationMsgs([]*entities.InfoNotificationMsg{notiMsg}), nil)
			},
		},
		{
			UserID: "user_id_1",
			Name:   "cannot find notification",
			Req: &npb.RetrieveNotificationDetailRequest{
				NotificationId: "notification_id_1",
			},
			Err: status.Error(codes.Internal, fmt.Sprintf("RetrieveNotificationDetail.FindNotification: InfoNotificationRepo.Find: %v", pgx.ErrNoRows)),
			Setup: func(ctx context.Context) {
				userNoti := utils.GenUserNotificationEntity()
				filter := repositories.NewFindNotificationFilter()
				filter.NotiIDs.Set([]string{"notification_id_1"})
				filter.Status.Set([]string{cpb.NotificationStatus_NOTIFICATION_STATUS_SENT.String()})

				findUserNotificationfilter := repositories.FindUserNotificationFilter{
					UserNotificationIDs: pgtype.TextArray{Status: pgtype.Null},
					UserIDs:             database.TextArray([]string{"user_id_1"}),
					NotiIDs:             database.TextArray([]string{"notification_id_1"}),
					UserStatus:          pgtype.TextArray{Status: pgtype.Null},
					Limit:               database.Int8(1),
					StudentID:           pgtype.Text{Status: pgtype.Null},
					ParentID:            pgtype.Text{Status: pgtype.Null},
					OffsetTime:          pgtype.Timestamptz{Status: pgtype.Null},
					OffsetText:          pgtype.Text{Status: pgtype.Null},
					IsImportant:         pgtype.Bool{Status: pgtype.Null},
				}

				userInfoNotificationRepo.On("Find", ctx, db, findUserNotificationfilter).Once().Return(entities.UserInfoNotifications{&userNoti}, nil)
				infoNotificationRepo.On("Find", ctx, db, filter).Once().Return(entities.InfoNotifications{}, pgx.ErrNoRows)
			},
		},
		{
			UserID: "user_id_1",
			Name:   "cannot find notification message",
			Req: &npb.RetrieveNotificationDetailRequest{
				NotificationId: "notification_id_1",
			},
			Err: status.Error(codes.Internal, fmt.Sprintf("RetrieveNotificationDetail.FindNotificationMsg: InfoNotificationMsgRepo.GetByIDs: %v", pgx.ErrNoRows)),
			Setup: func(ctx context.Context) {
				noti, _ := utils.GenSampleNotificationWithMsg()
				userNoti := utils.GenUserNotificationEntity()
				filter := repositories.NewFindNotificationFilter()
				filter.NotiIDs.Set([]string{"notification_id_1"})
				filter.Status.Set([]string{cpb.NotificationStatus_NOTIFICATION_STATUS_SENT.String()})

				findUserNotificationfilter := repositories.FindUserNotificationFilter{
					UserNotificationIDs: pgtype.TextArray{Status: pgtype.Null},
					UserIDs:             database.TextArray([]string{"user_id_1"}),
					NotiIDs:             database.TextArray([]string{"notification_id_1"}),
					UserStatus:          pgtype.TextArray{Status: pgtype.Null},
					StudentID:           pgtype.Text{Status: pgtype.Null},
					ParentID:            pgtype.Text{Status: pgtype.Null},
					Limit:               database.Int8(1),
					OffsetTime:          pgtype.Timestamptz{Status: pgtype.Null},
					OffsetText:          pgtype.Text{Status: pgtype.Null},
					IsImportant:         pgtype.Bool{Status: pgtype.Null},
				}

				userInfoNotificationRepo.On("Find", ctx, db, findUserNotificationfilter).Once().Return(entities.UserInfoNotifications{&userNoti}, nil)
				infoNotificationRepo.On("Find", ctx, db, filter).Once().Return(entities.InfoNotifications([]*entities.InfoNotification{noti}), nil)
				infoNotificationMsgRepo.On("GetByIDs", ctx, db, database.TextArray([]string{noti.NotificationMsgID.String})).Once().Return(entities.InfoNotificationMsgs{}, pgx.ErrNoRows)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			ctx := context.Background()
			ctx = interceptors.ContextWithUserID(ctx, testCase.UserID)
			ctx = metadata.AppendToOutgoingContext(ctx, "pkg", "manabie", "version", "1.0.0", "token", idutil.ULIDNow())
			testCase.Setup(ctx)
			_, err := svc.RetrieveNotificationDetail(ctx, testCase.Req)
			if testCase.Err == nil {
				assert.Nil(t, err)
			} else {
				assert.Equal(t, testCase.Err, err)
			}
		})
	}
}

func TestNotificationModifierService_RetrieveNotifications(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}

	infoNotificationRepo := &mock_repositories.MockInfoNotificationRepo{}
	infoNotificationMsgRepo := &mock_repositories.MockInfoNotificationMsgRepo{}
	userInfoNotificationRepo := &mock_repositories.MockUsersInfoNotificationRepo{}

	svc := &NotificationReaderService{
		DB:                       db,
		InfoNotificationRepo:     infoNotificationRepo,
		InfoNotificationMsgRepo:  infoNotificationMsgRepo,
		UserInfoNotificationRepo: userInfoNotificationRepo,
	}

	testCases := []struct {
		Name   string
		UserID string
		Req    *npb.RetrieveNotificationsRequest
		Err    error
		Setup  func(ctx context.Context)
	}{
		{
			Name:   "happy case",
			UserID: "user_id_1",
			Req:    &npb.RetrieveNotificationsRequest{},
			Setup: func(ctx context.Context) {
				noti, notiMsg := utils.GenSampleNotificationWithMsg()
				userNoti := &entities.UserInfoNotification{
					UserID:         database.Text("user_id_1"),
					NotificationID: noti.NotificationID,
				}

				filter := repositories.FindUserNotificationFilter{
					UserIDs:             database.TextArray([]string{"user_id_1"}),
					UserNotificationIDs: pgtype.TextArray{Status: pgtype.Null},
					NotiIDs:             pgtype.TextArray{Status: pgtype.Null},
					UserStatus:          pgtype.TextArray{Status: pgtype.Null},
					Limit:               database.Int8(100),
					OffsetTime:          pgtype.Timestamptz{Status: pgtype.Null},
					OffsetText:          pgtype.Text{Status: pgtype.Null},
					StudentID:           pgtype.Text{Status: pgtype.Null},
					ParentID:            pgtype.Text{Status: pgtype.Null},
					IsImportant:         pgtype.Bool{Status: pgtype.Null},
				}

				notiMap := make(map[string]*entities.InfoNotificationMsg)
				notiMap[noti.NotificationID.String] = notiMsg

				userInfoNotificationRepo.On("Find", ctx, db, filter).Once().Return(entities.UserInfoNotifications([]*entities.UserInfoNotification{userNoti}), nil)
				infoNotificationMsgRepo.On("GetByNotificationIDs", ctx, db, mock.Anything).Once().Return(notiMap, nil)
				infoNotificationRepo.On("Find", ctx, db, mock.Anything).Once().Return(entities.InfoNotifications{noti}, nil)
				infoNotificationMsgRepo.On("GetByIDs", ctx, db, database.TextArray([]string{noti.NotificationMsgID.String})).Once().Return(entities.InfoNotificationMsgs([]*entities.InfoNotificationMsg{notiMsg}), nil)
			},
		},
		{
			Name:   "cannot find notification",
			UserID: "user_id_2",
			Req:    &npb.RetrieveNotificationsRequest{},
			Err:    status.Error(codes.Internal, fmt.Sprintf("FindUserNotification UserInfoNotificationRepo.Find: %v", pgx.ErrNoRows)),
			Setup: func(ctx context.Context) {
				filter := repositories.FindUserNotificationFilter{
					UserNotificationIDs: pgtype.TextArray{Status: pgtype.Null},
					UserIDs:             database.TextArray([]string{"user_id_2"}),
					NotiIDs:             pgtype.TextArray{Status: pgtype.Null},
					UserStatus:          pgtype.TextArray{Status: pgtype.Null},
					Limit:               database.Int8(100),
					OffsetTime:          pgtype.Timestamptz{Status: pgtype.Null},
					OffsetText:          pgtype.Text{Status: pgtype.Null},
					StudentID:           pgtype.Text{Status: pgtype.Null},
					ParentID:            pgtype.Text{Status: pgtype.Null},
					IsImportant:         pgtype.Bool{Status: pgtype.Null},
				}

				userInfoNotificationRepo.On("Find", ctx, db, filter).Once().Return(entities.UserInfoNotifications{}, pgx.ErrNoRows)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			ctx := context.Background()
			ctx = interceptors.ContextWithUserID(ctx, testCase.UserID)
			ctx = metadata.AppendToOutgoingContext(ctx, "pkg", "manabie", "version", "1.0.0", "token", idutil.ULIDNow())
			testCase.Setup(ctx)
			notifications, err := svc.RetrieveNotifications(ctx, testCase.Req)
			if testCase.Err == nil {
				assert.Nil(t, err)
				for _, noti := range notifications.Items {
					if noti.UserNotification.Data != "hello world" || noti.UserNotification.Type.String() != cpb.NotificationType_NOTIFICATION_TYPE_COMPOSED.String() {
						assert.Error(t, fmt.Errorf("Data response error"))
					}
				}
			} else {
				assert.Equal(t, testCase.Err, err)
			}
		})
	}
}
