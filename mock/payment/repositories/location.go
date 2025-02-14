// Code generated by mockgen. DO NOT EDIT.
package mock_repositories

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/repositories"
)

type MockLocationRepo struct {
	mock.Mock
}

func (r *MockLocationRepo) GetByID(arg1 context.Context, arg2 database.QueryExecer, arg3 string) (entities.Location, error) {
	args := r.Called(arg1, arg2, arg3)
	return args.Get(0).(entities.Location), args.Error(1)
}

func (r *MockLocationRepo) GetByIDForUpdate(arg1 context.Context, arg2 database.QueryExecer, arg3 string) (entities.Location, error) {
	args := r.Called(arg1, arg2, arg3)
	return args.Get(0).(entities.Location), args.Error(1)
}

func (r *MockLocationRepo) GetByIDs(arg1 context.Context, arg2 database.QueryExecer, arg3 []string) ([]entities.Location, error) {
	args := r.Called(arg1, arg2, arg3)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entities.Location), args.Error(1)
}

func (r *MockLocationRepo) GetLowestGrantedLocationIDsByUserIDAndPermissions(arg1 context.Context, arg2 database.QueryExecer, arg3 repositories.GetGrantedLowestLevelLocationsParams) ([]string, error) {
	args := r.Called(arg1, arg2, arg3)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}
