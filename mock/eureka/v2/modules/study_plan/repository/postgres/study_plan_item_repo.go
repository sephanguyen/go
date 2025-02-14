// Code generated by mockgen. DO NOT EDIT.
package mock_postgres

import (
	"context"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/study_plan/repository/postgres/dto"
	"github.com/stretchr/testify/mock"
)

type MockStudyPlanItemRepo struct {
	mock.Mock
}

func (r *MockStudyPlanItemRepo) UpsertStudyPlanItems(arg1 context.Context, arg2 []*dto.StudyPlanItemDto) error {
	args := r.Called(arg1, arg2)
	return args.Error(0)
}
