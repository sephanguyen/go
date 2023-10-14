package controller

import (
	"context"
	"testing"

	mock_commands "github.com/manabie-com/backend/mock/notification/modules/system_notification/application/commands"
	"github.com/manabie-com/backend/mock/testutil"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestSystemNotificationModifier_SetSystemNotificationStatus(t *testing.T) {
	t.Parallel()
	mockDB := testutil.NewMockDB()
	systemNotificationCommandHandler := &mock_commands.MockSystemNotificationCommandHandler{}
	svc := &SystemNotificationModifierService{
		DB:                               mockDB.DB,
		SystemNotificationCommandHandler: systemNotificationCommandHandler,
	}

	testCases := []struct {
		Name     string
		Request  interface{}
		Response interface{}
		Error    error
		Setup    func(ctx context.Context)
	}{
		{
			Name: "must return error",
			Request: &npb.SetSystemNotificationStatusRequest{
				SystemNotificationId: "",
				Status:               npb.SystemNotificationStatus_SYSTEM_NOTIFICATION_STATUS_DONE,
			},
			Response: &npb.SetSystemNotificationStatusResponse{},
			Error:    status.Error(codes.InvalidArgument, "missing SystemNotificationID"),
			Setup: func(ctx context.Context) {
				systemNotificationCommandHandler.On("SetSystemNotificationStatus", ctx, mockDB.DB, mock.Anything).Once().
					Return(nil)
			},
		},
		{
			Name: "happy case",
			Request: &npb.SetSystemNotificationStatusRequest{
				SystemNotificationId: "sn-1",
				Status:               npb.SystemNotificationStatus_SYSTEM_NOTIFICATION_STATUS_DONE,
			},
			Response: &npb.SetSystemNotificationStatusResponse{},
			Error:    nil,
			Setup: func(ctx context.Context) {
				systemNotificationCommandHandler.On("SetSystemNotificationStatus", ctx, mockDB.DB, mock.Anything).Once().
					Return(nil)
			},
		},
	}

	for _, tc := range testCases {
		ctx := context.Background()
		tc.Setup(ctx)
		_, err := svc.SetSystemNotificationStatus(ctx, tc.Request.(*npb.SetSystemNotificationStatusRequest))
		assert.Equal(t, tc.Error, err)
	}
}
