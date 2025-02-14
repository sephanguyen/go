// Code generated by mockgen. DO NOT EDIT.
package mock_repositories

import (
	"context"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/mock"

	"github.com/manabie-com/backend/internal/golibs/database"
)

type MockQuizRepo struct {
	mock.Mock
}

func (r *MockQuizRepo) DeleteByExternalID(arg1 context.Context, arg2 database.QueryExecer, arg3 pgtype.Text) error {
	args := r.Called(arg1, arg2, arg3)
	return args.Error(0)
}
