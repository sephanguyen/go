package usecase

import (
	"context"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/book/domain"
	"github.com/manabie-com/backend/internal/eureka/v2/modules/book/repository"
)

type BookUsecase struct {
	BookRepo repository.BookRepo
}

type LearningMaterialUsecase struct {
	LearningMaterialRepo repository.LearningMaterialRepo
}

func NewBookUsecase(bookRepo repository.BookRepo) *BookUsecase {
	return &BookUsecase{
		BookRepo: bookRepo,
	}
}

func NewLearningMaterialUsecase(learningMaterialRepo repository.LearningMaterialRepo) *LearningMaterialUsecase {
	return &LearningMaterialUsecase{
		LearningMaterialRepo: learningMaterialRepo,
	}
}

type BookUpserter interface {
	UpsertBooks(ctx context.Context, books []domain.Book) error
}

type BookContentGetter interface {
	GetPublishedBookContent(ctx context.Context, bookID string) (domain.Book, error)
}

type BookHierarchyGetter interface {
	GetBookHierarchyFlattenByLearningMaterialID(ctx context.Context, learningMaterialID string) (domain.BookHierarchyFlatten, error)
}

type UpdatePublishStatusLearningMaterials interface {
	UpdatePublishStatusLearningMaterials(ctx context.Context, learningMaterials []domain.LearningMaterial) error
}
