// Code generated by mockgen. DO NOT EDIT.
package mock_repositories

import (
	"context"

	"github.com/stretchr/testify/mock"

	services "github.com/manabie-com/backend/internal/eureka/entities/items_bank"
	"github.com/manabie-com/backend/internal/golibs/learnosity"
)

type MockItemsBankRepo struct {
	mock.Mock
}

func (r *MockItemsBankRepo) ArchiveItems(arg1 context.Context, arg2 []string, arg3 string) error {
	args := r.Called(arg1, arg2, arg3)
	return args.Error(0)
}

func (r *MockItemsBankRepo) GetCurrentItemIDs(arg1 context.Context, arg2 []string) (map[string][]string, error) {
	args := r.Called(arg1, arg2)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string][]string), args.Error(1)
}

func (r *MockItemsBankRepo) GetExistedIDs(arg1 context.Context, arg2 []string) ([]string, error) {
	args := r.Called(arg1, arg2)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (r *MockItemsBankRepo) GetListItems(arg1 context.Context, arg2 []string, arg3 *string, arg4 uint32) (*learnosity.Result, error) {
	args := r.Called(arg1, arg2, arg3, arg4)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*learnosity.Result), args.Error(1)
}

func (r *MockItemsBankRepo) MapItemsByActivity(arg1 context.Context, arg2 string, arg3 map[string][]string) error {
	args := r.Called(arg1, arg2, arg3)
	return args.Error(0)
}

func (r *MockItemsBankRepo) UploadContentData(arg1 context.Context, arg2 string, arg3 map[string]*services.ItemsBankItem, arg4 []*services.ItemsBankQuestion) ([]string, error) {
	args := r.Called(arg1, arg2, arg3, arg4)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}
