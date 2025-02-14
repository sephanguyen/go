// Code generated by mockgen. DO NOT EDIT.
package mock_repositories

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/manabie-com/backend/internal/calendar/domain/dto"
	"github.com/manabie-com/backend/internal/golibs/database"
)

type MockSchedulerRepo struct {
	mock.Mock
}

func (r *MockSchedulerRepo) Create(arg1 context.Context, arg2 database.QueryExecer, arg3 *dto.CreateSchedulerParams) (string, error) {
	args := r.Called(arg1, arg2, arg3)
	return args.Get(0).(string), args.Error(1)
}

func (r *MockSchedulerRepo) CreateMany(arg1 context.Context, arg2 database.QueryExecer, arg3 []*dto.CreateSchedulerParamWithIdentity) (map[string]string, error) {
	args := r.Called(arg1, arg2, arg3)
	return args.Get(0).(map[string]string), args.Error(1)
}

func (r *MockSchedulerRepo) GetByID(arg1 context.Context, arg2 database.QueryExecer, arg3 string) (*dto.Scheduler, error) {
	args := r.Called(arg1, arg2, arg3)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.Scheduler), args.Error(1)
}

func (r *MockSchedulerRepo) Update(arg1 context.Context, arg2 database.QueryExecer, arg3 *dto.UpdateSchedulerParams, arg4 []string) error {
	args := r.Called(arg1, arg2, arg3, arg4)
	return args.Error(0)
}
