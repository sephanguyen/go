// Code generated by mockgen. DO NOT EDIT.
package mock_repositories

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/manabie-com/backend/internal/golibs/database"
)

type MockLiveRoomActivityLogRepo struct {
	mock.Mock
}

func (r *MockLiveRoomActivityLogRepo) CreateLog(arg1 context.Context, arg2 database.Ext, arg3 string, arg4 string, arg5 string) error {
	args := r.Called(arg1, arg2, arg3, arg4, arg5)
	return args.Error(0)
}
