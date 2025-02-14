// Code generated by mockgen. DO NOT EDIT.
package mock_commands

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/notification/modules/system_notification/application/commands/payloads"
)

type MockSystemNotificationCommandHandler struct {
	mock.Mock
}

func (r *MockSystemNotificationCommandHandler) SetSystemNotificationStatus(arg1 context.Context, arg2 database.QueryExecer, arg3 *payloads.SetSystemNotificationStatusPayload) error {
	args := r.Called(arg1, arg2, arg3)
	return args.Error(0)
}

func (r *MockSystemNotificationCommandHandler) UpsertSystemNotification(arg1 context.Context, arg2 database.QueryExecer, arg3 *payloads.UpsertSystemNotificationPayload) error {
	args := r.Called(arg1, arg2, arg3)
	return args.Error(0)
}
