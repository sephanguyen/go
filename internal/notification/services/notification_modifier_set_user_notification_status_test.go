package services

import (
	"context"
	"fmt"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_repositories "github.com/manabie-com/backend/mock/notification/repositories"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestNotificationModifierService_SetUserNotificationStatus(t *testing.T) {
	db := &mock_database.Ext{}
	userInfoNotificationRepo := &mock_repositories.MockUsersInfoNotificationRepo{}

	svc := &NotificationModifierService{
		DB:                   db,
		UserNotificationRepo: userInfoNotificationRepo,
	}

	userId := idutil.ULIDNow()
	notificationId := idutil.ULIDNow()

	testCases := []struct {
		Name    string
		Request *npb.SetUserNotificationStatusRequest
		Err     error
		Setup   func(ctx context.Context)
	}{
		{
			Name: "happy case set new status",
			Request: &npb.SetUserNotificationStatusRequest{
				NotificationIds: []string{notificationId},
				Status:          cpb.UserNotificationStatus_USER_NOTIFICATION_STATUS_NEW,
			},
			Err: nil,
			Setup: func(ctx context.Context) {
				userInfoNotificationRepo.On("SetStatusByNotificationIDs", ctx, db, database.Text(userId), database.TextArray([]string{notificationId}), database.Text(cpb.UserNotificationStatus_USER_NOTIFICATION_STATUS_NEW.String())).Once().Return(nil)
			},
		},
		{
			Name: "happy case set read status",
			Request: &npb.SetUserNotificationStatusRequest{
				NotificationIds: []string{notificationId},
				Status:          cpb.UserNotificationStatus_USER_NOTIFICATION_STATUS_READ,
			},
			Err: nil,
			Setup: func(ctx context.Context) {
				userInfoNotificationRepo.On("SetStatusByNotificationIDs", ctx, db, database.Text(userId), database.TextArray([]string{notificationId}), database.Text(cpb.UserNotificationStatus_USER_NOTIFICATION_STATUS_READ.String())).Once().Return(nil)
			},
		},
		{
			Name: "invalid status",
			Request: &npb.SetUserNotificationStatusRequest{
				NotificationIds: []string{notificationId},
				Status:          cpb.UserNotificationStatus_USER_NOTIFICATION_STATUS_NONE,
			},
			Err: status.Error(codes.InvalidArgument, fmt.Sprintf("invalid request Status %v", cpb.UserNotificationStatus_USER_NOTIFICATION_STATUS_NONE)),
			Setup: func(ctx context.Context) {
			},
		},
	}

	ctx := context.Background()
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			ctx = interceptors.ContextWithUserID(ctx, userId)
			ctx = metadata.AppendToOutgoingContext(ctx, "pkg", "manabie", "version", "1.0.0", "token", idutil.ULIDNow())
			testCase.Setup(ctx)
			_, err := svc.SetUserNotificationStatus(ctx, testCase.Request)
			if testCase.Err != nil {
				assert.Equal(t, testCase.Err, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
