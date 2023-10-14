package mock_service

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/aggregate"
)

type MockStudentValidationManager struct {
	mock.Mock
}

func (r *MockStudentValidationManager) FullyValidate(arg1 context.Context, arg2 database.Ext, arg3 aggregate.DomainStudents, arg4 bool) (aggregate.DomainStudents, aggregate.DomainStudents, []error) {
	args := r.Called(arg1, arg2, arg3, arg4)

	if args.Get(0) == nil {
		return nil, args.Get(1).(aggregate.DomainStudents), args.Get(2).([]error)
	}

	if args.Get(1) == nil {
		return nil, nil, args.Get(2).([]error)
	}

	return args.Get(0).(aggregate.DomainStudents), args.Get(1).(aggregate.DomainStudents), args.Get(2).([]error)
}
