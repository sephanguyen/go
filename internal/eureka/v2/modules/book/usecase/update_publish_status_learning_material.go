package usecase

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/book/domain"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (lm *LearningMaterialUsecase) UpdatePublishStatusLearningMaterials(ctx context.Context, learningMaterials []domain.LearningMaterial) error {
	if err := lm.LearningMaterialRepo.UpdatePublishStatusLearningMaterials(ctx, learningMaterials); err != nil {
		return status.Errorf(codes.Internal, fmt.Errorf("LearningMaterialRepo.UpdatePublishStatusLearningMaterials: %w", err).Error())
	}

	return nil
}
