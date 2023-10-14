package services

import (
	"context"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_repositories "github.com/manabie-com/backend/mock/notification/repositories"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
)

func TestNotificationModifierService_CountUserNotification(t *testing.T) {
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
		Req    *npb.CountUserNotificationRequest
		Err    error
		Setup  func(ctx context.Context)
	}{
		{
			Name:   "happy case",
			UserID: "user_id_1",
			Req: &npb.CountUserNotificationRequest{
				Status: cpb.UserNotificationStatus_USER_NOTIFICATION_STATUS_READ,
			},
			Setup: func(ctx context.Context) {
				readNotiCount := 3
				totalNoti := 5
				userInfoNotificationRepo.On("CountByStatus", ctx, db, database.Text("user_id_1"), database.Text(cpb.UserNotificationStatus_USER_NOTIFICATION_STATUS_READ.String())).Once().Return(readNotiCount, totalNoti, nil)
			},
		},
		{
			Name:   "happy case",
			UserID: "user_id_1",
			Req: &npb.CountUserNotificationRequest{
				Status: cpb.UserNotificationStatus_USER_NOTIFICATION_STATUS_NEW,
			},
			Setup: func(ctx context.Context) {
				newNotiCount := 2
				totalNoti := 5
				userInfoNotificationRepo.On("CountByStatus", ctx, db, database.Text("user_id_1"), database.Text(cpb.UserNotificationStatus_USER_NOTIFICATION_STATUS_NEW.String())).Once().Return(newNotiCount, totalNoti, nil)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			ctx := context.Background()
			ctx = interceptors.ContextWithUserID(ctx, testCase.UserID)
			ctx = metadata.AppendToOutgoingContext(ctx, "pkg", "manabie", "version", "1.0.0", "token", idutil.ULIDNow())
			testCase.Setup(ctx)
			_, err := svc.CountUserNotification(ctx, testCase.Req)
			if testCase.Err == nil {
				assert.Nil(t, err)
			} else {
				assert.Equal(t, testCase.Err, err)
			}
		})
	}
}
