// Code generated by mockgen. DO NOT EDIT.
package mock_repositories

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/notification/entities"
)

type MockNotificationInternalUserRepo struct {
	mock.Mock
}

func (r *MockNotificationInternalUserRepo) GetByOrgID(arg1 context.Context, arg2 database.QueryExecer, arg3 string) (*entities.NotificationInternalUser, error) {
	args := r.Called(arg1, arg2, arg3)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.NotificationInternalUser), args.Error(1)
}
