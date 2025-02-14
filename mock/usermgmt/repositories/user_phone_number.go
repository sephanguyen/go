// Code generated by mockgen. DO NOT EDIT.
package mock_repositories

import (
	"context"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/mock"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
)

type MockUserPhoneNumberRepo struct {
	mock.Mock
}

func (r *MockUserPhoneNumberRepo) FindByUserID(arg1 context.Context, arg2 database.QueryExecer, arg3 pgtype.Text) ([]*entity.UserPhoneNumber, error) {
	args := r.Called(arg1, arg2, arg3)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.UserPhoneNumber), args.Error(1)
}

func (r *MockUserPhoneNumberRepo) SoftDeleteByUserIDs(arg1 context.Context, arg2 database.QueryExecer, arg3 pgtype.TextArray) error {
	args := r.Called(arg1, arg2, arg3)
	return args.Error(0)
}

func (r *MockUserPhoneNumberRepo) Upsert(arg1 context.Context, arg2 database.QueryExecer, arg3 []*entity.UserPhoneNumber) error {
	args := r.Called(arg1, arg2, arg3)
	return args.Error(0)
}
