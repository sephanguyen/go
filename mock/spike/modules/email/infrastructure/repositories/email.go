// Code generated by mockgen. DO NOT EDIT.
package mock_repositories

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/spike/modules/email/domain/model"
)

type MockEmailRepo struct {
	mock.Mock
}

func (r *MockEmailRepo) UpdateEmail(arg1 context.Context, arg2 database.QueryExecer, arg3 string, arg4 map[string]interface{}) error {
	args := r.Called(arg1, arg2, arg3, arg4)
	return args.Error(0)
}

func (r *MockEmailRepo) UpsertEmail(arg1 context.Context, arg2 database.QueryExecer, arg3 *model.Email) error {
	args := r.Called(arg1, arg2, arg3)
	return args.Error(0)
}
