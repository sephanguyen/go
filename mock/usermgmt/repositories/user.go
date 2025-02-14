// Code generated by mockgen. DO NOT EDIT.
package mock_repositories

import (
	"context"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/mock"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
)

type MockUserRepo struct {
	mock.Mock
}

func (r *MockUserRepo) Create(arg1 context.Context, arg2 database.QueryExecer, arg3 *entity.LegacyUser) error {
	args := r.Called(arg1, arg2, arg3)
	return args.Error(0)
}

func (r *MockUserRepo) CreateMultiple(arg1 context.Context, arg2 database.QueryExecer, arg3 []*entity.LegacyUser) error {
	args := r.Called(arg1, arg2, arg3)
	return args.Error(0)
}

func (r *MockUserRepo) Find(arg1 context.Context, arg2 database.QueryExecer, arg3 *repository.UserFindFilter, arg4 ...string) ([]*entity.LegacyUser, error) {
	args := r.Called(arg1, arg2, arg3, arg4)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.LegacyUser), args.Error(1)
}

func (r *MockUserRepo) FindByIDUnscope(arg1 context.Context, arg2 database.QueryExecer, arg3 pgtype.Text) (*entity.LegacyUser, error) {
	args := r.Called(arg1, arg2, arg3)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.LegacyUser), args.Error(1)
}

func (r *MockUserRepo) Get(arg1 context.Context, arg2 database.QueryExecer, arg3 pgtype.Text) (*entity.LegacyUser, error) {
	args := r.Called(arg1, arg2, arg3)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.LegacyUser), args.Error(1)
}

func (r *MockUserRepo) GetBasicInfo(arg1 context.Context, arg2 database.QueryExecer, arg3 pgtype.Text) (*entity.LegacyUser, error) {
	args := r.Called(arg1, arg2, arg3)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.LegacyUser), args.Error(1)
}

func (r *MockUserRepo) GetByEmail(arg1 context.Context, arg2 database.QueryExecer, arg3 pgtype.TextArray) ([]*entity.LegacyUser, error) {
	args := r.Called(arg1, arg2, arg3)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.LegacyUser), args.Error(1)
}

func (r *MockUserRepo) GetByEmailInsensitiveCase(arg1 context.Context, arg2 database.QueryExecer, arg3 []string) ([]*entity.LegacyUser, error) {
	args := r.Called(arg1, arg2, arg3)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.LegacyUser), args.Error(1)
}

func (r *MockUserRepo) GetByPhone(arg1 context.Context, arg2 database.QueryExecer, arg3 pgtype.TextArray) ([]*entity.LegacyUser, error) {
	args := r.Called(arg1, arg2, arg3)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.LegacyUser), args.Error(1)
}

func (r *MockUserRepo) GetUserGroupMembers(arg1 context.Context, arg2 database.QueryExecer, arg3 pgtype.Text) ([]*entity.UserGroupMember, error) {
	args := r.Called(arg1, arg2, arg3)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.UserGroupMember), args.Error(1)
}

func (r *MockUserRepo) GetUserGroups(arg1 context.Context, arg2 database.QueryExecer, arg3 pgtype.Text) ([]*entity.UserGroupV2, error) {
	args := r.Called(arg1, arg2, arg3)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.UserGroupV2), args.Error(1)
}

func (r *MockUserRepo) GetUserRoles(arg1 context.Context, arg2 database.QueryExecer, arg3 pgtype.Text) (entity.Roles, error) {
	args := r.Called(arg1, arg2, arg3)
	return args.Get(0).(entity.Roles), args.Error(1)
}

func (r *MockUserRepo) GetUsersByUserGroupID(arg1 context.Context, arg2 database.QueryExecer, arg3 pgtype.Text) ([]*entity.LegacyUser, error) {
	args := r.Called(arg1, arg2, arg3)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.LegacyUser), args.Error(1)
}

func (r *MockUserRepo) ResourcePath(arg1 context.Context, arg2 database.QueryExecer, arg3 pgtype.Text) (string, error) {
	args := r.Called(arg1, arg2, arg3)
	return args.Get(0).(string), args.Error(1)
}

func (r *MockUserRepo) Retrieve(arg1 context.Context, arg2 database.QueryExecer, arg3 pgtype.TextArray, arg4 ...string) ([]*entity.LegacyUser, error) {
	args := r.Called(arg1, arg2, arg3, arg4)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.LegacyUser), args.Error(1)
}

func (r *MockUserRepo) SearchProfile(arg1 context.Context, arg2 database.QueryExecer, arg3 *repository.SearchProfileFilter) ([]*entity.LegacyUser, error) {
	args := r.Called(arg1, arg2, arg3)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.LegacyUser), args.Error(1)
}

func (r *MockUserRepo) SoftDelete(arg1 context.Context, arg2 database.QueryExecer, arg3 pgtype.TextArray) error {
	args := r.Called(arg1, arg2, arg3)
	return args.Error(0)
}

func (r *MockUserRepo) UpdateEmail(arg1 context.Context, arg2 database.QueryExecer, arg3 *entity.LegacyUser) error {
	args := r.Called(arg1, arg2, arg3)
	return args.Error(0)
}

func (r *MockUserRepo) UpdateLastLoginDate(arg1 context.Context, arg2 database.QueryExecer, arg3 *entity.LegacyUser) error {
	args := r.Called(arg1, arg2, arg3)
	return args.Error(0)
}

func (r *MockUserRepo) UpdateManyUserGroup(arg1 context.Context, arg2 database.QueryExecer, arg3 pgtype.TextArray, arg4 pgtype.Text) error {
	args := r.Called(arg1, arg2, arg3, arg4)
	return args.Error(0)
}

func (r *MockUserRepo) UpdateProfileV1(arg1 context.Context, arg2 database.QueryExecer, arg3 *entity.LegacyUser) error {
	args := r.Called(arg1, arg2, arg3)
	return args.Error(0)
}

func (r *MockUserRepo) UserGroup(arg1 context.Context, arg2 database.QueryExecer, arg3 pgtype.Text) (string, error) {
	args := r.Called(arg1, arg2, arg3)
	return args.Get(0).(string), args.Error(1)
}
