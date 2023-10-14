package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/book/domain"
	"github.com/manabie-com/backend/internal/eureka/v2/pkg/errors"
	mock_book_postgres "github.com/manabie-com/backend/mock/eureka/v2/modules/book/repository/postgres"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpsertBooksHandler_UpsertBooks(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	bookRepo := &mock_book_postgres.MockBookRepo{}

	t.Run("Upsert book successfully", func(t *testing.T) {
		// arrange
		handler := NewBookUsecase(bookRepo)
		books := []domain.Book{
			{
				Name: "book-1",
				ID:   "id-1",
			},
			{
				Name: "book-2",
				ID:   "id-2",
			},
		}
		bookRepo.On("Upsert", ctx, mock.Anything).
			Once().
			Run(func(args mock.Arguments) {
				actual := args[1].([]domain.Book)
				assert.Equal(t, books[0].Name, actual[0].Name)
				assert.Equal(t, books[0].ID, actual[0].ID)

				assert.Equal(t, books[1].Name, actual[1].Name)
				assert.Equal(t, books[1].ID, actual[1].ID)
			}).
			Return(nil)

		// act
		err := handler.UpsertBooks(ctx, books)

		// assert
		assert.Nil(t, err)
		mock.AssertExpectationsForObjects(t, bookRepo)
	})

	t.Run("Upsert book failed", func(t *testing.T) {
		// arrange
		handler := NewBookUsecase(bookRepo)
		books := []domain.Book{
			{
				Name: "book-3",
				ID:   "id-3",
			},
			{
				Name: "book-4",
				ID:   "id-4",
			},
		}
		repoErr := errors.New("any err", nil)
		bookRepo.On("Upsert", ctx, mock.Anything).
			Once().
			Run(func(args mock.Arguments) {
				actual := args[1].([]domain.Book)

				assert.Equal(t, books[0].Name, actual[0].Name)
				assert.Equal(t, books[0].ID, actual[0].ID)

				assert.Equal(t, books[1].Name, actual[1].Name)
				assert.Equal(t, books[1].ID, actual[1].ID)
			}).
			Return(repoErr)

		// act
		err := handler.UpsertBooks(ctx, books)

		// assert
		assert.Equal(t, err, errors.New("BookUsecase.UpsertBooks", repoErr))
		mock.AssertExpectationsForObjects(t, bookRepo)
	})
}
