// Code generated by mockgen. DO NOT EDIT.
package mock_repositories

import (
	"context"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/mock"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
)

type MockSchoolAdminRepo struct {
	mock.Mock
}

func (r *MockSchoolAdminRepo) CreateMultiple(arg1 context.Context, arg2 database.QueryExecer, arg3 []*entity.SchoolAdmin) error {
	args := r.Called(arg1, arg2, arg3)
	return args.Error(0)
}

func (r *MockSchoolAdminRepo) Get(arg1 context.Context, arg2 database.QueryExecer, arg3 pgtype.Text) (*entity.SchoolAdmin, error) {
	args := r.Called(arg1, arg2, arg3)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.SchoolAdmin), args.Error(1)
}

func (r *MockSchoolAdminRepo) SoftDelete(arg1 context.Context, arg2 database.QueryExecer, arg3 pgtype.Text) error {
	args := r.Called(arg1, arg2, arg3)
	return args.Error(0)
}

func (r *MockSchoolAdminRepo) SoftDeleteMultiple(arg1 context.Context, arg2 database.QueryExecer, arg3 pgtype.TextArray) error {
	args := r.Called(arg1, arg2, arg3)
	return args.Error(0)
}

func (r *MockSchoolAdminRepo) Upsert(arg1 context.Context, arg2 database.QueryExecer, arg3 *entity.SchoolAdmin) error {
	args := r.Called(arg1, arg2, arg3)
	return args.Error(0)
}

func (r *MockSchoolAdminRepo) UpsertMultiple(arg1 context.Context, arg2 database.QueryExecer, arg3 []*entity.SchoolAdmin) error {
	args := r.Called(arg1, arg2, arg3)
	return args.Error(0)
}
