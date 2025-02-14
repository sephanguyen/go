// Code generated by mockgen. DO NOT EDIT.
package mock_repositories

import (
	"context"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/mock"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/user/infrastructure/repo"
)

type MockUserRepo struct {
	mock.Mock
}

func (r *MockUserRepo) Find(arg1 context.Context, arg2 database.QueryExecer, arg3 *repo.UserFindFilter, arg4 ...string) ([]*repo.User, error) {
	args := r.Called(arg1, arg2, arg3, arg4)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*repo.User), args.Error(1)
}

func (r *MockUserRepo) GetTeacherByID(arg1 context.Context, arg2 database.QueryExecer, arg3 string) (*domain.Teacher, error) {
	args := r.Called(arg1, arg2, arg3)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Teacher), args.Error(1)
}

func (r *MockUserRepo) GetUsersByIDs(arg1 context.Context, arg2 database.QueryExecer, arg3 []string) (map[string]*domain.User, error) {
	args := r.Called(arg1, arg2, arg3)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]*domain.User), args.Error(1)
}

func (r *MockUserRepo) Retrieve(arg1 context.Context, arg2 database.QueryExecer, arg3 pgtype.TextArray, arg4 ...string) ([]*repo.User, error) {
	args := r.Called(arg1, arg2, arg3, arg4)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*repo.User), args.Error(1)
}

func (r *MockUserRepo) UserGroup(arg1 context.Context, arg2 database.QueryExecer, arg3 string) (string, error) {
	args := r.Called(arg1, arg2, arg3)
	return args.Get(0).(string), args.Error(1)
}
