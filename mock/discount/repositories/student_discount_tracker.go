// Code generated by mockgen. DO NOT EDIT.
package mock_repositories

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/manabie-com/backend/internal/discount/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
)

type MockStudentDiscountTrackerRepo struct {
	mock.Mock
}

func (r *MockStudentDiscountTrackerRepo) Create(arg1 context.Context, arg2 database.QueryExecer, arg3 *entities.StudentDiscountTracker) error {
	args := r.Called(arg1, arg2, arg3)
	return args.Error(0)
}

func (r *MockStudentDiscountTrackerRepo) GetActiveTrackingByStudentIDs(arg1 context.Context, arg2 database.QueryExecer, arg3 []string) ([]entities.StudentDiscountTracker, error) {
	args := r.Called(arg1, arg2, arg3)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entities.StudentDiscountTracker), args.Error(1)
}

func (r *MockStudentDiscountTrackerRepo) GetByID(arg1 context.Context, arg2 database.QueryExecer, arg3 string) (entities.StudentDiscountTracker, error) {
	args := r.Called(arg1, arg2, arg3)
	return args.Get(0).(entities.StudentDiscountTracker), args.Error(1)
}

func (r *MockStudentDiscountTrackerRepo) UpdateTrackingDurationByStudentProduct(arg1 context.Context, arg2 database.QueryExecer, arg3 entities.StudentProduct) error {
	args := r.Called(arg1, arg2, arg3)
	return args.Error(0)
}
