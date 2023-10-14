package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/book/domain"
	"github.com/manabie-com/backend/internal/eureka/v2/pkg/constants"
	"github.com/manabie-com/backend/internal/eureka/v2/pkg/errors"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	mock_book_postgres "github.com/manabie-com/backend/mock/eureka/v2/modules/book/repository/postgres"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestBookUsecase_GetPublishedBookContent(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	bookRepo := &mock_book_postgres.MockBookRepo{}

	t.Run("Get book content that unpublished are removed", func(t *testing.T) {
		// arrange
		handler := NewBookUsecase(bookRepo)
		bookID := idutil.ULIDNow()
		dbBook := domain.Book{
			ID:   bookID,
			Name: "Book ID",
			Chapters: []domain.Chapter{
				{
					ID:   "Chapter 1",
					Name: "Name 1",
					Topics: []domain.Topic{
						{
							ID:   "Topic 1",
							Name: "Name 1",
							LearningMaterials: []domain.LearningMaterial{
								{
									ID:        "LM ID 1",
									Name:      "LM Name 1",
									Type:      constants.FlashCard,
									Published: true,
								},
							},
						},
						{
							ID:   "Topic 2",
							Name: "Name 2",
							LearningMaterials: []domain.LearningMaterial{
								{
									ID:        "LM ID 2",
									Name:      "LM Name 2",
									Type:      constants.FlashCard,
									Published: true,
								},
							},
						},
					},
				},
			},
		}
		bookRepo.On("GetPublishedBookContent", ctx, bookID).
			Once().
			Return(dbBook, nil)

		// act
		actual, err := handler.GetPublishedBookContent(ctx, bookID)

		// assert
		assert.Nil(t, err)
		assert.Equal(t, dbBook, actual)
		mock.AssertExpectationsForObjects(t, bookRepo)
	})

	t.Run("Return error from Repo", func(t *testing.T) {
		// arrange
		handler := NewBookUsecase(bookRepo)
		bookID := idutil.ULIDNow()
		dbBook := domain.Book{}
		repoErr := errors.NewDBError("sample error", nil)
		expectedErr := errors.New("BookUsecase.GetPublishedBookContent", repoErr)
		bookRepo.On("GetPublishedBookContent", ctx, bookID).
			Once().
			Return(dbBook, repoErr)

		// act
		actual, err := handler.GetPublishedBookContent(ctx, bookID)

		// assert
		assert.Equal(t, expectedErr, err)
		assert.Equal(t, dbBook, actual)
		mock.AssertExpectationsForObjects(t, bookRepo)
	})

	t.Run("Return entity not found from Repo when receive NoRowsExistedError", func(t *testing.T) {
		// arrange
		handler := NewBookUsecase(bookRepo)
		bookID := idutil.ULIDNow()
		dbBook := domain.Book{}
		repoErr := errors.NewNoRowsExistedError("sample err no rows", nil)
		expectedErr := errors.NewEntityNotFoundError("BookUsecase.GetPublishedBookContent", repoErr)
		bookRepo.On("GetPublishedBookContent", ctx, bookID).
			Once().
			Return(dbBook, repoErr)

		// act
		actual, err := handler.GetPublishedBookContent(ctx, bookID)

		// assert
		assert.Equal(t, expectedErr, err)
		assert.Equal(t, dbBook, actual)
		mock.AssertExpectationsForObjects(t, bookRepo)
	})
}
