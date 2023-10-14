package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/book/domain"
	"github.com/manabie-com/backend/internal/eureka/v2/pkg/errors"
	mock_book_postgres "github.com/manabie-com/backend/mock/eureka/v2/modules/book/repository/postgres"

	"github.com/stretchr/testify/assert"
)

func TestBookUsecase_GetBookHierarchyFlattenByLearningMaterialID(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	bookRepo := &mock_book_postgres.MockBookRepo{}

	learningMaterialID := "learningMaterial_9"

	t.Run("happy case", func(t *testing.T) {
		// arrange
		bookHierarchyFlatten := domain.BookHierarchyFlatten{BookID: "b_id",
			ChapterID:          "chapter_id",
			TopicID:            "topic_id",
			LearningMaterialID: "learning_material_id",
		}

		handler := NewBookUsecase(bookRepo)
		bookRepo.On("GetBookHierarchyFlattenByLearningMaterialID", ctx, learningMaterialID).Once().Return(bookHierarchyFlatten, nil)

		// action
		actualBookHierarchyFlatten, err := handler.GetBookHierarchyFlattenByLearningMaterialID(ctx, learningMaterialID)

		// assert
		assert.Nil(t, err)
		assert.Equal(t, bookHierarchyFlatten, actualBookHierarchyFlatten)
	})

	t.Run("error on BookRepo.GetBookHierarchyFlattenByLearningMaterialID", func(t *testing.T) {
		// arrange
		handler := NewBookUsecase(bookRepo)
		repoError := errors.New("error from BookRepo", nil)
		expectedError := errors.New("BookUsecase.GetBookHierarchyFlattenByLearningMaterialID", repoError)

		bookRepo.On("GetBookHierarchyFlattenByLearningMaterialID", ctx, learningMaterialID).Once().Return(domain.BookHierarchyFlatten{}, repoError)

		// action
		actualBookHierarchyFlatten, err := handler.GetBookHierarchyFlattenByLearningMaterialID(ctx, learningMaterialID)

		// assert
		assert.Equal(t, expectedError, err)
		assert.Equal(t, domain.BookHierarchyFlatten{}, actualBookHierarchyFlatten)
	})
}
