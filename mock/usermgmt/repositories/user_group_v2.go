// Code generated by mockgen. DO NOT EDIT.
package mock_repositories

import (
	"context"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/mock"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
)

type MockUserGroupV2Repo struct {
	mock.Mock
}

func (r *MockUserGroupV2Repo) Create(arg1 context.Context, arg2 database.QueryExecer, arg3 *entity.UserGroupV2) error {
	args := r.Called(arg1, arg2, arg3)
	return args.Error(0)
}

func (r *MockUserGroupV2Repo) Find(arg1 context.Context, arg2 database.QueryExecer, arg3 pgtype.Text) (*entity.UserGroupV2, error) {
	args := r.Called(arg1, arg2, arg3)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.UserGroupV2), args.Error(1)
}

func (r *MockUserGroupV2Repo) FindAndMapUserGroupAndRolesByUserID(arg1 context.Context, arg2 database.QueryExecer, arg3 pgtype.Text) (map[entity.UserGroupV2][]*entity.Role, error) {
	args := r.Called(arg1, arg2, arg3)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[entity.UserGroupV2][]*entity.Role), args.Error(1)
}

func (r *MockUserGroupV2Repo) FindByIDs(arg1 context.Context, arg2 database.QueryExecer, arg3 []string) ([]*entity.UserGroupV2, error) {
	args := r.Called(arg1, arg2, arg3)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.UserGroupV2), args.Error(1)
}

func (r *MockUserGroupV2Repo) FindUserGroupAndRoleByUserID(arg1 context.Context, arg2 database.QueryExecer, arg3 pgtype.Text) (map[string][]*entity.Role, error) {
	args := r.Called(arg1, arg2, arg3)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string][]*entity.Role), args.Error(1)
}

func (r *MockUserGroupV2Repo) FindUserGroupByRoleName(arg1 context.Context, arg2 database.QueryExecer, arg3 string) (*entity.UserGroupV2, error) {
	args := r.Called(arg1, arg2, arg3)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.UserGroupV2), args.Error(1)
}

func (r *MockUserGroupV2Repo) Update(arg1 context.Context, arg2 database.QueryExecer, arg3 *entity.UserGroupV2) error {
	args := r.Called(arg1, arg2, arg3)
	return args.Error(0)
}
