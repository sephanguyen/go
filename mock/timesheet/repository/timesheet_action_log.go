// Code generated by mockgen. DO NOT EDIT.
package mock_repositories

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/timesheet/domain/entity"
)

type MockTimesheetActionLogRepoImpl struct {
	mock.Mock
}

func (r *MockTimesheetActionLogRepoImpl) Create(arg1 context.Context, arg2 database.QueryExecer, arg3 *entity.TimesheetActionLog) error {
	args := r.Called(arg1, arg2, arg3)
	return args.Error(0)
}
