package commands

import (
	"context"
	"fmt"
	"testing"

	"github.com/manabie-com/backend/internal/notification/modules/system_notification/application/commands/payloads"
	mock_repositories "github.com/manabie-com/backend/mock/notification/modules/system_notification/infrastructure/repo"
	"github.com/manabie-com/backend/mock/testutil"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestSystemNotificationCommandHandler_SetSystemNotificationStatus(t *testing.T) {
	t.Parallel()
	mockDB := testutil.NewMockDB()
	systemNotificationRepo := &mock_repositories.MockSystemNotificationRepo{}
	hdl := &SystemNotificationCommandHandler{
		SystemNotificationRepo: systemNotificationRepo,
	}

	t.Run("error not belong to system notification", func(t *testing.T) {
		payload := &payloads.SetSystemNotificationStatusPayload{
			SystemNotificationID: "sn-1",
			Status:               npb.SystemNotificationStatus_SYSTEM_NOTIFICATION_STATUS_DONE.String(),
		}
		ctx := context.Background()

		systemNotificationRepo.On("CheckUserBelongToSystemNotification", ctx, mockDB.DB, mock.Anything, "sn-1").Once().
			Return(false, nil)

		systemNotificationRepo.On("SetStatus", ctx, mockDB.DB, "sn-1", npb.SystemNotificationStatus_SYSTEM_NOTIFICATION_STATUS_DONE.String()).Once().
			Return(nil)

		err := hdl.SetSystemNotificationStatus(ctx, mockDB.DB, payload)
		assert.Equal(t, status.Error(codes.InvalidArgument, fmt.Sprintf("user does not belong to this system notification")), err)
	})

	t.Run("happy case", func(t *testing.T) {
		payload := &payloads.SetSystemNotificationStatusPayload{
			SystemNotificationID: "sn-1",
			Status:               npb.SystemNotificationStatus_SYSTEM_NOTIFICATION_STATUS_DONE.String(),
		}
		ctx := context.Background()

		systemNotificationRepo.On("CheckUserBelongToSystemNotification", ctx, mockDB.DB, mock.Anything, "sn-1").Once().
			Return(true, nil)

		systemNotificationRepo.On("SetStatus", ctx, mockDB.DB, "sn-1", npb.SystemNotificationStatus_SYSTEM_NOTIFICATION_STATUS_DONE.String()).Once().
			Return(nil)

		err := hdl.SetSystemNotificationStatus(ctx, mockDB.DB, payload)
		assert.Nil(t, err)
	})
}
