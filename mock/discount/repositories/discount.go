// Code generated by mockgen. DO NOT EDIT.
package mock_repositories

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/manabie-com/backend/internal/discount/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
)

type MockDiscountRepo struct {
	mock.Mock
}

func (r *MockDiscountRepo) GetByDiscountTagIDs(arg1 context.Context, arg2 database.QueryExecer, arg3 []string) ([]*entities.Discount, error) {
	args := r.Called(arg1, arg2, arg3)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Discount), args.Error(1)
}

func (r *MockDiscountRepo) GetByDiscountType(arg1 context.Context, arg2 database.QueryExecer, arg3 string) ([]*entities.Discount, error) {
	args := r.Called(arg1, arg2, arg3)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Discount), args.Error(1)
}

func (r *MockDiscountRepo) GetByID(arg1 context.Context, arg2 database.QueryExecer, arg3 string) (entities.Discount, error) {
	args := r.Called(arg1, arg2, arg3)
	return args.Get(0).(entities.Discount), args.Error(1)
}

func (r *MockDiscountRepo) GetMaxDiscountByTypeAndDiscountTagIDs(arg1 context.Context, arg2 database.QueryExecer, arg3 string, arg4 []string) (entities.Discount, error) {
	args := r.Called(arg1, arg2, arg3, arg4)
	return args.Get(0).(entities.Discount), args.Error(1)
}

func (r *MockDiscountRepo) GetMaxProductDiscountByProductID(arg1 context.Context, arg2 database.QueryExecer, arg3 string) (entities.Discount, error) {
	args := r.Called(arg1, arg2, arg3)
	return args.Get(0).(entities.Discount), args.Error(1)
}
