// Code generated by mockgen. DO NOT EDIT.
package mock_repositories

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/timesheet/domain/entity"
)

type MockPartnerAutoCreateTimesheetFlagRepoImpl struct {
	mock.Mock
}

func (r *MockPartnerAutoCreateTimesheetFlagRepoImpl) GetPartnerAutoCreateDefaultValue(arg1 context.Context, arg2 database.QueryExecer) (*entity.PartnerAutoCreateTimesheetFlag, error) {
	args := r.Called(arg1, arg2)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.PartnerAutoCreateTimesheetFlag), args.Error(1)
}
