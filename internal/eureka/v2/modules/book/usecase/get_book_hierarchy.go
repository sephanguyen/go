package usecase

import (
	"context"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/book/domain"
	"github.com/manabie-com/backend/internal/eureka/v2/pkg/errors"
)

func (b *BookUsecase) GetBookHierarchyFlattenByLearningMaterialID(ctx context.Context, learningMaterialID string) (domain.BookHierarchyFlatten, error) {
	bHierarchyFlatten, err := b.BookRepo.GetBookHierarchyFlattenByLearningMaterialID(ctx, learningMaterialID)

	if err != nil {
		if errors.CheckErrType(errors.ErrNoRowsExisted, err) {
			return bHierarchyFlatten, errors.NewEntityNotFoundError("BookUsecase.GetBookHierarchyFlattenByLearningMaterialID", err)
		}
		return bHierarchyFlatten, errors.New("BookUsecase.GetBookHierarchyFlattenByLearningMaterialID", err)
	}

	return bHierarchyFlatten, nil
}
