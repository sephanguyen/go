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

type MockOrganizationRepo struct {
	mock.Mock
}

func (r *MockOrganizationRepo) DefaultOrganizationAuthValues(arg1 string) string {
	args := r.Called(arg1)
	return args.Get(0).(string)
}

func (r *MockOrganizationRepo) Find(arg1 context.Context, arg2 database.QueryExecer, arg3 pgtype.Text) (*entity.Organization, error) {
	args := r.Called(arg1, arg2, arg3)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Organization), args.Error(1)
}

func (r *MockOrganizationRepo) GetAll(arg1 context.Context, arg2 database.QueryExecer, arg3 int) ([]*entity.OrganizationAuth, error) {
	args := r.Called(arg1, arg2, arg3)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.OrganizationAuth), args.Error(1)
}

func (r *MockOrganizationRepo) GetByDomainName(arg1 context.Context, arg2 database.QueryExecer, arg3 string) (*entity.Organization, error) {
	args := r.Called(arg1, arg2, arg3)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Organization), args.Error(1)
}

func (r *MockOrganizationRepo) GetByTenantID(arg1 context.Context, arg2 database.QueryExecer, arg3 string) (*entity.Organization, error) {
	args := r.Called(arg1, arg2, arg3)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Organization), args.Error(1)
}

func (r *MockOrganizationRepo) GetTenantIDByOrgID(arg1 context.Context, arg2 database.QueryExecer, arg3 string) (string, error) {
	args := r.Called(arg1, arg2, arg3)
	return args.Get(0).(string), args.Error(1)
}

func (r *MockOrganizationRepo) WithDefaultValue(arg1 string) *repository.OrganizationRepo {
	args := r.Called(arg1)

	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*repository.OrganizationRepo)
}
