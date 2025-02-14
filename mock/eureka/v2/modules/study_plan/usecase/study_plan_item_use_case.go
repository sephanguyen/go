// Code generated by mockgen. DO NOT EDIT.
package mock_usecase

import (
	"context"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/study_plan/domain"
	"github.com/stretchr/testify/mock"
)

type MockStudyPlanItemUseCase struct {
	mock.Mock
}

func (r *MockStudyPlanItemUseCase) UpsertStudyPlanItems(arg1 context.Context, arg2 []domain.StudyPlanItem) error {
	args := r.Called(arg1, arg2)
	return args.Error(0)
}
