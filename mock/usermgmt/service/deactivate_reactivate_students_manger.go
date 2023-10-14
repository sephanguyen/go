package mock_service

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/stretchr/testify/mock"
)

type MockStudentActivationStatusManager struct {
	mock.Mock
}

func (r *MockStudentActivationStatusManager) DeactivateAndReactivateStudents(arg1 context.Context, arg2 database.QueryExecer, arg3, arg4 []string) error {
	args := r.Called(arg1, arg2, arg3, arg4)
	return args.Error(0)
}
