// Code generated by mockgen. DO NOT EDIT.
package mock_repositories

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/payment/entities"
)

type MockStudentEnrollmentStatusHistoryRepo struct {
	mock.Mock
}

func (r *MockStudentEnrollmentStatusHistoryRepo) GetCurrentStatusByStudentIDAndLocationID(arg1 context.Context, arg2 database.QueryExecer, arg3 string, arg4 string) (entities.StudentEnrollmentStatusHistory, error) {
	args := r.Called(arg1, arg2, arg3, arg4)
	return args.Get(0).(entities.StudentEnrollmentStatusHistory), args.Error(1)
}

func (r *MockStudentEnrollmentStatusHistoryRepo) GetLatestStatusByStudentIDAndLocationID(arg1 context.Context, arg2 database.QueryExecer, arg3 string, arg4 string) (entities.StudentEnrollmentStatusHistory, error) {
	args := r.Called(arg1, arg2, arg3, arg4)
	return args.Get(0).(entities.StudentEnrollmentStatusHistory), args.Error(1)
}

func (r *MockStudentEnrollmentStatusHistoryRepo) GetLatestStatusEnrollmentByStudentIDAndLocationIDs(arg1 context.Context, arg2 database.QueryExecer, arg3 string, arg4 []string) ([]*entities.StudentEnrollmentStatusHistory, error) {
	args := r.Called(arg1, arg2, arg3, arg4)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.StudentEnrollmentStatusHistory), args.Error(1)
}

func (r *MockStudentEnrollmentStatusHistoryRepo) GetListEnrolledStatusByStudentIDAndTime(arg1 context.Context, arg2 database.QueryExecer, arg3 string, arg4 time.Time) ([]*entities.StudentEnrollmentStatusHistory, error) {
	args := r.Called(arg1, arg2, arg3, arg4)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.StudentEnrollmentStatusHistory), args.Error(1)
}

func (r *MockStudentEnrollmentStatusHistoryRepo) GetListEnrolledStudentEnrollmentStatusByStudentID(arg1 context.Context, arg2 database.QueryExecer, arg3 string) ([]*entities.StudentEnrollmentStatusHistory, error) {
	args := r.Called(arg1, arg2, arg3)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.StudentEnrollmentStatusHistory), args.Error(1)
}

func (r *MockStudentEnrollmentStatusHistoryRepo) GetListStudentEnrollmentStatusHistoryByStudentID(arg1 context.Context, arg2 database.QueryExecer, arg3 string) ([]*entities.StudentEnrollmentStatusHistory, error) {
	args := r.Called(arg1, arg2, arg3)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.StudentEnrollmentStatusHistory), args.Error(1)
}
