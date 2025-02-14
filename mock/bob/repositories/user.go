// Code generated by mockgen. DO NOT EDIT.
package mock_repositories

import (
	"context"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/mock"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
)

type MockUserRepo struct {
	mock.Mock
}

func (r *MockUserRepo) Create(arg1 context.Context, arg2 database.QueryExecer, arg3 *entities.User) error {
	args := r.Called(arg1, arg2, arg3)
	return args.Error(0)
}

func (r *MockUserRepo) CreateMultiple(arg1 context.Context, arg2 database.QueryExecer, arg3 []*entities.User) error {
	args := r.Called(arg1, arg2, arg3)
	return args.Error(0)
}

func (r *MockUserRepo) Find(arg1 context.Context, arg2 database.QueryExecer, arg3 *repositories.UserFindFilter, arg4 ...string) ([]*entities.User, error) {
	args := r.Called(arg1, arg2, arg3, arg4)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.User), args.Error(1)
}

func (r *MockUserRepo) FindByIDUnscope(arg1 context.Context, arg2 database.QueryExecer, arg3 pgtype.Text) (*entities.User, error) {
	args := r.Called(arg1, arg2, arg3)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.User), args.Error(1)
}

func (r *MockUserRepo) FindUsersMatchConditions(arg1 context.Context, arg2 database.Ext, arg3 entities.TargetConditions) ([]*entities.User, error) {
	args := r.Called(arg1, arg2, arg3)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.User), args.Error(1)
}

func (r *MockUserRepo) Get(arg1 context.Context, arg2 database.QueryExecer, arg3 pgtype.Text) (*entities.User, error) {
	args := r.Called(arg1, arg2, arg3)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.User), args.Error(1)
}

func (r *MockUserRepo) GetByEmail(arg1 context.Context, arg2 database.QueryExecer, arg3 pgtype.TextArray) ([]*entities.User, error) {
	args := r.Called(arg1, arg2, arg3)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.User), args.Error(1)
}

func (r *MockUserRepo) GetByPhone(arg1 context.Context, arg2 database.QueryExecer, arg3 pgtype.TextArray) ([]*entities.User, error) {
	args := r.Called(arg1, arg2, arg3)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.User), args.Error(1)
}

func (r *MockUserRepo) GetUsernameByUserID(arg1 context.Context, arg2 database.QueryExecer, arg3 string) (*entities.Username, error) {
	args := r.Called(arg1, arg2, arg3)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Username), args.Error(1)
}

func (r *MockUserRepo) ResourcePath(arg1 context.Context, arg2 database.QueryExecer, arg3 pgtype.Text) (string, error) {
	args := r.Called(arg1, arg2, arg3)
	return args.Get(0).(string), args.Error(1)
}

func (r *MockUserRepo) Retrieve(arg1 context.Context, arg2 database.QueryExecer, arg3 pgtype.TextArray, arg4 ...string) ([]*entities.User, error) {
	args := r.Called(arg1, arg2, arg3, arg4)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.User), args.Error(1)
}

func (r *MockUserRepo) SearchProfile(arg1 context.Context, arg2 database.QueryExecer, arg3 *repositories.SearchProfileFilter) ([]*entities.User, error) {
	args := r.Called(arg1, arg2, arg3)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.User), args.Error(1)
}

func (r *MockUserRepo) SoftDelete(arg1 context.Context, arg2 database.QueryExecer, arg3 pgtype.TextArray) error {
	args := r.Called(arg1, arg2, arg3)
	return args.Error(0)
}

func (r *MockUserRepo) StoreDeviceToken(arg1 context.Context, arg2 database.QueryExecer, arg3 *entities.User) error {
	args := r.Called(arg1, arg2, arg3)
	return args.Error(0)
}

func (r *MockUserRepo) Update(arg1 context.Context, arg2 database.QueryExecer, arg3 *entities.User) error {
	args := r.Called(arg1, arg2, arg3)
	return args.Error(0)
}

func (r *MockUserRepo) UpdateEmail(arg1 context.Context, arg2 database.QueryExecer, arg3 *entities.User) error {
	args := r.Called(arg1, arg2, arg3)
	return args.Error(0)
}

func (r *MockUserRepo) UpdateLastLoginDate(arg1 context.Context, arg2 database.QueryExecer, arg3 *entities.User) error {
	args := r.Called(arg1, arg2, arg3)
	return args.Error(0)
}

func (r *MockUserRepo) UpdateProfile(arg1 context.Context, arg2 database.QueryExecer, arg3 *entities.User) error {
	args := r.Called(arg1, arg2, arg3)
	return args.Error(0)
}

func (r *MockUserRepo) UpdateProfileV1(arg1 context.Context, arg2 database.QueryExecer, arg3 *entities.User) error {
	args := r.Called(arg1, arg2, arg3)
	return args.Error(0)
}

func (r *MockUserRepo) UserGroup(arg1 context.Context, arg2 database.QueryExecer, arg3 pgtype.Text) (string, error) {
	args := r.Called(arg1, arg2, arg3)
	return args.Get(0).(string), args.Error(1)
}
