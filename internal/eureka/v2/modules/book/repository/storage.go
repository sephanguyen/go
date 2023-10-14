package repository

import (
	"context"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/book/domain"
)

type BookRepo interface {
	Upsert(ctx context.Context, books []domain.Book) error
	GetPublishedBookContent(ctx context.Context, bookID string) (domain.Book, error)
	GetBookHierarchyFlattenByLearningMaterialID(ctx context.Context, learningMaterialID string) (domain.BookHierarchyFlatten, error)
}

type LearningMaterialRepo interface {
	// Only use ID, IsPublished field in domain.LearningMaterial
	UpdatePublishStatusLearningMaterials(ctx context.Context, learningMaterials []domain.LearningMaterial) error
	GetByID(ctx context.Context, id string) (domain.LearningMaterial, error)
	GetManyByIDs(ctx context.Context, ids []string) ([]domain.LearningMaterial, error)
}
