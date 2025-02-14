// Code generated by mockgen. DO NOT EDIT.
package mock_postgres

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/assessment/domain"
	"github.com/manabie-com/backend/internal/golibs/database"
)

type MockStudentEventLogRepo struct {
	mock.Mock
}

func (r *MockStudentEventLogRepo) GetManyByEventTypesAndLMs(arg1 context.Context, arg2 database.Ext, arg3 string, arg4 string, arg5 []string, arg6 []string) ([]domain.StudentEventLog, error) {
	args := r.Called(arg1, arg2, arg3, arg4, arg5, arg6)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.StudentEventLog), args.Error(1)
}
