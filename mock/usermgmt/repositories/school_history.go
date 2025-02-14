// Code generated by mockgen. DO NOT EDIT.
package mock_repositories

import (
	"context"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/mock"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
)

type MockSchoolHistoryRepo struct {
	mock.Mock
}

func (r *MockSchoolHistoryRepo) GetByStudentID(arg1 context.Context, arg2 database.QueryExecer, arg3 pgtype.Text) ([]*entity.SchoolHistory, error) {
	args := r.Called(arg1, arg2, arg3)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.SchoolHistory), args.Error(1)
}

func (r *MockSchoolHistoryRepo) GetCurrentSchoolByStudentID(arg1 context.Context, arg2 database.QueryExecer, arg3 pgtype.Text) ([]*entity.SchoolHistory, error) {
	args := r.Called(arg1, arg2, arg3)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.SchoolHistory), args.Error(1)
}

func (r *MockSchoolHistoryRepo) GetSchoolHistoriesByGradeIDAndStudentID(arg1 context.Context, arg2 database.QueryExecer, arg3 pgtype.Text, arg4 pgtype.Text, arg5 pgtype.Bool) ([]*entity.SchoolHistory, error) {
	args := r.Called(arg1, arg2, arg3, arg4, arg5)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.SchoolHistory), args.Error(1)
}

func (r *MockSchoolHistoryRepo) RemoveCurrentSchoolByStudentID(arg1 context.Context, arg2 database.QueryExecer, arg3 pgtype.Text) error {
	args := r.Called(arg1, arg2, arg3)
	return args.Error(0)
}

func (r *MockSchoolHistoryRepo) SetCurrentSchool(arg1 context.Context, arg2 database.QueryExecer, arg3 pgtype.Text) error {
	args := r.Called(arg1, arg2, arg3)
	return args.Error(0)
}

func (r *MockSchoolHistoryRepo) SetCurrentSchoolByStudentIDAndSchoolID(arg1 context.Context, arg2 database.QueryExecer, arg3 pgtype.Text, arg4 pgtype.Text) error {
	args := r.Called(arg1, arg2, arg3, arg4)
	return args.Error(0)
}

func (r *MockSchoolHistoryRepo) SoftDeleteByStudentIDs(arg1 context.Context, arg2 database.QueryExecer, arg3 pgtype.TextArray) error {
	args := r.Called(arg1, arg2, arg3)
	return args.Error(0)
}

func (r *MockSchoolHistoryRepo) UnsetCurrentSchoolByStudentID(arg1 context.Context, arg2 database.QueryExecer, arg3 pgtype.Text) error {
	args := r.Called(arg1, arg2, arg3)
	return args.Error(0)
}

func (r *MockSchoolHistoryRepo) Upsert(arg1 context.Context, arg2 database.QueryExecer, arg3 []*entity.SchoolHistory) error {
	args := r.Called(arg1, arg2, arg3)
	return args.Error(0)
}
