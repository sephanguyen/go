package usecase

import (
	"context"
	"fmt"
	"testing"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/book/domain"
	mock_postgres "github.com/manabie-com/backend/mock/eureka/v2/modules/book/repository/postgres"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestLearningMaterialUsecase_UpdatePublishStatusLearningMaterials(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	mockLearningMaterialRepo := &mock_postgres.MockLearningMaterialRepo{}

	t.Run("Update publish learning materials successfully", func(t *testing.T) {
		lms := []domain.LearningMaterial{{
			ID:        "1",
			Published: false,
		}, {
			ID:        "2",
			Published: true,
		}}

		usecase := NewLearningMaterialUsecase(
			mockLearningMaterialRepo,
		)

		mockLearningMaterialRepo.On("UpdatePublishStatusLearningMaterials", mock.Anything, lms).Once().Return(nil)

		err := usecase.UpdatePublishStatusLearningMaterials(ctx, lms)

		assert.Nil(t, err)
	})

	t.Run("Error: Repo.UpdatePublishStatusLearningMaterials failed", func(t *testing.T) {
		lms := []domain.LearningMaterial{{
			ID:        "1",
			Published: false,
		}, {
			ID:        "2",
			Published: true,
		}}

		usecase := NewLearningMaterialUsecase(
			mockLearningMaterialRepo,
		)

		errFromLmRepo := fmt.Errorf("UpdatePublishStatusLearningMaterials err")
		mockLearningMaterialRepo.On("UpdatePublishStatusLearningMaterials", mock.Anything, lms).Once().Return(errFromLmRepo)

		err := usecase.UpdatePublishStatusLearningMaterials(ctx, lms)

		assert.Equal(t, status.Errorf(codes.Internal, fmt.Errorf("LearningMaterialRepo.UpdatePublishStatusLearningMaterials: %w", errFromLmRepo).Error()), err)
	})
}
